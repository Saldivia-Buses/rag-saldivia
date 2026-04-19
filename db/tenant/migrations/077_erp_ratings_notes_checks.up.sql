-- 077_erp_ratings_notes_checks.up.sql
-- Phase 1 §Data migration — Pareto tail Grupo A (current-accounts /
-- treasury). Three tables that sit under the entity + invoice surfaces
-- already migrated, total ~237 K rows live:
--
--   REG_CUENTA_CALIFICACION  → erp_entity_credit_ratings   (136,064 rows)
--   REG_MOVIMIENTO_OBS       → erp_invoice_notes           ( 72,737 rows)
--   CARCHEHI                 → erp_check_history           ( 28,763 rows)
--
-- All three resolve their foreign keys against indexes built earlier
-- in the pipeline (REG_CUENTA entity cache, Phase 6 BuildRegMovimIndex,
-- ResolveEntityFlexible via nro_cuenta / id_regcuenta). No new hooks.

-- ---------------------------------------------------------------------
-- erp_entity_credit_ratings — REG_CUENTA_CALIFICACION (136 K)
-- Customer / supplier credit rating history.
-- Histrix shape: (id_regcalificacion PK, regcuenta_id FK, calificacion
-- VARCHAR(40), fecha_calificacion DATETIME, referencia_calificacion
-- VARCHAR(200)). The FK points at REG_CUENTA(id_regcuenta), which is
-- already cached in the entity domain — direct ResolveOptional.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_entity_credit_ratings (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          TEXT NOT NULL,
    legacy_id          BIGINT NOT NULL,                -- id_regcalificacion
    entity_id          UUID REFERENCES erp_entities(id),
    entity_legacy_id   INTEGER NOT NULL DEFAULT 0,     -- regcuenta_id raw
    rating             TEXT NOT NULL DEFAULT '',       -- calificacion
    rated_at           TIMESTAMPTZ,                    -- fecha_calificacion
    reference          TEXT NOT NULL DEFAULT '',       -- referencia_calificacion
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_entity_credit_ratings_entity
    ON erp_entity_credit_ratings(tenant_id, entity_id)
    WHERE entity_id IS NOT NULL;
CREATE INDEX idx_erp_entity_credit_ratings_date
    ON erp_entity_credit_ratings(tenant_id, rated_at DESC)
    WHERE rated_at IS NOT NULL;

ALTER TABLE erp_entity_credit_ratings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_entity_credit_ratings
    USING (tenant_id = current_setting('app.tenant_id', true));

-- ---------------------------------------------------------------------
-- erp_invoice_notes — REG_MOVIMIENTO_OBS (73 K)
-- Per-invoice / per-movement free-text observations attached to
-- REG_MOVIMIENTOS rows. Histrix stores one row per note with a
-- longtext `observacion`, the operator login, and a handful of
-- denormalised movement keys (ctacod + concod + movnpv + movnro +
-- siscod + movfec) that duplicate the FK. We keep them raw for
-- forensic purposes and resolve regmovim_id via the Phase 6 index.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_invoice_notes (
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                 TEXT NOT NULL,
    legacy_id                 BIGINT NOT NULL,             -- id_regmovimientoobs
    observation_date          DATE,                        -- fec_observacion (nullable for zero-date)
    observation_time          TIME,                        -- hora_observacion
    observation               TEXT NOT NULL DEFAULT '',    -- observacion longtext
    invoice_id                UUID REFERENCES erp_invoices(id),
    invoice_legacy_id         INTEGER NOT NULL DEFAULT 0,  -- regmovim_id raw
    login                     TEXT NOT NULL DEFAULT '',
    contact_legacy_id         INTEGER NOT NULL DEFAULT 0,  -- gencontacto_id
    source_table              TEXT NOT NULL DEFAULT '',    -- tabla_origen
    system_code               TEXT NOT NULL DEFAULT '',    -- siscod
    movement_date             DATE,                        -- movfec (nullable for zero-date)
    account_code              INTEGER NOT NULL DEFAULT 0,  -- ctacod
    concept_code              INTEGER NOT NULL DEFAULT 0,  -- concod
    movement_voucher_class    INTEGER NOT NULL DEFAULT 0,  -- movnpv
    movement_no               INTEGER NOT NULL DEFAULT 0,  -- movnro
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_invoice_notes_invoice
    ON erp_invoice_notes(tenant_id, invoice_id)
    WHERE invoice_id IS NOT NULL;
CREATE INDEX idx_erp_invoice_notes_date
    ON erp_invoice_notes(tenant_id, observation_date DESC)
    WHERE observation_date IS NOT NULL;
CREATE INDEX idx_erp_invoice_notes_source
    ON erp_invoice_notes(tenant_id, source_table)
    WHERE source_table <> '';

ALTER TABLE erp_invoice_notes ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_invoice_notes
    USING (tenant_id = current_setting('app.tenant_id', true));

-- ---------------------------------------------------------------------
-- erp_check_history — CARCHEHI (29 K)
-- Archived-check history. Sister of erp_checks (which is the active-
-- portfolio view of CARCHEQU). Composite natural key (carint, siscod,
-- succod) with 36 raw columns; we preserve the composite parts plus
-- the semantically meaningful columns and synthesise legacy_id via
-- hashCode("CARCHEHI:<carint>:<siscod>:<succod>") so the idempotency
-- UNIQUE (tenant_id, legacy_id) still works.
--
-- FK resolution: ctacod via ResolveEntityFlexible (id_regcuenta first,
-- then nro_cuenta). movnro+regmin is a soft pointer at REG_MOVIMIENTOS
-- composite keys — not resolved, preserved raw.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_check_history (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                TEXT NOT NULL,
    legacy_id                BIGINT NOT NULL,                -- hash(CARCHEHI:carint:siscod:succod)
    legacy_carint            INTEGER NOT NULL DEFAULT 0,
    legacy_siscod            TEXT NOT NULL DEFAULT '',
    legacy_succod            INTEGER NOT NULL DEFAULT 0,
    check_type               SMALLINT NOT NULL DEFAULT 0,    -- cartip
    number                   TEXT NOT NULL DEFAULT '',       -- carnro
    bank_name                TEXT NOT NULL DEFAULT '',       -- carbco
    amount                   NUMERIC(14,2) NOT NULL DEFAULT 0, -- carimp
    operation_date           DATE,                           -- carfec (nullable for zero)
    credited_at              DATE,                           -- caracr
    returned_at              DATE,                           -- cardev
    altered_at               DATE,                           -- caralt (nullable for zero)
    deposited_at             DATE,                           -- caring (nullable for zero)
    issue_date               DATE,                           -- fecha_emision
    description              TEXT NOT NULL DEFAULT '',       -- cardes
    observation              TEXT NOT NULL DEFAULT '',       -- carobv
    reference                TEXT NOT NULL DEFAULT '',       -- carref
    owner_ident              TEXT NOT NULL DEFAULT '',       -- carcui (CUIT)
    owner_mark               TEXT NOT NULL DEFAULT '',       -- carmar
    accredited               SMALLINT NOT NULL DEFAULT 0,    -- acreditado
    entity_legacy_code       INTEGER NOT NULL DEFAULT 0,     -- ctacod raw
    entity_id                UUID REFERENCES erp_entities(id),
    movement_no              INTEGER NOT NULL DEFAULT 0,     -- movnro
    movement_register        INTEGER NOT NULL DEFAULT 0,     -- regmin
    movement_voucher_class   INTEGER NOT NULL DEFAULT 0,     -- movnpv
    portfolio_id             INTEGER NOT NULL DEFAULT 0,     -- cartera_id
    branch                   INTEGER NOT NULL DEFAULT 0,     -- succod (duplicated for quick filter)
    system_code              TEXT NOT NULL DEFAULT '',       -- siscod
    concept_code             INTEGER NOT NULL DEFAULT 0,     -- concod
    operator_code            INTEGER NOT NULL DEFAULT 0,     -- opecod
    operator_class           TEXT NOT NULL DEFAULT '',       -- opecla
    plan_id                  INTEGER NOT NULL DEFAULT 0,     -- carpla
    pay_no                   INTEGER NOT NULL DEFAULT 0,     -- carpag
    received_no              INTEGER NOT NULL DEFAULT 0,     -- carrec
    check_counter            INTEGER NOT NULL DEFAULT 0,     -- carccb
    account_balance_ref      INTEGER NOT NULL DEFAULT 0,     -- ccbcod
    process_code             INTEGER NOT NULL DEFAULT 0,     -- procod
    circuit_code             INTEGER NOT NULL DEFAULT 0,     -- circod
    bcs_no                   INTEGER NOT NULL DEFAULT 0,     -- bcsnro
    cash_plan                INTEGER NOT NULL DEFAULT 0,     -- cajpla
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_check_history_number
    ON erp_check_history(tenant_id, number)
    WHERE number <> '';
CREATE INDEX idx_erp_check_history_operation_date
    ON erp_check_history(tenant_id, operation_date DESC)
    WHERE operation_date IS NOT NULL;
CREATE INDEX idx_erp_check_history_issue_date
    ON erp_check_history(tenant_id, issue_date DESC)
    WHERE issue_date IS NOT NULL;
CREATE INDEX idx_erp_check_history_entity
    ON erp_check_history(tenant_id, entity_id)
    WHERE entity_id IS NOT NULL;
CREATE INDEX idx_erp_check_history_amount
    ON erp_check_history(tenant_id, amount);

ALTER TABLE erp_check_history ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_check_history
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Reuses existing permission surfaces:
--   - erp.current_accounts.read / write for entity credit ratings + invoice notes.
--   - erp.treasury.read / write for check history.
-- No new permissions added.
