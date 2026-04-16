-- 020_erp_treasury.up.sql
-- Plan 17 Phase 4: Treasury (cash + banks + checks)
-- Replaces ~38 legacy tables: CAJ_*, CAR_*, BCS*, CHEQUERAS

-- Cuentas bancarias
CREATE TABLE IF NOT EXISTS erp_bank_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    bank_name       TEXT NOT NULL,
    branch          TEXT NOT NULL DEFAULT '',
    account_number  TEXT NOT NULL,
    cbu             TEXT,
    alias           TEXT,
    currency_id     UUID REFERENCES erp_catalogs(id),
    account_id      UUID REFERENCES erp_accounts(id),
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, account_number)
);

-- Puestos de caja
CREATE TABLE IF NOT EXISTS erp_cash_registers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    name        TEXT NOT NULL,
    account_id  UUID REFERENCES erp_accounts(id),
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Movimientos de tesoreria (unificado: caja + banco + cheques)
CREATE TABLE IF NOT EXISTS erp_treasury_movements (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    date             DATE NOT NULL,
    number           TEXT NOT NULL,
    movement_type    TEXT NOT NULL CHECK (movement_type IN (
        'cash_in', 'cash_out', 'bank_deposit', 'bank_withdrawal',
        'check_issued', 'check_received', 'transfer')),
    amount           NUMERIC(16,2) NOT NULL CHECK (amount > 0),
    currency_id      UUID REFERENCES erp_catalogs(id),
    bank_account_id  UUID REFERENCES erp_bank_accounts(id),
    cash_register_id UUID REFERENCES erp_cash_registers(id),
    entity_id        UUID REFERENCES erp_entities(id),
    concept_id       UUID REFERENCES erp_catalogs(id),
    payment_method   TEXT,
    reference_type   TEXT,
    reference_id     UUID,
    journal_entry_id UUID REFERENCES erp_journal_entries(id),
    user_id          TEXT NOT NULL,
    notes            TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('pending', 'confirmed', 'reversed')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_treasury_date ON erp_treasury_movements(tenant_id, date DESC);
CREATE INDEX idx_erp_treasury_entity ON erp_treasury_movements(tenant_id, entity_id)
    WHERE entity_id IS NOT NULL;

-- Immutability on confirmed/reversed movements
CREATE TRIGGER trg_treasury_immutable BEFORE UPDATE OR DELETE ON erp_treasury_movements
    FOR EACH ROW WHEN (OLD.status IN ('confirmed', 'reversed'))
    EXECUTE FUNCTION erp_prevent_financial_mutation();

-- Cheques (recibidos o emitidos)
CREATE TABLE IF NOT EXISTS erp_checks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    direction   TEXT NOT NULL CHECK (direction IN ('received', 'issued')),
    number      TEXT NOT NULL,
    bank_name   TEXT NOT NULL,
    amount      NUMERIC(16,2) NOT NULL CHECK (amount > 0),
    issue_date  DATE NOT NULL,
    due_date    DATE NOT NULL,
    entity_id   UUID REFERENCES erp_entities(id),
    status      TEXT NOT NULL DEFAULT 'in_portfolio' CHECK (status IN (
        'in_portfolio', 'deposited', 'cashed', 'rejected', 'endorsed')),
    movement_id UUID REFERENCES erp_treasury_movements(id),
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_checks_due ON erp_checks(tenant_id, due_date)
    WHERE status = 'in_portfolio';

-- Arqueos de caja
CREATE TABLE IF NOT EXISTS erp_cash_counts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    cash_register_id UUID NOT NULL REFERENCES erp_cash_registers(id),
    date             DATE NOT NULL,
    expected         NUMERIC(16,2) NOT NULL,
    counted          NUMERIC(16,2) NOT NULL,
    difference       NUMERIC(16,2) NOT NULL,
    user_id          TEXT NOT NULL,
    notes            TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- RLS
ALTER TABLE erp_bank_accounts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_bank_accounts USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_cash_registers ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_cash_registers USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_treasury_movements ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_treasury_movements USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_checks ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_checks USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_cash_counts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_cash_counts USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.treasury.read',    'Ver tesoreria',      'Consultar movimientos, cheques, saldos', 'erp'),
    ('erp.treasury.write',   'Gestionar tesoreria', 'Registrar movimientos, cheques',        'erp'),
    ('erp.treasury.confirm', 'Confirmar movimientos','Confirmar movimientos pendientes',      'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.treasury.%'
ON CONFLICT DO NOTHING;
