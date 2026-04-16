-- 056_event_outbox.down.sql
DROP TRIGGER IF EXISTS event_outbox_notify_insert ON event_outbox;
DROP FUNCTION IF EXISTS spine_outbox_notify();
DROP TABLE IF EXISTS event_outbox;
