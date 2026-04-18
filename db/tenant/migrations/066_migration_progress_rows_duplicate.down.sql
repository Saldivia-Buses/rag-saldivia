-- 066_migration_progress_rows_duplicate.down.sql
-- Reverses 066 — drops the ghost-row accounting column.

ALTER TABLE erp_migration_table_progress
    DROP COLUMN rows_duplicate;
