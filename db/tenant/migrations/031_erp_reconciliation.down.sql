-- 031_erp_reconciliation.down.sql
DROP TABLE IF EXISTS erp_bank_statement_lines;
DROP TABLE IF EXISTS erp_bank_reconciliations;

ALTER TABLE erp_treasury_movements
    DROP COLUMN IF EXISTS reconciled,
    DROP COLUMN IF EXISTS reconciliation_id;

-- Restore generic trigger
DROP TRIGGER IF EXISTS trg_treasury_immutable ON erp_treasury_movements;
DROP FUNCTION IF EXISTS erp_prevent_treasury_mutation();
CREATE TRIGGER trg_treasury_immutable BEFORE UPDATE OR DELETE ON erp_treasury_movements
    FOR EACH ROW WHEN (OLD.status IN ('confirmed', 'reversed'))
    EXECUTE FUNCTION erp_prevent_financial_mutation();

DELETE FROM role_permissions WHERE permission_id = 'erp.treasury.reconcile';
DELETE FROM permissions WHERE id = 'erp.treasury.reconcile';
