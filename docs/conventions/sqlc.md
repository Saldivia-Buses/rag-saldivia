---
title: Convention: sqlc Queries
audience: ai
last_reviewed: 2026-04-15
related:
  - ./migrations.md
  - ./security.md
  - ../architecture/database-postgres.md
  - ../architecture/multi-tenancy.md
---

Rules for `.sql` query files under `services/{name}/db/queries/`. sqlc reads these and generates type-safe Go in `services/{name}/internal/repository/`. Regenerate with `make sqlc` after any change.

## File organisation

DO group queries by entity in one `.sql` file per logical area (`auth.sql`, `chat.sql`, `audit.sql`).

DO put the comment `-- name: <FuncName> :<kind>` on the line before each query. Kinds: `:one`, `:many`, `:exec`, `:execrows`, `:batchexec`.

See `services/auth/db/queries/auth.sql:3` for the canonical pattern.

## Query annotations

DO match annotation to expected result:

| Annotation | When |
|---|---|
| `:one` | Exactly one row expected; returns `(Row, error)` with `pgx.ErrNoRows` if missing |
| `:many` | Zero-to-many rows; returns `([]Row, error)` |
| `:exec` | No result needed (INSERT/UPDATE/DELETE without RETURNING) |
| `:execrows` | Need the affected row count |

DON'T use `:one` for queries that may return zero rows unless callers explicitly handle `pgx.ErrNoRows`. Prefer `EXISTS(...)` returning a boolean.

## Parameters

DO use positional parameters (`$1`, `$2`) for simple cases. They map to function args in order.

DO use named parameters (`@max_failed::int`) when the same value is referenced multiple times in one query, or when the type cast clarifies intent. See `services/auth/db/queries/auth.sql:65`.

DO cast at the call site: `@user_id::text`, `@count::int`. sqlc infers types from the cast.

DON'T concatenate strings into queries — sqlc only generates safe code from `$N`/`@name` placeholders. Dynamic SQL is forbidden.

## Tenant isolation

This is an architectural invariant. Violations break tenant security and are blocked by review.

DO include `WHERE tenant_id = $1` (or whatever the param name is) in **every** query against a table that holds tenant-scoped data.

For per-tenant databases (the default in SDA — see [multi-tenancy](../architecture/multi-tenancy.md)), the database itself is the isolation boundary and no `tenant_id` column exists. The tenant pool is selected by `pkg/tenant.Resolver` from the JWT slug. DON'T ever look up a pool by a value that came from the request body — only from JWT claims validated by middleware.

For platform tables (single DB shared across tenants), the column is mandatory and must appear in every WHERE.

## Result types

DO return only the columns the caller needs. `SELECT id, name, email FROM users WHERE ...` not `SELECT *`.

DO use sqlc's automatic struct generation. Don't write hand-rolled scan code in `repository/`.

DO name returned columns deliberately when joining (`SELECT u.name AS user_name, r.name AS role_name`). sqlc derives Go field names from these.

## Mutations

DO use `RETURNING id, created_at` when the caller needs the generated ID or timestamp. Avoid a follow-up `SELECT`.

DO wrap multi-statement mutations in a transaction at the service layer (`pgx.Tx`), not in the SQL file. sqlc supports `WithTx(tx)` on the generated `Queries`.

DON'T expose row counts as the API success signal. A `:execrows` returning 0 means no row matched — surface as `httperr.NotFound`.

## Regeneration workflow

1. Edit `db/queries/<file>.sql`.
2. If the schema also changed, write the migration first (see [migrations](./migrations.md)).
3. Run `make sqlc` from the repo root.
4. Inspect the diff in `internal/repository/`. Don't hand-edit those files.
5. Commit the `.sql`, the migration (if any), and the generated Go in **one commit**.

The pre-commit invariant warns when query files are newer than generated code.

## Forbidden patterns

- `SELECT *` — explicit columns only
- Dynamic table names or column names (sqlc cannot generate types for them)
- Cross-tenant queries (no query joins data from multiple tenant DBs — that's NATS or platform-DB territory)
- Direct `pgx.Conn.Exec` in service code that bypasses generated `Queries` methods
