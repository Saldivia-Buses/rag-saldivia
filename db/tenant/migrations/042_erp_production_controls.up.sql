-- 042_erp_production_controls.up.sql
-- Controles de producción: causales de detención y registros de control por unidad/estación

-- Causals catalog for production stoppages and control events
CREATE TABLE erp_production_control_causals (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    code        TEXT NOT NULL,
    description TEXT NOT NULL,
    causal_type TEXT NOT NULL DEFAULT 'stoppage'
                CHECK (causal_type IN ('stoppage','rework','delay','quality','external')),
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);
ALTER TABLE erp_production_control_causals ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_control_causals
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Production control records per unit and station
CREATE TABLE erp_production_controls (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      TEXT NOT NULL,
    unit_id        UUID NOT NULL REFERENCES erp_manufacturing_units(id),
    station        TEXT NOT NULL,
    station_seq    INTEGER NOT NULL DEFAULT 0,
    responsible_id UUID REFERENCES erp_entities(id),
    planned_start  DATE,
    planned_end    DATE,
    actual_start   TIMESTAMPTZ,
    actual_end     TIMESTAMPTZ,
    status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','in_progress','completed','blocked','rework')),
    notes          TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, unit_id, station)
);
CREATE INDEX idx_prod_controls_unit   ON erp_production_controls (tenant_id, unit_id);
CREATE INDEX idx_prod_controls_status ON erp_production_controls (tenant_id, status);
ALTER TABLE erp_production_controls ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_controls
    USING (tenant_id = current_setting('app.tenant_id', true));
