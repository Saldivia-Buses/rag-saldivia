-- 029_erp_admin.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.admin.%';
DELETE FROM permissions WHERE id LIKE 'erp.admin.%';
DROP POLICY IF EXISTS tenant_isolation ON erp_survey_questions;
DROP POLICY IF EXISTS tenant_isolation ON erp_surveys;
DROP POLICY IF EXISTS tenant_isolation ON erp_calendar_events;
DROP POLICY IF EXISTS tenant_isolation ON erp_communication_recipients;
DROP POLICY IF EXISTS tenant_isolation ON erp_communications;
DROP TABLE IF EXISTS erp_survey_questions;
DROP TABLE IF EXISTS erp_surveys;
DROP TABLE IF EXISTS erp_calendar_events;
DROP TABLE IF EXISTS erp_communication_recipients;
DROP TABLE IF EXISTS erp_communications;
