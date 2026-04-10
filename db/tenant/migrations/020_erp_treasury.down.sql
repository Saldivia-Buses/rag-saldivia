-- 020_erp_treasury.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.treasury.%';
DELETE FROM permissions WHERE id LIKE 'erp.treasury.%';

DROP TRIGGER IF EXISTS trg_treasury_immutable ON erp_treasury_movements;

DROP POLICY IF EXISTS tenant_isolation ON erp_cash_counts;
DROP POLICY IF EXISTS tenant_isolation ON erp_checks;
DROP POLICY IF EXISTS tenant_isolation ON erp_treasury_movements;
DROP POLICY IF EXISTS tenant_isolation ON erp_cash_registers;
DROP POLICY IF EXISTS tenant_isolation ON erp_bank_accounts;

DROP TABLE IF EXISTS erp_cash_counts;
DROP TABLE IF EXISTS erp_checks;
DROP TABLE IF EXISTS erp_treasury_movements;
DROP TABLE IF EXISTS erp_cash_registers;
DROP TABLE IF EXISTS erp_bank_accounts;
