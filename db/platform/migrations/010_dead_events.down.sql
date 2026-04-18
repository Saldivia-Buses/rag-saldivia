-- 010_dead_events.down.sql
-- Permission rows are rolled back by the matching tenant migration
-- (064_dlq_admin_permissions.down.sql).
DROP TABLE IF EXISTS dead_events_replays;
DROP TABLE IF EXISTS dead_events;
