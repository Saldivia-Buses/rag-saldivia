-- Tenant DB — chat tables (applied on top of auth tables)

-- Chat sessions
CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL DEFAULT 'Nueva conversacion',
    collection  TEXT,                                    -- RAG collection name (NULL = no RAG)
    is_saved    BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id, updated_at DESC);

-- Chat messages
CREATE TABLE IF NOT EXISTS messages (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    session_id  TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content     TEXT NOT NULL,
    sources     JSONB,                                   -- RAG citations [{document_name, content, score}]
    metadata    JSONB NOT NULL DEFAULT '{}',              -- focus_mode, model, tokens, etc.
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, created_at);

-- Session tags
CREATE TABLE IF NOT EXISTS tags (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    session_id  TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (session_id, name)
);

-- Chat feedback (thumbs up/down on messages)
CREATE TABLE IF NOT EXISTS chat_feedback (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    message_id  TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL REFERENCES users(id),
    thumbs      TEXT NOT NULL CHECK (thumbs IN ('up', 'down')),
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (message_id, user_id)
);
