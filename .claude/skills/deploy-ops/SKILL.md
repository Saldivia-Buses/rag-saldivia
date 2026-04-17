---
name: deploy-ops
description: Use when touching deploy/, Dockerfiles, docker-compose, Traefik config, workstation operations, or production rollout. Covers the inhouse-workstation deploy pipeline, service versioning, health checks, rollback, the make deploy flow, and incident triage.
---

# deploy-ops

Scope: `deploy/`, `**/Dockerfile`, `.github/workflows/`, any service `VERSION` file,
`Makefile` targets that affect runtime.

## Where things run

| Layer | Where | Why |
|---|---|---|
| Postgres, NATS, Redis, MinIO, Traefik | inhouse workstation (Docker) | low-latency to services |
| Go services | inhouse workstation (Docker, one per service) | colocated with data |
| SGLang / GPU inference | inhouse workstation (GPU) | bandwidth to GPU |
| Next.js frontend | CDN (built in CI) | cheap static + edge |
| CI | GitHub Actions | builds, tests, pushes images |

Single-box backend — no Kubernetes, no cluster. Compose + Traefik is the runtime.

## Silo deploys: one container per tenant (ADR 022 + 023)

Each tenant is a **single all-in-one container** that runs the Go app plus its
own Postgres, NATS, Redis, and MinIO under `s6-overlay`. Code is single-tenant;
the tenant **is** the container. Volumes on the host hold all persistent state so
the container itself is disposable.

### Image

- `deploy/Dockerfile.all-in-one` — multi-stage:
  1. `golang:1.25` builder → app binary.
  2. Final stage based on `debian:12-slim` (or distroless-compatible variant if
     it supports embedded Postgres/NATS/Redis/MinIO).
  3. `s6-overlay` installed; `/etc/s6-overlay/s6-rc.d/` has one directory per service:
     `postgres`, `nats`, `redis`, `minio`, `app`.
  4. App launches last; depends on Postgres readiness.
- Single tag: `sda:<version>`. One image serves every tenant.

### Volume layout (host → container)

```
/data/<tenant>/postgres   → /var/lib/postgresql/data
/data/<tenant>/nats       → /var/lib/nats/jetstream
/data/<tenant>/redis      → /var/lib/redis
/data/<tenant>/minio      → /var/lib/minio
/data/<tenant>/uploads    → /var/lib/sda/uploads
```

**Rule: never `docker rm` a tenant container without confirming the volumes are
mounted to the host.** The container is replaceable; the `/data/<tenant>/` tree
is not.

### Layout

```
deploy/
├── Dockerfile.all-in-one
├── s6-overlay/
│   ├── postgres/run
│   ├── nats/run
│   ├── redis/run
│   ├── minio/run
│   └── app/run
└── tenants/
    ├── saldivia/
    │   ├── .env                  # SDA_TENANT=saldivia, brand, feature flags
    │   └── run.sh                # thin `docker run` wrapper
    └── <next-tenant>/
        └── …
```

No per-tenant `docker-compose.yml`.

### Env vars that matter per tenant

- `SDA_TENANT` — slug, used for log context, branding, storage paths.
- `SDA_BRAND_*` — logo URL, display name, palette overrides.
- Feature flags — booleans (`FEATURE_ERP=1`, `FEATURE_BIGBROTHER=0`).
- Internal service URLs are **not** env vars anymore — everything listens on
  localhost inside the container (Postgres on Unix socket, NATS/Redis/MinIO on
  `127.0.0.1:<port>`).

Secrets (DB passwords for internal services, JWT signing key) are generated at
first-boot if absent and stored under the volume, never in the image.

### Commands

```bash
make build-image                 # builds sda:<version> from Dockerfile.all-in-one
make deploy TENANT=saldivia      # stop+rm+run for one tenant (data volume preserved)
make deploy TENANT=all           # iterate over deploy/tenants/*/
make status TENANT=saldivia      # docker ps + /readyz probe
make logs TENANT=saldivia        # s6-overlay prefixes each internal service
```

### Upgrade order

1. `make build-image` — tag a new `sda:<version>`.
2. For each tenant in `deploy/tenants/*/`:
   a. `docker stop sda-<tenant>` (volumes stay mounted).
   b. `docker rm sda-<tenant>`.
   c. `docker run --name sda-<tenant> -v /data/<tenant>/…:… sda:<version>`.
   d. Wait for `/readyz`; if red within 60 s, roll back to the prior image for
      this tenant and halt the upgrade — remaining tenants stay on the old version.
