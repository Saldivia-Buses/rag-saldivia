-- 015_erp_infra.up.sql
-- Plan 17 Phase 0: ERP infrastructure — audit_log fix, financial triggers, sequences

-- next_erp_sequence below references erp_sequences, which is created in
-- 016_erp_catalogs. LANGUAGE sql parses the body at CREATE time (unlike
-- LANGUAGE plpgsql which defers to call time), so this migration cannot
-- validate the body on a fresh database. Defer validation — the table
-- will exist by the time any caller runs the function. Scope is LOCAL
-- so the session restores the default after commit.
SET LOCAL check_function_bodies = off;

-- Fix: audit_log.tenant_id was added as UUID in migration 013 but pkg/audit
-- writes TEXT slugs (e.g., "saldivia", "dev"). Change to TEXT so audit writes work.
ALTER TABLE audit_log ALTER COLUMN tenant_id TYPE TEXT USING tenant_id::TEXT;
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
-- Auto-creates the sequence row if it doesn't exist (upsert).
-- Usage: SELECT next_erp_sequence('tenant-slug', 'invoice', '0001');
CREATE OR REPLACE FUNCTION next_erp_sequence(
    p_tenant TEXT,
    p_domain TEXT,
    p_prefix TEXT DEFAULT ''
) RETURNS BIGINT AS $$
    INSERT INTO erp_sequences (tenant_id, domain, prefix, next_value)
    VALUES (p_tenant, p_domain, p_prefix, 2)
    ON CONFLICT (tenant_id, domain, prefix)
    DO UPDATE SET next_value = erp_sequences.next_value + 1
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
