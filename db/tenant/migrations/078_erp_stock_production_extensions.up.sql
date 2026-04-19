-- 078_erp_stock_production_extensions.up.sql
-- Phase 1 §Data migration — Pareto tail Grupo B (stock / production
-- extensions). Three tables, ~176 K rows live combined:
--
--   STK_COSTO_REPOSICION_HIST → erp_article_replacement_cost_history
--                                                        (109,123 rows)
--   ACCESORIOS_COCHE          → erp_unit_accessories     ( 37,909 rows)
--   COTIZOPMOVIM              → erp_quotation_section_items( 28,573 rows)
--
-- FK resolution leans on earlier phases:
--   - entity cache (Phase 2) for `regcuenta_id`
--   - currency domain (catalogs) for `moneda_id`
--   - stock STK_ARTICULOS default-subsystem lookup for `artcod`
--   - production CHASIS for `nrofab`
--   - sales COTIZACION + PEDCOTIZ for `cotizacion_id` + `ficha_id`
--   - productos PRODUCTO_SECCION for `prdseccion_id`
-- None add new hooks. `costoreposicion_id` (parent STK_COSTO_REPOSICION)
-- is currently surfaced only as metadata-enriched JSON on erp_articles;
-- we preserve the raw integer in `replacement_cost_legacy_id`.

