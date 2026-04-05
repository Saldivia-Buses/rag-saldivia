-- Plan 06: Intelligence schema — replaces ingest_jobs with richer document model.
-- documents + document_pages + document_trees + collections + tool_calls

CREATE TABLE IF NOT EXISTS documents (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name        TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    file_type   TEXT NOT NULL,
    file_hash   TEXT NOT NULL,
    size_bytes  BIGINT NOT NULL,
    total_pages INT,
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'extracting', 'indexing', 'ready', 'error')),
    metadata    JSONB DEFAULT '{}',
    uploaded_by TEXT NOT NULL,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_hash ON documents(file_hash);
CREATE INDEX IF NOT EXISTS idx_documents_uploaded_by ON documents(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);

CREATE TABLE IF NOT EXISTS document_pages (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    document_id   TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_number   INT NOT NULL,
    text          TEXT NOT NULL DEFAULT '',
    tables        JSONB DEFAULT '[]',
    images        JSONB DEFAULT '[]',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(document_id, page_number)
);

CREATE INDEX IF NOT EXISTS idx_document_pages_doc ON document_pages(document_id, page_number);

CREATE TABLE IF NOT EXISTS document_trees (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    document_id     TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tree            JSONB NOT NULL,
    doc_description TEXT,
    tree_version    INT NOT NULL DEFAULT 1,
    model_used      TEXT NOT NULL,
    node_count      INT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_document_trees_doc ON document_trees(document_id);

CREATE TABLE IF NOT EXISTS collections (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS collection_documents (
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    document_id   TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, document_id)
);

CREATE TABLE IF NOT EXISTS tool_calls (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    trace_id    TEXT NOT NULL,
    session_id  TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    tool_name   TEXT NOT NULL,
    input       JSONB NOT NULL,
    output      JSONB,
    status      TEXT NOT NULL CHECK (status IN ('success', 'error', 'timeout', 'denied')),
    duration_ms INT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tool_calls_session ON tool_calls(session_id);
CREATE INDEX IF NOT EXISTS idx_tool_calls_user ON tool_calls(user_id, created_at);
