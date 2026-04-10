-- 018_erp_stock.up.sql
-- Plan 17 Phase 2: Stock & Warehouse
-- Replaces ~72 legacy tables: STK_*, STKINSUM, HERRAMIENTAS, etc.

-- Articulos (maestro de productos, insumos, repuestos, herramientas)
CREATE TABLE IF NOT EXISTS erp_articles (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    code          TEXT NOT NULL,
    name          TEXT NOT NULL,
    family_id     UUID REFERENCES erp_catalogs(id),
    category_id   UUID REFERENCES erp_catalogs(id),
    unit_id       UUID REFERENCES erp_catalogs(id),
    article_type  TEXT NOT NULL DEFAULT 'material'
        CHECK (article_type IN ('material', 'product', 'tool', 'spare', 'consumable')),
    min_stock     NUMERIC(14,4) NOT NULL DEFAULT 0,
    max_stock     NUMERIC(14,4) NOT NULL DEFAULT 0,
    reorder_point NUMERIC(14,4) NOT NULL DEFAULT 0,
    last_cost     NUMERIC(14,4) NOT NULL DEFAULT 0,
    avg_cost      NUMERIC(14,4) NOT NULL DEFAULT 0,
    metadata      JSONB NOT NULL DEFAULT '{}',
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_erp_articles_search ON erp_articles(tenant_id, active, name);
CREATE INDEX idx_erp_articles_type ON erp_articles(tenant_id, article_type) WHERE active = true;

-- Depositos/Almacenes
CREATE TABLE IF NOT EXISTS erp_warehouses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    location    TEXT NOT NULL DEFAULT '',
    active      BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, code)
);

-- Stock actual por articulo x deposito (denormalized cache)
CREATE TABLE IF NOT EXISTS erp_stock_levels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    article_id      UUID NOT NULL REFERENCES erp_articles(id),
    warehouse_id    UUID NOT NULL REFERENCES erp_warehouses(id),
    quantity        NUMERIC(14,4) NOT NULL DEFAULT 0,
    reserved        NUMERIC(14,4) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, article_id, warehouse_id)
);

-- Movimientos de stock (inmutable — append-only)
CREATE TABLE IF NOT EXISTS erp_stock_movements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    article_id      UUID NOT NULL REFERENCES erp_articles(id),
    warehouse_id    UUID NOT NULL REFERENCES erp_warehouses(id),
    movement_type   TEXT NOT NULL CHECK (movement_type IN ('in', 'out', 'transfer', 'adjustment')),
    quantity        NUMERIC(14,4) NOT NULL,
    unit_cost       NUMERIC(14,4) NOT NULL DEFAULT 0,
    reference_type  TEXT,
    reference_id    UUID,
    concept_id      UUID REFERENCES erp_catalogs(id),
    user_id         TEXT NOT NULL,
    notes           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_stock_movements_article ON erp_stock_movements(tenant_id, article_id, created_at DESC);
CREATE INDEX idx_erp_stock_movements_ref ON erp_stock_movements(reference_type, reference_id)
    WHERE reference_id IS NOT NULL;

-- BOM (bill of materials)
CREATE TABLE IF NOT EXISTS erp_bom (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    parent_id   UUID NOT NULL REFERENCES erp_articles(id),
    child_id    UUID NOT NULL REFERENCES erp_articles(id),
    quantity    NUMERIC(14,4) NOT NULL,
    unit_id     UUID REFERENCES erp_catalogs(id),
    sort_order  INT NOT NULL DEFAULT 0,
    notes       TEXT NOT NULL DEFAULT '',
    UNIQUE(tenant_id, parent_id, child_id)
);
CREATE INDEX idx_erp_bom_parent ON erp_bom(tenant_id, parent_id);

-- Fotos de articulos
CREATE TABLE IF NOT EXISTS erp_article_photos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    article_id  UUID NOT NULL REFERENCES erp_articles(id) ON DELETE CASCADE,
    file_key    TEXT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_article_photos ON erp_article_photos(tenant_id, article_id);

-- RLS
ALTER TABLE erp_articles ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_articles USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_warehouses ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_warehouses USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_stock_levels ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_stock_levels USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_stock_movements ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_stock_movements USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_bom ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_bom USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_article_photos ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_article_photos USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.stock.read',  'Ver stock',      'Consultar articulos, movimientos, niveles', 'erp'),
    ('erp.stock.write', 'Gestionar stock', 'Registrar movimientos, crear articulos',   'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.stock.%'
ON CONFLICT DO NOTHING;
