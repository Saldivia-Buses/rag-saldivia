-- 021_erp_purchasing.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.purchasing.%';
DELETE FROM permissions WHERE id LIKE 'erp.purchasing.%';

DROP POLICY IF EXISTS tenant_isolation ON erp_purchase_receipt_lines;
DROP POLICY IF EXISTS tenant_isolation ON erp_purchase_receipts;
DROP POLICY IF EXISTS tenant_isolation ON erp_purchase_order_lines;
DROP POLICY IF EXISTS tenant_isolation ON erp_purchase_orders;

DROP TABLE IF EXISTS erp_purchase_receipt_lines;
DROP TABLE IF EXISTS erp_purchase_receipts;
DROP TABLE IF EXISTS erp_purchase_order_lines;
DROP TABLE IF EXISTS erp_purchase_orders;
