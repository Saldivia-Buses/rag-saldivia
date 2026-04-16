-- 022_erp_sales.up.sql
-- Plan 17 Phase 6: Sales & CRM
-- Replaces ~64 legacy tables: FICHAVENTAS, PEDCOTIZ, COTIZACION, LISTA*, CRM*, etc.

-- Cotizaciones
CREATE TABLE IF NOT EXISTS erp_quotations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    number      TEXT NOT NULL,
    date        DATE NOT NULL,
    customer_id UUID NOT NULL REFERENCES erp_entities(id),
    status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','sent','approved','rejected','expired')),
    currency_id UUID REFERENCES erp_catalogs(id),
    total       NUMERIC(16,2) NOT NULL DEFAULT 0,
    valid_until DATE,
    notes       TEXT NOT NULL DEFAULT '',
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, number)
);
CREATE INDEX idx_erp_quotations_date ON erp_quotations(tenant_id, date DESC);

-- Lineas de cotizacion
CREATE TABLE IF NOT EXISTS erp_quotation_lines (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    quotation_id UUID NOT NULL REFERENCES erp_quotations(id) ON DELETE CASCADE,
    article_id   UUID REFERENCES erp_articles(id),
    description  TEXT NOT NULL,
    quantity     NUMERIC(14,4) NOT NULL CHECK (quantity > 0),
    unit_price   NUMERIC(14,4) NOT NULL CHECK (unit_price >= 0),
    sort_order   INT NOT NULL DEFAULT 0,
    metadata     JSONB NOT NULL DEFAULT '{}'
);
CREATE INDEX idx_erp_quotation_lines ON erp_quotation_lines(tenant_id, quotation_id);

-- Pedidos (internos o de clientes)
CREATE TABLE IF NOT EXISTS erp_orders (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    number       TEXT NOT NULL,
    date         DATE NOT NULL,
    order_type   TEXT NOT NULL CHECK (order_type IN ('customer', 'internal')),
    customer_id  UUID REFERENCES erp_entities(id),
    quotation_id UUID REFERENCES erp_quotations(id),
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','shipped','delivered','cancelled')),
    total        NUMERIC(16,2) NOT NULL DEFAULT 0,
    user_id      TEXT NOT NULL,
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, number)
);
CREATE INDEX idx_erp_orders_date ON erp_orders(tenant_id, date DESC);
CREATE INDEX idx_erp_orders_customer ON erp_orders(tenant_id, customer_id) WHERE customer_id IS NOT NULL;

-- Listas de precios
CREATE TABLE IF NOT EXISTS erp_price_lists (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    name        TEXT NOT NULL,
    currency_id UUID REFERENCES erp_catalogs(id),
    valid_from  DATE,
    valid_until DATE,
    active      BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS erp_price_list_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    price_list_id UUID NOT NULL REFERENCES erp_price_lists(id) ON DELETE CASCADE,
    article_id    UUID REFERENCES erp_articles(id),
    description   TEXT,
    price         NUMERIC(16,2) NOT NULL,
    CHECK (article_id IS NOT NULL OR description IS NOT NULL)
);
CREATE INDEX idx_erp_price_list_items ON erp_price_list_items(tenant_id, price_list_id);

-- RLS
ALTER TABLE erp_quotations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_quotations USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_quotation_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_quotation_lines USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_orders USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_price_lists ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_price_lists USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_price_list_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_price_list_items USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.sales.read',  'Ver ventas',      'Consultar cotizaciones, pedidos, precios', 'erp'),
    ('erp.sales.write', 'Gestionar ventas', 'Crear cotizaciones y pedidos',            'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.sales.%'
ON CONFLICT DO NOTHING;
