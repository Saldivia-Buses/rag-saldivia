---
name: database
description: Use when touching SQL migrations, sqlc queries, or Postgres schema. Covers the migration numbering convention (platform vs tenant), the up/down pairing invariant, sqlc code generation, the tools/cli/sda migrate pipeline, and how to evolve schemas without breaking running deployments.
---

# database

Scope: `deploy/migrations/`, `services/**/migrations/`, `sqlc.yaml`, anything called
`query.sql` under `services/**/internal/repository/`.

## One Postgres per container (ADR 022 + 023)

The deployment is a single all-in-one container per tenant. Postgres runs
**inside that container** with its data on a host volume (`/data/<tenant>/postgres`).
The app connects via a **Unix socket** (`/var/run/postgresql`) — lower latency
than TCP, zero network surface, and no password needed from the app.

One Postgres instance, one migration tree, no `tenant_id` columns, no RLS for
tenant boundaries.

- **All tables live in one DB.** No `platform` vs `tenant` split. Schemas group
  by domain (`auth.*`, `rag.*`, `erp.*`) if it helps readability, not for isolation.
- **No `tenant_id` column.** The data in this DB belongs to this deployment's
  tenant. Period. If you catch yourself adding a tenant column, stop — the silo
  contract is violated.
- **RLS is optional, for user-scoped rows.** If you want defense-in-depth against
  a bug that would let user A read user B's rows within the same tenant, RLS keyed
  on `user_id` via `current_setting('app.user_id')` is fine. Not required.

Migrations live in `deploy/migrations/` as a single sequence.

Filenames: `NNNN_slug.up.sql` + `NNNN_slug.down.sql`. `NNNN` is a zero-padded sequence
integer, unique across the whole project.

## The invariant

**Every `.up.sql` has a matching `.down.sql`. Sequence numbers are contiguous.**

If you add `0042_add_chat_tags.up.sql` you also add `0042_add_chat_tags.down.sql`.
The down must genuinely reverse the up — a stub `-- noop` is only acceptable for
additive inserts.

The pre-commit check catches missing pairs. If it fails, don't skip it — fix the pair.

## Writing a migration

1. Look at the highest number in the target tree: `ls deploy/migrations/tenant/ | tail`
2. Pick the next integer.
3. Write `up` and `down` together. Verify they compose (`up` + `down` = nothing changed).
4. Never re-number a merged migration. Append only.
5. Never edit a merged migration. If it was wrong, write a new one that fixes it.

## Schema evolution without downtime

For renames / non-null additions on populated tables:

1. **Add** the new column / table / index — nullable, default-safe.
2. **Backfill** in a script or in code, behind a feature flag.
3. **Switch** reads and writes to the new shape.
4. **Drop** the old column / table.

Four migrations, four deploys. Never combine.

## sqlc

- Config: `sqlc.yaml` at repo root — one config, multiple packages.
- Queries live in `services/<svc>/internal/repository/query.sql`, one directive per query:

  ```sql
  -- name: CreateChat :one
  INSERT INTO chats (id, user_id, ...) VALUES ($1, $2, ...) RETURNING *;

  -- name: ListChats :many
  SELECT * FROM chats ORDER BY created_at DESC;
  ```

- Regenerate with `make sqlc`. Commit the generated code (`*.sql.go`).
- Never hand-edit the generated files.
- No `tenant_id` parameter. Scoping to a tenant is handled by the deployment,
  not by the query (ADR 022).

## Running migrations

```bash
make migrate                          # this deployment's DB
make migrate TENANT=saldivia          # scope to a specific deployment
tools/cli/sda migrate                 # same, direct CLI invocation
```

In production: each tenant deployment runs its own migrations as part of the
`make deploy TENANT=<slug>` pipeline — see `deploy-ops`.

## Indexes and performance

- Every foreign-key column has an index.
- Every `WHERE tenant_id = $1 AND …` query has an index with `tenant_id` as the
  leading column.
- Add indexes in their own migration — never bundled with schema changes.

## Testing migrations

- `go test ./tools/cli/internal/migration/...` exercises every up+down in a
  throwaway Postgres via testcontainers.
- For a new migration, add a case to the migrators test that asserts the
  final schema matches expectations.

## Known anti-patterns in this repo (fix on contact)

### Two migration trees (ADR 022 violation)

The codebase still has `db/platform/migrations/` + `db/tenant/migrations/` (10+
files each). Post-ADR 025 fusion, the bigbrother migrations that used to live
at `services/bigbrother/db/migrations/` were folded into the main tenant tree
(now `db/tenant/migrations/013_bigbrother.up.sql` + `053_bigbrother_permissions.up.sql`).
Under the silo model there is still **one** tree per deployment target:

- Collapse `db/platform/` + `db/tenant/` into `db/migrations/` as a single sequence.
- Drop the `PLATFORM_TENANT_SLUG=platform` hack in compose envs.

Migration pairs are currently well-maintained (no broken pairs detected). **Keep
it that way** — any new `.up.sql` lands with its `.down.sql` in the same commit.

### `tenant_id` columns

Any table with a `tenant_id` column is a relic. Silo = single tenant per DB.
Schedule the column for removal in the next migration that touches the table.

### `pkg/tenant.Resolve` in repository code

Repository code must take a `*pgxpool.Pool` (or `*sql.DB`) from the service's
constructor. **Never resolve a pool per-request by tenant slug.** If you see
that pattern, it belongs to the old pool-tenant era.
