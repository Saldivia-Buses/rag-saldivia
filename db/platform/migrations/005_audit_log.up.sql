-- Platform DB — audit log for platform admin actions.
-- Unlike the tenant audit_log (which has FK to users), platform audit_log
-- stores user_id as plain TEXT since platform admins live in tenant DBs.

CREATE TABLE IF NOT EXISTS audit_log (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id     TEXT,                                     -- platform admin who did it (no FK — lives in tenant DB)
    action      TEXT NOT NULL,                            -- 'tenant.created', 'module.enabled', etc.
    resource    TEXT,                                     -- what was affected
    details     JSONB NOT NULL DEFAULT '{}',              -- action-specific data
    ip_address  TEXT,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_platform_audit_log_user ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_platform_audit_log_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_platform_audit_log_created ON audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_platform_audit_log_action_created ON audit_log(action, created_at DESC);

-- Match the action format constraint from tenant migration 002
ALTER TABLE audit_log ADD CONSTRAINT audit_log_action_format
    CHECK (action ~ '^[a-z]+\.[a-z_.]+$');
