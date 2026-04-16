-- Health check snapshots (one per service per check cycle)
CREATE TABLE IF NOT EXISTS health_snapshots (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    service     TEXT NOT NULL,
    status      TEXT NOT NULL CHECK (status IN ('healthy', 'degraded', 'unhealthy', 'offline')),
    response_ms INT,
    version     TEXT,
    details     JSONB,         -- arbitrary health data from /health
    checked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_health_snapshots_service ON health_snapshots(service, checked_at DESC);

-- Retention: daily cleanup via HealthWatch service cron (NOT a DB trigger).
-- HealthWatch runs cleanup once/day at startup or via internal scheduler:
--   DELETE FROM health_snapshots WHERE checked_at < now() - interval '7 days';
-- (~14 services * 1 check/min * 60 * 24 * 7 = 141K rows max at steady state)

-- Triage records (AI-generated analysis)
CREATE TABLE IF NOT EXISTS triage_records (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    severity      TEXT NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    title         TEXT NOT NULL,
    analysis      TEXT NOT NULL,  -- AI-generated analysis (scrubbed, no raw errors)
    services      TEXT[],         -- affected services
    github_issue  INT,            -- GitHub issue number if created
    status        TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'dismissed')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at   TIMESTAMPTZ
);
