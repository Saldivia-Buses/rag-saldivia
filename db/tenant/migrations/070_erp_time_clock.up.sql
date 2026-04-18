-- 070_erp_time_clock.up.sql
-- Phase 1 §Data migration: FICHADAS + PERSONAL_TARJETA (Pareto #2)
-- Together these close ~41 % of the remaining Phase 1 row-volume gap
-- (FICHADAS alone = 1.46 M rows, #1 in the post-2.0.8 Pareto).
--
-- Histrix context:
--   - PERSONAL_TARJETA (1,403 rows) is the versioned card-to-employee
--     assignment table. Cards get reassigned over time; a row is
--     (idPersona, tarjeta VARCHAR(20), fechaDesde DATE).
--   - FICHADAS (1,465,002 rows) is the raw clock-punch stream from the
--     physical terminals. Every marcaje has its own row with
--     (tarjeta INT, fecha DATE, hora TIME, reloj VARCHAR(5), codigo,
--     marca, borrado, insertkey UNIQUE, id_fichada AUTO_INC).
--   - FICHADADIA (already migrated to erp_attendance) is the DAILY
--     rollup computed from FICHADAS in Histrix. It keeps hours worked
--     and the first four punches per day. FICHADAS is the source of
--     truth; the XML-form scrape shows 116 direct references to the
--     raw table across rrhh, sueldos, dashboard, estadisticas,
--     rh_evaluaciones — raw events are NOT dead weight.
--
-- FK chain: FICHADAS.tarjeta (INT) → PERSONAL_TARJETA.tarjeta (VARCHAR,
-- date-versioned by fechaDesde) → PERSONAL.idPersona → erp_entities.
-- The migrator resolves at row-time using the largest fechaDesde <=
-- FICHADAS.fecha; orphan tarjetas are migrated with entity_id NULL
-- (preserving the event for forensic/audit; the employee just can't
-- be tied back).

-- ---------------------------------------------------------------------
-- Card-to-employee assignments (PERSONAL_TARJETA).
-- Date-versioned: a single card may reassign across employees over
-- time. Queries resolve "who had card X on date Y" via the row with
-- the largest effective_from <= Y.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_employee_cards (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      TEXT NOT NULL,
    entity_id      UUID NOT NULL REFERENCES erp_entities(id),
    card_code      TEXT NOT NULL,
    effective_from DATE NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, entity_id, card_code, effective_from)
);
CREATE INDEX idx_erp_employee_cards_card_date ON erp_employee_cards(tenant_id, card_code, effective_from DESC);
CREATE INDEX idx_erp_employee_cards_entity ON erp_employee_cards(tenant_id, entity_id);

-- ---------------------------------------------------------------------
-- Raw clock-punch events (FICHADAS).
-- One row per physical marcaje. entity_id is NULLABLE so orphan
-- tarjetas (card never assigned to anyone in PERSONAL_TARJETA) still
-- migrate — we preserve the event and surface "unknown employee" at
-- query time rather than dropping the row.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS erp_time_clock_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    entity_id     UUID REFERENCES erp_entities(id),
    card_code     TEXT NOT NULL,
    event_time    TIMESTAMPTZ,
    event_type    TEXT NOT NULL DEFAULT '',
    terminal      TEXT NOT NULL DEFAULT '',
    marca         SMALLINT NOT NULL DEFAULT 0,
    deleted_flag  SMALLINT NOT NULL DEFAULT 0,
    insert_key    TEXT NOT NULL,
    legacy_id     BIGINT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, insert_key)
);
CREATE INDEX idx_erp_time_clock_event_time ON erp_time_clock_events(tenant_id, event_time DESC) WHERE event_time IS NOT NULL;
CREATE INDEX idx_erp_time_clock_entity ON erp_time_clock_events(tenant_id, entity_id, event_time DESC) WHERE entity_id IS NOT NULL;
CREATE INDEX idx_erp_time_clock_terminal ON erp_time_clock_events(tenant_id, terminal, event_time DESC) WHERE terminal <> '';
CREATE INDEX idx_erp_time_clock_card ON erp_time_clock_events(tenant_id, card_code, event_time DESC);
CREATE INDEX idx_erp_time_clock_active ON erp_time_clock_events(tenant_id) WHERE deleted_flag = 0;

-- RLS (silo-compliant, same pattern as 069).
ALTER TABLE erp_employee_cards ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_employee_cards USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_time_clock_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_time_clock_events USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions.
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.time_clock.read',  'Ver fichadas',      'Consultar eventos de reloj y asignaciones de tarjeta', 'erp'),
    ('erp.time_clock.write', 'Gestionar fichadas', 'Corregir/anular eventos de reloj y asignar tarjetas', 'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.time_clock.%'
ON CONFLICT DO NOTHING;
