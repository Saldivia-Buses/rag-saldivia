-- Migration 062 — universal legacy archive
--
-- Goal: every row in the Histrix MySQL that we chose not to (or cannot yet)
-- model explicitly in SDA still has a place to land. No silent data loss,
-- ever. Operators can query `erp_legacy_archive` to answer any "where did
-- the old row for X go?" question, and the data is available for future
-- promotion into a first-class SDA table without re-reading MySQL.
--
-- Schema choice: wide JSONB payload. We trade schema enforcement for
-- universality — legacy has 620 table shapes, many of them layout-only, and
-- it is not worth hand-modelling the 478 we decided to skip for now.
--
-- Indexing: (tenant, legacy_table) — the dominant query pattern is "show me
-- all archived FOO rows for tenant X". GIN on data is opt-in later because
-- populating the archive already writes a lot of bytes.

CREATE TABLE IF NOT EXISTS erp_legacy_archive (
    id            UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id     TEXT        NOT NULL,
    legacy_table  TEXT        NOT NULL,
    legacy_pk     TEXT,                          -- composite PKs collapsed to a canonical string
    legacy_pk_num BIGINT,                        -- for single-BIGINT PKs, duplicated for fast range queries
    data          JSONB       NOT NULL,
    migrated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_erp_legacy_archive_tenant_table
    ON erp_legacy_archive (tenant_id, legacy_table);

-- legacy_pk_num exists only on tables where the PK fits in a BIGINT. Filtered
-- index keeps the non-numeric rows out of the B-tree so size stays sane.
CREATE INDEX IF NOT EXISTS idx_erp_legacy_archive_pk_num
    ON erp_legacy_archive (tenant_id, legacy_table, legacy_pk_num)
    WHERE legacy_pk_num IS NOT NULL;

-- Unique per (tenant, table, legacy_pk) so re-runs are idempotent.
CREATE UNIQUE INDEX IF NOT EXISTS uq_erp_legacy_archive_natural
    ON erp_legacy_archive (tenant_id, legacy_table, legacy_pk);

-- RLS: same tenant isolation policy as every other tenant table so the
-- archive cannot be read across tenants.
ALTER TABLE erp_legacy_archive ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_legacy_archive
    USING (tenant_id = current_setting('app.tenant_id', true));
