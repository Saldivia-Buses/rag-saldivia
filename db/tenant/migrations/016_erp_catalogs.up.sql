-- 016_erp_catalogs.up.sql
-- Plan 17 Phase 0: Generic catalog system replacing 79 legacy lookup tables

CREATE TABLE IF NOT EXISTS erp_catalogs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    type        TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    parent_id   UUID REFERENCES erp_catalogs(id),
    sort_order  INT NOT NULL DEFAULT 0,
    active      BOOLEAN NOT NULL DEFAULT true,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, type, code),
    CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX IF NOT EXISTS idx_erp_catalogs_type
    ON erp_catalogs(tenant_id, type, active);
CREATE INDEX IF NOT EXISTS idx_erp_catalogs_parent
    ON erp_catalogs(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_erp_catalogs_meta
    ON erp_catalogs USING GIN(metadata) WHERE metadata != '{}'::jsonb;

-- RLS (pattern P1)
ALTER TABLE erp_catalogs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_catalogs
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Sequences table for document numbering
CREATE TABLE IF NOT EXISTS erp_sequences (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    domain      TEXT NOT NULL,
    prefix      TEXT NOT NULL DEFAULT '',
    next_value  BIGINT NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, domain, prefix)
);

ALTER TABLE erp_sequences ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_sequences
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Seed ERP permissions (pattern P5)
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.catalogs.read',  'Ver catalogos',      'Listar catalogos ERP',       'erp'),
    ('erp.catalogs.write', 'Gestionar catalogos', 'Crear/editar catalogos ERP', 'erp')
ON CONFLICT (id) DO NOTHING;

-- Grant catalog permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE category = 'erp'
ON CONFLICT DO NOTHING;
