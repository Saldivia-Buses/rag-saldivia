-- 072_erp_tools.down.sql — reverse of 072_erp_tools.up.sql

DELETE FROM role_permissions WHERE permission_id LIKE 'erp.tools.%';
DELETE FROM permissions WHERE id LIKE 'erp.tools.%';

DROP TABLE IF EXISTS erp_tool_movements;
DROP TABLE IF EXISTS erp_tools;