3. Migrations run **inside the new container at startup** (the Go app owns them
   and the DB is local). First-boot migrations are part of `s6-overlay/app/run`.

### Customer data is scoped by the container, not by code

A common mistake: adding "just one column to filter by tenant" in the shared
codebase. That reintroduces the pool-tenant posture (ADR 022 supersedes it; ADR
023 reinforces it). The answer is always: route the need through env vars +
per-container isolation, never through a tenant predicate in a shared query.

### Workstation specs

- **CPU:** AMD Threadripper Pro 9975WX — 32 cores / 64 threads
- **RAM:** 256 GB DDR5 ECC
- **Storage:** 8 TB M.2 NVMe
- **GPU:** 1× NVIDIA RTX PRO 6000 Blackwell Q Edition — 96 GB VRAM

Implications when choosing defaults:

- **Postgres** can be tuned generously (`shared_buffers=64GB`, `work_mem=256MB`, huge pages).
  Don't leave it on container defaults — the hardware is wasted.
- **`GOMAXPROCS` is 64** unless explicitly constrained. `go test` and build pipelines
  can use `-p 32` safely.
- **A single RTX PRO 6000 with 96 GB VRAM** fits a 70-72B model in FP8/NVFP4 with a
  large context. No need to shard across GPUs; pick models that fit rather than splitting.
- **Network latency between containers is ~µs** (same kernel, same NUMA). Don't design
  around gRPC/HTTP cost for things that used to be function calls.
- **Storage** is fast enough to not need aggressive caching; `io_uring` / `O_DIRECT`
  patterns are overkill for this workload.

## Workstation access

Host: `srv-ia-01` (`172.22.100.23`), user `sistemas`. SSH. Credentials live in
your local `~/.ssh/config` or password manager — **never committed**.

```
Host srv-ia-01
  HostName 172.22.100.23
  User sistemas
  IdentityFile ~/.ssh/id_ed25519
```

From outside the office, the WireGuard VPN must be up first. For network layout,
VPN config, server inventory, and how to handle credentials, read the
`infrastructure-access` skill — that is the authoritative source for anything
touching `172.22.x.x`.

### Frontend in the browser

Once the VPN is up, the frontend is reachable at the workstation's internal
hostname. If the VPN routes traffic but the browser won't load:

```bash
ssh -L 3000:localhost:3000 -L 8080:localhost:8080 srv-ia-01
# then open http://localhost:3000
```

Forward every port you need to probe: Traefik (`:8080`/`:80`), frontend (`:3000`),
direct service ports (`:8001-:8013`).

## Commands

```bash
make deploy               # full deploy: build, push, pull on workstation, restart
make status               # per-service health + GPU check
make versions             # running version vs repo VERSION per service
```

## Versioning

Each service has a plain-text `services/<name>/VERSION` file. Bump it in the same
PR as the change. `make deploy` tags images with that version + git SHA.

## Dockerfiles

- Multi-stage: `golang:<version> AS build` → `gcr.io/distroless/base` runtime.
- `COPY --from=build /out/<svc> /<svc>` — only the binary, no sources.
- Non-root user.
- `HEALTHCHECK` hitting `/healthz` (which every service exposes via `pkg/health`).

## Traefik

- Config in `deploy/traefik/`.
- Routes by host + path: `api.sda.local/auth/*` → auth service.
- JWT is **not** verified at the edge — each service does its own verification (see
  `auth-security`). Edge trusts nothing.

## Rollout

1. CI builds + pushes on tag.
2. `make deploy` SSHes to the workstation, pulls the new image, recreates the
   container, waits for `/healthz`.
3. On failure: container stays on the old image; nothing rolls forward.

## Rollback

```bash
tools/cli/sda deploy rollback <service> <version>
# or:
docker compose -f deploy/docker-compose.prod.yml up -d --no-deps <service>
# after pinning the image tag in the compose file
```

## Health

Every service exposes:

- `GET /healthz` — liveness (returns 200 as long as the process runs).
- `GET /readyz` — readiness (checks DB, NATS, upstream deps).

