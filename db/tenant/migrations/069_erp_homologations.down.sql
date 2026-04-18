-- 069_erp_homologations.down.sql — reverse of 069_erp_homologations.up.sql

DELETE FROM role_permissions WHERE permission_id LIKE 'erp.homologations.%';
DELETE FROM permissions WHERE id LIKE 'erp.homologations.%';

DROP TABLE IF EXISTS erp_homologation_revision_lines;
DROP TABLE IF EXISTS erp_homologation_revisions;
DROP TABLE IF EXISTS erp_homologations;
