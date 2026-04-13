-- 054_migration_schema_fixes.up.sql
-- Schema fixes required by Plan 21B (full legacy migration).
-- Addresses 7 Critical findings from gateway review.

-- =============================================================================
-- C1: Expand erp_invoices.invoice_type CHECK for AFIP comprobante variants
-- =============================================================================
ALTER TABLE erp_invoices DROP CONSTRAINT IF EXISTS erp_invoices_invoice_type_check;
ALTER TABLE erp_invoices ADD CONSTRAINT erp_invoices_invoice_type_check
    CHECK (invoice_type IN (
        'invoice_a','invoice_b','invoice_c','invoice_e',
        'credit_note_a','credit_note_b','credit_note_c',
        'credit_note',
        'debit_note_a','debit_note_b','debit_note_c',
        'debit_note',
        'delivery_note'
    ));

-- =============================================================================
-- C2: Create erp_bom_history for versioned BOM records (3.3M rows)
-- erp_bom keeps current BOM with UNIQUE(tenant_id, parent_id, child_id).
-- History gets its own table without that constraint.
-- =============================================================================
CREATE TABLE IF NOT EXISTS erp_bom_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    parent_id     UUID NOT NULL REFERENCES erp_articles(id),
    child_id      UUID NOT NULL REFERENCES erp_articles(id),
    quantity      NUMERIC(14,4) NOT NULL,
    unit_id       UUID REFERENCES erp_catalogs(id),
    version       INT NOT NULL DEFAULT 1,
    effective_date DATE,
    replaced_date  DATE,
    legacy_id     BIGINT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_bom_hist_tenant ON erp_bom_history(tenant_id);
CREATE INDEX idx_erp_bom_hist_parent ON erp_bom_history(tenant_id, parent_id);
CREATE INDEX idx_erp_bom_hist_dates ON erp_bom_history(tenant_id, effective_date);

ALTER TABLE erp_bom_history ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_bom_history
    USING (tenant_id = current_setting('app.tenant_id', true));

-- =============================================================================
-- C3: Add metadata JSONB to 6 tables that need it for legacy data enrichment
-- =============================================================================
ALTER TABLE erp_account_movements
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';

ALTER TABLE erp_controlled_documents
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';

ALTER TABLE erp_audits
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';

ALTER TABLE erp_quotations
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';

ALTER TABLE erp_production_orders
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';

ALTER TABLE erp_production_inspections
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';

-- =============================================================================
-- C5: Make supplier_id nullable for internal requisitions (PEDIDOINT)
-- =============================================================================
ALTER TABLE erp_purchase_orders
    ALTER COLUMN supplier_id DROP NOT NULL;

-- =============================================================================
-- C6: Add force_password_reset to users table for legacy migration
-- =============================================================================
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS force_password_reset BOOLEAN NOT NULL DEFAULT false;

-- =============================================================================
-- H1: Expand treasury status CHECK to include 'cancelled' for legacy anulados
-- Semantically distinct from 'reversed' (reversed implies accounting reversal,
-- cancelled means voided before confirmation).
-- =============================================================================
ALTER TABLE erp_treasury_movements DROP CONSTRAINT IF EXISTS erp_treasury_movements_status_check;
ALTER TABLE erp_treasury_movements ADD CONSTRAINT erp_treasury_movements_status_check
    CHECK (status IN ('pending', 'confirmed', 'reversed', 'cancelled'));

-- =============================================================================
-- H5: Add number generation helper comment (number is NOT NULL, migrators
-- must generate MOV-{legacy_id} for treasury movements)
-- =============================================================================
COMMENT ON COLUMN erp_treasury_movements.number IS
    'Unique movement number. Legacy migration uses MOV-{legacy_id} format.';

-- =============================================================================
-- Risk agents catalog for safety exposures (C7 dependency)
-- RIESGOS (212 rows) maps to erp_risk_agents which already exists.
-- No schema change needed, just catalog data migration.
-- =============================================================================
