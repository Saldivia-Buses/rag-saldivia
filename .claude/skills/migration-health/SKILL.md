---
name: migration-health
description: Use when touching the Histrix → SDA migration pipeline (tools/cli/internal/migration/*), running or validating a migration, or investigating row-count discrepancies in the migrated tenant DB. Owns Phase 0 of ADR 027 ("data integrity = religion") — enforces the `rows_read == rows_written + rows_skipped` invariant and the "zero `rows_written=0` migrators" invariant.
---

# migration-health

Scope: the migration pipeline (`tools/cli/internal/migration/`,
`tools/cli/cmd/migrate_legacy.go`), the tenant migrations at
`db/tenant/migrations/` that create the `erp_*` schema, and the
live `erp_migration_*` tables on the workstation.

## The two hard invariants (Phase 0, ADR 027)

Any work that touches migration is valid only if these hold after the
session:

1. **Bookkeeping invariant.** For every row in
   `erp_migration_table_progress` (latest run):
   `rows_read = rows_written + rows_skipped`.
   Any row with `rows_read > rows_written + rows_skipped` is a **ghost
   row** — data was read from Histrix, silently dropped somewhere in
   the pipeline, and nothing noticed. Bugs of this kind are **blockers**.

2. **No-op completion invariant.** No row in
   `erp_migration_table_progress` has `rows_read > 0 AND rows_written = 0
   AND status = 'completed'`. That means the transform returns `nil`
   for every row — the migrator exists but silently emits nothing.

Zero ghost rows + zero no-op completions = Phase 0 integrity gate green.

## Known open bugs (as of 2026-04-18)

Run of `2026-04-17 05:28` (prod, completed in 31 min, 23.9M read → 21.98M
written):

| Bug | Evidence | Fix direction |
|---|---|---|
| FACDETAL → erp_invoice_lines: 198K read, 0 written, 198K skipped | Transform rejects every row | Inspect `migrators_phase7_8.go` / `migrators.go` for `NewFACDETALMigrator` (or equivalent) — fix `Transform` |
| REMDETAL → erp_invoice_lines: 5K read, 0 written, 5K skipped | Same pattern | Same migrator family likely |
| FACREMIT → erp_invoices: 199K read, 80K written, **119K unaccounted** | Ghost rows | Instrument writer; likely silent `pgx.Tx` retry/dedup eats rows |
| IVACOMPRAS → erp_invoices: 125K read, 50K written, **75K unaccounted** | Ghost rows | Same pattern |
| STKPIEZA → erp_bom: 7K unaccounted | Ghost rows | Same pattern |
| REG_CUENTA → erp_entities: 5547 read, 464 written, **5083 unaccounted** | Ghost rows | Only 8% entity coverage — likely a filter on a wrong field |
| CTB_CUENTAS → erp_accounts: 3932 unaccounted | Ghost rows | Same |
| CHASIS → erp_units: 3751 unaccounted | Ghost rows | Carrocería-specific, investigate unit model mapping |

## Canonical diagnostic queries

Run these against the tenant DB (on workstation: `docker exec
deploy-postgres-1 psql -U sda -d sda_tenant_saldivia_bench`).

### Ghost rows (bookkeeping invariant)

```sql
SELECT domain, legacy_table, sda_table, rows_read, rows_written, rows_skipped,
       (rows_read - rows_written - rows_skipped) AS unaccounted, status
FROM erp_migration_table_progress
WHERE run_id = (SELECT id FROM erp_migration_runs ORDER BY started_at DESC LIMIT 1)
  AND (rows_read - rows_written - rows_skipped) <> 0
ORDER BY unaccounted DESC;
```

Expected: zero rows. Any row returned is a bug.

### No-op completions (silent skippers)

```sql
SELECT domain, legacy_table, sda_table, rows_read, rows_skipped, status
FROM erp_migration_table_progress
WHERE run_id = (SELECT id FROM erp_migration_runs ORDER BY started_at DESC LIMIT 1)
  AND rows_read > 0 AND rows_written = 0
ORDER BY rows_read DESC;
```

Expected: zero rows.

### Per-domain totals

```sql
SELECT domain,
       COUNT(*) AS migrators,
       SUM(rows_read) AS read,
       SUM(rows_written) AS written,
       SUM(rows_skipped) AS skipped,
       SUM(rows_read - rows_written - rows_skipped) AS ghost
FROM erp_migration_table_progress
WHERE run_id = (SELECT id FROM erp_migration_runs ORDER BY started_at DESC LIMIT 1)
GROUP BY domain ORDER BY domain;
```

### SDA table live counts (sanity check)

```sql
SELECT relname, n_live_tup
FROM pg_stat_user_tables
WHERE schemaname = 'public' AND relname LIKE 'erp_%'
ORDER BY n_live_tup DESC LIMIT 50;
```

## Fixing a broken migrator

Root-cause-first (see `systematic-debugging`):

1. **Reproduce locally.** `make dev` + point the CLI at the dev tenant.
   Run only the failing migrator:
   `sda migrate-legacy --only-table <LEGACY> --dry-run`.
2. **Inspect the transform.** Find `New<Table>Migrator` in
   `tools/cli/internal/migration/migrators*.go`. Look at the
   `Transform` function. The usual suspects:
   - A `return nil` guard that matches every row (wrong field check).
   - A FK lookup that fails and is silently `return nil, nil`.
   - A type conversion that errors and is silently swallowed.
3. **Verify the fix** on a sample: `--only-table X --limit 100 --prod`.
4. **Re-run that table** with `--resume` so the orchestrator picks up
   where it left off.
5. **Assert the invariants** via the canonical queries above.

## Fixing ghost-row bookkeeping

Ghost rows are **not** a transform bug — they mean the writer lost
rows between "I got them" and "I committed them", and neither the
reader nor the progress tracker noticed.

Likely places in `tools/cli/internal/migration/pipeline.go`:

- `sendTransformed` loop — a writer error that cancels context but
  doesn't decrement the in-flight counter.
- `runTablePipeline` accounting — the progress increment happens on
  successful `WriteBatch`; an error path that resets the tx may leave
  rows "read" but not "skipped".
- `ParallelCopyWriter` fan-out — worker dies mid-chunk, chunk is
  retried, old rows are silently dropped.

When you fix it, **add a test** that reproduces ghost rows on a
forced write failure, and the fix produces a `rows_skipped`
increment (or a re-read, but the invariant must hold).

## Rescue hooks

For orphan-FK cases (parent rows missing when child rows arrive),
the pipeline has `rescue.go`:

- `RescueBOMOrphanParents` — creates ghost articles for STK_BOM_HIST
  parents that never existed.
- `RescueCCTIMPUTOrphanMovements` — synthesizes account_movements
  for dangling ledger imputations.
- `RescueLegacyAuthAll` — moves HTXPROFILES into `roles.metadata`.

Rescue hooks run via `--after-table <table>` in the phase order. If a
new migrator produces orphans, write a rescue and wire it, don't
silently drop rows.

## When to block a deploy

- Ghost rows > 0 in latest prod run → block.
- No-op completions > 0 in latest prod run → block.
- New migration changes pipeline logic without a test covering the
  ghost-row / no-op scenarios → block.
- New migrator added without `--dry-run` evidence on full Histrix →
  block.

## Integration with ADR 027

This skill owns the Phase 0 data-integrity items. When a session ships
a migration-health fix:

- Tick the relevant Phase 0 item in `docs/decisions/027-mvp-success-criteria.md`.
- If the fix resolves a Phase 1 data-parity gap, tick that too.
- If the fix requires a schema change, update the `database` skill's
  migration tree and link the migration number in the commit message.

## Don't

- Don't mark a migrator `completed` with `rows_written = 0` as "works
  as intended". That's the failure case.
- Don't fix ghost rows by swallowing the read counter — the read
  counter is correct; the writer is lying.
- Don't skip rescue hooks for orphan FKs by making the FK nullable
  "just because". Nullable FKs in ERP land are data debt.
- Don't trust `status=completed` without running the canonical
  queries — completion is a transform opinion, not a fact.
