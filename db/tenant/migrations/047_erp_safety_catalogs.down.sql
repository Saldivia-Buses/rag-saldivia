-- 047_erp_safety_catalogs.down.sql

DELETE FROM role_permissions WHERE permission_id LIKE 'erp.safety.%';
DELETE FROM permissions WHERE id LIKE 'erp.safety.%';

DROP TABLE IF EXISTS erp_risk_agents;
DROP TABLE IF EXISTS erp_body_parts;
DROP TABLE IF EXISTS erp_accident_types;
