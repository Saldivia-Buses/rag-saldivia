---
title: Convention: Database Migrations
audience: ai
last_reviewed: 2026-04-15
related:
  - ./sqlc.md
  - ../architecture/database-postgres.md
  - ../architecture/multi-tenancy.md
---

Rules for PostgreSQL migrations under `db/platform/migrations/` (platform admin DB) and `db/tenant/migrations/` (every tenant's DB). Per-service migrations live under `services/{name}/db/migrations/` only when explicitly scoped to that service.

## File naming

`NNN_<short_description>.up.sql` paired with `NNN_<short_description>.down.sql`.

DO use the next sequential 3-digit number with no gaps. Look at the target migrations dir and increment from the highest existing number.

DO match the `up` and `down` base names exactly: `001_auth_init.up.sql` ↔ `001_auth_init.down.sql`.

DON'T renumber existing migrations. Once a migration ships, the number is permanent.

DON'T commit an `.up.sql` without its `.down.sql`. The pre-commit invariant check fails the commit.

## Pair completeness

DO write a real reversal in every `down.sql`. `DROP TABLE IF EXISTS users CASCADE;` for table creation, `DROP COLUMN` for column adds, `DELETE` for seeded rows.

DON'T leave `down.sql` empty or marked TODO. If a change is genuinely irreversible (data loss), document why in a comment and emit `RAISE EXCEPTION` so the migration cannot be rolled back accidentally.

## Idempotency

DO use `IF NOT EXISTS` on `CREATE TABLE`, `CREATE INDEX`, and `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`. Migrations may re-run during recovery — they must be safe.

DO use `IF EXISTS` on `DROP` statements in `down.sql`.

See `db/tenant/migrations/001_auth_init.up.sql:5` for the canonical pattern.

## Schema rules

DO use `TIMESTAMPTZ` for time columns with `DEFAULT now()` and `NOT NULL` where applicable. Never `TIMESTAMP` (timezone-naive).

DO use `TEXT` (not `VARCHAR(n)`) for variable strings; PostgreSQL stores them identically. Add `CHECK (length(col) <= N)` if a hard upper bound is required.

DO use `gen_random_uuid()::text` for primary keys unless a natural key is well-justified.

DO add `CHECK` constraints for enums and value ranges (`CHECK (status IN ('pending','active','revoked'))`). Never enforce at the application layer alone.

DO use `ON DELETE CASCADE` for child rows whose existence depends entirely on a parent (e.g. `user_roles → users`). Use `ON DELETE SET NULL` for nullable references.

## Permissions and RBAC

DO insert a `role_permissions` row for every new permission introduced by the migration. New endpoints check permissions at the handler — without the seeded mapping, every user is denied.

DO use `INSERT ... ON CONFLICT DO NOTHING` for seeded permissions and roles so the migration is replay-safe.

## sqlc regeneration

DO run `make sqlc` after any schema change touching tables referenced in `db/queries/*.sql`. Commit the regenerated code in the same commit as the migration.

DON'T edit `internal/repository/*.sql.go` by hand — sqlc owns those files. The pre-commit invariant compares query timestamps vs generated code timestamps and warns on drift.

See [sqlc](./sqlc.md) for query authoring rules.

## Multi-tenant scope

`db/tenant/migrations/` runs against every tenant's database (per-tenant PostgreSQL). Schema must be identical across tenants. No `tenant_id` column needed in tenant-DB tables — isolation is at the database level.

`db/platform/migrations/` runs once on the platform DB. Tables here track tenants, deploys, infra alerts, healthwatch snapshots.

DO place a comment header at the top of every `up.sql` declaring scope and intent (see `db/tenant/migrations/001_auth_init.up.sql:1`).

DON'T mix platform and tenant tables in one migration file.

## Indexes

DO create indexes in the same migration that creates the column they support. Use `IF NOT EXISTS`.

DON'T add `CREATE INDEX CONCURRENTLY` inside a transaction-wrapped migration runner — it errors. If the table is large in production, ship it as a separate migration documented as concurrent.

## Testing migrations

Bring up a fresh database with `make migrate`. Verify the `down.sql` by running migrate down then up — both must succeed without errors. CI runs the full migration sequence on a clean database in the Test gate.
