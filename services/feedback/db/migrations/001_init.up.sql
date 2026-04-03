-- Feedback events (all types in one table, discriminated by category)
CREATE TABLE IF NOT EXISTS feedback_events (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    category    TEXT NOT NULL,
    module      TEXT NOT NULL,
    user_id     TEXT,
    score       INTEGER,
    thumbs      TEXT,
    severity    TEXT,
    status      TEXT NOT NULL DEFAULT 'open',
    context     JSONB NOT NULL DEFAULT '{}',
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_feedback_events_category
    ON feedback_events(category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_feedback_events_module
    ON feedback_events(module, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_feedback_events_user
    ON feedback_events(user_id, created_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_feedback_events_open_errors
    ON feedback_events(status, created_at DESC)
    WHERE category = 'error_report' AND status = 'open';

CREATE INDEX IF NOT EXISTS idx_feedback_events_open_features
    ON feedback_events(status, created_at DESC)
    WHERE category = 'feature_request' AND status = 'open';
