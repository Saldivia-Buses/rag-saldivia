-- 075_erp_production_extensions.up.sql
-- Phase 1 §Data migration: Pareto #7 (PROD_CONTROL_HOMOLOG) + Pareto #8
-- (STK_COSTO_HIST). Two small simple tables joining existing erp_*
-- surfaces: production inspections ↔ homologations, and monthly cost
-- history per article. ~507 K rows combined.
--
-- Histrix context:
--   - PROD_CONTROL_HOMOLOG (403,028 rows live — scrape was 105,683,
--     +282 % growth) is a 3-column join table linking PROD_CONTROLES
--     (already migrated to erp_production_inspections in Phase 7/8) to
--     HOMOLOGMOD (erp_homologations, 2.0.8). Live Histrix queries show
--     zero orphans on both FKs — every row resolves cleanly.
--   - STK_COSTO_HIST (103,799 rows live — scrape 95,217) is monthly
--     cost snapshots per article. Composite PK (articulo_id, anio,
--     mes). Already consumed by the metadata_enricher
--     articleCostHistory spec for JSONB attach on erp_articles; this
--     migration adds the native relational shape for UIs that want
--     structured history without parsing JSON.

-- ---------------------------------------------------------------------
-- erp_production_inspection_homologations (PROD_CONTROL_HOMOLOG, 403 K)
-- Join table: inspection templates × homologated vehicle models.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_production_inspection_homologations (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              TEXT NOT NULL,
    legacy_id              BIGINT NOT NULL,               -- id_controlhomolog
    inspection_id          UUID REFERENCES erp_production_inspections(id),
    inspection_legacy_id   INTEGER NOT NULL DEFAULT 0,    -- prodcontrol_id raw
    homologation_id        UUID REFERENCES erp_homologations(id),
    homologation_legacy_id INTEGER NOT NULL DEFAULT 0,    -- homologacion_id raw
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_pi_homolog_inspection
    ON erp_production_inspection_homologations(tenant_id, inspection_id)
    WHERE inspection_id IS NOT NULL;
CREATE INDEX idx_erp_pi_homolog_homologation
    ON erp_production_inspection_homologations(tenant_id, homologation_id)
    WHERE homologation_id IS NOT NULL;

ALTER TABLE erp_production_inspection_homologations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_inspection_homologations
    USING (tenant_id = current_setting('app.tenant_id', true));

-- ---------------------------------------------------------------------
-- erp_article_cost_history (STK_COSTO_HIST, 104 K)
-- Monthly cost snapshots per article. Composite natural key
-- (articulo_id, anio, mes) in Histrix — we keep that uniqueness via a
-- deterministic legacy_id hash so migrations stay idempotent.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_article_cost_history (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    legacy_id        BIGINT NOT NULL,           -- hash(articulo_id|anio|mes)
    article_code     TEXT NOT NULL DEFAULT '',  -- articulo_id varchar(12)
    article_id       UUID REFERENCES erp_articles(id),
    year             INTEGER NOT NULL DEFAULT 0,-- anio_hist
    month            SMALLINT NOT NULL DEFAULT 0, -- mes_hist
    cost             NUMERIC(14,4) NOT NULL DEFAULT 0, -- costo_hist
    period_code      TEXT NOT NULL DEFAULT '',  -- periodo_hist (YYYYMM)
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_article_cost_history_article_code
    ON erp_article_cost_history(tenant_id, article_code)
    WHERE article_code <> '';
CREATE INDEX idx_erp_article_cost_history_article_id
    ON erp_article_cost_history(tenant_id, article_id)
    WHERE article_id IS NOT NULL;
CREATE INDEX idx_erp_article_cost_history_period
    ON erp_article_cost_history(tenant_id, period_code);

ALTER TABLE erp_article_cost_history ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_article_cost_history
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Both tables use existing permission surfaces:
--   - erp.production.read / write for inspection-homolog links (Phase 7/8).
--   - erp.stock.read / write for article cost history (Phase 4).
-- No new permissions added.
