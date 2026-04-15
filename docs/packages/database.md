---
title: Package: pkg/database
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./tenant.md
  - ../architecture/database-postgres.md
---

## Purpose

Thin pgxpool helpers for SDA services. `NewPool` constructs a connection pool
from a Postgres URL — ready to be wrapped in OpenTelemetry tracing once
`otelpgx` is added to `go.mod`. `SetTenantID` sets PostgreSQL Row-Level
Security context inside a transaction. Import this when wiring a service's
tenant pool or when relying on RLS policies.

## Public API

Source: `pkg/database/pool.go:10`

| Symbol | Kind | Description |
|--------|------|-------------|
| `NewPool(ctx, dbURL)` | func | Creates a `*pgxpool.Pool` from a Postgres URL |
| `SetTenantID(ctx, tx, tenantID)` | func | `SET LOCAL app.tenant_id = $1` for RLS-filtered queries |

## Usage

```go
pool, err := database.NewPool(ctx, dbURL)
if err != nil { return err }
defer pool.Close()

tx, _ := pool.Begin(ctx)
defer tx.Rollback(ctx)
if err := database.SetTenantID(ctx, tx, slug); err != nil { return err }
// all queries in tx are now filtered by current_setting('app.tenant_id')
tx.Commit(ctx)
```

## Invariants

- Pools should be per-tenant (use `pkg/tenant.Resolver`) or per-platform — never
  share one pool across tenants.
- `SetTenantID` uses `SET LOCAL`, so the binding only lives until the
  transaction ends. Call it at the start of every RLS-scoped transaction.
- The TODO at `pkg/database/pool.go:28` notes that uncommenting one line will
  add OpenTelemetry query tracing once `otelpgx` is in `go.mod` — every SQL
  query then appears as a span in Tempo.

## Importers

`services/auth`, `astro`, `bigbrother`, `chat`, `erp`, `feedback`, `healthwatch`,
`ingest`, `notification`, `platform`, `traces`, `search` (typically only in
`cmd/main.go`).