-- ---------------------------------------------------------------------
-- erp_article_replacement_cost_history — STK_COSTO_REPOSICION_HIST (109 K)
-- Rolling log of supplier replacement-cost changes. Extends the 075
-- erp_article_cost_history family with currency + origin + incoterm +
-- import-cost breakdown.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_article_replacement_cost_history (
    id                           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                    TEXT NOT NULL,
    legacy_id                    BIGINT NOT NULL,                -- id_costoreposicion_hist
    replacement_cost_legacy_id   INTEGER NOT NULL DEFAULT 0,     -- costoreposicion_id raw
    supplier_entity_id           UUID REFERENCES erp_entities(id),
    supplier_legacy_id           INTEGER NOT NULL DEFAULT 0,     -- regcuenta_id raw
    currency_id                  UUID,                           -- resolved via GEN_MONEDAS mapper
    currency_legacy_id           INTEGER NOT NULL DEFAULT 0,     -- moneda_id raw
    exchange_rate                NUMERIC(14,4) NOT NULL DEFAULT 0, -- cotizacion
    supplier_cost                NUMERIC(12,4) NOT NULL DEFAULT 0, -- costo_proveedor
    origin                       TEXT NOT NULL DEFAULT '',       -- origen (1-char: N / I)
    incoterm                     TEXT NOT NULL DEFAULT '',       -- incoterm_id varchar(3)
    import_expenses              NUMERIC(10,2) NOT NULL DEFAULT 0, -- gasto_importacion
    local_freight                NUMERIC(10,2) NOT NULL DEFAULT 0, -- flete_local_ars
    modified_at                  TIMESTAMPTZ,                    -- modificado
    discount_1                   NUMERIC(10,2) NOT NULL DEFAULT 0,
    discount_2                   NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_article_replacement_cost_history_parent
    ON erp_article_replacement_cost_history(tenant_id, replacement_cost_legacy_id)
    WHERE replacement_cost_legacy_id > 0;
CREATE INDEX idx_erp_article_replacement_cost_history_supplier
    ON erp_article_replacement_cost_history(tenant_id, supplier_entity_id)
    WHERE supplier_entity_id IS NOT NULL;
CREATE INDEX idx_erp_article_replacement_cost_history_modified
    ON erp_article_replacement_cost_history(tenant_id, modified_at DESC)
    WHERE modified_at IS NOT NULL;

ALTER TABLE erp_article_replacement_cost_history ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_article_replacement_cost_history
    USING (tenant_id = current_setting('app.tenant_id', true));

-- ---------------------------------------------------------------------
-- erp_unit_accessories — ACCESORIOS_COCHE (38 K)
-- Per-unit accessory lines: article + qty + price for a specific
-- vehicle unit (nrofab), an order (PEDCOTIZ = ficha) and a quotation
-- (COTIZACION). Multi-domain bridge table so it carries FKs to all
-- four parent domains.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_unit_accessories (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                  TEXT NOT NULL,
    legacy_id                  BIGINT NOT NULL,                -- id_accesorio
    unit_id                    UUID,                           -- resolved via CHASIS (nrofab)
    unit_legacy_id             INTEGER NOT NULL DEFAULT 0,     -- nrofab raw
    article_code               TEXT NOT NULL DEFAULT '',       -- artcod varchar(10)
    article_id                 UUID REFERENCES erp_articles(id),
    article_description        TEXT NOT NULL DEFAULT '',       -- artdes longtext
    accessory_date             DATE,                           -- fecha (nullable for zero)
    quotation_id               UUID,                           -- resolved via COTIZACION
    quotation_legacy_id        INTEGER NOT NULL DEFAULT 0,     -- cotizacion_id raw
    order_id                   UUID,                           -- resolved via PEDCOTIZ
    order_legacy_id            INTEGER NOT NULL DEFAULT 0,     -- ficha_id raw
    status                     INTEGER NOT NULL DEFAULT 0,     -- estado
    additional_price           NUMERIC(10,2),                  -- precio_adicional (nullable)
    quantity                   INTEGER NOT NULL DEFAULT 0,
    approved_at                DATE,                           -- aprobado (nullable for zero)
    unit_price                 NUMERIC(10,2) NOT NULL DEFAULT 0,
    product_section_id         UUID,                           -- resolved via PRODUCTO_SECCION
    product_section_legacy_id  INTEGER NOT NULL DEFAULT 0,     -- prdseccion_id raw
    observations               TEXT NOT NULL DEFAULT '',       -- observaciones longtext
    show_on_fv                 SMALLINT NOT NULL DEFAULT 1,    -- muestra_fv (0/1)
    show_on_ft                 SMALLINT NOT NULL DEFAULT 1,    -- muestra_ft (0/1)
    accessory_state_legacy_id  INTEGER NOT NULL DEFAULT 0,     -- fc_estado_acc_id raw
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_unit_accessories_unit
    ON erp_unit_accessories(tenant_id, unit_id)
    WHERE unit_id IS NOT NULL;
CREATE INDEX idx_erp_unit_accessories_article
    ON erp_unit_accessories(tenant_id, article_id)
    WHERE article_id IS NOT NULL;
CREATE INDEX idx_erp_unit_accessories_order
    ON erp_unit_accessories(tenant_id, order_id)
    WHERE order_id IS NOT NULL;
CREATE INDEX idx_erp_unit_accessories_quotation
    ON erp_unit_accessories(tenant_id, quotation_id)
    WHERE quotation_id IS NOT NULL;
CREATE INDEX idx_erp_unit_accessories_date
    ON erp_unit_accessories(tenant_id, accessory_date DESC)
    WHERE accessory_date IS NOT NULL;

ALTER TABLE erp_unit_accessories ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_unit_accessories
    USING (tenant_id = current_setting('app.tenant_id', true));

-- ---------------------------------------------------------------------
-- erp_quotation_section_items — COTIZOPMOVIM (29 K)
-- "OPCIONES POR COTIZACION" — free-text option lines per quotation
-- section. 4-column shape: FK idCotiz + idSeccion discriminator +
-- descripcion + idMovim PK.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_quotation_section_items (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              TEXT NOT NULL,
    legacy_id              BIGINT NOT NULL,                -- idMovim
    quotation_id           UUID REFERENCES erp_quotations(id),
    quotation_legacy_id    INTEGER NOT NULL DEFAULT 0,     -- idCotiz raw
    section_legacy_id      INTEGER NOT NULL DEFAULT 0,     -- idSeccion raw (no target table)
    description            TEXT NOT NULL DEFAULT '',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_quotation_section_items_quotation
    ON erp_quotation_section_items(tenant_id, quotation_id)
    WHERE quotation_id IS NOT NULL;
CREATE INDEX idx_erp_quotation_section_items_section
    ON erp_quotation_section_items(tenant_id, section_legacy_id);

ALTER TABLE erp_quotation_section_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_quotation_section_items
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Reuses existing permission surfaces:
--   - erp.stock.read / write for replacement-cost history.
--   - erp.production.read / write for unit accessories.
--   - erp.sales.read / write for quotation section items.
-- No new permissions added.
