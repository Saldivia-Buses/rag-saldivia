-- 010_dead_events.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'admin.dlq.%';
DELETE FROM permissions WHERE id LIKE 'admin.dlq.%';
DROP TABLE IF EXISTS dead_events_replays;
DROP TABLE IF EXISTS dead_events;
