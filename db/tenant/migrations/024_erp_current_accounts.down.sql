-- 024_erp_current_accounts.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.accounts.%';
DELETE FROM permissions WHERE id LIKE 'erp.accounts.%';
DROP TRIGGER IF EXISTS trg_acct_mov_immutable ON erp_account_movements;
DROP POLICY IF EXISTS tenant_isolation ON erp_payment_allocations;
DROP POLICY IF EXISTS tenant_isolation ON erp_account_movements;
DROP TABLE IF EXISTS erp_payment_allocations;
DROP TABLE IF EXISTS erp_account_movements;
