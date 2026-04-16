-- 028_erp_maintenance.up.sql
-- Plan 17 Phase 11: Maintenance
-- Replaces ~33 legacy tables: MANT_*, SERVICE_*, GPS_*, TACOGRAFOS, etc.

CREATE TABLE IF NOT EXISTS erp_maintenance_assets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    asset_type  TEXT NOT NULL CHECK (asset_type IN ('vehicle','machine','tool','facility')),
    unit_id     UUID REFERENCES erp_units(id),
    location    TEXT NOT NULL DEFAULT '',
    metadata    JSONB NOT NULL DEFAULT '{}',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_erp_maint_assets ON erp_maintenance_assets(tenant_id, active);

CREATE TABLE IF NOT EXISTS erp_maintenance_plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    asset_id        UUID NOT NULL REFERENCES erp_maintenance_assets(id),
    name            TEXT NOT NULL,
    frequency_days  INT,
    frequency_km    INT,
    frequency_hours INT,
    last_done       DATE,
    next_due        DATE,
    active          BOOLEAN NOT NULL DEFAULT true
);
CREATE INDEX idx_erp_maint_plans ON erp_maintenance_plans(tenant_id, asset_id);

CREATE TABLE IF NOT EXISTS erp_work_orders (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    number       TEXT NOT NULL,
    asset_id     UUID NOT NULL REFERENCES erp_maintenance_assets(id),
    plan_id      UUID REFERENCES erp_maintenance_plans(id),
    date         DATE NOT NULL,
    work_type    TEXT NOT NULL CHECK (work_type IN ('preventive','corrective','inspection')),
    description  TEXT NOT NULL,
    assigned_to  UUID REFERENCES erp_entities(id),
    status       TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','in_progress','completed','cancelled')),
    priority     TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low','normal','high','urgent')),
    completed_at TIMESTAMPTZ,
    user_id      TEXT NOT NULL,
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, number)
);
CREATE INDEX idx_erp_wo_status ON erp_work_orders(tenant_id, status) WHERE status NOT IN ('completed', 'cancelled');
CREATE INDEX idx_erp_wo_asset ON erp_work_orders(tenant_id, asset_id);

CREATE TABLE IF NOT EXISTS erp_work_order_parts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    work_order_id UUID NOT NULL REFERENCES erp_work_orders(id) ON DELETE CASCADE,
    article_id    UUID NOT NULL REFERENCES erp_articles(id),
    quantity      NUMERIC(14,4) NOT NULL CHECK (quantity > 0)
);
CREATE INDEX idx_erp_wo_parts ON erp_work_order_parts(tenant_id, work_order_id);

CREATE TABLE IF NOT EXISTS erp_fuel_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    asset_id    UUID NOT NULL REFERENCES erp_maintenance_assets(id),
    date        DATE NOT NULL,
    liters      NUMERIC(10,2) NOT NULL CHECK (liters > 0),
    km_reading  INT,
    cost        NUMERIC(14,2),
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_fuel ON erp_fuel_logs(tenant_id, asset_id, date DESC);

-- RLS
ALTER TABLE erp_maintenance_assets ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_maintenance_assets USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_maintenance_plans ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_maintenance_plans USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_work_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_work_orders USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_work_order_parts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_work_order_parts USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_fuel_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_fuel_logs USING (tenant_id = current_setting('app.tenant_id', true));

INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.maintenance.read',  'Ver mantenimiento',      'Consultar equipos, OT, planes', 'erp'),
    ('erp.maintenance.write', 'Gestionar mantenimiento', 'Crear OT, registrar combustible','erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.maintenance.%'
ON CONFLICT DO NOTHING;
