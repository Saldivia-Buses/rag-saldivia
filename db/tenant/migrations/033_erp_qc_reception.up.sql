-- 033_erp_qc_reception.up.sql
-- Plan 18 Fase 3: Recepción con control de calidad
-- QC inspections (accept/reject per receipt line) + supplier demerits

CREATE TABLE erp_qc_inspections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    receipt_id      UUID NOT NULL REFERENCES erp_purchase_receipts(id),
    receipt_line_id UUID NOT NULL REFERENCES erp_purchase_receipt_lines(id),
    article_id      UUID NOT NULL REFERENCES erp_articles(id),
    quantity        NUMERIC(14,4) NOT NULL,
    accepted_qty    NUMERIC(14,4) NOT NULL DEFAULT 0,
    rejected_qty    NUMERIC(14,4) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed')),
    inspector_id    TEXT NOT NULL,
    notes           TEXT NOT NULL DEFAULT '',
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (accepted_qty + rejected_qty <= quantity),
    CHECK (accepted_qty >= 0 AND rejected_qty >= 0)
);
CREATE INDEX idx_erp_qc_receipt ON erp_qc_inspections(tenant_id, receipt_id);

CREATE TABLE erp_supplier_demerits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    supplier_id     UUID NOT NULL REFERENCES erp_entities(id),
    inspection_id   UUID NOT NULL REFERENCES erp_qc_inspections(id),
    points          INT NOT NULL,
    reason          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_demerits_supplier ON erp_supplier_demerits(tenant_id, supplier_id);

-- RLS
ALTER TABLE erp_qc_inspections ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_supplier_demerits ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_qc_inspections
    USING (tenant_id = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON erp_supplier_demerits
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Permission
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.purchasing.inspect', 'Inspeccionar recepciones', 'Control de calidad en recepcion', 'erp')
ON CONFLICT (id) DO NOTHING;

-- Grant to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id = 'erp.purchasing.inspect'
ON CONFLICT DO NOTHING;
