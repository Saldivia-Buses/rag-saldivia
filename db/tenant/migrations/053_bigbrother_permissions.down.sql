-- 053_bigbrother_permissions.down.sql

DELETE FROM role_permissions WHERE permission_id IN (
    'bigbrother.read', 'bigbrother.admin', 'bigbrother.exec',
    'bigbrother.plc.read', 'bigbrother.plc.write'
);

DELETE FROM permissions WHERE category = 'bigbrother';
