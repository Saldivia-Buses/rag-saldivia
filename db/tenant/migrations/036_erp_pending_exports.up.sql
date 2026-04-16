-- 036_erp_pending_exports.up.sql
-- Plan 20: BI & Analytics — async export queue for large datasets

CREATE TABLE erp_pending_exports (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    user_id      TEXT NOT NULL,
    export_name  TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','running','ready','failed')),
    row_count    INT NOT NULL DEFAULT 0,
    format       TEXT NOT NULL DEFAULT 'excel'
                 CHECK (format IN ('excel','csv')),
    params       JSONB NOT NULL DEFAULT '{}',
    columns_def  JSONB NOT NULL DEFAULT '[]',
    file_key     TEXT,
    error        TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at   TIMESTAMPTZ,
    ready_at     TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours')
);
CREATE INDEX idx_erp_pending_exports_user ON erp_pending_exports(tenant_id, user_id, status);

ALTER TABLE erp_pending_exports ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_pending_exports
    USING (tenant_id = current_setting('app.tenant_id', true));
