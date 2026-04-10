-- 015_erp_infra.up.sql
-- Plan 17 Phase 0: ERP infrastructure — audit_log fix, financial triggers, sequences

-- Fix: audit_log table is missing tenant_id column (code already writes it)
ALTER TABLE audit_log ADD COLUMN IF NOT EXISTS tenant_id TEXT;
CREATE INDEX IF NOT EXISTS idx_audit_log_tenant ON audit_log(tenant_id) WHERE tenant_id IS NOT NULL;

-- Immutability trigger for financial records (pattern P3)
-- Prevents UPDATE/DELETE on records with status in ('posted', 'confirmed', 'paid', 'reversed')
CREATE OR REPLACE FUNCTION erp_prevent_financial_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.status IN ('posted', 'confirmed', 'paid', 'reversed') THEN
            RAISE EXCEPTION 'cannot delete financial record with status %', OLD.status;
        END IF;
        RETURN OLD;
    END IF;
    -- UPDATE
    IF OLD.status IN ('posted', 'confirmed', 'paid', 'reversed') THEN
        RAISE EXCEPTION 'cannot modify financial record with status %', OLD.status;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Atomic sequence generator (pattern P10)
-- Usage: SELECT next_erp_sequence('tenant-slug', 'invoice', '0001');
CREATE OR REPLACE FUNCTION next_erp_sequence(
    p_tenant TEXT,
    p_domain TEXT,
    p_prefix TEXT DEFAULT ''
) RETURNS BIGINT AS $$
    UPDATE erp_sequences
    SET next_value = next_value + 1
    WHERE tenant_id = p_tenant AND domain = p_domain AND prefix = p_prefix
    RETURNING next_value - 1;
$$ LANGUAGE sql;

-- Journal entry balance validation (pattern P13)
-- Validates sum(debit) = sum(credit) when entry is posted
CREATE OR REPLACE FUNCTION erp_validate_journal_balance() RETURNS trigger AS $$
DECLARE
    total_debit NUMERIC;
    total_credit NUMERIC;
BEGIN
    SELECT COALESCE(SUM(debit), 0), COALESCE(SUM(credit), 0)
    INTO total_debit, total_credit
    FROM erp_journal_lines WHERE entry_id = NEW.id AND tenant_id = NEW.tenant_id;

    IF total_debit != total_credit THEN
        RAISE EXCEPTION 'journal entry % is unbalanced: debit=% credit=%',
            NEW.number, total_debit, total_credit;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- RLS on existing ERP tables (suggestions)
ALTER TABLE erp_suggestions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_suggestions
    USING (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE erp_suggestion_responses ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_suggestion_responses
    USING (tenant_id = current_setting('app.tenant_id', true));
