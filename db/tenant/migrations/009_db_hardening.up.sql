-- Plan 08 Phase 3: Database hardening

-- L7: FK documents.uploaded_by → users(id)
ALTER TABLE documents ADD CONSTRAINT fk_documents_uploaded_by
    FOREIGN KEY (uploaded_by) REFERENCES users(id);

-- L8: Generated column for feedback_events latency (avoids JSONB extraction in queries)
-- CASE guard prevents insert failures when latency_ms is missing or non-numeric
ALTER TABLE feedback_events ADD COLUMN IF NOT EXISTS latency_ms NUMERIC
    GENERATED ALWAYS AS (
        CASE WHEN context->>'latency_ms' ~ '^\d+(\.\d+)?$'
             THEN (context->>'latency_ms')::numeric
             ELSE NULL
        END
    ) STORED;
CREATE INDEX IF NOT EXISTS idx_feedback_events_latency ON feedback_events(latency_ms)
    WHERE category = 'usage';

-- L11: Drop redundant index (UNIQUE constraint on token_hash already creates one)
DROP INDEX IF EXISTS idx_refresh_tokens_hash;
