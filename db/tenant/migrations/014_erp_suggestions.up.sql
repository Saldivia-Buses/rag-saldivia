-- 014_erp_suggestions.up.sql
-- ERP Module: Suggestions (buzón de sugerencias)
-- Replaces legacy SUGERENCIAS + SUGERESP tables from Histrix

CREATE TABLE IF NOT EXISTS erp_suggestions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,              -- tenant slug (single-tenant V1)
    user_id     TEXT NOT NULL DEFAULT '',   -- user identifier from JWT
    origin      TEXT NOT NULL DEFAULT '',   -- department/area
    body        TEXT NOT NULL,              -- suggestion content
    is_read     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_erp_suggestions_tenant ON erp_suggestions(tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS erp_suggestion_responses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    suggestion_id   UUID NOT NULL REFERENCES erp_suggestions(id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL DEFAULT '',
    body            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_erp_suggestion_responses_suggestion
    ON erp_suggestion_responses(tenant_id, suggestion_id, created_at);
