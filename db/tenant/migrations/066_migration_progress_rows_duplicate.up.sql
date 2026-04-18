-- 066_migration_progress_rows_duplicate.up.sql
-- Phase 0 migration-integrity: restore the bookkeeping invariant.
--
-- Before this migration, erp_migration_table_progress tracked (rows_read,
-- rows_written, rows_skipped). The writer's `rows_written` counter uses
-- tag.RowsAffected() after INSERT ... ON CONFLICT DO NOTHING, which only
-- counts newly inserted rows — conflict-dedup'd rows are invisible. The
-- prod saldivia run on 2026-04-17 had 214K such "ghost rows" across 13
-- migrators, violating `rows_read = rows_written + rows_skipped`.
--
-- This column tracks conflict-dedup'd rows as a first-class category so
-- the invariant becomes `rows_read = rows_written + rows_skipped + rows_duplicate`.

ALTER TABLE erp_migration_table_progress
    ADD COLUMN rows_duplicate INT NOT NULL DEFAULT 0;
