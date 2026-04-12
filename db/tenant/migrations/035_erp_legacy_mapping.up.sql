-- 035_erp_legacy_mapping.up.sql
-- Plan 21: Data migration infrastructure — Histrix MySQL → SDA PostgreSQL
-- Tables: legacy ID mapping, migration run tracking, per-table progress, validation issues

-- Tabla de mapping legacy INT → SDA UUID
CREATE TABLE erp_legacy_mapping (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    domain          TEXT NOT NULL,
    legacy_table    TEXT NOT NULL,
    legacy_id       BIGINT NOT NULL,
    sda_id          UUID NOT NULL,
    legacy_created_by TEXT,
    migrated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, domain, legacy_table, legacy_id)
);
CREATE INDEX idx_erp_legacy_map ON erp_legacy_mapping(tenant_id, domain, legacy_id);
CREATE INDEX idx_erp_legacy_sda ON erp_legacy_mapping(sda_id);

-- Migration run log
CREATE TABLE erp_migration_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'running'
                    CHECK (status IN ('running','completed','failed','dry_run_ok','dry_run_failed')),
    mode            TEXT NOT NULL DEFAULT 'dry_run'
                    CHECK (mode IN ('dry_run','prod')),
    current_domain  TEXT,
    current_table   TEXT,
    error_message   TEXT,
    stats           JSONB DEFAULT '{}'
);
CREATE INDEX idx_erp_mig_runs_tenant ON erp_migration_runs(tenant_id, status);

-- Progress per-table (granular resume)
CREATE TABLE erp_migration_table_progress (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    run_id          UUID NOT NULL REFERENCES erp_migration_runs(id),
    domain          TEXT NOT NULL,
    legacy_table    TEXT NOT NULL,
    sda_table       TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','in_progress','completed','failed','skipped')),
    last_legacy_key TEXT NOT NULL DEFAULT '',
    rows_read       INT NOT NULL DEFAULT 0,
    rows_written    INT NOT NULL DEFAULT 0,
    rows_skipped    INT NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    error_message   TEXT,
    UNIQUE(tenant_id, run_id, legacy_table)
);
CREATE INDEX idx_erp_mig_progress_run ON erp_migration_table_progress(run_id, status);

-- Pre-validation findings
CREATE TABLE erp_migration_validation_issues (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    run_id          UUID NOT NULL REFERENCES erp_migration_runs(id),
    domain          TEXT NOT NULL,
    legacy_table    TEXT NOT NULL,
    legacy_id       BIGINT NOT NULL,
    constraint_name TEXT NOT NULL,
    details         JSONB NOT NULL,
    resolution      TEXT NOT NULL DEFAULT 'pending'
                    CHECK (resolution IN ('pending','fix_manual','skip','auto_fixed')),
    resolved_by     TEXT,
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_mig_issues_run ON erp_migration_validation_issues(run_id, resolution);

-- NO RLS on migration infrastructure tables:
-- These tables are used exclusively by the CLI `sda migrate-legacy`, a binary
-- operated by sysadmins. The CLI validates tenant in app layer via WHERE tenant_id=$1.
-- Business tables (erp_invoices etc.) keep their RLS.
