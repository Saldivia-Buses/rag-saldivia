-- 019_erp_accounting.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.accounting.%';
DELETE FROM permissions WHERE id LIKE 'erp.accounting.%';

DROP TRIGGER IF EXISTS trg_journal_balance ON erp_journal_entries;
DROP TRIGGER IF EXISTS trg_journal_immutable ON erp_journal_entries;

DROP POLICY IF EXISTS tenant_isolation ON erp_journal_lines;
DROP POLICY IF EXISTS tenant_isolation ON erp_journal_entries;
DROP POLICY IF EXISTS tenant_isolation ON erp_fiscal_years;
DROP POLICY IF EXISTS tenant_isolation ON erp_accounts;
DROP POLICY IF EXISTS tenant_isolation ON erp_cost_centers;

DROP TABLE IF EXISTS erp_journal_lines;
DROP TABLE IF EXISTS erp_journal_entries;
DROP TABLE IF EXISTS erp_fiscal_years;
DROP TABLE IF EXISTS erp_accounts;
DROP TABLE IF EXISTS erp_cost_centers;
