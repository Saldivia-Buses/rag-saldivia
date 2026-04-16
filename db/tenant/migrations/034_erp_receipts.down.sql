-- 034_erp_receipts.down.sql
DROP TABLE IF EXISTS erp_receipt_withholdings;
DROP TABLE IF EXISTS erp_receipt_allocations;
DROP TABLE IF EXISTS erp_receipt_payments;
DROP TRIGGER IF EXISTS trg_receipt_immutable ON erp_receipts;
DROP TABLE IF EXISTS erp_receipts;
DELETE FROM role_permissions WHERE permission_id = 'erp.treasury.receipt';
DELETE FROM permissions WHERE id = 'erp.treasury.receipt';
