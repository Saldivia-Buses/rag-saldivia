-- 076_erp_bank_imports.up.sql
-- Phase 1 §Data migration: BCS_IMPORTACION (rank 1 of the remaining
-- long tail post-Pareto #8, 91,959 rows live / 84,492 scrape).
--
-- Histrix context:
--   - BCS_IMPORTACION is bank-statement import staging: rows arrive
--     from supplier CSV/XLS dumps of bank account movements and sit
--     here until the concil screens match them against internal
--     REG_MOVIMIENTOS (treasury moves). cod_movimiento + nombre_concepto
--     capture the bank's own concept code (CHEQUE CAMARA, TRANSF, etc.)
--     with debit/credit/saldo per line.
--   - Live UI forms in bancos_local/: bcs_importacion_qry (main view),
--     bcsmovim_importacion_auto_ins (insert from import file),
--     bcsmovim_conci_auto (auto-conciliation), bcsmovim_borrar_auto_ins
--     (mass delete), qry/noacreditados (pending credits).
--   - Resolves two FKs to surfaces already migrated:
--       regmovim_id  → erp_invoices / treasury movement (via
--                      BuildRegMovimIndex from Phase 6, which pulled
--                      IVAVENTAS + IVACOMPRAS into the index).
--       nro_cuenta   → erp_entities (via BuildNroCuentaIndex from
--                      Phase 2's REG_CUENTA AfterTableHook).

CREATE TABLE IF NOT EXISTS erp_bank_imports (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             TEXT NOT NULL,
    legacy_id             BIGINT NOT NULL,              -- id_importacion
    movement_date         DATE,                         -- fecha_movimiento (NULL for zero dates)
    concept_name          TEXT NOT NULL DEFAULT '',     -- nombre_concepto
    movement_no           INTEGER NOT NULL DEFAULT 0,   -- nro_movimiento
    amount                NUMERIC(14,2) NOT NULL DEFAULT 0, -- importe
    debit                 NUMERIC(14,2) NOT NULL DEFAULT 0,
    credit                NUMERIC(14,2) NOT NULL DEFAULT 0,
    balance               NUMERIC(14,2) NOT NULL DEFAULT 0, -- saldo
    movement_code         TEXT NOT NULL DEFAULT '',     -- cod_movimiento
    treasury_movement_id  UUID REFERENCES erp_invoices(id),  -- regmovim_id resolved (nullable)
    treasury_legacy_id    INTEGER NOT NULL DEFAULT 0,   -- regmovim_id raw preserved
    imported_at           TIMESTAMPTZ,                  -- importado (original import timestamp)
    account_number        INTEGER NOT NULL DEFAULT 0,   -- nro_cuenta raw
    account_entity_id     UUID REFERENCES erp_entities(id), -- nullable, resolved via nro_cuenta index
    processed             INTEGER NOT NULL DEFAULT 0,   -- procesado flag (1=matched, 2=unmatched, etc.)
    comments              TEXT NOT NULL DEFAULT '',
    internal_no           INTEGER NOT NULL DEFAULT 0,   -- nro_interno
    branch                TEXT NOT NULL DEFAULT '',     -- sucursal
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, legacy_id)
);

CREATE INDEX idx_erp_bank_imports_date
    ON erp_bank_imports(tenant_id, movement_date DESC)
    WHERE movement_date IS NOT NULL;
CREATE INDEX idx_erp_bank_imports_account
    ON erp_bank_imports(tenant_id, account_number)
    WHERE account_number > 0;
CREATE INDEX idx_erp_bank_imports_account_entity
    ON erp_bank_imports(tenant_id, account_entity_id)
    WHERE account_entity_id IS NOT NULL;
CREATE INDEX idx_erp_bank_imports_treasury
    ON erp_bank_imports(tenant_id, treasury_movement_id)
    WHERE treasury_movement_id IS NOT NULL;
CREATE INDEX idx_erp_bank_imports_processed
    ON erp_bank_imports(tenant_id, processed, movement_date DESC);

ALTER TABLE erp_bank_imports ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_bank_imports
    USING (tenant_id = current_setting('app.tenant_id', true));

-- No new permissions: reuses erp.treasury.read / erp.treasury.write
-- (bank-statement import lives within the treasury scope).
