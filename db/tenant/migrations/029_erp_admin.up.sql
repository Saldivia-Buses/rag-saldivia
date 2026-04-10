-- 029_erp_admin.up.sql
-- Plan 17 Phase 12: Admin (completing remaining modules)
-- Replaces ~8 legacy tables: COMUNICACIONES, CALENDAR, EVENTOS, ENCUESTA, etc.

-- Internal communications
CREATE TABLE IF NOT EXISTS erp_communications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    subject     TEXT NOT NULL,
    body        TEXT NOT NULL,
    sender_id   TEXT NOT NULL,
    priority    TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low','normal','high','urgent')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_comms ON erp_communications(tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS erp_communication_recipients (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    communication_id UUID NOT NULL REFERENCES erp_communications(id) ON DELETE CASCADE,
    recipient_id     TEXT NOT NULL,
    read_at          TIMESTAMPTZ
);
CREATE INDEX idx_erp_comm_recipients ON erp_communication_recipients(tenant_id, recipient_id, read_at);

-- Calendar events
CREATE TABLE IF NOT EXISTS erp_calendar_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ,
    all_day     BOOLEAN NOT NULL DEFAULT false,
    entity_id   UUID REFERENCES erp_entities(id),
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_calendar ON erp_calendar_events(tenant_id, start_at);

-- Surveys
CREATE TABLE IF NOT EXISTS erp_surveys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','closed')),
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_survey_questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    survey_id   UUID NOT NULL REFERENCES erp_surveys(id) ON DELETE CASCADE,
    question    TEXT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0
);

-- RLS
ALTER TABLE erp_communications ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_communications USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_communication_recipients ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_communication_recipients USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_calendar_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_calendar_events USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_surveys ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_surveys USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_survey_questions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_survey_questions USING (tenant_id = current_setting('app.tenant_id', true));

INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.admin.read',  'Ver admin ERP',      'Comunicaciones, calendario, encuestas', 'erp'),
    ('erp.admin.write', 'Gestionar admin ERP', 'Crear comunicaciones, eventos',        'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.admin.%'
ON CONFLICT DO NOTHING;
