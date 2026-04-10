-- 027_erp_quality.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.quality.%';
DELETE FROM permissions WHERE id LIKE 'erp.quality.%';
DROP POLICY IF EXISTS tenant_isolation ON erp_controlled_documents;
DROP POLICY IF EXISTS tenant_isolation ON erp_audit_findings;
DROP POLICY IF EXISTS tenant_isolation ON erp_audits;
DROP POLICY IF EXISTS tenant_isolation ON erp_corrective_actions;
DROP POLICY IF EXISTS tenant_isolation ON erp_nonconformities;
DROP TABLE IF EXISTS erp_controlled_documents;
DROP TABLE IF EXISTS erp_audit_findings;
DROP TABLE IF EXISTS erp_audits;
DROP TABLE IF EXISTS erp_corrective_actions;
DROP TABLE IF EXISTS erp_nonconformities;
