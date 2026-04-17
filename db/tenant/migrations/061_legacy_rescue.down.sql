-- Down migration for 061_legacy_rescue
--
-- We re-tighten the quantity/amount checks. Any rescued zero-amount rows that
-- were inserted while 061 was up will violate the restored CHECK — operators
-- must delete or patch them before running this down migration. Failing
-- loudly is intentional: silent data loss is what 061 was built to prevent.

DROP INDEX IF EXISTS idx_stock_movements_legacy_notes;
DROP INDEX IF EXISTS idx_treasury_movements_legacy_notes;
DROP INDEX IF EXISTS idx_account_movements_legacy_notes;
DROP INDEX IF EXISTS idx_users_legacy_login;
DROP INDEX IF EXISTS idx_roles_metadata_legacy;

ALTER TABLE users DROP COLUMN IF EXISTS legacy_login;
ALTER TABLE users DROP COLUMN IF EXISTS legacy_profile_id;
ALTER TABLE roles DROP COLUMN IF EXISTS metadata;

ALTER TABLE erp_stock_movements DROP CONSTRAINT IF EXISTS erp_stock_movements_quantity_check;
ALTER TABLE erp_stock_movements ADD CONSTRAINT erp_stock_movements_quantity_check CHECK (quantity > 0);

ALTER TABLE erp_treasury_movements DROP CONSTRAINT IF EXISTS erp_treasury_movements_amount_check;
ALTER TABLE erp_treasury_movements ADD CONSTRAINT erp_treasury_movements_amount_check CHECK (amount > 0);

ALTER TABLE erp_account_movements DROP CONSTRAINT IF EXISTS erp_account_movements_amount_check;
ALTER TABLE erp_account_movements ADD CONSTRAINT erp_account_movements_amount_check CHECK (amount > 0);

ALTER TABLE erp_withholdings DROP CONSTRAINT IF EXISTS erp_withholdings_amount_check;
ALTER TABLE erp_withholdings ADD CONSTRAINT erp_withholdings_amount_check CHECK (amount > 0);

ALTER TABLE erp_payment_allocations DROP CONSTRAINT IF EXISTS erp_payment_allocations_amount_check;
ALTER TABLE erp_payment_allocations ADD CONSTRAINT erp_payment_allocations_amount_check CHECK (amount > 0);

ALTER TABLE erp_tax_entries ALTER COLUMN invoice_id SET NOT NULL;
