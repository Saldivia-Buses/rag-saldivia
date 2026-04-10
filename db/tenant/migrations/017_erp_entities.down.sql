-- 017_erp_entities.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'erp.entities.%';
DELETE FROM permissions WHERE id LIKE 'erp.entities.%';

DROP POLICY IF EXISTS tenant_isolation ON erp_entity_notes;
DROP POLICY IF EXISTS tenant_isolation ON erp_entity_relations;
DROP POLICY IF EXISTS tenant_isolation ON erp_entity_documents;
DROP POLICY IF EXISTS tenant_isolation ON erp_entity_contacts;
DROP POLICY IF EXISTS tenant_isolation ON erp_entities;

DROP TABLE IF EXISTS erp_entity_notes;
DROP TABLE IF EXISTS erp_entity_relations;
DROP TABLE IF EXISTS erp_entity_documents;
DROP TABLE IF EXISTS erp_entity_contacts;
DROP TABLE IF EXISTS erp_entities;
