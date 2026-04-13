-- 052_erp_customer_vehicles.down.sql
DROP TABLE IF EXISTS erp_vehicle_incidents;
DROP TABLE IF EXISTS erp_vehicle_incident_types;
DROP TABLE IF EXISTS erp_customer_vehicles;

DELETE FROM role_permissions WHERE permission_id LIKE 'erp.maintenance.%';
DELETE FROM permissions WHERE id LIKE 'erp.maintenance.%';
