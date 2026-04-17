---
name: backend-go
description: Use when editing Go code in services/** or pkg/**. Covers service layout, chi routing, JWT middleware wiring, sqlc usage, error handling with pkg/httperr, NATS publish/subscribe patterns, testcontainers-based testing, Go workspace conventions, commit style, and blast-radius discipline for changes to pkg/*.
---

# backend-go

Scope: `services/**/*.go`, `pkg/**/*.go`, `tools/cli/**/*.go`, `go.work`, `go.mod`.

## Before you create anything new

This project is already over-serviced for a one-dev box. Before adding a new
service or `pkg/` package, answer **all three**:

1. **Can this be a function or a sub-package inside an existing service?** If yes, do that.
2. **Is there a `pkg/*` that already owns this concern?** If kind-of, extend it instead of forking.
3. **Does this justify a process boundary?** Separate binaries pay a cost (deploy, health,
   logs, JWT pass-through, RPC serialization). The reward has to be real — independent
   release cadence, different runtime, different scaling profile — not "it feels cleaner".

If you can't defend the third point out loud, the new service/package doesn't happen.
Add it as a module inside an existing one.

## Service skeleton (non-negotiable)

Every service in `services/<name>/` has:

```
cmd/main.go          # entry; wires config, DB, NATS, server
internal/            # handlers, service, repository (not importable)
internal/handler/    # chi routes; one file per entity
internal/service/    # business logic; no HTTP concerns
internal/repository/ # sqlc-generated + thin wrappers
Dockerfile
VERSION              # plain text, semver, bumped per release
README.md            # what the service does + endpoints
```

Registered in `go.work`. If it is not in `go.work`, `make build` skips it.

## Commands

```bash
make build-<svc>            # binary in bin/<svc>
make test-<svc>             # go test ./services/<svc>/...
go test ./pkg/<pkg>/...     # package tests
go test -run TestName ./... # single test
go test -race ./...         # race detector — use before merge
```

## Routing (chi)

```go
r := chi.NewRouter()
r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)
r.Use(tenantmw.Resolve(cfg.TenantResolver))
r.Group(func(r chi.Router) {
    r.Use(jwtmw.Verify(cfg.JWTPublicKey))
    r.Post("/v1/chat", h.CreateChat)
})
```

Order matters: `Recoverer` outermost, `tenant` before `jwt` (tenant is a URL concern,
JWT is request identity).

## JWT

- `pkg/jwt` verifies ed25519 signatures. Public key comes from config.
- `pkg/middleware/jwtmw` extracts claims into `ctx`.
- **No tenant middleware.** Since ADR 022 the deployment is single-tenant — no
  `pkg/tenant`, no slug resolution, no pool-per-tenant. The DB pool is singular
  and lives in the service's config.

Read the `auth-security` skill before touching anything in this area.

## sqlc

- Queries in `services/<svc>/internal/repository/query.sql` (or split per entity).
- Config in `sqlc.yaml` at repo root.
- `make sqlc` regenerates. Commit the generated code.
- Never hand-edit `*.sql.go`.

Read the `database` skill for migration + sqlc details.

## NATS + outbox

```go
import "github.com/Camionerou/rag-saldivia/pkg/nats"

publisher := nats.NewPublisher(conn, "chat.message.created")
_ = publisher.Publish(ctx, event)
```

Subject format is **flat** (ADR 022): `{service}.{entity}[.{action}]`. The NATS
process runs **inside the tenant container** (ADR 023) on `127.0.0.1:4222`. It
is tenant-scoped by construction — the process only sees this tenant's traffic.
Use `pkg/nats.Subject()` builder or constants, never ad-hoc concatenation.

Once the consolidation from ADR 021 lands, evaluate whether an in-process Go
event bus can replace NATS entirely inside the container. Track as follow-up
work; don't remove NATS speculatively.

For cross-service writes that need at-least-once semantics: use `pkg/outbox`.

### Idempotency is the consumer's job

NATS JetStream guarantees at-least-once, not exactly-once. A consumer **must**
tolerate the same message arriving twice. Pattern:

- Every event carries a stable `event_id` (UUID) set by the publisher.
- The consumer has an `idempotency_keys` table keyed by `(consumer, event_id)`.
- The consumer's handler wraps DB work + insert-into-idempotency-keys in one
  transaction. If the insert conflicts (duplicate), the effect is a no-op.

Never use Postgres `LISTEN/NOTIFY` for cross-service delivery — it drops messages
under disconnect and breaks the at-least-once guarantee.

## Errors

- Domain errors: `pkg/httperr` with a stable `Code` string.
- At the handler edge: `httperr.Render(w, err)` — never `http.Error(...)`.
- `errors.Is`/`errors.As` for propagation. Never string-match.
- Wrap with `%w`, add context: `fmt.Errorf("fetch chat %s: %w", id, err)`.

## Context

- Every function that does I/O takes `ctx context.Context` as its first param.
- Never store ctx in a struct. Never pass `context.Background()` from a handler.
- Propagate deadlines: if an upstream has 5s, give downstream 4.5s.

## Logging (slog)

- `log/slog` with a JSON handler in prod. Never `log.Printf`, never `fmt.Println`.
- Use a **context-aware handler** that pulls `tenant_id`, `request_id`, `user_id`,
  and OTel `trace_id` from `ctx` automatically — so every call site writes
  `slog.InfoContext(ctx, "msg", "k", v)` and the common fields appear without
  boilerplate.
- `slog.With(...)` returns a scoped logger for loops/services; bind once, reuse.
- Never log tokens, passwords, full JWTs, user PII, full document contents.
- Error level: `slog.ErrorContext(ctx, "msg", "err", err)` — `%w`-wrapped err still
  renders readably.

## Testing

- Table-driven by default. Testify for assertions.
- Integration tests hit a real Postgres + NATS via `testcontainers-go`. Never mock them.
- Name: `TestXxx` for unit, `TestXxx_Integration` with `//go:build integration`.
- One `_test.go` file per production file.

### testcontainers — shared container pattern

Don't spin up one container per test — it's slow and wastes resources. Boot once
per package in `TestMain`, hand each test a **unique schema** or **unique DB name**:

```go
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
    ctx := context.Background()
    pgC, _ := postgres.Run(ctx, "postgres:17-alpine",
        postgres.WithDatabase("test"), postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    defer pgC.Terminate(ctx)
    dsn, _ := pgC.ConnectionString(ctx, "sslmode=disable")
    testDB, _ = pgxpool.New(ctx, dsn)
    os.Exit(m.Run())
}

func TestCreateChat_Integration(t *testing.T) {
    schema := "t_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
    _, _ = testDB.Exec(ctx, "CREATE SCHEMA "+schema)
    t.Cleanup(func() { _, _ = testDB.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") })
    // run migrations into `schema`, then test
}
```

Tests in the same package run sequentially by default; use `t.Parallel()` when the
schema isolation holds.

## Commits

- Style: `type(scope): subject` — e.g. `fix(auth): rotate refresh token on login`.
- Subject ≤ 72 chars, imperative, no trailing period.
- One logical change per commit. If the diff needs two paragraphs to explain, split it.
- Merges into `main`: squash.

## Blast-radius rule for `pkg/*`

Before changing any exported symbol in `pkg/`:

1. `rg -l "pkg/<name>" services/ pkg/` — list importers.
2. If ≥3 importers: break the change into two commits (add new, migrate callers, remove old).
3. Prefer additive changes over breaking renames.

## Known anti-patterns in this repo (fix on contact)

Audited and real. When you touch one of these, you fix it.

### Tenant-awareness residual (ADR 022 violation)

**~2,480 matches** of `tenant_id` / `tenant_slug` / `pkg/tenant` / `TenantSlug`
still in the tree. Representative offenders:

- `gen/go/auth/v1/auth.pb.go` + `gen/go/chat/v1/chat.pb.go` — `TenantSlug` in
  protobuf messages. Regenerate once the `.proto` drops the field.
- `services/feedback/internal/service/aggregator.go:40` — `Start(ctx, tenantID, tenantSlug)`.
- `services/ingest/internal/service/documents.go:36` — constructor takes tenant.
- `pkg/tenant/context.go` — the whole package is scheduled for deletion.
- `pkg/metrics/business.go:44–50` — every metric has a `tenant_slug` label; all
  useless in the silo model.

Rule: **any new code that writes a `tenant_id`, `tenant_slug`, or imports
`pkg/tenant` is a blocking review finding.** Existing instances are removed
progressively via `continuous-improvement` — they are listed in the anti-pattern
hunt list.

### Goroutines without a stop mechanism

Production goroutines need **ctx + a visible shutdown path**. No fire-and-forget.

```go
// Wrong — leaks on shutdown
go func() {
    for {
        time.Sleep(5 * time.Second)
        doWork()
    }
}()

// Correct
func (s *Service) runTicker(ctx context.Context) {
    t := time.NewTicker(5 * time.Second)
    defer t.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-t.C:
            s.doWork(ctx)
        }
    }
}
// launched as: go s.runTicker(ctx)
```

Current offenders to clean up on contact:
- `services/healthwatch/internal/service/healthwatch.go:280`
- `services/feedback/internal/service/aggregator.go:43`
- `services/chat/cmd/main.go:98`
- `services/ingest/internal/service/extractor_consumer.go:83`
- `services/erp/internal/handler/analytics.go:868` (has WaitGroup but no timeout)

### `panic()` in request paths

**Never panic in a handler or service method.** Return a `pkg/httperr` error.
Panic is reserved for program-wide invariant violations at startup (bad config,
missing key file) — never for runtime validation.

Current offenders:
- `services/ingest/internal/service/documents.go:38` — tenant validation.
- `services/bigbrother/internal/scanner/stub.go:34` — MAC parsing util.

### `fmt.Println` / `log.Printf` outside `tools/cli`

Production services use `slog.*Context`. CLI tools (`tools/cli/`) may use
`fmt.Println` for user-facing output, but **migrate to `slog.InfoContext` with a
text handler** when the output is a log rather than a report.

### main.go boilerplate across 14 services

Each service's `cmd/main.go` duplicates ~300–400 LOC of config load, DB pool, NATS
connect, health, middleware wire, server start. This is the single largest
consolidation opportunity in the codebase. **Target:** a shared
`pkg/server.Bootstrap(cfg)` that returns a configured `chi.Router` + shutdown
func; each service's main becomes <50 lines. Track as a `continuous-improvement`
target until all services migrate.

### `pkg/*` to delete

Confirmed candidates (0 live importers, or only dead references):

- `pkg/approval` — delete.
- `pkg/featureflags` — delete.
- `pkg/metrics` — keep the core, but strip `tenant_slug` labels (ADR 022).
- `pkg/cache` — inspect; Redis access happens through `pkg/database/pool` instead.
