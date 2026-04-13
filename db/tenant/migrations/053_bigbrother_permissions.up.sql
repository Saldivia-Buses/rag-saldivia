-- 053_bigbrother_permissions.up.sql
-- Add missing permissions for BigBrother service (should have been in 013).

INSERT INTO permissions (id, name, description, category) VALUES
    ('bigbrother.read',      'Ver dispositivos',       'Listar dispositivos, topología, eventos, stats',      'bigbrother'),
    ('bigbrother.admin',     'Admin BigBrother',       'Gestionar credenciales, scans, modos',                'bigbrother'),
    ('bigbrother.exec',      'Ejecutar comandos',      'Ejecutar comandos remotos en dispositivos',           'bigbrother'),
    ('bigbrother.plc.read',  'Leer registros PLC',     'Listar registros Modbus/OPC-UA',                      'bigbrother'),
    ('bigbrother.plc.write', 'Escribir registros PLC', 'Escribir valores en registros PLC (two-person rule)', 'bigbrother')
ON CONFLICT (id) DO NOTHING;

-- Grant all bigbrother permissions to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE category = 'bigbrother'
ON CONFLICT DO NOTHING;

-- Grant read-only to manager
INSERT INTO role_permissions (role_id, permission_id)
VALUES
    ('role-manager', 'bigbrother.read'),
    ('role-manager', 'bigbrother.plc.read')
ON CONFLICT DO NOTHING;
