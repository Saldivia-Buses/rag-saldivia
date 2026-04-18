-- 069_erp_homologations.up.sql
-- Phase 1 §Data migration: HOMOLOGMOD + STK_ARTICULO_PROCESO_HIST + STK_ARTICULO_PROCESO_HIST_DETALLE
-- Together these close 42.7 % of the Phase 1 §Data migration row-volume gap
-- (STK_ARTICULO_PROCESO_HIST_DETALLE alone = 2.6 M rows, #1 in the Pareto).
--
-- Domain: production (UX lives in .intranet-scrape/xml-forms/produccion/linea/ —
-- the STK_ prefix on the Histrix tables is misleading).
-- Histrix title: "HOMOLOGACION POR MODELO" (materiales_proc_art_exp_qry.xml).

-- Vehicle model homologations (HOMOLOGMOD — 585 rows).
-- Minimal shape — keeps the identity + the columns the XML-form queries most.
-- Extra Histrix columns (chassis axle weights, VIN, industry filings, etc.) are
-- left for a future extension migration once the Phase 1 UI needs them.
CREATE TABLE IF NOT EXISTS erp_homologations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    plano           TEXT NOT NULL DEFAULT '',
    expte           TEXT NOT NULL DEFAULT '',
    dispos          TEXT NOT NULL DEFAULT '',
    fecha_aprob     DATE,
    fecha_vto       DATE,
    seats           INT NOT NULL DEFAULT 0,
    seats_lower     INT NOT NULL DEFAULT 0,
    weight_tare     NUMERIC(10,2) NOT NULL DEFAULT 0,
    weight_gross    NUMERIC(10,2) NOT NULL DEFAULT 0,
    vin             TEXT NOT NULL DEFAULT '',
    commercial_code TEXT NOT NULL DEFAULT '',
    commercial_desc TEXT NOT NULL DEFAULT '',
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_homologations_active ON erp_homologations(tenant_id, active) WHERE active;
CREATE INDEX idx_erp_homologations_commercial ON erp_homologations(tenant_id, commercial_code) WHERE commercial_code <> '';

-- Process/cost revisions per homologation (STK_ARTICULO_PROCESO_HIST — 1,173 rows).
-- Each row is a dated snapshot of the process + article cost structure for one homologation.
CREATE TABLE IF NOT EXISTS erp_homologation_revisions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    homologation_id UUID NOT NULL REFERENCES erp_homologations(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    notes           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_homologation_revisions ON erp_homologation_revisions(tenant_id, homologation_id, date DESC);

-- Detail lines per revision (STK_ARTICULO_PROCESO_HIST_DETALLE — 2,640,976 rows, the Pareto #1).
-- Each row: one article × one cascaded process path × cost breakdown for one revision.
-- article_id is nullable because artcod may reference articles that never made
-- it into erp_articles (e.g. deleted from STK_ARTICULOS after this line was recorded).
CREATE TABLE IF NOT EXISTS erp_homologation_revision_lines (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                TEXT NOT NULL,
    revision_id              UUID NOT NULL REFERENCES erp_homologation_revisions(id) ON DELETE CASCADE,
    article_id               UUID REFERENCES erp_articles(id),
    article_code             TEXT NOT NULL DEFAULT '',
    article_desc             TEXT NOT NULL DEFAULT '',
    article_unit             TEXT NOT NULL DEFAULT '',
    process_1                TEXT NOT NULL DEFAULT '',
    process_2                TEXT NOT NULL DEFAULT '',
    process_3                TEXT NOT NULL DEFAULT '',
    process_4                TEXT NOT NULL DEFAULT '',
    multiplier               NUMERIC(10,4) NOT NULL DEFAULT 0,
    quantity                 NUMERIC(10,2) NOT NULL DEFAULT 0,
    replacement_cost         NUMERIC(10,2) NOT NULL DEFAULT 0,
    replacement_partial      NUMERIC(10,2) NOT NULL DEFAULT 0,
    replacement_cost_desc    NUMERIC(10,2) NOT NULL DEFAULT 0,
    replacement_partial_desc NUMERIC(10,2) NOT NULL DEFAULT 0,
    account_code             TEXT NOT NULL DEFAULT '',
    account_name             TEXT NOT NULL DEFAULT '',
    partial_with_surcharge   NUMERIC(10,2) NOT NULL DEFAULT 0,
    region_percentage        NUMERIC(10,2) NOT NULL DEFAULT 0,
    partial_clog             NUMERIC(10,2) NOT NULL DEFAULT 0,
    partial_surcharge_log    NUMERIC(10,2) NOT NULL DEFAULT 0,
    logistics_cost           NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_hom_rev_lines_revision ON erp_homologation_revision_lines(tenant_id, revision_id);
CREATE INDEX idx_erp_hom_rev_lines_article ON erp_homologation_revision_lines(tenant_id, article_id) WHERE article_id IS NOT NULL;
CREATE INDEX idx_erp_hom_rev_lines_artcode ON erp_homologation_revision_lines(tenant_id, article_code) WHERE article_code <> '';

-- RLS (silo-compliant, same pattern as 023_erp_invoicing.up.sql).
ALTER TABLE erp_homologations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_homologations USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_homologation_revisions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_homologation_revisions USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_homologation_revision_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_homologation_revision_lines USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions (same erp.production namespace used by related modules).
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.homologations.read',  'Ver homologaciones',      'Consultar homologaciones + revisiones',         'erp'),
    ('erp.homologations.write', 'Gestionar homologaciones', 'Crear/editar homologaciones y sus revisiones', 'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.homologations.%'
ON CONFLICT DO NOTHING;
