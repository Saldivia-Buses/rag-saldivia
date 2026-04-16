-- 018_erp_stock.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.stock.%';
DELETE FROM permissions WHERE id LIKE 'erp.stock.%';

DROP POLICY IF EXISTS tenant_isolation ON erp_article_photos;
DROP POLICY IF EXISTS tenant_isolation ON erp_bom;
DROP POLICY IF EXISTS tenant_isolation ON erp_stock_movements;
DROP POLICY IF EXISTS tenant_isolation ON erp_stock_levels;
DROP POLICY IF EXISTS tenant_isolation ON erp_warehouses;
DROP POLICY IF EXISTS tenant_isolation ON erp_articles;

DROP TABLE IF EXISTS erp_article_photos;
DROP TABLE IF EXISTS erp_bom;
DROP TABLE IF EXISTS erp_stock_movements;
DROP TABLE IF EXISTS erp_stock_levels;
DROP TABLE IF EXISTS erp_warehouses;
DROP TABLE IF EXISTS erp_articles;
