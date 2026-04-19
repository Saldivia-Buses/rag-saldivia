-- 072_erp_tools.up.sql
-- Phase 1 §Data migration: HERRAMIENTAS + HERRMOVS (Pareto #4, 400 K rows
-- combined).
--
-- Histrix context:
--   - HERRAMIENTAS (389,253 rows live) is mis-named historically. Despite
--     the name it is NOT a workshop tools catalog; it is the serialized
--     inventory tag ledger — one row per physical item received, each with
--     its own UNIQUE barcode-like code (id_herramienta varchar(25)) and AI
--     PK (id_etiqueta). Live xml-form scrape shows it used across
--     recepcion/, almacen/, herramientas/, mantenimiento/, help_local/ —
--     the "Herramientas" menu groups serialized inventory UX.
--     Typical content: "VENTANA CENTRAL CAME 1400X320 IZQUIERDA",
--     "VIDRIO TRASERO IZQUIERDO NUEVO ARIES" — part-level items with OCP
--     (orden de compra) + REM (remito) traceability.
--   - HERRMOVS (11,680 rows live) is the lending ledger: employees take
--     out / return / damage items. CONCHERR defines the 4 movement
--     concepts (1=Devol. Rotura, 2=Devolucion, 3=A Cargo, 7=Prestamo)
--     with movtip IN/OUT flag — we inline these 4 values as integer code
--     + derived direction rather than a separate table.
--   - HERRMOVS.id_herramienta ALSO resolves to MANT_EQUIPOS.numero_serie
--     in some live XML queries (mixed-use lending ledger). We resolve
--     primarily against HERRAMIENTAS; the 1,566 orphan movements (13 %)
--     migrate with tool_id NULL, preserving the raw code — same forensic
--     pattern as FICHADAS orphan tarjetas.
--
-- Naming rationale: keep erp_tools / erp_tool_movements for operational
-- parity with the Histrix "Herramientas" menu that users navigate. The
-- domain comment here and in the commit message carries the "this is
-- serialized inventory, not workshop tools" caveat.

-- ---------------------------------------------------------------------
-- erp_tools — serialized inventory tags (HERRAMIENTAS)
-- One row per physical item; id_herramienta (varchar(25)) is the human-
-- readable unique code stamped on the item, id_etiqueta (bigint AI) is
-- the system-internal tag PK preserved as legacy_id.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_tools (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          TEXT NOT NULL,
    legacy_id          BIGINT NOT NULL,              -- id_etiqueta
    code               TEXT NOT NULL,                -- id_herramienta (the stamped barcode code)
    article_code       TEXT NOT NULL DEFAULT '',     -- artcod (STK_ARTICULOS.id_stkarticulo)
    article_id         UUID REFERENCES erp_articles(id), -- nullable: resolved via stock domain
    inventory_code     TEXT NOT NULL DEFAULT '',     -- invcod
    name               TEXT NOT NULL DEFAULT '',     -- nomherr
    characteristic     TEXT NOT NULL DEFAULT '',     -- caract
    group_code         SMALLINT NOT NULL DEFAULT 0,  -- grucod
    tool_type          SMALLINT NOT NULL DEFAULT 0,  -- tipoherr
    status_code        INTEGER NOT NULL DEFAULT 0,   -- codest (1=active, 0/2/3=other states)
    purchase_order_no  INTEGER NOT NULL DEFAULT 0,   -- ocpnro
    purchase_order_date DATE,                        -- ocpfec
    delivery_note_date DATE,                         -- remfec
    delivery_note_post INTEGER NOT NULL DEFAULT 0,   -- remnpv
    delivery_note_no   INTEGER NOT NULL DEFAULT 0,   -- remnro
    supplier_code      INTEGER NOT NULL DEFAULT 0,   -- ctacod (proveedor)
    pending_oc         NUMERIC(14,2) NOT NULL DEFAULT 0, -- pendiente_oc
    observation        TEXT NOT NULL DEFAULT '',     -- observacion
    manufacture_no     INTEGER NOT NULL DEFAULT 0,   -- nrofab
    generated_at       TIMESTAMPTZ,                  -- generada (NULL for 0000 rows)
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);

CREATE INDEX idx_erp_tools_code ON erp_tools(tenant_id, code);
CREATE INDEX idx_erp_tools_article_code ON erp_tools(tenant_id, article_code);
CREATE INDEX idx_erp_tools_article_id ON erp_tools(tenant_id, article_id) WHERE article_id IS NOT NULL;
CREATE INDEX idx_erp_tools_status ON erp_tools(tenant_id, status_code);
CREATE INDEX idx_erp_tools_oc ON erp_tools(tenant_id, purchase_order_no) WHERE purchase_order_no > 0;

-- ---------------------------------------------------------------------
-- erp_tool_movements — lending ledger (HERRMOVS)
-- Each row: someone takes out, returns, or damages a tool/item. tool_id
-- nullable so the 13 % orphan movements (id_herramienta not in HERRAMIENTAS
-- — probably MANT_EQUIPOS.numero_serie cross-references) still migrate.
-- The 4-row CONCHERR enum is inlined as concept_code SMALLINT + the
-- canonical name resolved at query time (stable hardcoded mapping).
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_tool_movements (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      TEXT NOT NULL,
    legacy_id      BIGINT NOT NULL,                 -- id_herrmovs (AI PK)
    tool_id        UUID REFERENCES erp_tools(id),   -- nullable — orphan movs preserve tool_code
    tool_code      TEXT NOT NULL DEFAULT '',        -- raw id_herramienta
    user_code      TEXT NOT NULL DEFAULT '',        -- HERRMOVS.usuario (PERSONAL.opecod)
    quantity       INTEGER NOT NULL DEFAULT 0,
    movement_date  DATE,                            -- movfec (NULL for 0000-00-00)
    concept_code   SMALLINT NOT NULL DEFAULT 0,     -- movher (CONCHERR.movher: 1/2=IN, 3/7=OUT)
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);

CREATE INDEX idx_erp_tool_movements_tool ON erp_tool_movements(tenant_id, tool_id, movement_date DESC) WHERE tool_id IS NOT NULL;
CREATE INDEX idx_erp_tool_movements_tool_code ON erp_tool_movements(tenant_id, tool_code);
CREATE INDEX idx_erp_tool_movements_user ON erp_tool_movements(tenant_id, user_code, movement_date DESC) WHERE user_code <> '';
CREATE INDEX idx_erp_tool_movements_date ON erp_tool_movements(tenant_id, movement_date DESC) WHERE movement_date IS NOT NULL;

-- RLS (silo-compliant, same pattern as 071).
ALTER TABLE erp_tools ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_tools
    USING (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE erp_tool_movements ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_tool_movements
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions.
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.tools.read',  'Ver herramientas', 'Consultar etiquetas serializadas y movimientos de herramientas', 'erp'),
    ('erp.tools.write', 'Gestionar herramientas', 'Asignar/devolver herramientas y corregir movimientos', 'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.tools.%'
ON CONFLICT DO NOTHING;
