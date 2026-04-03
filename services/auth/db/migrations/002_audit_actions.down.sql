DROP INDEX IF EXISTS idx_audit_log_action_created;
ALTER TABLE audit_log DROP CONSTRAINT IF EXISTS audit_log_action_format;
