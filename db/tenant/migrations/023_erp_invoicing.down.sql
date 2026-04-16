-- 023_erp_invoicing.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.invoicing.%';
DELETE FROM permissions WHERE id LIKE 'erp.invoicing.%';
DROP TRIGGER IF EXISTS trg_invoice_immutable ON erp_invoices;
DROP POLICY IF EXISTS tenant_isolation ON erp_withholdings;
DROP POLICY IF EXISTS tenant_isolation ON erp_tax_entries;
DROP POLICY IF EXISTS tenant_isolation ON erp_invoice_lines;
DROP POLICY IF EXISTS tenant_isolation ON erp_invoices;
DROP TABLE IF EXISTS erp_withholdings;
DROP TABLE IF EXISTS erp_tax_entries;
DROP TABLE IF EXISTS erp_invoice_lines;
DROP TABLE IF EXISTS erp_invoices;
