-- Reverse Plan 08 Phase 3 DB hardening

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);

DROP INDEX IF EXISTS idx_feedback_events_latency;
ALTER TABLE feedback_events DROP COLUMN IF EXISTS latency_ms;

ALTER TABLE documents DROP CONSTRAINT IF EXISTS fk_documents_uploaded_by;
