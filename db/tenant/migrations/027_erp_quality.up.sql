-- 027_erp_quality.up.sql
-- Plan 17 Phase 10: Quality
-- Replaces ~42 legacy tables: CAL_*, RIESGOS, CONTROLCALIDAD, INSNORMA, etc.

CREATE TABLE IF NOT EXISTS erp_nonconformities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    number      TEXT NOT NULL,
    date        DATE NOT NULL,
    type_id     UUID REFERENCES erp_catalogs(id),
    origin_id   UUID REFERENCES erp_catalogs(id),
    description TEXT NOT NULL,
    severity    TEXT NOT NULL DEFAULT 'minor' CHECK (severity IN ('minor','major','critical')),
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','investigating','corrective_action','closed')),
    assigned_to UUID REFERENCES erp_entities(id),
    closed_at   TIMESTAMPTZ,
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, number)
);
CREATE INDEX idx_erp_nc_status ON erp_nonconformities(tenant_id, status) WHERE status != 'closed';

CREATE TABLE IF NOT EXISTS erp_corrective_actions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      TEXT NOT NULL,
    nc_id          UUID NOT NULL REFERENCES erp_nonconformities(id),
    action_type    TEXT NOT NULL CHECK (action_type IN ('corrective','preventive')),
    description    TEXT NOT NULL,
    responsible_id UUID REFERENCES erp_entities(id),
    due_date       DATE,
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','completed','verified')),
    completed_at   TIMESTAMPTZ,
    effectiveness  TEXT CHECK (effectiveness IS NULL OR effectiveness IN ('effective','ineffective','pending_review'))
);
CREATE INDEX idx_erp_ca_nc ON erp_corrective_actions(tenant_id, nc_id);

CREATE TABLE IF NOT EXISTS erp_audits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    number          TEXT NOT NULL,
    date            DATE NOT NULL,
    audit_type      TEXT NOT NULL CHECK (audit_type IN ('internal','external','supplier')),
    scope           TEXT NOT NULL DEFAULT '',
    lead_auditor_id UUID REFERENCES erp_entities(id),
    status          TEXT NOT NULL DEFAULT 'planned' CHECK (status IN ('planned','in_progress','completed')),
    score           NUMERIC(5,2),
    notes           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_audit_findings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    audit_id     UUID NOT NULL REFERENCES erp_audits(id) ON DELETE CASCADE,
    finding_type TEXT NOT NULL CHECK (finding_type IN ('observation','minor_nc','major_nc','opportunity')),
    description  TEXT NOT NULL,
    nc_id        UUID REFERENCES erp_nonconformities(id)
);
CREATE INDEX idx_erp_audit_findings ON erp_audit_findings(tenant_id, audit_id);

CREATE TABLE IF NOT EXISTS erp_controlled_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    code        TEXT NOT NULL,
    title       TEXT NOT NULL,
    revision    INT NOT NULL DEFAULT 1,
    doc_type_id UUID REFERENCES erp_catalogs(id),
    file_key    TEXT NOT NULL,
    approved_by UUID REFERENCES erp_entities(id),
    approved_at TIMESTAMPTZ,
    status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved','obsolete')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, code, revision)
);

-- RLS
ALTER TABLE erp_nonconformities ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_nonconformities USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_corrective_actions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_corrective_actions USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_audits ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_audits USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_audit_findings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_audit_findings USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_controlled_documents ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_controlled_documents USING (tenant_id = current_setting('app.tenant_id', true));

INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.quality.read',  'Ver calidad',      'Consultar NC, auditorias, documentos', 'erp'),
    ('erp.quality.write', 'Gestionar calidad', 'Crear NC, acciones, auditorias',      'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.quality.%'
ON CONFLICT DO NOTHING;
