-- 070_erp_time_clock.down.sql — reverse of 070_erp_time_clock.up.sql

DELETE FROM role_permissions WHERE permission_id LIKE 'erp.time_clock.%';
DELETE FROM permissions WHERE id LIKE 'erp.time_clock.%';

DROP TABLE IF EXISTS erp_time_clock_events;
DROP TABLE IF EXISTS erp_employee_cards;
