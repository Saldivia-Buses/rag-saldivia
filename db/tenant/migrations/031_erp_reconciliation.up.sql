-- 031_erp_reconciliation.up.sql
-- Plan 18 Fase 1: Reconciliación bancaria
-- Treasury-specific trigger + reconciliation tables + ALTER treasury_movements

-- Treasury-specific trigger: allows reconciliation flag updates on confirmed movements.
-- The generic erp_prevent_financial_mutation() blocks ALL updates on confirmed records.
-- Treasury needs to allow setting reconciled/reconciliation_id without changing core fields.
CREATE OR REPLACE FUNCTION erp_prevent_treasury_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'cannot delete treasury movement with status %', OLD.status;
    END IF;
    -- Allow voiding: confirmed → cancelled/reversed
    IF NEW.status IN ('cancelled', 'reversed') AND OLD.status IN ('confirmed', 'pending') THEN
        RETURN NEW;
    END IF;
    -- Allow reconciliation updates (only reconciled + reconciliation_id may change)
    IF OLD.amount = NEW.amount AND OLD.movement_type = NEW.movement_type
       AND OLD.status = NEW.status AND OLD.date = NEW.date AND OLD.number = NEW.number THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'cannot modify treasury movement with status %', OLD.status;
END;
$$ LANGUAGE plpgsql;

-- Replace generic trigger with treasury-specific one
DROP TRIGGER trg_treasury_immutable ON erp_treasury_movements;
CREATE TRIGGER trg_treasury_immutable BEFORE UPDATE OR DELETE ON erp_treasury_movements
    FOR EACH ROW WHEN (OLD.status IN ('confirmed', 'reversed'))
    EXECUTE FUNCTION erp_prevent_treasury_mutation();

-- Reconciliaciones bancarias
CREATE TABLE erp_bank_reconciliations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    bank_account_id UUID NOT NULL REFERENCES erp_bank_accounts(id),
    period          TEXT NOT NULL,           -- '2026-04'
    statement_balance NUMERIC(16,2) NOT NULL, -- saldo segun extracto
    book_balance    NUMERIC(16,2) NOT NULL,  -- saldo segun tesoreria
    status          TEXT NOT NULL DEFAULT 'draft',  -- 'draft', 'confirmed'
    user_id         TEXT NOT NULL,
    confirmed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, bank_account_id, period)
);

-- Items del extracto bancario (importados)
CREATE TABLE erp_bank_statement_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    reconciliation_id UUID NOT NULL REFERENCES erp_bank_reconciliations(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    description     TEXT NOT NULL,
    amount          NUMERIC(16,2) NOT NULL,  -- positivo=credito, negativo=debito
    reference       TEXT NOT NULL DEFAULT '', -- nro cheque, referencia bancaria
    matched         BOOLEAN NOT NULL DEFAULT false,
    movement_id     UUID REFERENCES erp_treasury_movements(id),  -- match
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_bank_stmt_recon ON erp_bank_statement_lines(tenant_id, reconciliation_id);

-- Flag de conciliacion en treasury movements
ALTER TABLE erp_treasury_movements
    ADD COLUMN reconciled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN reconciliation_id UUID REFERENCES erp_bank_reconciliations(id);

-- RLS
ALTER TABLE erp_bank_reconciliations ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_bank_statement_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_bank_reconciliations
    USING (tenant_id = current_setting('app.tenant_id', true));
CREATE POLICY tenant_isolation ON erp_bank_statement_lines
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Permission
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.treasury.reconcile', 'Reconciliar banco', 'Conciliar extractos bancarios', 'erp')
ON CONFLICT (id) DO NOTHING;

-- Grant to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id = 'erp.treasury.reconcile'
ON CONFLICT DO NOTHING;
