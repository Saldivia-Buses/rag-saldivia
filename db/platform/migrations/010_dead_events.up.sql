-- 010_dead_events.up.sql
-- Plan 26 Fase 4: dead-letter persistence for the spine DLQ supervisor.
--
-- Healthwatch's DLQ consumer reads from dlq.> JetStream stream and persists
-- each entry here. Operators can list, replay, or drop dead events via the
-- admin API (/v1/admin/dlq).
--
-- Lives in platform DB (cross-tenant operational data). Tenant context is
-- preserved inside the original envelope JSON.

CREATE TABLE dead_events (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    original_subject   text NOT NULL,
    original_stream    text NOT NULL,
    consumer_name      text NOT NULL,
    tenant_id          text,
    event_type         text,
    delivery_count     int NOT NULL DEFAULT 0,
    last_error         text NOT NULL DEFAULT '',
    dead_at            timestamptz NOT NULL DEFAULT now(),
    envelope           jsonb NOT NULL,
    headers            jsonb,
    replay_count       int NOT NULL DEFAULT 0,
    last_replayed_at   timestamptz,
    dropped_at         timestamptz
);

CREATE INDEX dead_events_tenant ON dead_events (tenant_id, dead_at DESC)
    WHERE dropped_at IS NULL;
CREATE INDEX dead_events_consumer ON dead_events (consumer_name, dead_at DESC)
    WHERE dropped_at IS NULL;

-- Replay history (separate table per plan review — auditable, supports
-- multiple replays of the same dead event).
CREATE TABLE dead_events_replays (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dead_event_id       uuid NOT NULL REFERENCES dead_events(id),
    replayed_at         timestamptz NOT NULL DEFAULT now(),
    replayed_by_user_id text NOT NULL,
    new_event_id        uuid NOT NULL,
    status              text NOT NULL CHECK (status IN ('pending','succeeded','failed'))
);

CREATE INDEX dead_events_replays_dead ON dead_events_replays (dead_event_id, replayed_at);

-- DLQ admin permissions are seeded in the tenant DB by
-- db/tenant/migrations/064_dlq_admin_permissions.up.sql — keep this
-- file pure platform DDL per ADR 023's split-database silo model.
