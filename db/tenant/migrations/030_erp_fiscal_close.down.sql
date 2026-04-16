-- 030_erp_fiscal_close.down.sql
ALTER TABLE erp_fiscal_years
    DROP CONSTRAINT IF EXISTS erp_fiscal_years_result_account_required;

ALTER TABLE erp_fiscal_years
    DROP COLUMN IF EXISTS result_account_id,
    DROP COLUMN IF EXISTS opening_entry_id,
    DROP COLUMN IF EXISTS closing_entry_id,
    DROP COLUMN IF EXISTS closed_at,
    DROP COLUMN IF EXISTS closed_by;

DELETE FROM role_permissions WHERE permission_id = 'erp.accounting.close';
DELETE FROM permissions WHERE id = 'erp.accounting.close';
