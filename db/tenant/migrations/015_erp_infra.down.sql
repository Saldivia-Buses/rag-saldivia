-- 015_erp_infra.down.sql

DROP POLICY IF EXISTS tenant_isolation ON erp_suggestion_responses;
ALTER TABLE erp_suggestion_responses DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON erp_suggestions;
ALTER TABLE erp_suggestions DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS erp_validate_journal_balance();
DROP FUNCTION IF EXISTS next_erp_sequence(TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS erp_prevent_financial_mutation();

ALTER TABLE audit_log DROP COLUMN IF EXISTS tenant_id;
