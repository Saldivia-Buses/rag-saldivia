-- Tenant DB — notification tables (applied on top of auth tables)

-- Notifications
CREATE TABLE notifications (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,                                    -- 'chat.new_message', 'auth.login_new_ip', 'ingest.completed', etc.
    title       TEXT NOT NULL,
    body        TEXT NOT NULL DEFAULT '',
    data        JSONB NOT NULL DEFAULT '{}',                      -- action-specific payload (session_id, job_id, etc.)
    channel     TEXT NOT NULL DEFAULT 'in_app',                   -- 'in_app', 'email', 'both'
    is_read     BOOLEAN NOT NULL DEFAULT false,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id) WHERE is_read = false;

-- User notification preferences
CREATE TABLE notification_preferences (
    user_id         TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_enabled   BOOLEAN NOT NULL DEFAULT true,
    in_app_enabled  BOOLEAN NOT NULL DEFAULT true,
    quiet_start     TIME,                                          -- quiet hours start (e.g., '22:00')
    quiet_end       TIME,                                          -- quiet hours end (e.g., '08:00')
    muted_types     TEXT[] NOT NULL DEFAULT '{}',                   -- types the user has muted
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
