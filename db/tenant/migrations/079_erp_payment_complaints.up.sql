-- 079_erp_payment_complaints.up.sql
-- Phase 1 §Data migration — RECLAMOPAGOS (Pareto #20 of the
-- post-2.0.10 gap, 15,463 rows live). The "reclamo de pagos"
-- log: one row per supplier-payment complaint / reclamation the
-- accounts-payable team opens, with a tiny flag (`marca`) toggling
-- "pendiente" (0) ↔ "cumplida" (1).
--
-- Histrix context:
--   - Live form: `xml-forms/reclamos/reclamopagos.xml` ("abm-mini"),
--     plus `reclamopagos_ing.xml` and `reclamopagos_ingmov.xml` for
--     the entry-from-movement flow.
--   - 6-column source (idReclamo AI, fecha, ctacod, observacion
--     longtext, marca tinyint, login). `ctacod` resolves through
--     ResolveEntityFlexible against the REG_CUENTA / nro_cuenta
--     indexes already built in Phase 2.

CREATE TABLE IF NOT EXISTS erp_payment_complaints (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            TEXT NOT NULL,
    legacy_id            BIGINT NOT NULL,              -- idReclamo
    complaint_date       DATE,                         -- fecha (nullable for zero-date)
    entity_legacy_code   INTEGER NOT NULL DEFAULT 0,   -- ctacod raw
    entity_id            UUID REFERENCES erp_entities(id),
    observation          TEXT NOT NULL DEFAULT '',     -- observacion longtext
    status_flag          SMALLINT NOT NULL DEFAULT 0,  -- marca (0=pending, 1=done)
    login                TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);
CREATE INDEX idx_erp_payment_complaints_entity
    ON erp_payment_complaints(tenant_id, entity_id)
    WHERE entity_id IS NOT NULL;
CREATE INDEX idx_erp_payment_complaints_status
    ON erp_payment_complaints(tenant_id, status_flag, complaint_date DESC);
CREATE INDEX idx_erp_payment_complaints_date
    ON erp_payment_complaints(tenant_id, complaint_date DESC)
    WHERE complaint_date IS NOT NULL;

ALTER TABLE erp_payment_complaints ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_payment_complaints
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Reuses erp.current_accounts.read / erp.current_accounts.write.
-- No new permissions added.
