-- Ingest service schema: job tracking and connectors

CREATE TABLE IF NOT EXISTS ingest_jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    collection  TEXT NOT NULL,
    file_name   TEXT NOT NULL,
    file_size   BIGINT NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ingest_jobs_user ON ingest_jobs (user_id, created_at DESC);
CREATE INDEX idx_ingest_jobs_status ON ingest_jobs (status) WHERE status IN ('pending', 'processing');

CREATE TABLE IF NOT EXISTS connectors (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('gdrive', 'onedrive', 's3', 'local')),
    config      JSONB NOT NULL DEFAULT '{}',
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
