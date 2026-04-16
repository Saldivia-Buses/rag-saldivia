-- 040_erp_manufacturing_units.up.sql
-- Unidades productivas: órdenes de trabajo de fabricación de carrocerías

CREATE TABLE erp_manufacturing_units (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             TEXT NOT NULL,
    work_order_number     INTEGER NOT NULL,
    chassis_serial        TEXT NOT NULL DEFAULT '',
    engine_number         TEXT NOT NULL DEFAULT '',
    chassis_brand_id      UUID REFERENCES erp_chassis_brands(id),
    chassis_model_id      UUID REFERENCES erp_chassis_models(id),
    carroceria_model_id   UUID REFERENCES erp_carroceria_models(id),
    customer_id           UUID REFERENCES erp_entities(id),
    entry_date            DATE,
    expected_completion   DATE,
    actual_completion     DATE,
    exit_date             DATE,
    tachograph_id         INTEGER,
    tachograph_serial     TEXT NOT NULL DEFAULT '',
    invoice_reference     TEXT,
    observations          TEXT NOT NULL DEFAULT '',
    status                TEXT NOT NULL DEFAULT 'pending'
                          CHECK (status IN ('pending','in_production','completed','delivered','returned')),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, work_order_number)
);
CREATE INDEX idx_mfg_units_status ON erp_manufacturing_units (tenant_id, status);
CREATE INDEX idx_mfg_units_entry  ON erp_manufacturing_units (entry_date);
ALTER TABLE erp_manufacturing_units ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_manufacturing_units
    USING (tenant_id = current_setting('app.tenant_id', true));
