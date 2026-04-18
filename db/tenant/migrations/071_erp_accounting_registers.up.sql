-- 071_erp_accounting_registers.up.sql
-- Phase 1 §Data migration: CTBREGIS (Pareto #3, ~604 K rows live — scrape
-- at 572 K, table still growing in prod).
--
-- Histrix context:
--   - CTBREGIS is the LEAF-LEVEL accounting log: one row per debe/haber line
--     across every minuta (asiento) in every subsystem. PK id_ctbregis
--     AUTO_INCREMENT; natural traversal is (siscod, regmin, regord).
--   - The newer schema CTB_MOVIMIENTOS (cabecera) + CTB_DETALLES (detalle)
--     was introduced later and already covered by 019_erp_accounting
--     (→ erp_journal_entries + erp_journal_lines). But CTBREGIS IS STILL
--     ALIVE — xml-form scrape shows 59 references; the live query
--     contabilidad/qry/libro_diario_qry.xml joins
--     CTB_MOVIMIENTOS ← CTB_DETALLES → CTBREGIS via CTB_DETALLES.ctbregis_id
--     to pull the legacy ctbcod. proveedores_loc/ctbregis_update.xml,
--     clientes_local/ctbregis_update.xml, ordenpago/ordenpago_lote_ctbregis_*,
--     iva/iva_ventas_tauro_qry.xml, estadisticas/evolutivo_ctbregis_grupo.xml,
--     anulaciones/del_ctbmovimientos.xml etc. all read/write this table
--     directly. Not dead weight, not a duplicate of the new flow — the new
--     flow POINTS BACK at CTBREGIS for the legacy account code.
--
-- Shape from live Histrix (172.22.100.99, saldivia):
--   id_ctbregis INT AI PRI, siscod VARCHAR(2), regfec DATE (0000-00-00 for
--   122 rows), regmin MEDIUMINT, regtip TINYINT, regnro INT,
--   ctbcod VARCHAR(12), regdoh TINYINT (1=debe/2=haber, 8 edge rows =0),
--   regimp DECIMAL(13,2), regref VARCHAR(60), regfco DATE NULL,
--   regpoa VARCHAR(1) (A=604,457 / blank=122 / no P), coscod, impcod,
--   idcos, idimpu, regcta, regnpv, regord, regsub, reguni.
--
-- Target design: erp_accounting_registers keeps the full row for forensic /
-- audit use (the libro diario screens still read it). Two FKs are resolved
-- during migration:
--   - account_id via BuildCodeIndex("accounting","erp_accounts","code")
--     — NULLABLE because CTBREGIS has historical codes not in the current
--     plan de cuentas. Preserving account_code as varchar lets downstream
--     reconciliation work without losing the raw value.
--   - (cost center / imputation left as raw SMALLINTs for now — low read
--     value in the live XML scrape; revisit if a parity screen needs them.)

CREATE TABLE IF NOT EXISTS erp_accounting_registers (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             TEXT NOT NULL,
    legacy_id             BIGINT NOT NULL,
    subsystem_code        TEXT NOT NULL DEFAULT '',
    reg_date              DATE,                  -- NULL for 0000-00-00 rows (122)
    voucher_date          DATE,                  -- regfco (NULL in Histrix allowed)
    minuta_number         INTEGER NOT NULL DEFAULT 0,
    comprobante_type      SMALLINT NOT NULL DEFAULT 0,
    comprobante_number    INTEGER NOT NULL DEFAULT 0,
    account_code          TEXT NOT NULL DEFAULT '',
    account_id            UUID REFERENCES erp_accounts(id),  -- NULL when code not in plan
    entry_side            SMALLINT NOT NULL DEFAULT 0,       -- 1=debe, 2=haber, 0=edge
    amount                NUMERIC(16,2) NOT NULL DEFAULT 0,
    reference             TEXT NOT NULL DEFAULT '',
    status                TEXT NOT NULL DEFAULT '',          -- regpoa (A / blank)
    cost_center_code      SMALLINT NOT NULL DEFAULT 0,
    imputation_code       SMALLINT NOT NULL DEFAULT 0,
    legacy_cost_center_id INTEGER NOT NULL DEFAULT 0,
    legacy_imputation_id  INTEGER NOT NULL DEFAULT 0,
    legacy_account_id     INTEGER NOT NULL DEFAULT 0,
    post_number           SMALLINT NOT NULL DEFAULT 0,
    entry_order           SMALLINT NOT NULL DEFAULT 0,
    subdiary_code         SMALLINT NOT NULL DEFAULT 0,
    physical_units        NUMERIC(16,2) NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);

CREATE INDEX idx_erp_accounting_registers_date
    ON erp_accounting_registers(tenant_id, reg_date DESC)
    WHERE reg_date IS NOT NULL;
CREATE INDEX idx_erp_accounting_registers_minuta
    ON erp_accounting_registers(tenant_id, minuta_number);
CREATE INDEX idx_erp_accounting_registers_account_code
    ON erp_accounting_registers(tenant_id, account_code);
CREATE INDEX idx_erp_accounting_registers_account_id
    ON erp_accounting_registers(tenant_id, account_id)
    WHERE account_id IS NOT NULL;
CREATE INDEX idx_erp_accounting_registers_subsystem
    ON erp_accounting_registers(tenant_id, subsystem_code)
    WHERE subsystem_code <> '';

-- RLS (silo-compliant, same pattern as 070).
ALTER TABLE erp_accounting_registers ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_accounting_registers
    USING (tenant_id = current_setting('app.tenant_id', true));

-- No new permissions: reuses erp.accounting.read / erp.accounting.write
-- from migration 019. CTBREGIS is the same functional scope as journal
-- entries/lines from the user's point of view.
