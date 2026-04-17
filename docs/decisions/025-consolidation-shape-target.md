# ADR 025 — Consolidation shape target: 5 domain modules in 1 Go binary

**Status:** accepted
**Date:** 2026-04-17
**Deciders:** Enzo Saldivia
**Refines:** ADR 021 (reduce before add) — fills in the concrete shape
**Relates to:** ADR 023 (one container per tenant), ADR 024 (frontend inside tenant container)

## Context

ADR 021 set the direction: 13 Go services → 3–5 domain modules, "one area at
a time", no big-bang rewrite. It deliberately did not name the groups —
that was deferred until enough signal existed to decide well.

After the all-in-one Dockerfile prototype landed (commits `ee7a909b..e3beb3f6`
on `2.0.6`), the signal is in hand. Mapping the 13 services produced:

- **Zero cross-service imports at the Go level.** Every service module is
  an island; services communicate only via HTTP, gRPC, or NATS. Consolidation
  is *move packages*, not *refactor dependencies*.
- **Runtime coupling graph** (who calls whom at runtime):
  - `agent` → `search`, `ingest`, `notification`   (HTTP + gRPC)
  - `ws` → `chat`   (gRPC)
  - Everyone else: standalone or NATS-consumer only.
- **gRPC server-side:** only `chat` and `search` expose gRPC.
- **NATS subscribers:** `feedback`, `healthwatch`, `ingest`, `notification`,
  `traces`, `ws`.
- **Service sizes (LOC of `.go`, excluding sqlc-generated `db/`):**
  - `erp` 37,609 — more than half the whole Go codebase
  - `ingest` 6,623 · `feedback` 6,171 · `auth` 5,614 · `notification` 5,474
  - `chat` 4,049 · `bigbrother` 3,494 · `platform` 3,307 · `agent` 2,990
  - `healthwatch` 2,056 · `ws` 1,833 · `traces` 1,790 · `search` 1,422

## Decision

### Target shape: **5 internal modules in 1 Go binary**

```
services/app/
├── cmd/main.go                 single entrypoint, owns :80 (chi router)
├── go.mod
└── internal/
    ├── core/                   auth + platform + feedback
    ├── rag/                    ingest + search + agent
    ├── realtime/               chat + ws + notification
    ├── ops/                    bigbrother + healthwatch + traces
    └── erp/                    erp (unchanged, just moved)
```

Grouping rationale — **by domain, not by noun** (per ADR 021):

| Module | Services absorbed | ≈LOC | Why |
|---|---|---|---|
| `core` | auth, platform, feedback | 15 k | JWT-gated CRUD, zero runtime coupling to others. Minimal blast radius. |
| `rag` | ingest, search, agent | 11 k | The product pipeline. `agent → search + ingest` already a call graph — in-process removes 2 HTTP hops. |
| `realtime` | chat, ws, notification | 11 k | User-facing realtime surface. `ws → chat` is gRPC today; notification consumes chat events. |
| `ops` | bigbrother, healthwatch, traces | 7 k | Passive NATS consumers + admin HTTP. Platform-lifecycle, not product. |
| `erp` | erp | 38 k | Isolated domain (Histrix legacy integration). Size makes folding risky. Stays alone. |

### Why 1 binary (not 5)

CLAUDE.md already committed to this: *"Inside the all-in-one container
(ADR 024) these are internal modules of one binary, not separate processes."*
This ADR ratifies that stance with the concrete list and spells out the
consequences:

- The **`deploy/frontdoor` binary dissolves** — its port-mapping routing becomes
  chi routing inside `cmd/main.go`. `/v1/core/*`, `/v1/rag/*`, etc. are just
  chi sub-routers. `/ws` is a chi handler. `/*` still reverse-proxies to
  Next.js on :3000 but from inside the monolith.
- **s6-overlay runs fewer services:** today 5 infra (postgres, nats, redis,
  minio, nextjs) + 1 frontdoor + eventually 13 Go backends. After
  consolidation: 5 infra + 1 `app` binary + nextjs = 7 services under s6.
- **Inter-service calls collapse to method calls.** `agent→search` goes
  from HTTP JSON to a Go function. The `pkg/sdagrpc`, HTTP clients, and
  the URLs in env config all go away.

### Migration sequence

The first fusion (`ops`) is the pilot. Subsequent fusions follow the same
pattern, in this order, each its own session / PR:

