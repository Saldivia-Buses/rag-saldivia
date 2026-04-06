-- Hourly aggregated metrics per tenant+module+category
CREATE TABLE IF NOT EXISTS feedback_metrics (
    tenant_id       TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module          TEXT NOT NULL,
    category        TEXT NOT NULL,
    period          TIMESTAMPTZ NOT NULL,
    total_events    INTEGER NOT NULL DEFAULT 0,
    positive        INTEGER NOT NULL DEFAULT 0,
    negative        INTEGER NOT NULL DEFAULT 0,
    avg_score       REAL,
    p50_latency_ms  REAL,
    p95_latency_ms  REAL,
    p99_latency_ms  REAL,
    error_count     INTEGER NOT NULL DEFAULT 0,
    metadata        JSONB NOT NULL DEFAULT '{}',
    PRIMARY KEY (tenant_id, module, category, period)
);

CREATE INDEX IF NOT EXISTS idx_feedback_metrics_tenant
    ON feedback_metrics(tenant_id, period DESC);

CREATE INDEX IF NOT EXISTS idx_feedback_metrics_category
    ON feedback_metrics(category, period DESC);

-- Tenant health scores (composite, updated hourly)
CREATE TABLE IF NOT EXISTS tenant_health_scores (
    tenant_id           TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    period              TIMESTAMPTZ NOT NULL,
    overall_score       REAL NOT NULL,
    ai_quality_score    REAL NOT NULL DEFAULT 100,
    error_rate_score    REAL NOT NULL DEFAULT 100,
    usage_score         REAL NOT NULL DEFAULT 100,
    performance_score   REAL NOT NULL DEFAULT 100,
    security_score      REAL NOT NULL DEFAULT 100,
    nps_score           REAL,
    details             JSONB NOT NULL DEFAULT '{}',
    PRIMARY KEY (tenant_id, period)
);

CREATE INDEX IF NOT EXISTS idx_tenant_health_scores_period
    ON tenant_health_scores(period DESC);

-- Active alerts (cross-tenant, managed by feedback service)
CREATE TABLE IF NOT EXISTS feedback_alerts (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id       TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    alert_type      TEXT NOT NULL,
    severity        TEXT NOT NULL DEFAULT 'warning',
    module          TEXT,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL,
    threshold       TEXT,
    current_value   TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ,
    acknowledged_by TEXT
);

CREATE INDEX IF NOT EXISTS idx_feedback_alerts_active
    ON feedback_alerts(status, created_at DESC)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_feedback_alerts_tenant
    ON feedback_alerts(tenant_id, created_at DESC);
