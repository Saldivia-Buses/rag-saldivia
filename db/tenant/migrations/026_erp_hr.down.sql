-- 026_erp_hr.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.hr.%';
DELETE FROM permissions WHERE id LIKE 'erp.hr.%';
DROP POLICY IF EXISTS tenant_isolation ON erp_attendance;
DROP POLICY IF EXISTS tenant_isolation ON erp_training_attendees;
DROP POLICY IF EXISTS tenant_isolation ON erp_training;
DROP POLICY IF EXISTS tenant_isolation ON erp_hr_events;
DROP POLICY IF EXISTS tenant_isolation ON erp_employee_details;
DROP POLICY IF EXISTS tenant_isolation ON erp_departments;
DROP TABLE IF EXISTS erp_attendance;
DROP TABLE IF EXISTS erp_training_attendees;
DROP TABLE IF EXISTS erp_training;
DROP TABLE IF EXISTS erp_hr_events;
DROP TABLE IF EXISTS erp_employee_details;
DROP TABLE IF EXISTS erp_departments;
