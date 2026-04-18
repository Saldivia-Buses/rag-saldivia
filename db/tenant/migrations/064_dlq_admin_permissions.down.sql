-- 064_dlq_admin_permissions.down.sql
DELETE FROM role_permissions WHERE permission_id LIKE 'admin.dlq.%';
DELETE FROM permissions WHERE id LIKE 'admin.dlq.%';
