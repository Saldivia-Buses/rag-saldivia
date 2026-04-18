-- 074_erp_products.up.sql
-- Phase 1 §Data migration: the full PRODUCTO_* cluster (Pareto #6 +
-- Pareto #18 in one migration, ~406 K rows across 6 tables).
--
-- Histrix context:
--   - PRODUCTOS (4,108 rows) = master catalog of "productos terminados"
--     (bus/unit models). descripcion_producto is a short code like "434"
--     that ALSO appears in STK_ARTICULOS.artcod — a bus model is both a
--     product in this domain AND an article in stock. Historical
--     metadata-enricher (tools/cli/internal/migration/metadata_enrichment.go
--     articleProductAttributes) already joined through this to attach
--     JSONB metadata to erp_articles. That path stays; we now ALSO
--     materialize the full relational shape.
--   - PRODUCTO_SECCION (10 rows) = sections of a product spec (Datos
--     Generales, Aire Acondicionado, Calefacción, Pintura, Asientos…).
--   - PRODUCTO_ATRIBUTOS (415 rows) = attribute definitions per section
--     (Cucheta, Posición, Color Llantas, Cartel Ruta…).
--   - PRODUCTO_ATRIB_OPCIONES (147 rows) = enumerated option lists for
--     enum-typed attributes (e.g. id_prdatributo=2: 0=Sin Definir,
--     1=Preparado, 2=Colocado, 3=No).
--   - PRODUCTO_ATRIB_VALORES (353,936 rows) = the actual key-value
--     payload — one row per (product, attribute) snapshot with its
--     current value and timestamp. 89 % have a resolvable producto_id,
--     11 % orphan rows migrate with product_id NULL preserving the raw
--     value. This is Pareto #6 of the post-2.0.10 gap.
--   - PRODUCTO_ATRIBUTO_HOMOLOGACION (47,189 rows) = join table linking
--     attribute definitions to homologations (erp_homologations from
--     2.0.8). Rank 18 in the original Pareto, closes cleanly as part of
--     this cluster.
--
-- XML-form scrape: producto/, producto/ing/, producto/producto_atrib*,
-- ventas/ficha_venta_seccion_qry, producto_panas, productos_duplicar_ins,
-- producto_atributos_homologacion_qry. The UI walks this cluster end-to-end.

CREATE TABLE IF NOT EXISTS erp_product_sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    legacy_id       BIGINT NOT NULL,           -- id_prdseccion
    name            TEXT NOT NULL DEFAULT '',  -- nombre_seccion
    sort_order      INTEGER NOT NULL DEFAULT 0,-- orden_seccion
    rubro_id        INTEGER NOT NULL DEFAULT 0,-- legacy rubro FK (preserved raw)
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);