Traefik uses `readyz` for routing. If `readyz` is red, the container is not in the
load-balancer.

## Logs

- `docker logs -f sda-<svc>` on the workstation.
- Logs are JSON (slog). Ship to a collector if one is configured (check
  `deploy/docker-compose.prod.yml` for a `vector` or `loki` container).

## Incident triage

When something is on fire:

1. `make status` — which service is red?
2. `make versions` — did we just deploy something?
3. `docker logs --tail 200 sda-<svc>` — what is it complaining about?
4. Hit `readyz` directly — which dependency is down?
5. If the service is looping on start: read the first 100 log lines, not the last.

Read the `systematic-debugging` skill if you are investigating a non-obvious issue.

## GPU

- `nvidia-smi` on the workstation — check VRAM usage and temps.
- SGLang servers boot from `deploy/docker-compose.dev.yml --profile gpu` in dev.
- If a model OOMs on load: the first thing to check is batch size in the SGLang
  startup args, not the model itself.

## Don't

- Don't run migrations as part of a service's startup — migrations are a separate
  step in the deploy pipeline.
- Don't pin `:latest` in compose — always a version tag.
- Don't SSH and edit files on the workstation. Change the repo, deploy, verify.

## Known anti-patterns in this repo (fix on contact)

Audited and real; whenever you touch a file below, fix the pattern. These are
not "nice to have" — they are the reason the harness principle #0 (reduce) exists.

### Images

- **`:latest` tags** in `deploy/docker-compose.dev.yml` (mailpit, minio, sglang ×2)
  and `deploy/docker-compose.prod.yml` (crowdsec). **Pin a version** on sight.
- **Go Dockerfiles without `HEALTHCHECK`.** Every Go service image must declare
  one hitting `/readyz`. Traefik can mitigate but the image must still be honest.
- **`python:3.12-slim`** in `services/extractor/Dockerfile`. Go services → **distroless**.
  Python stays slim-ish but must not `apt install` in runtime stage.

### Makefile (root)

- **`.PHONY` is incomplete** (missing `events-gen`, `sqlc`, `proto`, deploy-*,
  `rollback-%`, `test-%`, `lint-%`, `build-%`, etc.). **Keep it updated.**
- **Hardcoded JWT base64 keys at lines 47–48** — they are "dev-only" but violate
  the "no credentials in repo" rule. Move to `deploy/.env.dev.example` (not
  committed as `.env`, only as `.example`).
- **`cd` inside recipes without subshell** (`@cd apps/web && ...`). Always use
  subshell form: `(cd apps/web && ...)` so the `cd` doesn't leak.
- **`sleep 15 && ...`** as synchronization in `dev-all`. **Never use sleep to wait
  for readiness.** Loop on `curl -sf .../readyz` with a retry count.

### docker compose layout

- **4 compose files** (`dev`, `prod`, `test`, `observability`) with overlapping
  service declarations. **Target:** one `docker-compose.yml` with `profiles:`
  (`[dev, prod, gpu, full, observability, test]`). Until that refactor lands,
  any edit to a service spec updates **all** files that declare it.
- When the silo model lands (ADR 022), each tenant's stack inherits a single
  `compose.base.yml` and layers its own `compose.yml` + `.env`. Don't fork the
  base per tenant.

### CI

- **docker-build runs after go-test sequentially** (`ci.yml`). Matrix can run in
  parallel with security, not after it. Unblock the dag.
- **go-build and go-vet are parallel** but `go-test` needs *both*. Fine, but
  `go-test-integration` only needs `go-build` — shave a step there.
- **Discovery of services uses a grep+jq trick** that hardcodes the `extractor`
  exclusion. When the service list changes the matrix must not silently drop a service.

### Scripts

- All non-trivial bash scripts need `set -euo pipefail` and `IFS=$'\n\t'` at the top.
  `deploy/scripts/*.sh` mostly have it; any new script that doesn't **fails review**.
- Quote every variable expansion that could contain a space or an empty value.
- No `sleep N` as synchronization. Use a bounded readiness loop.

### Observability

- **OTel endpoint is configured in compose but no collector is running by default**
  in dev. Either run the observability stack as a profile (`--profile observability`)
  whenever services are up, or remove the env var so we're not lying about tracing.
- No Prometheus scrape config in dev. Add it under the observability profile or
  drop the metrics endpoints.
