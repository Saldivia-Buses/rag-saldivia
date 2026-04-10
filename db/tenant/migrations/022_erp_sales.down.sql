-- 022_erp_sales.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.sales.%';
DELETE FROM permissions WHERE id LIKE 'erp.sales.%';
DROP POLICY IF EXISTS tenant_isolation ON erp_price_list_items;
DROP POLICY IF EXISTS tenant_isolation ON erp_price_lists;
DROP POLICY IF EXISTS tenant_isolation ON erp_orders;
DROP POLICY IF EXISTS tenant_isolation ON erp_quotation_lines;
DROP POLICY IF EXISTS tenant_isolation ON erp_quotations;
DROP TABLE IF EXISTS erp_price_list_items;
DROP TABLE IF EXISTS erp_price_lists;
DROP TABLE IF EXISTS erp_orders;
DROP TABLE IF EXISTS erp_quotation_lines;
DROP TABLE IF EXISTS erp_quotations;
