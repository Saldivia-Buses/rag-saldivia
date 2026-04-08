-- Astro sessions (conversation threads for astrological consultations)
CREATE TABLE IF NOT EXISTS astro_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL,
    contact_id  UUID REFERENCES contacts(id) ON DELETE SET NULL,
    title       TEXT NOT NULL DEFAULT 'Nueva consulta',
    pinned      BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_astro_sessions_user
    ON astro_sessions(tenant_id, user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_astro_sessions_pinned
    ON astro_sessions(tenant_id, user_id, pinned, updated_at DESC);

-- Astro messages (linked to sessions)
-- role: only 'user' and 'assistant' from API (no 'system' — C3 review fix)
CREATE TABLE IF NOT EXISTS astro_messages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    session_id  UUID NOT NULL REFERENCES astro_sessions(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content     TEXT NOT NULL,
    thinking    TEXT,
    techniques  TEXT[],
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_astro_messages_session
    ON astro_messages(session_id, created_at);
CREATE INDEX IF NOT EXISTS idx_astro_messages_tenant
    ON astro_messages(tenant_id, session_id);
