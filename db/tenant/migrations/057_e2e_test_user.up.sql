-- 053_e2e_test_user.up.sql
-- Creates a fixed test user used by the end-to-end and API smoke suites
-- (apps/web/e2e/api/, apps/web/e2e/workstation/).
--
-- The hash below is bcrypt(testpassword123, cost=12). Generated with:
--   python3 -c "import bcrypt; print(bcrypt.hashpw(b'testpassword123', bcrypt.gensalt(12)).decode())"
-- Anyone reading this migration can log in as e2e-test@saldivia.local
-- with password 'testpassword123'. Acceptable: the user only exists in
-- dev/workstation tenants, not in any production tenant.
--
-- Linked to role-admin so the smoke suite can hit every ERP endpoint
-- without permission noise. RBAC-specific tests should create their own
-- scoped users instead of reusing this one.

INSERT INTO users (id, email, name, password_hash, is_active)
VALUES (
    'u-e2e-test',
    'e2e-test@saldivia.local',
    'E2E Test User',
    '$2b$12$0ztHvtq4n1HuN9u2ScrYl.KGzfqb6O50UaR3qJZ5qMawigzahcqAC',
    true
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
VALUES ('u-e2e-test', 'role-admin')
ON CONFLICT DO NOTHING;
