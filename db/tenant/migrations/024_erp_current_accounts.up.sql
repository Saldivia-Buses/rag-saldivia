-- 024_erp_current_accounts.up.sql
-- Plan 17 Phase 7b: Current Accounts (accounts receivable/payable)
-- Replaces ~12 legacy tables: REG_MOVIMIENTOS, REG_IMPUTACIONES, etc.

-- Movimientos de cuenta corriente
CREATE TABLE IF NOT EXISTS erp_account_movements (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    entity_id        UUID NOT NULL REFERENCES erp_entities(id),
    date             DATE NOT NULL,
    movement_type    TEXT NOT NULL CHECK (movement_type IN ('invoice','payment','credit_note','debit_note','adjustment')),
    direction        TEXT NOT NULL CHECK (direction IN ('receivable', 'payable')),
    amount           NUMERIC(16,2) NOT NULL CHECK (amount > 0),
    balance          NUMERIC(16,2) NOT NULL,
    invoice_id       UUID REFERENCES erp_invoices(id),
    treasury_id      UUID REFERENCES erp_treasury_movements(id),
    journal_entry_id UUID REFERENCES erp_journal_entries(id),
    notes            TEXT NOT NULL DEFAULT '',
    user_id          TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_acct_mov_entity ON erp_account_movements(tenant_id, entity_id, date DESC);
CREATE INDEX idx_erp_acct_mov_balance ON erp_account_movements(tenant_id, entity_id) WHERE balance > 0;

-- Immutability (financial record)
CREATE TRIGGER trg_acct_mov_immutable BEFORE UPDATE OR DELETE ON erp_account_movements
    FOR EACH ROW EXECUTE FUNCTION erp_prevent_financial_mutation();

-- Imputaciones: que pago aplica a que factura
CREATE TABLE IF NOT EXISTS erp_payment_allocations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    payment_id  UUID NOT NULL REFERENCES erp_account_movements(id),
    invoice_id  UUID NOT NULL REFERENCES erp_account_movements(id),
    amount      NUMERIC(16,2) NOT NULL CHECK (amount > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, payment_id, invoice_id)
);

-- RLS
ALTER TABLE erp_account_movements ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_account_movements USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_payment_allocations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_payment_allocations USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.accounts.read',  'Ver cuentas corrientes',      'Consultar saldos y movimientos', 'erp'),
    ('erp.accounts.write', 'Gestionar cuentas corrientes', 'Imputar pagos',                 'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.accounts.%'
ON CONFLICT DO NOTHING;
