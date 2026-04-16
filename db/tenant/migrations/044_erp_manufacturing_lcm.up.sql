-- 044_erp_manufacturing_lcm.up.sql
-- LCM (Libro de Control de Materiales): consumo de materiales por unidad productiva

-- LCM header: material consumption record per manufacturing unit
CREATE TABLE erp_manufacturing_lcm (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    unit_id         UUID NOT NULL REFERENCES erp_manufacturing_units(id),
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    issued_by       UUID REFERENCES erp_entities(id),
    warehouse_id    UUID REFERENCES erp_warehouses(id),
    reference       TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','issued','closed','cancelled')),
    notes           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_lcm_unit   ON erp_manufacturing_lcm (tenant_id, unit_id);
CREATE INDEX idx_lcm_status ON erp_manufacturing_lcm (tenant_id, status);
ALTER TABLE erp_manufacturing_lcm ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_manufacturing_lcm
    USING (tenant_id = current_setting('app.tenant_id', true));

-- LCM line items: articles consumed per LCM
CREATE TABLE erp_manufacturing_lcm_models (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    lcm_id       UUID NOT NULL REFERENCES erp_manufacturing_lcm(id) ON DELETE CASCADE,
    article_id   UUID NOT NULL REFERENCES erp_articles(id),
    bom_qty      NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (bom_qty >= 0),
    issued_qty   NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (issued_qty >= 0),
    returned_qty NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (returned_qty >= 0),
    unit_cost    NUMERIC(15,4) NOT NULL DEFAULT 0,
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, lcm_id, article_id)
);
CREATE INDEX idx_lcm_models_lcm     ON erp_manufacturing_lcm_models (tenant_id, lcm_id);
CREATE INDEX idx_lcm_models_article ON erp_manufacturing_lcm_models (tenant_id, article_id);
ALTER TABLE erp_manufacturing_lcm_models ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_manufacturing_lcm_models
    USING (tenant_id = current_setting('app.tenant_id', true));
