DELETE FROM role_permissions WHERE permission_id IN (
    'erp.manufacturing.read',
    'erp.manufacturing.write',
    'erp.manufacturing.control',
    'erp.manufacturing.certify'
);
DELETE FROM permissions WHERE id IN (
    'erp.manufacturing.read',
    'erp.manufacturing.write',
    'erp.manufacturing.control',
    'erp.manufacturing.certify'
);
DROP TABLE IF EXISTS erp_carroceria_models;
DROP TABLE IF EXISTS erp_chassis_models;
DROP TABLE IF EXISTS erp_chassis_brands;
