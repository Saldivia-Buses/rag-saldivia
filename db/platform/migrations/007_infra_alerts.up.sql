-- Infrastructure alerts from Alertmanager.
-- Stored in the platform DB because alerts are infra-level, not tenant-scoped.
-- Exception to NATS invariant 4: no tenant context → no NATS event.

CREATE TABLE IF NOT EXISTS infra_alerts (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    fingerprint TEXT NOT NULL,
    status      TEXT NOT NULL,                              -- 'firing', 'resolved'
    severity    TEXT NOT NULL,                              -- 'critical', 'warning', 'info'
    alertname   TEXT NOT NULL,
    service     TEXT,                                       -- service_name label
    summary     TEXT,                                       -- annotations.summary
    description TEXT,                                       -- annotations.description
    labels      JSONB NOT NULL DEFAULT '{}',
    annotations JSONB NOT NULL DEFAULT '{}',
    starts_at   TIMESTAMPTZ NOT NULL,
    ends_at     TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Dedup: same alert firing can be re-sent by Alertmanager on each repeat_interval.
-- UPSERT by (fingerprint, starts_at) prevents unbounded duplicate rows.
CREATE UNIQUE INDEX IF NOT EXISTS idx_infra_alerts_dedup ON infra_alerts(fingerprint, starts_at);
CREATE INDEX IF NOT EXISTS idx_infra_alerts_received ON infra_alerts(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_infra_alerts_severity ON infra_alerts(severity, received_at DESC);
