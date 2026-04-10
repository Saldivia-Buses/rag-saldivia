-- 028_erp_maintenance.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.maintenance.%';
DELETE FROM permissions WHERE id LIKE 'erp.maintenance.%';
DROP POLICY IF EXISTS tenant_isolation ON erp_fuel_logs;
DROP POLICY IF EXISTS tenant_isolation ON erp_work_order_parts;
DROP POLICY IF EXISTS tenant_isolation ON erp_work_orders;
DROP POLICY IF EXISTS tenant_isolation ON erp_maintenance_plans;
DROP POLICY IF EXISTS tenant_isolation ON erp_maintenance_assets;
DROP TABLE IF EXISTS erp_fuel_logs;
DROP TABLE IF EXISTS erp_work_order_parts;
DROP TABLE IF EXISTS erp_work_orders;
DROP TABLE IF EXISTS erp_maintenance_plans;
DROP TABLE IF EXISTS erp_maintenance_assets;