CREATE TABLE IF NOT EXISTS erp_products (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            TEXT NOT NULL,
    legacy_id            BIGINT NOT NULL,              -- id_producto
    description          TEXT NOT NULL DEFAULT '',     -- descripcion_producto
    supplier_entity_id   UUID REFERENCES erp_entities(id), -- regcuenta_id
    supplier_code        INTEGER NOT NULL DEFAULT 0,   -- regcuenta_id raw preserved
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_products_description ON erp_products(tenant_id, description);

CREATE TABLE IF NOT EXISTS erp_product_attributes (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                  TEXT NOT NULL,
    legacy_id                  BIGINT NOT NULL,            -- id_prdatributo
    name                       TEXT NOT NULL DEFAULT '',   -- nombre_atributo
    attribute_type             TEXT NOT NULL DEFAULT '',   -- tipo_atributo (check/varchar/file/etc)
    section_id                 UUID REFERENCES erp_product_sections(id),
    section_legacy_id          INTEGER NOT NULL DEFAULT 0, -- prdseccion_id preserved raw
    article_code               TEXT NOT NULL DEFAULT '',   -- stkarticulo_id varchar
    helper_xml                 TEXT NOT NULL DEFAULT '',
    helper_dir                 TEXT NOT NULL DEFAULT '',
    parameters                 TEXT NOT NULL DEFAULT '',
    sort_order                 INTEGER NOT NULL DEFAULT 0,
    active                     BOOLEAN NOT NULL DEFAULT true,
    print_label                BOOLEAN NOT NULL DEFAULT true,
    print_value                BOOLEAN NOT NULL DEFAULT true,
    active_in_quote            BOOLEAN NOT NULL DEFAULT false, -- activo_cotizacion
    active_in_tech_sheet       BOOLEAN NOT NULL DEFAULT false, -- activo_fichatecnica
    quote_description          TEXT NOT NULL DEFAULT '',       -- descrip_cotizacion
    define_before_section_id   INTEGER NOT NULL DEFAULT 0,
    standard_additional        SMALLINT NOT NULL DEFAULT 0,
    code                       TEXT NOT NULL DEFAULT '',
    print_section_id           TEXT NOT NULL DEFAULT '',
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_product_attributes_section ON erp_product_attributes(tenant_id, section_id) WHERE section_id IS NOT NULL;
CREATE INDEX idx_erp_product_attributes_name ON erp_product_attributes(tenant_id, name);
CREATE INDEX idx_erp_product_attributes_active ON erp_product_attributes(tenant_id) WHERE active = true;

CREATE TABLE IF NOT EXISTS erp_product_attribute_options (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    legacy_id       BIGINT NOT NULL,            -- id_atribopcion
    attribute_id    UUID REFERENCES erp_product_attributes(id),
    attribute_legacy_id INTEGER NOT NULL DEFAULT 0, -- prdatributo_id raw
    option_name     TEXT NOT NULL DEFAULT '',   -- nombre_opcion
    option_value    TEXT NOT NULL DEFAULT '',   -- valor_opcion
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_product_attribute_options_attr ON erp_product_attribute_options(tenant_id, attribute_id) WHERE attribute_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS erp_product_attribute_values (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              TEXT NOT NULL,
    legacy_id              BIGINT NOT NULL,             -- id_atribvalor
    product_id             UUID REFERENCES erp_products(id),           -- nullable for orphan producto_id (11 %)
    product_legacy_id      INTEGER NOT NULL DEFAULT 0,  -- producto_id raw preserved
    attribute_id           UUID REFERENCES erp_product_attributes(id), -- 100 % resolvable
    attribute_legacy_id    INTEGER NOT NULL DEFAULT 0,  -- prdatributo_id raw
    value                  TEXT NOT NULL DEFAULT '',    -- valor_atributo
    quantity               INTEGER NOT NULL DEFAULT 0,  -- cantidad_atributo
    quote_legacy_id        INTEGER NOT NULL DEFAULT 0,  -- cotizacion_id (raw, no FK resolution yet)
    recorded_at            TIMESTAMPTZ,                 -- timestamp_atributo
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_pav_product ON erp_product_attribute_values(tenant_id, product_id) WHERE product_id IS NOT NULL;
CREATE INDEX idx_erp_pav_attribute ON erp_product_attribute_values(tenant_id, attribute_id) WHERE attribute_id IS NOT NULL;
CREATE INDEX idx_erp_pav_quote ON erp_product_attribute_values(tenant_id, quote_legacy_id) WHERE quote_legacy_id > 0;

CREATE TABLE IF NOT EXISTS erp_product_attribute_homologations (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              TEXT NOT NULL,
    legacy_id              BIGINT NOT NULL,             -- id_atrib_homolog
    attribute_id           UUID REFERENCES erp_product_attributes(id),
    attribute_legacy_id    INTEGER NOT NULL DEFAULT 0,  -- prdatributo_id raw
    homologation_id        UUID REFERENCES erp_homologations(id),
    homologation_legacy_id INTEGER NOT NULL DEFAULT 0,  -- homologacion_id raw
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_pah_attr ON erp_product_attribute_homologations(tenant_id, attribute_id) WHERE attribute_id IS NOT NULL;
CREATE INDEX idx_erp_pah_homolog ON erp_product_attribute_homologations(tenant_id, homologation_id) WHERE homologation_id IS NOT NULL;

-- RLS (silo-compliant)
ALTER TABLE erp_product_sections ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_product_sections USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_products ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_products USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_product_attributes ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_product_attributes USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_product_attribute_options ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_product_attribute_options USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_product_attribute_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_product_attribute_values USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_product_attribute_homologations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_product_attribute_homologations USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions.
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.products.read',  'Ver productos', 'Consultar productos terminados y sus atributos', 'erp'),
    ('erp.products.write', 'Gestionar productos', 'Crear/modificar productos, atributos y valores', 'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.products.%'
ON CONFLICT DO NOTHING;
