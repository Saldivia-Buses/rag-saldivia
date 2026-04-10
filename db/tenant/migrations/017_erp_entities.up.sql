-- 017_erp_entities.up.sql
-- Plan 17 Phase 1: Unified entities (employees, customers, suppliers)
-- Replaces ~45 legacy tables: PERSONAL, REG_CUENTA, contacts, docs, relations

-- Entidad unificada
CREATE TABLE IF NOT EXISTS erp_entities (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    type             TEXT NOT NULL,              -- 'employee', 'customer', 'supplier'
    code             TEXT NOT NULL,              -- legajo, CUIT, codigo proveedor
    name             TEXT NOT NULL,
    encrypted_tax_id BYTEA,                     -- CUIT/CUIL (envelope encrypted, pattern P8)
    tax_id_hash      TEXT,                       -- SHA-256 hash for search without decrypting
    email            TEXT,
    phone            TEXT,
    address          JSONB DEFAULT '{}',         -- {street, city, province, zip, country}
    metadata         JSONB DEFAULT '{}',
    active           BOOLEAN NOT NULL DEFAULT true,
    deleted_at       TIMESTAMPTZ,                -- soft delete
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, type, code),
    CHECK (jsonb_typeof(address) = 'object'),
    CHECK (jsonb_typeof(metadata) = 'object')
);
CREATE INDEX idx_erp_entities_type ON erp_entities(tenant_id, type, active) WHERE deleted_at IS NULL;
CREATE INDEX idx_erp_entities_tax ON erp_entities(tenant_id, tax_id_hash) WHERE tax_id_hash IS NOT NULL;
CREATE INDEX idx_erp_entities_name ON erp_entities(tenant_id, name) WHERE deleted_at IS NULL;

-- Contactos de una entidad
CREATE TABLE IF NOT EXISTS erp_entity_contacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    entity_id   UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,              -- 'phone', 'email', 'address', 'bank_account'
    label       TEXT NOT NULL DEFAULT '',
    value       TEXT NOT NULL,
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_entity_contacts ON erp_entity_contacts(tenant_id, entity_id);

-- Documentos adjuntos
CREATE TABLE IF NOT EXISTS erp_entity_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    entity_id   UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    doc_type    TEXT NOT NULL,              -- 'certificate', 'contract', 'photo', 'dni'
    file_key    TEXT NOT NULL,              -- MinIO object key
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_entity_docs ON erp_entity_documents(tenant_id, entity_id);

-- Relaciones entre entidades
CREATE TABLE IF NOT EXISTS erp_entity_relations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    from_id     UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
    to_id       UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,              -- 'parent', 'spouse', 'representative', 'branch'
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, from_id, to_id, type)
);

-- Notas/eventos sobre una entidad
CREATE TABLE IF NOT EXISTS erp_entity_notes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    entity_id   UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,
    type        TEXT NOT NULL DEFAULT 'note', -- 'note', 'event', 'call', 'visit'
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_entity_notes ON erp_entity_notes(tenant_id, entity_id, created_at DESC);

-- RLS (pattern P1)
ALTER TABLE erp_entities ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_entities USING (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE erp_entity_contacts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_entity_contacts USING (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE erp_entity_documents ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_entity_documents USING (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE erp_entity_relations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_entity_relations USING (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE erp_entity_notes ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_entity_notes USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions (pattern P5)
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.entities.read',  'Ver entidades',      'Listar personal, clientes, proveedores', 'erp'),
    ('erp.entities.write', 'Gestionar entidades', 'Crear/editar entidades',                'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.entities.%'
ON CONFLICT DO NOTHING;
