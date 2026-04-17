# ADR 023 — One container per tenant (all-in-one image)

**Status:** accepted
**Date:** 2026-04-16
**Deciders:** Enzo Saldivia
**Refines:** ADR 022 (silo model) — the silo **is** a single container.

## Context

ADR 022 committed to silo deployment: one isolated stack per tenant. It left open
the shape of that stack — "its own Postgres, NATS, Redis, MinIO, services, frontend".
In practice that meant a `docker compose` file per tenant with 5–8 containers each.

Reviewing that choice against the project's actual constraints:

- **The workstation is one vertical-scale box** (Threadripper 9975WX, 256 GB RAM,
  96 GB VRAM, 8 TB NVMe). There is no cluster, no node-to-pod distribution to optimise.
- **The project is maintained by one person.** Operational complexity is the
  dominant cost, not runtime efficiency.
- **Principle #0 (reduce before add)** applies: if the shape can be smaller with
  the same behaviour, it should be.
- **The product may ship on-premise** to future B2B-enterprise customers. A
  portable, single-artifact container is the easiest delivery mechanism possible.

A `docker compose` stack of 5+ containers per tenant carries real overhead per
tenant: 5 service definitions to keep in sync, 5 images to pull, network config,
logging per container, ordered startup, inter-container health checks. Each
avoidable for a one-person project serving a handful of tenants on a box that
fits them all.

## Decision

**One container per tenant. Everything runs inside it.**

Runtime contents of the single image:

- **App** — the Go monolith (successor binary once the consolidation from
  ADR 021 completes). Single process, `chi` router, all domain modules.
- **Postgres** — serves only this tenant. Listens on an internal Unix socket.
- **NATS JetStream** — serves only this tenant. Localhost binding.
- **Redis** — localhost binding.
- **MinIO** — localhost binding, object storage.
- **Traefik entrypoint** — optional; if the workstation already fronts tenants
  via a shared Traefik, the container can expose only the app's HTTP port.

Process supervision: **`s6-overlay`** (small, deterministic, well-documented;
used by LinuxServer.io images). `supervisord` is an acceptable fallback.

### Volume layout (state lives on the host)

```
/data/<tenant>/
├── postgres/      ← /var/lib/postgresql/data inside container
├── nats/          ← JetStream file store
├── redis/         ← dump.rdb / AOF
├── minio/         ← object buckets
└── uploads/       ← raw staging for ingest
```

The container is **reproducible and replaceable**. State is on the host.
`docker rm <container>` must never destroy data — that lives in `/data/<tenant>/`.

### Deploy layout

```
deploy/
├── Dockerfile.all-in-one       # multi-stage; builds app + bakes Postgres, NATS, Redis, MinIO
├── s6-overlay/                 # init scripts for each service
│   ├── postgres/
│   ├── nats/
│   ├── redis/
│   ├── minio/
│   └── app/
└── tenants/
    ├── saldivia/
    │   ├── .env                # SDA_TENANT=saldivia, branding, feature flags, secrets refs
    │   └── run.sh              # thin wrapper: docker run with the right -v and --env-file
    └── <next-tenant>/
```

No per-tenant `docker-compose.yml` needed. A single `docker run` (or a Makefile
target) is the whole deploy surface.

### Commands

```bash
make build-image               # builds deploy/Dockerfile.all-in-one → sda:<version>
make deploy TENANT=saldivia    # pull/use image, docker stop+rm, docker run with volumes
make deploy TENANT=all         # iterate over deploy/tenants/*/
make status TENANT=saldivia    # docker ps + readyz probe
make logs TENANT=saldivia      # docker logs (s6-overlay prefixes each service)
```

## Consequences

**Positive**

- **Operational simplicity is maximal.** One image to build, one command to
  deploy, one artifact to back up, one process graph to reason about.
- **Trivial on-premise shipping.** A new customer receives the image + a volume
  template; they are running in minutes.
- **Backup / restore per tenant is a volume snapshot.** No coordination across
  Postgres dumps + NATS stream snapshots + Redis RDB.
- **No inter-container networking per tenant.** Postgres via Unix socket,
  NATS/Redis via `localhost` — lower latency, less config.
- **Aligns with principle #0.** Every line of compose boilerplate per tenant
  disappears.

**Negative**

- **Breaks the "one process per container" dogma.** Mitigated by `s6-overlay`,
  which is purpose-built for exactly this and does not bleed concerns across
  services. The dogma is a guideline, not a rule — GitLab, Sentry, Plausible, and
  Mattermost all ship this way for self-hosted.
- **Image is ~2–3 GB.** Irrelevant for an inhouse deployment (no push across the
  internet). Matters only if/when we push to a customer registry.
- **A container restart restarts all internal services.** Postgres cold-start is
  3–5 s; the app itself is ~1 s. Acceptable for a product that does not claim
  four-9s uptime.
- **Postgres major-version upgrades require rebuilding the image + running
  `pg_upgrade` against the volume.** This is a planned operation, done every
  few years, not per-deploy.
- **Logs are interleaved** by default. `s6-overlay` prefixes each service's
  output so `docker logs` remains readable; downstream log shipping can split by
  prefix if needed.

## Alternatives considered

1. **`docker compose` stack per tenant (ADR 022 as literally written).**
   Rejected. 5–8 containers of ceremony per tenant, for no benefit the hardware
   or the product actually uses.

2. **Compose with profiles + one Postgres shared across tenants (with databases
   per tenant).**
   Rejected. Reintroduces a pool surface (shared Postgres) that contradicts the
   silo commitment and complicates backup/restore/on-premise delivery.

3. **Hybrid: app container + sidecar Postgres container per tenant.**
   Rejected. Moves half the complexity into scope without removing it. If
   Postgres is going to ride with the app anyway, it is cleaner to have them in
   the same image.

4. **Separate images for app / Postgres / NATS deployed by an orchestrator.**
   Rejected. Requires an orchestrator the project does not have (K8s/Nomad) and
   optimises for a scale profile the project does not have.

## Open items

- Choose between `s6-overlay` (current preference) and `supervisord` once the
  Dockerfile lands. Decide in the implementing PR, not here.
- Metrics / tracing collection inside the container is an open question —
  likely OTLP to the host via a socket, with the collector outside the tenant
  container to stay single-tenant-scoped per deployment.
- NATS JetStream vs. in-process event bus becomes re-examinable. If the monolith
  consolidation (ADR 021) lands fully, a Go in-process event bus could replace
  NATS entirely inside the container. Track as a follow-up, not part of this ADR.
