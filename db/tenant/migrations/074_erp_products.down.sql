-- 074_erp_products.down.sql — reverse of 074_erp_products.up.sql

DELETE FROM role_permissions WHERE permission_id LIKE 'erp.products.%';
DELETE FROM permissions WHERE id LIKE 'erp.products.%';

DROP TABLE IF EXISTS erp_product_attribute_homologations;
DROP TABLE IF EXISTS erp_product_attribute_values;
DROP TABLE IF EXISTS erp_product_attribute_options;
DROP TABLE IF EXISTS erp_product_attributes;
DROP TABLE IF EXISTS erp_products;
DROP TABLE IF EXISTS erp_product_sections;
