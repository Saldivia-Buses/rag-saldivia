# ADR 029 — Histrix MySQL as the ERP backing database

**Status:** Accepted (2026-04-20)
**Supersedes:** parts of ADR 026 about the migrator
**Related:** ADR 022 (silo), ADR 026 (SDA replaces Histrix), ADR 028
(eliminate intra-silo tenancy), ADR 027 (MVP success criteria)

## Context

Until today, the SDA architecture was:

```
Histrix MySQL (legacy server) → migrator (Go pipeline) → SDA Postgres
                                                           ↓
                                                       SDA app reads
```

Two databases, with a continuous translation layer (the migrator)
keeping them in sync. The migrator was supposed to:

- Map every Histrix table to an `erp_*` Postgres table.
- Translate Histrix's idiosyncratic schemas (XML-form-driven, decades
  of accretion) into a clean, sqlc-friendly Postgres schema.
- Handle data integrity (Phase 0 of ADR 027): zero ghost rows, every
  read accounted for.

Reality after months of work:

- Migrator has known bugs (15 FK orphans found 2026-04-20 in the
  bench restore — see `docs/parity/2026-04-20-phase0-fk-orphans-bench.md`).
- Schema drift: the SDA Postgres schema is a *guess* of what Histrix
  has; XML-forms are the source of truth, but the Postgres mapping
  is hand-written and lags.
- Every "new ERP feature" requires three steps: read the XML-form,
  design a Postgres table, write the migrator path. Most of the
  effort is in step 2 — recreating data Histrix already has.
- The `.intranet-scrape/` (676 tables, ~4500 XML-forms across 99 area
  groups) is the parity contract. The Postgres rewrite is a lossy
  reinterpretation.

The 2026-04-20 prod cutover made the cost obvious: data was migrated,
then the cutover surfaced 7 distinct bugs (see ADR 028) on top of the
data-integrity bugs (ghost rows). The migrator is a permanent source
of bugs that wouldn't exist without it.

## Decision

**Use the Histrix MySQL database directly as the ERP backing store.**
The MySQL instance runs **inside the SDA container** (docker-compose
service alongside `app`, `erp`, `postgres`). The data stays in its
native schema; SDA reads and writes via standard MySQL queries.

```
                ┌── SDA container (silo) ────────────────────┐
                │                                            │
                │   web ── traefik ── app ── postgres        │
                │                       │       (SDA-native: │
                │                       │        platform,   │
                │                       │        chat,       │
                │                       │        collections,│
                │                       │        suggestions)│
                │                       │                    │
                │                       └── erp ── mysql     │
                │                                  (Histrix  │
                │                                   schema   │
                │                                   1:1)     │
                │                                            │
                └────────────────────────────────────────────┘
```

The Histrix legacy server (`172.22.100.99`) eventually powers off
once the dump has been imported into the in-container MySQL.
ADR 026's "Histrix powers off" is preserved — it's the *server* that
powers off, not the *data*.

## Consequences

### Positive

- **Migrator deleted.** `tools/cli/internal/migration/*` and the
  Go service that runs it are removed entirely. ~thousands of LOC
  gone. Phase 0 of ADR 027 (data integrity) becomes trivially
  satisfied — there is no translation layer to introduce ghosts.
- **Schema = Histrix.** No more "what does this table mean?" guessing.
  `.intranet-scrape/xml-forms/<area>/*.xml` is the contract; SQL
  hits the same columns Histrix uses.
- **Parity is structural.** When SDA's UI for an area is built from
  the XML-form, it reads/writes the *exact* fields Histrix used.
  Round-trip Histrix → SDA → Histrix is a no-op.
- **One source of truth.** No drift between SDA Postgres and Histrix
  MySQL because there is no SDA Postgres for ERP data anymore.
- **Cutover from current state**: dump MySQL once from the legacy
  server, restore into the in-container MySQL, point ERP service at
  it. Done. No ongoing sync.
- **Histrix server powers off.** The legacy hardware is freed; the
  data lives in the silo container.

### Negative

- **Go database driver changes.** ERP service swaps `pgx` for
  `go-sql-driver/mysql`. sqlc gets a `mysql` engine config.
  Some SQL syntax differences (LIMIT/OFFSET vs FETCH, RETURNING
  unsupported, etc.). Estimated ~1 week of mechanical work.
- **Postgres remains in the container** for the SDA-native domain
  (platform metadata, chat, collections, suggestions, audit log).
  ERP queries hit MySQL; SDA-native queries hit Postgres. Two
  drivers live in the binary. Acceptable.
- **Backups change.** `mysqldump` per-silo, plus `pg_dump` for the
  Postgres slice. The deploy runbook gets a section.
- **Existing Postgres ERP tables become orphans.** They get dropped
  in a single migration once the cutover is done. No data is lost
  because nothing was ever written there in production prior to
  this decision (the bench was a test target).
- **MySQL inside a per-silo container** uses ~200-500 MB resident,
  manageable on the workstation specs.

### Interaction with ADR 028

ADR 028 (eliminate intra-silo tenancy) is a **prerequisite**. With
Histrix MySQL as backing, the data has Histrix's columns, none of
which are `tenant_id`. So Postgres-side queries that previously
filtered by `tenant_id` simply have no equivalent on the MySQL
side — confirming that the tenancy boundary belongs to the
container, not the schema.

### Interaction with ADR 026

ADR 026 said "SDA replaces Histrix; the Histrix server eventually
powers off." This ADR amends only the **mechanism**: the Histrix
server still powers off, but its data lives on as the in-container
MySQL. The user-visible promise (one product, modern UI, AI agent,
etc.) is unchanged.

The migrator section of ADR 026 ("zero-loss migration verifiable
end-to-end") becomes "one-shot mysqldump | mysql restore", which
is trivially verifiable.

### Phased rollout

1. **Phase A — infra**: add `mysql:8` service to
   `deploy/docker-compose.dev.yml`. Restore the Histrix dump into
   it. Verify ERP service can connect.
2. **Phase B — first read path**: rewrite one ERP query (entities)
   to hit MySQL. Frontend page works. Validate.
3. **Phase C — bulk read paths**: PR per area (compras, ventas,
   …). sqlc with mysql engine.
4. **Phase D — write paths**: invoicing, treasury, etc. hit MySQL.
5. **Phase E — drop SDA Postgres ERP tables**: after all paths
   migrated. Single migration. The Postgres instance survives for
   platform/chat/collections/suggestions.
6. **Phase F — delete the migrator**: `tools/cli/internal/migration/*`
   and related code is removed. Power off the Histrix server.

Each phase is independently shippable. Phase A is hours; Phases
B-D are weeks of mechanical refactor.

### Priority

**Maximum, after ADR 028**. ADR 028 is the prerequisite (without
it, the MySQL queries would have to deal with the same tenant_id
nonsense, which doesn't exist on the MySQL side). Then this ADR
unblocks the long-term plan.

## Out of scope

- **Schema reform inside Histrix MySQL.** We accept the legacy
  schema as-is. Indexes can be added; column renames are forbidden
  (they break the parity contract).
- **MySQL replication, HA, etc.** Single-instance per silo,
  matching the silo deployment model.
- **Other RDBMS engines.** MySQL is what Histrix uses; we follow.
