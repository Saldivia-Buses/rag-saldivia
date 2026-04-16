-- 039_erp_manufacturing_catalogs.up.sql
-- Catálogos base de fabricación: marcas y modelos de chasis, modelos de carrocería

-- Chassis brands catalog
CREATE TABLE erp_chassis_brands (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  TEXT NOT NULL,
    code       TEXT NOT NULL,
    name       TEXT NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);
ALTER TABLE erp_chassis_brands ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_chassis_brands
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Chassis models catalog
CREATE TABLE erp_chassis_models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    brand_id        UUID NOT NULL REFERENCES erp_chassis_brands(id),
    model_code      TEXT NOT NULL,
    description     TEXT NOT NULL,
    traction        TEXT NOT NULL DEFAULT 'trasera'
                    CHECK (traction IN ('delantera','trasera','media')),
    engine_location TEXT NOT NULL DEFAULT 'rear'
                    CHECK (engine_location IN ('front','rear','mid')),
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, brand_id, model_code)
);
ALTER TABLE erp_chassis_models ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_chassis_models
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Carrocería (body) models catalog
CREATE TABLE erp_carroceria_models (
    id                               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                        TEXT NOT NULL,
    code                             TEXT NOT NULL,
    model_code                       TEXT NOT NULL,
    description                      TEXT NOT NULL,
    abbreviation                     TEXT NOT NULL DEFAULT '',
    double_deck                      BOOLEAN NOT NULL DEFAULT false,
    axle_weight_pct                  NUMERIC(10,9) NOT NULL DEFAULT 0,
    productive_hours_per_station     INTERVAL,
    active                           BOOLEAN NOT NULL DEFAULT true,
    tech_sheet_image                 TEXT NOT NULL DEFAULT '',
    created_at                       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);
ALTER TABLE erp_carroceria_models ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_carroceria_models
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.manufacturing.read',    'Ver fabricación',           'Consultar órdenes, controles, certificados', 'erp'),
    ('erp.manufacturing.write',   'Gestionar fabricación',     'Crear y editar unidades, BOM, controles',    'erp'),
    ('erp.manufacturing.control', 'Control producción',        'Registrar ejecuciones y retrabajos',         'erp'),
    ('erp.manufacturing.certify', 'Certificar unidades',       'Emitir y gestionar certificados CNRT',       'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.manufacturing.%'
ON CONFLICT DO NOTHING;
