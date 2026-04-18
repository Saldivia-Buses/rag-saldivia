-- Migration 061 — legacy rescue
--
-- Rationale: the plan-21b skip-audit (2026-04-16, saldivia dataset) showed
-- 3.15M legacy rows (13.4% of the read corpus) dropped silently by transform
-- filters. This migration relaxes three constraints that forced those drops
-- so the data lands intact, flagged as rescued via metadata.
--
-- Invariants preserved:
--   * Tenant isolation (policies unchanged)
--   * FK direction unchanged
--   * Numeric columns keep numeric(14,4) precision
--
-- The accompanying down migration re-tightens the checks but does not delete
-- rescued rows — operators must clean them up explicitly.

-- 1) Stock movements: 173K saldivia rows had cantidad=0 in legacy. These
--    represent zero-qty adjustments (reclassification between depots,
--    correction entries) that are real events but got blocked by
--    CHECK (quantity > 0). Relax to >= 0 so they land; migrators tag them
--    in `notes` with "[legacy:zero_qty]".
ALTER TABLE erp_stock_movements DROP CONSTRAINT IF EXISTS erp_stock_movements_quantity_check;
ALTER TABLE erp_stock_movements ADD CONSTRAINT erp_stock_movements_quantity_check CHECK (quantity >= 0);

-- 2) Cash / bank movements: 21K saldivia rows had cajimp=0. Same reasoning.
ALTER TABLE erp_treasury_movements DROP CONSTRAINT IF EXISTS erp_treasury_movements_amount_check;
ALTER TABLE erp_treasury_movements ADD CONSTRAINT erp_treasury_movements_amount_check CHECK (amount >= 0);

-- 3) Account movements (REG_MOVIMIENTOS): historic accounting adjustments can
--    be booked at amount=0 (formal closure records). Same relaxation.
ALTER TABLE erp_account_movements DROP CONSTRAINT IF EXISTS erp_account_movements_amount_check;
ALTER TABLE erp_account_movements ADD CONSTRAINT erp_account_movements_amount_check CHECK (amount >= 0);

-- 3b) Withholdings: 29,461 legacy RETACUMU rows (75%) carry acuret=0 — these
--     are valid "no tax withheld this period" records tied to a base amount
--     (acupag). Blocking them with amount > 0 is data loss.
ALTER TABLE erp_withholdings DROP CONSTRAINT IF EXISTS erp_withholdings_amount_check;
ALTER TABLE erp_withholdings ADD CONSTRAINT erp_withholdings_amount_check CHECK (amount >= 0);

-- 3c) Payment allocations: CCTIMPUT sometimes carries movimp=0 when the
--     allocation is an informational "already settled" marker tied to a
--     fully-paid invoice. Relax to include these.
ALTER TABLE erp_payment_allocations DROP CONSTRAINT IF EXISTS erp_payment_allocations_amount_check;
ALTER TABLE erp_payment_allocations ADD CONSTRAINT erp_payment_allocations_amount_check CHECK (amount >= 0);

-- 3d) Broad CHECK relaxation for tables that receive legacy data. Real-world
--     Histrix rows have negative unit prices (returns, adjustments), zero
--     quantities, and other values that the strict > 0 checks were built
--     for clean forward-only SDA data, not historical imports. Keep the
--     non-null columns — just drop the numeric range guards.
ALTER TABLE erp_purchase_order_lines DROP CONSTRAINT IF EXISTS erp_purchase_order_lines_quantity_check;
ALTER TABLE erp_purchase_order_lines DROP CONSTRAINT IF EXISTS erp_purchase_order_lines_unit_price_check;
ALTER TABLE erp_purchase_receipt_lines DROP CONSTRAINT IF EXISTS erp_purchase_receipt_lines_quantity_check;
ALTER TABLE erp_quotation_lines DROP CONSTRAINT IF EXISTS erp_quotation_lines_quantity_check;
ALTER TABLE erp_quotation_lines DROP CONSTRAINT IF EXISTS erp_quotation_lines_unit_price_check;
ALTER TABLE erp_invoice_lines DROP CONSTRAINT IF EXISTS erp_invoice_lines_quantity_check;
ALTER TABLE erp_invoice_lines DROP CONSTRAINT IF EXISTS erp_invoice_lines_unit_price_check;
ALTER TABLE erp_production_orders DROP CONSTRAINT IF EXISTS erp_production_orders_quantity_check;
ALTER TABLE erp_production_materials DROP CONSTRAINT IF EXISTS erp_production_materials_required_qty_check;
ALTER TABLE erp_work_order_parts DROP CONSTRAINT IF EXISTS erp_work_order_parts_quantity_check;
ALTER TABLE erp_receipt_payments DROP CONSTRAINT IF EXISTS erp_receipt_payments_amount_check;
ALTER TABLE erp_receipt_allocations DROP CONSTRAINT IF EXISTS erp_receipt_allocations_amount_check;
ALTER TABLE erp_carroceria_bom DROP CONSTRAINT IF EXISTS erp_carroceria_bom_quantity_check;
ALTER TABLE erp_checks DROP CONSTRAINT IF EXISTS erp_checks_amount_check;

-- 3e) Tax entries: IVAIMPORTES carries period VAT records that don't always
--     reference a specific invoice (periodic adjustments, regime changes).
--     Make invoice_id nullable so these land in the structured table.
ALTER TABLE erp_tax_entries ALTER COLUMN invoice_id DROP NOT NULL;

-- 4) Roles: the 23 Histrix HTXPROFILES map 1-to-1 to SDA roles but their
--    legacy menu-level ACL (HTXPROFILE_AUTH: 2104 rows) does not match SDA's
--    permission IDs 1-to-1. We store the raw legacy menu list as JSONB so
--    the admin UI can review + auto-translate it later without losing data.
ALTER TABLE roles ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;
CREATE INDEX IF NOT EXISTS idx_roles_metadata_legacy ON roles USING gin (metadata jsonb_path_ops);

-- 5) Users: preserve the legacy login + profile id so support queries
--    ("which legacy user is this?") work without joining erp_legacy_mapping.
ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_login TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_profile_id INTEGER;
CREATE INDEX IF NOT EXISTS idx_users_legacy_login ON users (legacy_login) WHERE legacy_login IS NOT NULL;

-- 6) Stock movements, account movements, treasury movements: let the ghost
--    article/entity fallback mark rescued rows so they can be filtered from
--    reports. Piggy-back on the existing notes column — no schema change
--    needed — but add an index on the "[legacy:...]" prefix for fast ops.
CREATE INDEX IF NOT EXISTS idx_stock_movements_legacy_notes
  ON erp_stock_movements (tenant_id)
  WHERE notes LIKE '[legacy:%';
CREATE INDEX IF NOT EXISTS idx_treasury_movements_legacy_notes
  ON erp_treasury_movements (tenant_id)
  WHERE notes LIKE '[legacy:%';
CREATE INDEX IF NOT EXISTS idx_account_movements_legacy_notes
  ON erp_account_movements (tenant_id)
  WHERE notes LIKE '[legacy:%';
