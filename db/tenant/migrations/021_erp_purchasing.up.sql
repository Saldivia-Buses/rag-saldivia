-- 021_erp_purchasing.up.sql
-- Plan 17 Phase 5: Purchasing
-- Replaces ~14 legacy tables: CPS*

-- Ordenes de compra
CREATE TABLE IF NOT EXISTS erp_purchase_orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    number      TEXT NOT NULL,
    date        DATE NOT NULL,
    supplier_id UUID NOT NULL REFERENCES erp_entities(id),
    status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved','partial','received','cancelled')),
    currency_id UUID REFERENCES erp_catalogs(id),
    total       NUMERIC(16,2) NOT NULL DEFAULT 0,
    notes       TEXT NOT NULL DEFAULT '',
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, number)
);
CREATE INDEX idx_erp_po_date ON erp_purchase_orders(tenant_id, date DESC);
CREATE INDEX idx_erp_po_supplier ON erp_purchase_orders(tenant_id, supplier_id);

-- Lineas de la OC
CREATE TABLE IF NOT EXISTS erp_purchase_order_lines (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    order_id     UUID NOT NULL REFERENCES erp_purchase_orders(id) ON DELETE CASCADE,
    article_id   UUID NOT NULL REFERENCES erp_articles(id),
    quantity     NUMERIC(14,4) NOT NULL CHECK (quantity > 0),
    unit_price   NUMERIC(14,4) NOT NULL CHECK (unit_price >= 0),
    received_qty NUMERIC(14,4) NOT NULL DEFAULT 0,
    sort_order   INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_erp_po_lines ON erp_purchase_order_lines(tenant_id, order_id);

-- Recepciones
CREATE TABLE IF NOT EXISTS erp_purchase_receipts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    order_id    UUID NOT NULL REFERENCES erp_purchase_orders(id),
    date        DATE NOT NULL,
    number      TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Lineas de recepcion
CREATE TABLE IF NOT EXISTS erp_purchase_receipt_lines (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    receipt_id    UUID NOT NULL REFERENCES erp_purchase_receipts(id) ON DELETE CASCADE,
    order_line_id UUID NOT NULL REFERENCES erp_purchase_order_lines(id),
    article_id    UUID NOT NULL REFERENCES erp_articles(id),
    quantity      NUMERIC(14,4) NOT NULL CHECK (quantity > 0)
);
CREATE INDEX idx_erp_receipt_lines ON erp_purchase_receipt_lines(tenant_id, receipt_id);

-- RLS
ALTER TABLE erp_purchase_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_purchase_orders USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_purchase_order_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_purchase_order_lines USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_purchase_receipts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_purchase_receipts USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_purchase_receipt_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_purchase_receipt_lines USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.purchasing.read',    'Ver compras',      'Consultar ordenes de compra y recepciones', 'erp'),
    ('erp.purchasing.write',   'Gestionar compras', 'Crear ordenes de compra',                  'erp'),
    ('erp.purchasing.approve', 'Aprobar OC',        'Aprobar ordenes de compra',                'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.purchasing.%'
ON CONFLICT DO NOTHING;
