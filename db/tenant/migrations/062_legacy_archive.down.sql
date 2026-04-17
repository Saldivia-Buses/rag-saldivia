DROP INDEX IF EXISTS uq_erp_legacy_archive_natural;
DROP INDEX IF EXISTS idx_erp_legacy_archive_pk_num;
DROP INDEX IF EXISTS idx_erp_legacy_archive_tenant_table;
DROP POLICY IF EXISTS tenant_isolation ON erp_legacy_archive;
DROP TABLE IF EXISTS erp_legacy_archive;
