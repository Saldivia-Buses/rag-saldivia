-- 041_erp_carroceria_bom.up.sql
-- Bill of Materials (BOM) por modelo de carrocería

CREATE TABLE erp_carroceria_bom (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            TEXT NOT NULL,
    carroceria_model_id  UUID NOT NULL REFERENCES erp_carroceria_models(id),
    article_id           UUID NOT NULL REFERENCES erp_articles(id),
    quantity             NUMERIC(10,2) NOT NULL CHECK (quantity > 0),
    unit_of_use          TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, carroceria_model_id, article_id)
);
CREATE INDEX idx_carroceria_bom_model ON erp_carroceria_bom (tenant_id, carroceria_model_id);
ALTER TABLE erp_carroceria_bom ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_carroceria_bom
    USING (tenant_id = current_setting('app.tenant_id', true));
