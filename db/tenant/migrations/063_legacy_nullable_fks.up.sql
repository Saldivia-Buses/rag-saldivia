-- Migration 063 — nullable legacy FK columns
--
-- Continues the legacy-rescue thread started in 061/062: any SDA table that
-- accepts imported Histrix data and has a NOT NULL FK that the legacy row
-- may not resolve (order deleted, article retired, unit scrapped) gets the
-- constraint relaxed so the row lands instead of being dropped.
--
-- Empirical trigger: OCPRECIB (purchase receipts, 320K rows) failed prod
-- migration because some rows don't have an `idPed` that resolves to
-- PEDIDOINT. Same pattern is expected in production_inspections,
-- production_materials, production_steps, purchase_order_lines, etc. — all
-- rescue-style relaxations.
--
-- Where a NULL FK would leave a row meaningless (erp_invoice_lines without
-- invoice_id, erp_bom without both parent/child), we leave the constraint
-- intact — those must either be rescued via ghost parents or land in the
-- archive.

ALTER TABLE erp_purchase_receipts      ALTER COLUMN order_id      DROP NOT NULL;
ALTER TABLE erp_production_inspections ALTER COLUMN order_id      DROP NOT NULL;
ALTER TABLE erp_production_materials   ALTER COLUMN order_id      DROP NOT NULL;
ALTER TABLE erp_production_steps       ALTER COLUMN order_id      DROP NOT NULL;
ALTER TABLE erp_production_orders      ALTER COLUMN product_id    DROP NOT NULL;
ALTER TABLE erp_purchase_order_lines   ALTER COLUMN order_id      DROP NOT NULL;
ALTER TABLE erp_quotation_lines        ALTER COLUMN quotation_id  DROP NOT NULL;
ALTER TABLE erp_cnrt_work_orders       ALTER COLUMN unit_id       DROP NOT NULL;
ALTER TABLE erp_manufacturing_certificates ALTER COLUMN unit_id   DROP NOT NULL;
ALTER TABLE erp_manufacturing_lcm      ALTER COLUMN unit_id       DROP NOT NULL;
ALTER TABLE erp_production_controls    ALTER COLUMN unit_id       DROP NOT NULL;
ALTER TABLE erp_unit_photos            ALTER COLUMN unit_id       DROP NOT NULL;
ALTER TABLE erp_receipt_allocations    ALTER COLUMN invoice_id    DROP NOT NULL;
