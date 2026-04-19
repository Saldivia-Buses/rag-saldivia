-- 080_erp_stock_cost_movements.up.sql
-- Phase 1 §Data migration — STK_COSTOS (Pareto #21 of the post-2.0.10
-- gap, 15,066 rows live / scrape). Priced stock-movement ledger: one
-- row per stock movement with the computed costs (precio_costo,
-- precio_venta, precio_total, precio_promedio) alongside all the
-- domain FKs (article, entity, deposit, sector, family, rubro, list,
-- concept, unit, reference invoice/movement/order/cash).
--
-- Consumed by 15 live XML-forms: presup/* (budgeting), estadisticas/
-- evolutivo_costo (cost evolution charts), costos/qry/aumento_costo4,
-- produccion/linea/presup_cons_exp + listado_expandido, stock_local/
-- stkinmov_ingresos, plus the VISTA_STK_COSTOS derived view.
--
-- Not waivable — closing this with migration `080` finishes the 2.0.11
-- residual and leaves the business-data gap at ZERO tables.

CREATE TABLE IF NOT EXISTS erp_stock_cost_movements (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                  TEXT NOT NULL,
    legacy_id                  BIGINT NOT NULL,                 -- id_stkmovimiento

    -- Article (resolved via STK_ARTICULOS default-subsystem lookup)
    article_code               TEXT NOT NULL DEFAULT '',        -- stkarticulo_id raw
    article_id                 UUID REFERENCES erp_articles(id),

    -- Entity (resolved via ResolveEntityFlexible)
    entity_legacy_id           INTEGER NOT NULL DEFAULT 0,      -- regcuenta_id raw
    entity_id                  UUID REFERENCES erp_entities(id),
    account_legacy_code        INTEGER NOT NULL DEFAULT 0,      -- ctacod raw

    -- Stock hierarchy (kept raw — secondary catalogs not first-class yet)
    deposit_legacy_id          INTEGER NOT NULL DEFAULT 0,      -- stkdeposito_id
    sector_legacy_id           INTEGER NOT NULL DEFAULT 0,      -- stksector_id
    family_legacy_id           INTEGER NOT NULL DEFAULT 0,      -- stkfamilia_id
    rubro_legacy_id            INTEGER NOT NULL DEFAULT 0,      -- stkrubro_id
    list_legacy_id             INTEGER NOT NULL DEFAULT 0,      -- stklista_id
    concept_legacy_id          INTEGER NOT NULL DEFAULT 0,      -- stkconcepto_id
    unit_legacy_id             INTEGER NOT NULL DEFAULT 0,      -- unidad_id
    subsystem_code             TEXT NOT NULL DEFAULT '',        -- subsistema_id

    -- Dates
    movement_date              DATE,                            -- fecha_movimiento
    registered_date            DATE,                            -- alta_movimiento
    invoice_date               DATE,                            -- facfec

    -- Identifiers
    station                    INTEGER NOT NULL DEFAULT 0,      -- puesto_movimiento
    movement_no                INTEGER NOT NULL DEFAULT 0,      -- numero_movimiento
    movement_order             INTEGER NOT NULL DEFAULT 0,      -- orden_movimiento
    reference                  TEXT NOT NULL DEFAULT '',        -- referencia
    barcode                    TEXT NOT NULL DEFAULT '',
    description                TEXT NOT NULL DEFAULT '',        -- descripcion
    title_code                 TEXT NOT NULL DEFAULT '',        -- titcod
    operator_class             TEXT NOT NULL DEFAULT '',        -- opecla
    operator_code              INTEGER NOT NULL DEFAULT 0,      -- opecod
    register_min               INTEGER NOT NULL DEFAULT 0,      -- regmin
    branch_code                INTEGER NOT NULL DEFAULT 0,      -- succod
    unit_type                  SMALLINT NOT NULL DEFAULT 0,     -- tipuni

    -- Quantities + prices
    quantity                   NUMERIC(10,3) NOT NULL DEFAULT 0,
    cost_price                 NUMERIC(10,3) NOT NULL DEFAULT 0,
    sale_price                 NUMERIC(10,3) NOT NULL DEFAULT 0,
    total_price                NUMERIC(10,3) NOT NULL DEFAULT 0,
    average_price              NUMERIC(10,3) NOT NULL DEFAULT 0,
    bonus_pct                  NUMERIC(5,2) NOT NULL DEFAULT 0,   -- movbon
    purchase_amount            NUMERIC(9,2) NOT NULL DEFAULT 0,   -- movcps
    pending_amount             NUMERIC(12,2) NOT NULL DEFAULT 0,  -- movpen
    peso_amount                NUMERIC(12,2) NOT NULL DEFAULT 0,  -- movpes
    usage_amount               NUMERIC(9,2) NOT NULL DEFAULT 0,   -- movuso
    sale_ref                   INTEGER NOT NULL DEFAULT 0,        -- movven
    chassis_no                 INTEGER NOT NULL DEFAULT 0,        -- nrocha
    order_cps_no               INTEGER NOT NULL DEFAULT 0,        -- ocpnro

    -- Cross-domain FKs (invoice resolved; others preserved raw)
    invoice_id                 UUID REFERENCES erp_invoices(id),
    invoice_legacy_id          INTEGER NOT NULL DEFAULT 0,      -- regmovimiento_id raw
    invoice_line_legacy_id     INTEGER NOT NULL DEFAULT 0,      -- regdetalle_id raw
    cps_movement_legacy_id     INTEGER NOT NULL DEFAULT 0,      -- cpsmovimiento_id
    cps_detail_legacy_id       INTEGER NOT NULL DEFAULT 0,      -- cpsdetalle_id
    order_detail_legacy_id     INTEGER NOT NULL DEFAULT 0,      -- ordendetalle_id
    cash_movement_legacy_id    INTEGER NOT NULL DEFAULT 0,      -- cajmovimiento_id
    order_legacy_id            INTEGER NOT NULL DEFAULT 0,      -- pedido_id
    user_legacy_id             INTEGER NOT NULL DEFAULT 0,      -- user_id

    created_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_stock_cost_movements_article
    ON erp_stock_cost_movements(tenant_id, article_id)
    WHERE article_id IS NOT NULL;
CREATE INDEX idx_erp_stock_cost_movements_article_code
    ON erp_stock_cost_movements(tenant_id, article_code)
    WHERE article_code <> '';
CREATE INDEX idx_erp_stock_cost_movements_entity
    ON erp_stock_cost_movements(tenant_id, entity_id)
    WHERE entity_id IS NOT NULL;
CREATE INDEX idx_erp_stock_cost_movements_invoice
    ON erp_stock_cost_movements(tenant_id, invoice_id)
    WHERE invoice_id IS NOT NULL;
CREATE INDEX idx_erp_stock_cost_movements_date
    ON erp_stock_cost_movements(tenant_id, movement_date DESC)
    WHERE movement_date IS NOT NULL;
CREATE INDEX idx_erp_stock_cost_movements_deposit
    ON erp_stock_cost_movements(tenant_id, deposit_legacy_id);
CREATE INDEX idx_erp_stock_cost_movements_chassis
    ON erp_stock_cost_movements(tenant_id, chassis_no)
    WHERE chassis_no > 0;

ALTER TABLE erp_stock_cost_movements ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_stock_cost_movements
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Reuses erp.stock.read / erp.stock.write.
-- No new permissions added.
