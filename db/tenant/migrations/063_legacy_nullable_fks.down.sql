-- Revert the legacy NULL tolerance. Any row with a NULL FK must be cleaned
-- before running this down migration or the ALTER will fail.

ALTER TABLE erp_purchase_receipts      ALTER COLUMN order_id      SET NOT NULL;
ALTER TABLE erp_production_inspections ALTER COLUMN order_id      SET NOT NULL;
ALTER TABLE erp_production_materials   ALTER COLUMN order_id      SET NOT NULL;
ALTER TABLE erp_production_steps       ALTER COLUMN order_id      SET NOT NULL;
ALTER TABLE erp_production_orders      ALTER COLUMN product_id    SET NOT NULL;
ALTER TABLE erp_purchase_order_lines   ALTER COLUMN order_id      SET NOT NULL;
ALTER TABLE erp_quotation_lines        ALTER COLUMN quotation_id  SET NOT NULL;
ALTER TABLE erp_cnrt_work_orders       ALTER COLUMN unit_id       SET NOT NULL;
ALTER TABLE erp_manufacturing_certificates ALTER COLUMN unit_id   SET NOT NULL;
ALTER TABLE erp_manufacturing_lcm      ALTER COLUMN unit_id       SET NOT NULL;
ALTER TABLE erp_production_controls    ALTER COLUMN unit_id       SET NOT NULL;
ALTER TABLE erp_unit_photos            ALTER COLUMN unit_id       SET NOT NULL;
ALTER TABLE erp_receipt_allocations    ALTER COLUMN invoice_id    SET NOT NULL;