1. **`ops`** (pilot) — smallest, NATS-consumer shape, admin-only HTTP.
   Proves the pattern with minimum blast radius. **✅ done (2026-04-17).**
   bigbrother + healthwatch + traces absorbed into
   `services/app/internal/ops/{bigbrother,healthwatch,traces}/`. The three
   old `services/*` shells are deleted, `go.work` holds one `services/app`
   entry, the frontdoor map shrank 13 → 10 and Makefile / compose / deploy
   scripts route the three old names through `app` (port 8020). `app`
   binary is not yet under s6 supervision — that lands when more modules
   have fused (separate session). Followups surfaced: NATS per-service
   users, traefik dynamic configs, Grafana dashboards and `.env.example`
   still reference the 3 old names; harmless today but will need a sweep.
2. **`core`** — internal identity + config. **✅ done (2026-04-17).**
   auth + platform + feedback absorbed into
   `services/app/internal/core/{auth,platform,feedback}/`. The three
   old `services/*` shells are deleted; `go.work` holds 7 standalone
   service entries (agent, chat, erp, ingest, notification, search, ws)
   plus `services/app`; the frontdoor map shrank 10 → 7; Makefile /
   compose / deploy scripts drop the three names. Paths preserved
   (`/v1/auth/*`, `/v1/modules/*`, `/v1/platform/*`, `/v1/flags/*`,
   `/v1/feedback/*`, `/v1/platform/feedback/*`) so clients see zero
   change. Single-tenant path hardening shipped with the move:
   `handler.NewMultiTenantAuth` + `pkg/tenant.Resolver` wiring deleted
   from auth (dead code under ADR 022), replaced by a
   `handler.SetJWTConfig` seam called in `wireAuth`. One
   `pkg/outbox.NewDrainer` per silo takes over from the registry of
   per-tenant drainers. Same open followups as ops pilot (NATS per-
   service users, traefik dynamic, Grafana, `.env.example`); non-
   blocking, tracked for the session that lands the `app` binary
   under s6.
3. **`rag`** — removes the biggest runtime coupling (agent→search+ingest).
   **✅ done (2026-04-17).** ingest + search + agent absorbed into
   `services/app/internal/rag/{ingest,search,agent}/`. The three old
   `services/*` shells are deleted; `go.work` holds four standalone
   service entries (chat, erp, notification, ws) plus `services/app`;
   the frontdoor map shrank 7 → 4; Makefile / compose / deploy scripts
   drop the three names. Paths preserved (`/v1/ingest/*`, `/v1/search/*`,
   `/v1/agent/*`) so clients see zero change. First fusion that
   deletes process-boundary scaffolding that used to be load-bearing:
   - agent→search gRPC is gone. The `:50051` listener,
     `searchv1.RegisterSearchServiceServer`, `handler/grpc.go` and the
     agent-side `GRPCSearchClient` (~170 LOC) are all deleted because
     agent — search's only consumer — now calls `SearchDocuments`
     in-process. `pkg/grpc` survives because chat↔ws still uses it;
     that link collapses in the realtime fusion.
   - `agenttools.Executor` grew a `SearchBackend` / `IngestBackend`
     seam. Core tools (search_documents, check_job_status) dispatch
     in-process via the struct-concrete path; cross-module tools
     (notification, bigbrother, erp) still ride HTTP.
   - Env vars `SEARCH_SERVICE_URL`, `SEARCH_GRPC_URL`,
     `INGEST_SERVICE_URL` are no longer read anywhere and got dropped
     from compose + Makefile.
   Same open followups as ops + core pilots (NATS per-service users,
   traefik dynamic, Grafana, `.env.example`); non-blocking, tracked
   for the session that lands the `app` binary under s6.
4. **`realtime`** — user-facing, ws/chat gRPC removal. Test coverage matters
   here; do it after `rag` validates the pattern end-to-end.
5. **`erp`** — only if/when ERP shrinks. 38 k LOC folded in one move is
   too big; defer or split `erp` internally first.

Each fusion session:

- Create `services/app/internal/<module>/` by moving code from the N
  absorbed services.
- Wire its chi sub-router into `cmd/main.go`.
- Wire its NATS subscribers / gRPC handlers directly.
- Delete the absorbed `services/<svc>/` directories.
- Update `go.work` (remove N entries, add `services/app` once).
- Update `deploy/Dockerfile.all-in-one`: one binary to build, one s6
  longrun, frontdoor upstream map shrinks per fusion (deleted entirely
  once all 13 are absorbed).
