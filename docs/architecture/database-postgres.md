---
title: Database (PostgreSQL per-tenant)
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/database.md
  - ../packages/tenant.md
  - ../conventions/sqlc.md
  - ../conventions/migrations.md
  - multi-tenancy.md
---

This document describes how SDA partitions data across PostgreSQL
databases. Read it before adding a new table, changing the migration
layout, or wiring a service to a new pool — the platform-vs-tenant split
and the migration ordering are invariants the rest of the system assumes.

## Two schemas, two roles

| Database         | Role                                  | Owner package(s)                |
|------------------|---------------------------------------|---------------------------------|
| `sda_platform`   | Control plane: tenants, modules, plans, feature flags, deploy log, traces metadata, audit log, intelligence config, healthwatch state. | `pkg/tenant`, `services/platform`, `services/traces`, `services/healthwatch` |
| Per-tenant DB    | Application data: users, roles, sessions, messages, documents, document_trees, notifications, feedback, ERP, astro contacts, etc. | `services/auth`, `services/chat`, `services/ingest`, `services/notification`, `services/erp`, `services/astro`, ... |

Each tenant's connection URL lives in `tenants.postgres_url` /
`postgres_url_enc` in the platform DB
(`db/platform/migrations/001_init.up.sql:19`,
`db/platform/migrations/003_encrypted_credentials.up.sql`). Tenants do not
share a database — adding a tenant means provisioning a new PostgreSQL
instance / database and writing the row.

## Pool resolution

- **Platform DB:** opened once at service startup via
  `database.NewPool(ctx, POSTGRES_PLATFORM_URL)`
  (`pkg/database/pool.go:22`). Used directly by `platform`, `traces`,
  `healthwatch`, and by `tenant.Resolver` for tenant lookups.
- **Tenant DB:** resolved per request through `tenant.Resolver`
  (`pkg/tenant/resolver.go:32`). Pools are cached by slug, capped at
  `PoolMaxConns = 4` per tenant by default, and health-checked
  periodically. See multi-tenancy.md.

`ensureSSLMode` injects `sslmode=require` whenever the URL omits one
(`pkg/tenant/resolver.go:357`) — production connections are always
encrypted in transit.

## Migration layout

```
db/
  platform/migrations/   # NNN_topic.up.sql + NNN_topic.down.sql
  tenant/migrations/     # NNN_topic.up.sql + NNN_topic.down.sql
```

Numbers are sequential, no gaps allowed (`bash .claude/hooks/check-invariants.sh`
enforces). Every `*.up.sql` has a paired `*.down.sql`. Today: 18 platform
migrations, 108 tenant migrations.

The first three tenant migrations stand up the base every tenant has:
`001_auth_init` (users, roles, refresh tokens, audit), `003_chat_init`
(sessions, messages), `006_ingest_init` (documents, trees, pages). Later
migrations layer on notifications, intelligence, ERP modules, astro, and
bigbrother.

## sqlc

Each Go service that touches its database has `db/queries/*.sql` and
`db/sqlc.yaml`; running `make sqlc` generates a typed `repository` package
in `services/{name}/internal/repository`. The schema input always points to
the central migrations directory — services do not own schema files
(`services/auth/db/sqlc.yaml:5`). Generated code lives in-repo so the
build never depends on `sqlc` being installed.

When changing a query: edit `db/queries/*.sql`, run `make sqlc`, commit
both the SQL and the generated `.go` together. The pre-commit invariant
check warns if generated code is older than the queries.

## Row-Level Security (where it exists)

The ERP modules enable PostgreSQL RLS on every table with a
`tenant_isolation` policy keyed on `current_setting('app.tenant_id', true)`
(e.g. `db/tenant/migrations/017_erp_entities.up.sql:82`). Services activate
it per transaction with `database.SetTenantID(ctx, tx, tenantID)`
(`pkg/database/pool.go:54`). RLS is **defence in depth** — the primary
isolation is the per-tenant pool; you should not rely on RLS to fix a
missing `tenant_id` filter in application code.

## Tracing and pooling

`pkg/database/pool.go` is intentionally thin. The package note documents
the planned hook for `otelpgx` so SQL queries appear as spans in Tempo
once that dependency lands. Until then, query latency is part of the HTTP
handler span.

## What you must never do

- Open a raw PostgreSQL pool against a tenant DB outside `tenant.Resolver`.
- Hand-write Go SQL — use sqlc-generated repositories.
- Skip the `.down.sql`. The invariant check blocks the commit.
- Add a tenant-scoped table to the platform DB or vice versa.
- Reuse a tenant pool for a different slug.
