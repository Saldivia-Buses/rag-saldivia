-- 010_dead_events.down.sql
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'permissions'
    ) THEN
        DELETE FROM role_permissions WHERE permission_id LIKE 'admin.dlq.%';
        DELETE FROM permissions WHERE id LIKE 'admin.dlq.%';
    END IF;
END $$;
DROP TABLE IF EXISTS dead_events_replays;
DROP TABLE IF EXISTS dead_events;