- Update ADR status/tracking in this file (tick off the module).

### What the frontdoor's upstream map becomes

Today, `deploy/frontdoor/main.go` has 13 entries. After each fusion:

- After `ops`:      map shrinks by 3 → 10 entries. `ops` routes land in app.
- After `core`:     map shrinks by 3 → 7 entries.
- After `rag`:      map shrinks by 3 → 4 entries. **✅ done.**
- After `realtime`: map shrinks by 3 → 1 entry (`erp` only).
- After `erp`:      **map empty, frontdoor binary deleted.**

`cmd/sda/main.go` then owns `:80` directly; `/*` proxy to Next.js lives
in its chi router as one final handler.

## Consequences

**Positive**

- Single Go binary to build, deploy, version. One set of flags, one process
  tree, one logging stream. Matches one-dev operational reality.
- Inter-service HTTP/gRPC hops disappear. RAG pipeline in particular
  (agent→search, agent→ingest) becomes nanosecond-scale function calls.
- `pkg/sdagrpc`, per-service HTTP clients, service-URL env vars
  (`SEARCH_SERVICE_URL`, `INGEST_SERVICE_URL`, etc.) all become deletable.
- Frontdoor binary and its upstream map both go to zero. `chi` owns
  the routing tree end to end.
- Tests simplify: no cross-service mocking, no testcontainers per
  service. One set of integration tests against one binary.
- Fewer s6 services to supervise; faster container boot.

**Negative**

- Short-term refactor cost, paid in 5 sessions (one per module).
  Accepted per ADR 021.
- Single binary = single restart impact. Today, restarting `auth` leaves
  `chat` untouched; after fusion, any rebuild restarts everything. For
  a one-dev inhouse deploy this is a feature, not a bug — one thing to
  restart. For a hypothetical future where a module needs independent
  release cadence, *that* module can be re-extracted.
- Internal name collisions (every absorbed service has a `handler`,
  `service`, `repository` package) must be resolved on the way in.
  Handled per fusion — rename to `internal/<module>/{handler,service,repo}`
  and package paths handle the rest.
- `erp` at 38 k LOC is too big to fold naively. Deferred until `erp`
  itself is consolidated internally.

**Neutral**

- Does not affect ADR 022 (silo model) or ADR 023 (one container per
  tenant) — those are deployment decisions; this is a code-shape one.
- `apps/web` (Next.js) is unchanged — this ADR only touches Go code.

## Alternatives considered

1. **5 binaries, one per module, each as its own s6 longrun.**
   Rejected. Every counter-argument in ADR 021 still applies: network
   serialization between processes on the same kernel is pure tax; 5
   Dockerfiles / version files / health probes multiply boilerplate.
   CLAUDE.md explicitly states "internal modules of one binary".

2. **Keep 13 services, fuse only shared code into `pkg/`.**
   Rejected. That's what the current shape already is and it doesn't
   buy the consolidation ADR 021 called for. Per-service main.go,
   per-service Dockerfile, per-service JWT wiring, per-service HTTP
   client to every other service remains.

3. **Collapse all 13 into one module (no sub-modules).**
   Rejected. The five domains are genuinely distinct; flattening loses
   the grouping signal that makes `internal/` self-documenting. The
   cost of 5 sub-packages is near zero.

4. **Different module names** (e.g. `identity` instead of `core`,
   `documents` instead of `rag`).
   Considered; rejected for churn. `core/rag/realtime/ops/erp` fit the
   project's existing vocabulary and are short enough for daily use.

5. **Fuse `erp` first since it's the biggest.**
   Rejected. `erp` is the highest-risk and has the most internal
   structure already; it's the *last* candidate, not the first. Ship
   the pattern on the smallest domain (`ops`) first.

## Open items

- Whether `erp` ever actually folds in, or stays as a dedicated
  deployable forever. Decide after `realtime` lands — by then `services/app`
  will have real production mileage.
- Migration of service-owned migrations (`services/<svc>/db/migrations/`)
  into a single tenant migrations tree under `services/app/db/`. Handle in
  the pilot session; the migration numbering convention (ADR from the
  `database` skill) already supports this.
- Whether `/v1/core/*` becomes the external URL or we preserve the
  original path prefixes (`/v1/auth/*`, `/v1/platform/*`, etc.) to avoid
  breaking clients. Lean toward preservation — chi sub-routers can mount
  at the old paths with zero client-side change.
