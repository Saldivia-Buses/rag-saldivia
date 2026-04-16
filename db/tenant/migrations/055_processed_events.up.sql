-- 055_processed_events.up.sql
-- Plan 26 Fase 2: idempotency for spine consumers.
--
-- Each tenant DB gets this table. The spine consumer framework inserts a row
-- inside the same tx as the handler so the (event_id, consumer_name) pair is
-- atomically marked processed. Duplicate redeliveries are no-ops because the
-- INSERT ... ON CONFLICT DO NOTHING short-circuits.
--
-- TTL: a nightly job (added in Fase 5) deletes rows where processed_at >
-- 30 days. The retention horizon matches JetStream stream retention so a
-- replay older than 30 days CAN re-execute (intentional — see plan section
-- "Replay no es time-travel").

CREATE TABLE processed_events (
    event_id      uuid NOT NULL,
    consumer_name text NOT NULL,
    processed_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, consumer_name)
);

CREATE INDEX processed_events_ttl ON processed_events (processed_at);
