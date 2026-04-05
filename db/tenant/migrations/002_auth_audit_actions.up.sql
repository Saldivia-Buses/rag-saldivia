-- Strengthen audit_log: add action format validation and composite index.
-- The table is created in 001_init.up.sql.

-- Validate action follows service.entity.verb format (e.g., user.login, chat.session.create)
ALTER TABLE audit_log ADD CONSTRAINT audit_log_action_format
    CHECK (action ~ '^[a-z]+\.[a-z_.]+$');

-- Composite index for filtering by action type with time ordering
CREATE INDEX IF NOT EXISTS idx_audit_log_action_created
    ON audit_log(action, created_at DESC);
