-- 064_dlq_admin_permissions.up.sql
-- Seeds the DLQ admin permissions introduced by platform migration
-- 010_dead_events. Lives in the tenant DB because `permissions` and
-- `role_permissions` are tenant tables (created in 001_auth_init).
-- Assigned to role-admin by default — consistent with every other
-- admin.* permission in the tenant bootstrap.

INSERT INTO permissions (id, name, description, category) VALUES
    ('admin.dlq.read',   'Ver dead events',    'Listar eventos muertos del DLQ',                         'admin'),
    ('admin.dlq.replay', 'Replay dead events', 'Re-publicar eventos muertos al subject original',        'admin'),
    ('admin.dlq.drop',   'Drop dead events',   'Descartar eventos muertos definitivamente',              'admin')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'admin.dlq.%'
ON CONFLICT DO NOTHING;
