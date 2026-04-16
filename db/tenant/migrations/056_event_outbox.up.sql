-- 056_event_outbox.up.sql
-- Plan 26 Fase 3: transactional outbox for spine events.
--
-- Each tenant DB gets this table. Services INSERT into event_outbox within the
-- same tx as the business write (e.g. INSERT messages). The DrainerWorker
-- polls for unpublished rows, publishes to NATS, and marks them published.
--
-- Concurrency between replicas is handled via SELECT ... FOR UPDATE SKIP LOCKED
-- in the drainer query (see pkg/outbox/worker.go).
--
-- The NOTIFY trigger wakes up any drainer listening on this tenant pool so it
-- can drain immediately instead of waiting for the next poll cycle.

CREATE TABLE event_outbox (
    id               uuid PRIMARY KEY,
    subject          text NOT NULL,
    payload          jsonb NOT NULL,
    headers          jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at       timestamptz NOT NULL DEFAULT now(),
    published_at     timestamptz,
    attempts         int NOT NULL DEFAULT 0,
    last_error       text,
    next_attempt_at  timestamptz
);

CREATE INDEX event_outbox_drainable
    ON event_outbox (created_at)
    WHERE published_at IS NULL;

-- Trigger to NOTIFY drainers immediately on INSERT.
CREATE OR REPLACE FUNCTION spine_outbox_notify() RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('spine_outbox_new', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER event_outbox_notify_insert
    AFTER INSERT ON event_outbox
    FOR EACH ROW EXECUTE FUNCTION spine_outbox_notify();
