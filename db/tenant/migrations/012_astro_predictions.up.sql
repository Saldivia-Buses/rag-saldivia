-- Prediction tracking (for verification and ROI)
CREATE TABLE IF NOT EXISTS astro_predictions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    user_id       UUID NOT NULL,
    session_id    UUID REFERENCES astro_sessions(id) ON DELETE SET NULL,
    contact_id    UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    category      TEXT NOT NULL CHECK (category IN ('timing', 'event', 'financial', 'relational', 'health', 'general')),
    description   TEXT NOT NULL,
    date_from     DATE NOT NULL,
    date_to       DATE NOT NULL,
    techniques    TEXT[],
    outcome       TEXT DEFAULT 'pending' CHECK (outcome IN ('correct', 'incorrect', 'partial', 'pending')),
    outcome_notes TEXT,
    verified_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_astro_predictions_user
    ON astro_predictions(tenant_id, user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_astro_predictions_contact
    ON astro_predictions(contact_id, date_from);
CREATE INDEX IF NOT EXISTS idx_astro_predictions_pending
    ON astro_predictions(tenant_id, user_id, outcome) WHERE outcome = 'pending';

-- Feedback (thumbs up/down on messages)
CREATE TABLE IF NOT EXISTS astro_feedback (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    message_id  UUID NOT NULL REFERENCES astro_messages(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    thumbs      TEXT NOT NULL CHECK (thumbs IN ('up', 'down')),
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (message_id, user_id)
);

-- Business follow-ups (generic, tenant-isolated)
CREATE TABLE IF NOT EXISTS astro_followups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    user_id         UUID NOT NULL,
    company_id      UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    counterparty_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    due_date        DATE,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'done', 'dismissed')),
    category        TEXT NOT NULL DEFAULT 'general',
    astro_basis     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_astro_followups_user
    ON astro_followups(tenant_id, user_id, status, due_date);

-- Daily usage tracking
CREATE TABLE IF NOT EXISTS astro_usage (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    user_id    UUID NOT NULL,
    date       DATE NOT NULL DEFAULT CURRENT_DATE,
    queries    INT NOT NULL DEFAULT 0,
    tokens_in  INT NOT NULL DEFAULT 0,
    tokens_out INT NOT NULL DEFAULT 0,
    UNIQUE (tenant_id, user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_astro_usage_date
    ON astro_usage(tenant_id, user_id, date DESC);
