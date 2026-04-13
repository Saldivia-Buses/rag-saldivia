-- 045_erp_cnrt_work_orders.up.sql
-- Órdenes de trabajo CNRT: habilitación y verificación técnica obligatoria (Argentina)

CREATE TABLE erp_cnrt_work_orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           TEXT NOT NULL,
    unit_id             UUID NOT NULL REFERENCES erp_manufacturing_units(id),
    cnrt_number         TEXT NOT NULL DEFAULT '',
    inspection_type     TEXT NOT NULL DEFAULT 'initial'
                        CHECK (inspection_type IN ('initial','periodic','extraordinary','modification')),
    inspector_name      TEXT NOT NULL DEFAULT '',
    inspection_date     DATE,
    approved            BOOLEAN,
    approval_date       DATE,
    expiry_date         DATE,
    observations        TEXT NOT NULL DEFAULT '',
    rejection_reasons   TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','scheduled','inspected','approved','rejected','expired')),
    document_url        TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_cnrt_unit   ON erp_cnrt_work_orders (tenant_id, unit_id);
CREATE INDEX idx_cnrt_status ON erp_cnrt_work_orders (tenant_id, status);
CREATE INDEX idx_cnrt_expiry ON erp_cnrt_work_orders (expiry_date);
ALTER TABLE erp_cnrt_work_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_cnrt_work_orders
    USING (tenant_id = current_setting('app.tenant_id', true));
