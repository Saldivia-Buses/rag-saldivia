-- 047_erp_safety_catalogs.up.sql
-- Catálogos base de seguridad: tipos de accidente, partes del cuerpo, agentes de riesgo

-- Accident types catalog
CREATE TABLE erp_accident_types (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    name         TEXT NOT NULL,
    abbreviation TEXT NOT NULL DEFAULT '',
    severity_idx NUMERIC(3,2) NOT NULL DEFAULT 1.0,
    active       BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);
ALTER TABLE erp_accident_types ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_accident_types
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Body parts catalog
CREATE TABLE erp_body_parts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, description)
);
ALTER TABLE erp_body_parts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_body_parts
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Risk agents catalog
CREATE TABLE erp_risk_agents (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  TEXT NOT NULL,
    name       TEXT NOT NULL,
    risk_type  TEXT NOT NULL DEFAULT 'physical'
               CHECK (risk_type IN ('physical','chemical','ergonomic','biological','psychosocial')),
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);
ALTER TABLE erp_risk_agents ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_risk_agents
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.safety.read',  'Ver seguridad',      'Consultar accidentes, exposiciones, licencias médicas', 'erp'),
    ('erp.safety.write', 'Gestionar seguridad', 'Crear y editar registros de seguridad y salud laboral', 'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.safety.%'
ON CONFLICT DO NOTHING;
