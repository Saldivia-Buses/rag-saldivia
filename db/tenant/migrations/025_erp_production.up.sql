-- 025_erp_production.up.sql
-- Plan 17 Phase 8: Production & MRP
-- Replaces ~84 legacy tables: PROD_*, MRP_*, FSM*, CERT*, CHASIS, CARROCERIA, etc.

-- Centros productivos
CREATE TABLE IF NOT EXISTS erp_production_centers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, code)
);

-- Ordenes de produccion
CREATE TABLE IF NOT EXISTS erp_production_orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    number      TEXT NOT NULL,
    date        DATE NOT NULL,
    product_id  UUID NOT NULL REFERENCES erp_articles(id),
    center_id   UUID REFERENCES erp_production_centers(id),
    quantity    NUMERIC(14,4) NOT NULL CHECK (quantity > 0),
    status      TEXT NOT NULL DEFAULT 'planned' CHECK (status IN ('planned','in_progress','completed','cancelled')),
    priority    INT NOT NULL DEFAULT 0,
    order_id    UUID REFERENCES erp_orders(id),
    start_date  DATE,
    end_date    DATE,
    user_id     TEXT NOT NULL,
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, number)
);
CREATE INDEX idx_erp_prod_orders_date ON erp_production_orders(tenant_id, date DESC);
CREATE INDEX idx_erp_prod_orders_status ON erp_production_orders(tenant_id, status) WHERE status != 'cancelled';

-- Materiales requeridos (explode del BOM)
CREATE TABLE IF NOT EXISTS erp_production_materials (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    order_id     UUID NOT NULL REFERENCES erp_production_orders(id) ON DELETE CASCADE,
    article_id   UUID NOT NULL REFERENCES erp_articles(id),
    required_qty NUMERIC(14,4) NOT NULL CHECK (required_qty > 0),
    consumed_qty NUMERIC(14,4) NOT NULL DEFAULT 0,
    warehouse_id UUID REFERENCES erp_warehouses(id)
);
CREATE INDEX idx_erp_prod_materials ON erp_production_materials(tenant_id, order_id);

-- Pasos de produccion (trazabilidad)
CREATE TABLE IF NOT EXISTS erp_production_steps (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    order_id     UUID NOT NULL REFERENCES erp_production_orders(id) ON DELETE CASCADE,
    step_name    TEXT NOT NULL,
    sort_order   INT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','completed','skipped')),
    assigned_to  UUID REFERENCES erp_entities(id),
    started_at   TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    notes        TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_erp_prod_steps ON erp_production_steps(tenant_id, order_id);

-- Inspecciones de calidad en produccion
CREATE TABLE IF NOT EXISTS erp_production_inspections (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    order_id     UUID NOT NULL REFERENCES erp_production_orders(id),
    step_id      UUID REFERENCES erp_production_steps(id),
    inspector_id UUID REFERENCES erp_entities(id),
    result       TEXT NOT NULL CHECK (result IN ('pass', 'fail', 'rework')),
    observations TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_prod_inspections ON erp_production_inspections(tenant_id, order_id);

-- Unidades/Vehiculos (chasis + carroceria unificado)
CREATE TABLE IF NOT EXISTS erp_units (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           TEXT NOT NULL,
    chassis_number      TEXT NOT NULL,
    internal_number     TEXT,
    model               TEXT,
    customer_id         UUID REFERENCES erp_entities(id),
    order_id            UUID REFERENCES erp_orders(id),
    production_order_id UUID REFERENCES erp_production_orders(id),
    patent              TEXT,
    status              TEXT NOT NULL DEFAULT 'in_production' CHECK (status IN ('in_production','ready','delivered')),
    engine_brand        TEXT,
    body_style          TEXT,
    seat_count          INT,
    year                INT,
    metadata            JSONB NOT NULL DEFAULT '{}',
    delivered_at        DATE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, chassis_number)
);
CREATE INDEX idx_erp_units_status ON erp_units(tenant_id, status);

-- Fotos de unidades
CREATE TABLE IF NOT EXISTS erp_unit_photos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    unit_id     UUID NOT NULL REFERENCES erp_units(id) ON DELETE CASCADE,
    photo_type  TEXT NOT NULL CHECK (photo_type IN ('homologation','delivery','production','patent','general')),
    file_key    TEXT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_unit_photos ON erp_unit_photos(tenant_id, unit_id);

-- RLS
ALTER TABLE erp_production_centers ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_centers USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_production_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_orders USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_production_materials ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_materials USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_production_steps ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_steps USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_production_inspections ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_inspections USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_units ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_units USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_unit_photos ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_unit_photos USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.production.read',  'Ver produccion',      'Consultar ordenes, unidades, trazabilidad', 'erp'),
    ('erp.production.write', 'Gestionar produccion', 'Crear ordenes, registrar avance',          'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.production.%'
ON CONFLICT DO NOTHING;
