-- 034_erp_receipts.up.sql
-- Plan 18 Fase 4: Recibos de cobro y pago
-- Receipt workflow: treasury + accounting + current accounts in one TX

-- Recibos (cobro o pago)
CREATE TABLE erp_receipts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    number          TEXT NOT NULL,
    date            DATE NOT NULL,
    receipt_type    TEXT NOT NULL,           -- 'collection' (cobro), 'payment' (pago)
    entity_id       UUID NOT NULL REFERENCES erp_entities(id),
    total           NUMERIC(16,2) NOT NULL CHECK (total > 0),
    journal_entry_id UUID REFERENCES erp_journal_entries(id),
    user_id         TEXT NOT NULL,
    notes           TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'confirmed',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, receipt_type, number)
);
CREATE INDEX idx_erp_receipts_date ON erp_receipts(tenant_id, date DESC);

-- Medios de pago del recibo
CREATE TABLE erp_receipt_payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    receipt_id      UUID NOT NULL REFERENCES erp_receipts(id) ON DELETE RESTRICT,
    payment_method  TEXT NOT NULL,           -- 'cash', 'check', 'transfer', 'echeq'
    amount          NUMERIC(16,2) NOT NULL CHECK (amount > 0),
    treasury_movement_id UUID REFERENCES erp_treasury_movements(id),
    check_id        UUID REFERENCES erp_checks(id),
    bank_account_id UUID REFERENCES erp_bank_accounts(id),
    notes           TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_erp_receipt_payments ON erp_receipt_payments(tenant_id, receipt_id);

-- Facturas imputadas en el recibo
CREATE TABLE erp_receipt_allocations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    receipt_id      UUID NOT NULL REFERENCES erp_receipts(id) ON DELETE RESTRICT,
    invoice_id      UUID NOT NULL REFERENCES erp_invoices(id),
    amount          NUMERIC(16,2) NOT NULL CHECK (amount > 0),
    account_movement_id UUID REFERENCES erp_account_movements(id)
);
CREATE INDEX idx_erp_receipt_alloc ON erp_receipt_allocations(tenant_id, receipt_id);

-- Retenciones en el recibo
CREATE TABLE erp_receipt_withholdings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    receipt_id      UUID NOT NULL REFERENCES erp_receipts(id) ON DELETE RESTRICT,
    withholding_id  UUID NOT NULL REFERENCES erp_withholdings(id)
);

-- Inmutabilidad: confirmed receipts can only transition to cancelled
CREATE TRIGGER trg_receipt_immutable BEFORE UPDATE OR DELETE ON erp_receipts
    FOR EACH ROW WHEN (OLD.status = 'confirmed')
    EXECUTE FUNCTION erp_prevent_financial_mutation();

-- RLS
ALTER TABLE erp_receipts ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_receipt_payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_receipt_allocations ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_receipt_withholdings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_receipts
    USING (tenant_id = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON erp_receipt_payments
    USING (tenant_id = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON erp_receipt_allocations
    USING (tenant_id = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON erp_receipt_withholdings
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Permission
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.treasury.receipt', 'Registrar recibos', 'Cobros y pagos con imputacion', 'erp')
ON CONFLICT (id) DO NOTHING;

-- Grant to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id = 'erp.treasury.receipt'
ON CONFLICT DO NOTHING;
