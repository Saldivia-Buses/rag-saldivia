-- 033_erp_qc_reception.down.sql
DROP TABLE IF EXISTS erp_supplier_demerits;
DROP TABLE IF EXISTS erp_qc_inspections;
DELETE FROM role_permissions WHERE permission_id = 'erp.purchasing.inspect';
DELETE FROM permissions WHERE id = 'erp.purchasing.inspect';
