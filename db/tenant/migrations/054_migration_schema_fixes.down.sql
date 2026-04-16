-- 054_migration_schema_fixes.down.sql

-- Revert users
ALTER TABLE users DROP COLUMN IF EXISTS force_password_reset;

-- Revert purchase orders
ALTER TABLE erp_purchase_orders ALTER COLUMN supplier_id SET NOT NULL;

-- Revert metadata columns
ALTER TABLE erp_production_inspections DROP COLUMN IF EXISTS metadata;
ALTER TABLE erp_production_orders DROP COLUMN IF EXISTS metadata;
ALTER TABLE erp_quotations DROP COLUMN IF EXISTS metadata;
ALTER TABLE erp_audits DROP COLUMN IF EXISTS metadata;
ALTER TABLE erp_controlled_documents DROP COLUMN IF EXISTS metadata;
ALTER TABLE erp_account_movements DROP COLUMN IF EXISTS metadata;

-- Revert BOM history
DROP POLICY IF EXISTS tenant_isolation ON erp_bom_history;
DROP TABLE IF EXISTS erp_bom_history;

-- Revert treasury status CHECK
ALTER TABLE erp_treasury_movements DROP CONSTRAINT IF EXISTS erp_treasury_movements_status_check;
ALTER TABLE erp_treasury_movements ADD CONSTRAINT erp_treasury_movements_status_check
    CHECK (status IN ('pending', 'confirmed', 'reversed'));

-- Revert invoice type CHECK
ALTER TABLE erp_invoices DROP CONSTRAINT IF EXISTS erp_invoices_invoice_type_check;
ALTER TABLE erp_invoices ADD CONSTRAINT erp_invoices_invoice_type_check
    CHECK (invoice_type IN (
        'invoice_a','invoice_b','invoice_c','invoice_e',
        'credit_note','debit_note','delivery_note'
    ));
