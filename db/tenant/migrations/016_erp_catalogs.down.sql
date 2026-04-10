-- 016_erp_catalogs.down.sql

DELETE FROM role_permissions WHERE permission_id IN ('erp.catalogs.read', 'erp.catalogs.write');
DELETE FROM permissions WHERE id IN ('erp.catalogs.read', 'erp.catalogs.write');

DROP POLICY IF EXISTS tenant_isolation ON erp_sequences;
DROP TABLE IF EXISTS erp_sequences;

DROP POLICY IF EXISTS tenant_isolation ON erp_catalogs;
DROP TABLE IF EXISTS erp_catalogs;
