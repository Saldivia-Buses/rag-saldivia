-- 019_erp_accounting.up.sql
-- Plan 17 Phase 3: Accounting
-- Replaces ~31 legacy tables: CTB_*, CCT*, OBL*

-- Centros de costos (created before accounts, which references it)
CREATE TABLE IF NOT EXISTS erp_cost_centers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    parent_id   UUID REFERENCES erp_cost_centers(id),
    active      BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_erp_cost_centers ON erp_cost_centers(tenant_id, active);

-- Plan de cuentas (arbol jerarquico)
CREATE TABLE IF NOT EXISTS erp_accounts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      TEXT NOT NULL,
    code           TEXT NOT NULL,
    name           TEXT NOT NULL,
    parent_id      UUID REFERENCES erp_accounts(id),
    account_type   TEXT NOT NULL CHECK (account_type IN ('asset', 'liability', 'equity', 'income', 'expense')),
    is_detail      BOOLEAN NOT NULL DEFAULT true,
    cost_center_id UUID REFERENCES erp_cost_centers(id),
    active         BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_erp_accounts ON erp_accounts(tenant_id, active);

-- Ejercicios contables
CREATE TABLE IF NOT EXISTS erp_fiscal_years (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    year        INT NOT NULL,
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    UNIQUE(tenant_id, year)
);

-- Asientos contables (cabecera)
CREATE TABLE IF NOT EXISTS erp_journal_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    number          TEXT NOT NULL,
    date            DATE NOT NULL,
    fiscal_year_id  UUID REFERENCES erp_fiscal_years(id),
    concept         TEXT NOT NULL,
    entry_type      TEXT NOT NULL DEFAULT 'manual' CHECK (entry_type IN ('manual', 'auto', 'adjustment')),
    reference_type  TEXT,
    reference_id    UUID,
    user_id         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'posted', 'reversed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, number)
);
CREATE INDEX idx_erp_journal_date ON erp_journal_entries(tenant_id, date DESC);

-- Detalle del asiento (debe/haber)
CREATE TABLE IF NOT EXISTS erp_journal_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    entry_id        UUID NOT NULL REFERENCES erp_journal_entries(id) ON DELETE RESTRICT,
    account_id      UUID NOT NULL REFERENCES erp_accounts(id),
    cost_center_id  UUID REFERENCES erp_cost_centers(id),
    entry_date      DATE NOT NULL,
    debit           NUMERIC(16,2) NOT NULL DEFAULT 0,
    credit          NUMERIC(16,2) NOT NULL DEFAULT 0,
    description     TEXT NOT NULL DEFAULT '',
    sort_order      INT NOT NULL DEFAULT 0,
    CHECK (debit >= 0 AND credit >= 0),
    CHECK (NOT (debit > 0 AND credit > 0))
);
CREATE INDEX idx_erp_journal_lines_entry ON erp_journal_lines(tenant_id, entry_id);
CREATE INDEX idx_erp_journal_lines_account ON erp_journal_lines(tenant_id, account_id, entry_date DESC);

-- Immutability trigger (pattern P3) — prevents modifying posted/reversed entries
CREATE TRIGGER trg_journal_immutable BEFORE UPDATE OR DELETE ON erp_journal_entries
    FOR EACH ROW WHEN (OLD.status IN ('posted', 'reversed'))
    EXECUTE FUNCTION erp_prevent_financial_mutation();

-- Balance validation (pattern P13) — validates when entry is posted
CREATE CONSTRAINT TRIGGER trg_journal_balance
    AFTER UPDATE ON erp_journal_entries
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW
    WHEN (NEW.status = 'posted' AND OLD.status != 'posted')
    EXECUTE FUNCTION erp_validate_journal_balance();

-- RLS
ALTER TABLE erp_cost_centers ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_cost_centers USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_accounts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_accounts USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_fiscal_years ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_fiscal_years USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_journal_entries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_journal_entries USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_journal_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_journal_lines USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.accounting.read',    'Ver contabilidad',     'Consultar asientos y plan de cuentas', 'erp'),
    ('erp.accounting.write',   'Gestionar contabilidad','Crear asientos',                      'erp'),
    ('erp.accounting.reverse', 'Reversar asientos',    'Crear contra-asientos',                'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.accounting.%'
ON CONFLICT DO NOTHING;
