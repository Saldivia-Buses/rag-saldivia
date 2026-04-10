-- 023_erp_invoicing.up.sql
-- Plan 17 Phase 7: Invoicing & Tax (IVA)
-- Replaces ~34 legacy tables: FAC*, REM*, IVA*, RET*

-- Comprobantes (facturas, NC, ND, remitos)
CREATE TABLE IF NOT EXISTS erp_invoices (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    number           TEXT NOT NULL,
    date             DATE NOT NULL,
    due_date         DATE,
    invoice_type     TEXT NOT NULL CHECK (invoice_type IN (
        'invoice_a','invoice_b','invoice_c','invoice_e',
        'credit_note','debit_note','delivery_note')),
    direction        TEXT NOT NULL CHECK (direction IN ('issued', 'received')),
    entity_id        UUID NOT NULL REFERENCES erp_entities(id),
    currency_id      UUID REFERENCES erp_catalogs(id),
    subtotal         NUMERIC(16,2) NOT NULL DEFAULT 0,
    tax_amount       NUMERIC(16,2) NOT NULL DEFAULT 0,
    total            NUMERIC(16,2) NOT NULL DEFAULT 0,
    order_id         UUID REFERENCES erp_orders(id),
    journal_entry_id UUID REFERENCES erp_journal_entries(id),
    afip_cae         TEXT,
    afip_cae_due     DATE,
    status           TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','posted','paid','cancelled')),
    user_id          TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, invoice_type, number)
);
CREATE INDEX idx_erp_invoices_date ON erp_invoices(tenant_id, date DESC);
CREATE INDEX idx_erp_invoices_entity ON erp_invoices(tenant_id, entity_id);

-- Immutability on posted/paid invoices
CREATE TRIGGER trg_invoice_immutable BEFORE UPDATE OR DELETE ON erp_invoices
    FOR EACH ROW WHEN (OLD.status IN ('posted', 'paid'))
    EXECUTE FUNCTION erp_prevent_financial_mutation();

-- Immutability for invoice lines — prevents modifying lines of posted invoices
CREATE OR REPLACE FUNCTION erp_prevent_invoice_line_mutation() RETURNS trigger AS $$
DECLARE parent_status TEXT;
BEGIN
    SELECT status INTO parent_status FROM erp_invoices
    WHERE id = COALESCE(OLD.invoice_id, NEW.invoice_id);
    IF parent_status IN ('posted', 'paid') THEN
        RAISE EXCEPTION 'cannot modify lines of % invoice', parent_status;
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Detalle del comprobante
CREATE TABLE IF NOT EXISTS erp_invoice_lines (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    invoice_id  UUID NOT NULL REFERENCES erp_invoices(id) ON DELETE RESTRICT,
    article_id  UUID REFERENCES erp_articles(id),
    description TEXT NOT NULL,
    quantity    NUMERIC(14,4) NOT NULL CHECK (quantity > 0),
    unit_price  NUMERIC(14,4) NOT NULL CHECK (unit_price >= 0),
    tax_rate    NUMERIC(5,2) NOT NULL DEFAULT 21.00,
    tax_amount  NUMERIC(16,2) NOT NULL DEFAULT 0,
    line_total  NUMERIC(16,2) NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_erp_invoice_lines ON erp_invoice_lines(tenant_id, invoice_id);

CREATE TRIGGER trg_invoice_lines_immutable BEFORE UPDATE OR DELETE ON erp_invoice_lines
    FOR EACH ROW EXECUTE FUNCTION erp_prevent_invoice_line_mutation();

-- Libro IVA (generado al contabilizar)
CREATE TABLE IF NOT EXISTS erp_tax_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    invoice_id  UUID NOT NULL REFERENCES erp_invoices(id) ON DELETE RESTRICT,
    period      TEXT NOT NULL,
    direction   TEXT NOT NULL CHECK (direction IN ('purchases', 'sales')),
    net_amount  NUMERIC(16,2) NOT NULL,
    tax_rate    NUMERIC(5,2) NOT NULL,
    tax_amount  NUMERIC(16,2) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_tax_period ON erp_tax_entries(tenant_id, period, direction);

-- Retenciones
CREATE TABLE IF NOT EXISTS erp_withholdings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    invoice_id      UUID REFERENCES erp_invoices(id) ON DELETE RESTRICT,
    movement_id     UUID REFERENCES erp_treasury_movements(id) ON DELETE RESTRICT,
    entity_id       UUID NOT NULL REFERENCES erp_entities(id),
    type            TEXT NOT NULL CHECK (type IN ('iibb', 'gains', 'iva', 'suss')),
    rate            NUMERIC(5,2) NOT NULL,
    base_amount     NUMERIC(16,2) NOT NULL,
    amount          NUMERIC(16,2) NOT NULL CHECK (amount > 0),
    certificate_num TEXT,
    date            DATE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_withholdings ON erp_withholdings(tenant_id, entity_id, date DESC);

-- RLS
ALTER TABLE erp_invoices ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_invoices USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_invoice_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_invoice_lines USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_tax_entries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_tax_entries USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_withholdings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_withholdings USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.invoicing.read',  'Ver facturacion',      'Consultar comprobantes y libro IVA', 'erp'),
    ('erp.invoicing.write', 'Gestionar facturacion', 'Crear comprobantes',                'erp'),
    ('erp.invoicing.post',  'Contabilizar facturas', 'Postear comprobantes',              'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.invoicing.%'
ON CONFLICT DO NOTHING;
