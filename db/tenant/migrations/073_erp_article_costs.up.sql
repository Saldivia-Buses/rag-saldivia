-- 073_erp_article_costs.up.sql
-- Phase 1 §Data migration: STKINSPR (Pareto #5, 189,863 rows live).
--
-- Histrix context:
--   - STKINSPR (most likely "STock INSumo PRecios" historically — the
--     name is opaque) is a per-supplier cost ledger: one row per
--     (artcod, ctacod) cost snapshot. artcos = cost value, fecult = last
--     update date, ctacod = supplier entity. Used by the cost-update
--     screens (stock/costos/, costos/, evolucion_costos), statistics
--     (estadisticas/evolutivo_costo), and written by invoice-import
--     (remitos/factura_stkinspr_ingmov) + the recalc flag for periodic
--     re-cost runs.
--   - Single subsystem in live data (siscod='01' on 100 % of rows)
--     but kept as a preserved column for forensic / future multi-sub.
--   - Sister table STK_COSTO_HIST (Pareto #8, 95 K rows) is the cost
--     *history*; STKINSPR holds the most recent per-(article, supplier).
--     Both ride the same erp_articles FK.
--
-- Shape from live Histrix:
--   idCosto INT AI PRI, artcod VARCHAR(20) MUL, siscod VARCHAR(2) MUL,
--   artcos DECIMAL(10,3), artpor__1/__2/__3 DECIMAL(6,2), artpro
--   VARCHAR(20) (supplier's own article code), ctacod MEDIUMINT (ctacod
--   → REG_CUENTA via id_regcuenta or nro_cuenta), fecfac DATE (~all zero
--   — legacy), fecult DATE MUL (business "as-of"), movnro/movnpv INT +
--   movfec DATE (movement reference), recalc INT (periodic re-cost flag).

CREATE TABLE IF NOT EXISTS erp_article_costs (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              TEXT NOT NULL,
    legacy_id              BIGINT NOT NULL,                 -- idCosto
    article_code           TEXT NOT NULL DEFAULT '',        -- artcod
    article_id             UUID REFERENCES erp_articles(id), -- nullable, stock default-sub resolve
    subsystem_code         TEXT NOT NULL DEFAULT '',        -- siscod
    cost                   NUMERIC(14,3) NOT NULL DEFAULT 0,-- artcos
    percentage_1           NUMERIC(8,2)  NOT NULL DEFAULT 0,-- artpor__1
    percentage_2           NUMERIC(8,2)  NOT NULL DEFAULT 0,-- artpor__2
    percentage_3           NUMERIC(8,2)  NOT NULL DEFAULT 0,-- artpor__3
    supplier_article_code  TEXT NOT NULL DEFAULT '',        -- artpro (supplier's SKU)
    supplier_code          INTEGER NOT NULL DEFAULT 0,      -- ctacod (raw code, preserved)
    supplier_entity_id     UUID REFERENCES erp_entities(id),-- nullable, via ResolveEntityFlexible
    invoice_date           DATE,                            -- fecfac (~99.9 % zero)
    last_update_date       DATE,                            -- fecult (business as-of)
    movement_no            INTEGER NOT NULL DEFAULT 0,      -- movnro
    movement_post          INTEGER NOT NULL DEFAULT 0,      -- movnpv
    movement_date          DATE,                            -- movfec
    recalc_flag            INTEGER NOT NULL DEFAULT 0,      -- recalc
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);

CREATE INDEX idx_erp_article_costs_article_code
    ON erp_article_costs(tenant_id, article_code);
CREATE INDEX idx_erp_article_costs_article_id
    ON erp_article_costs(tenant_id, article_id)
    WHERE article_id IS NOT NULL;
CREATE INDEX idx_erp_article_costs_supplier
    ON erp_article_costs(tenant_id, supplier_code)
    WHERE supplier_code > 0;
CREATE INDEX idx_erp_article_costs_supplier_entity
    ON erp_article_costs(tenant_id, supplier_entity_id)
    WHERE supplier_entity_id IS NOT NULL;
CREATE INDEX idx_erp_article_costs_last_update
    ON erp_article_costs(tenant_id, last_update_date DESC)
    WHERE last_update_date IS NOT NULL;

ALTER TABLE erp_article_costs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_article_costs
    USING (tenant_id = current_setting('app.tenant_id', true));

-- No new permissions: reuses erp.stock.read / erp.stock.write (migration
-- 017_erp_stock). Cost tracking is the same functional scope as stock
-- from the user's point of view — costs are stock attributes.
