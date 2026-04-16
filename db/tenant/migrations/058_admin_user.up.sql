-- 058_admin_user.up.sql
-- Seeds the canonical admin user every tenant gets out of the box. The same
-- user used to live only in deploy/scripts/seed.sh, which the workstation
-- deploy never runs — meaning admin@sda.local effectively did not exist on
-- production-style installs and operators couldn't get in without first
-- shelling onto the host.
--
-- Password is 'admin123' (bcrypt cost 12, matching deploy/scripts/seed.sh).
-- Same caveat as 057: only acceptable because dev/workstation tenants are
-- behind a VPN. Rotate via the user-update API before exposing publicly.

INSERT INTO users (id, email, name, password_hash, is_active)
VALUES (
    'u-admin',
    'admin@sda.local',
    'Enzo Saldivia',
    '$2b$12$EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.',
    true
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
VALUES ('u-admin', 'role-admin')
ON CONFLICT DO NOTHING;
