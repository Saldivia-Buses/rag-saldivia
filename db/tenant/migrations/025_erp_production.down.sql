-- 025_erp_production.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.production.%';
DELETE FROM permissions WHERE id LIKE 'erp.production.%';
DROP POLICY IF EXISTS tenant_isolation ON erp_unit_photos;
DROP POLICY IF EXISTS tenant_isolation ON erp_units;
DROP POLICY IF EXISTS tenant_isolation ON erp_production_inspections;
DROP POLICY IF EXISTS tenant_isolation ON erp_production_steps;
DROP POLICY IF EXISTS tenant_isolation ON erp_production_materials;
DROP POLICY IF EXISTS tenant_isolation ON erp_production_orders;
DROP POLICY IF EXISTS tenant_isolation ON erp_production_centers;
DROP TABLE IF EXISTS erp_unit_photos;
DROP TABLE IF EXISTS erp_units;
DROP TABLE IF EXISTS erp_production_inspections;
DROP TABLE IF EXISTS erp_production_steps;
DROP TABLE IF EXISTS erp_production_materials;
DROP TABLE IF EXISTS erp_production_orders;
DROP TABLE IF EXISTS erp_production_centers;
