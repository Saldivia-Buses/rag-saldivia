-- 068_drop_default_admin_users.up.sql
-- Remove the publicly-documented admin backdoors seeded by 057 / 058.
--
-- Migrations 057_e2e_test_user and 058_admin_user unconditionally INSERT two
-- role-admin users with bcrypt hashes of known plaintext passwords
-- ('testpassword123' / 'admin123') that are committed to git. The migration
-- comments rationalised the risk as "dev/workstation only behind VPN" and
-- "rotate before exposing publicly" — but deploy/scripts/migrate.sh has no
-- env gate, and deploy/s6-overlay/scripts/db-init.sh runs every tenant
-- migration on every container cold-start. Under the ADR 022/023 one-silo-
-- per-tenant model this ships the backdoors to every production tenant the
-- moment Traefik routes to :443.
--
-- Strategy:
--   DELETE only when password_hash matches the committed default. An operator
--   who rotated the password via the user-update API has a different hash, so
--   their legitimate admin stays untouched — INCLUDING its role binding.
--
-- Both DELETEs (user_roles + users) must be gated on the SAME hash check,
-- or we'd strip the role from a rotated admin while leaving the user row.
-- Use a CTE-less correlated EXISTS so both statements agree atomically.
--
-- After this migration, future fresh silos apply the whole chain in order:
-- 057 seeds u-e2e-test, 058 seeds u-admin, 068 deletes both — so the user
-- rows never outlast the transaction-adjacent gap. A separate follow-up PR
-- should move the e2e/admin seeds out of `db/tenant/migrations/` entirely
-- (into `deploy/scripts/seed.sh`, which only runs for dev/workstation) so
-- no new silo ever even briefly holds the backdoor.

-- u-e2e-test (see 057).
DELETE FROM user_roles
WHERE user_id = 'u-e2e-test'
  AND EXISTS (
    SELECT 1 FROM users u
    WHERE u.id = 'u-e2e-test'
      AND u.password_hash = '$2b$12$0ztHvtq4n1HuN9u2ScrYl.KGzfqb6O50UaR3qJZ5qMawigzahcqAC'
  );

DELETE FROM users
WHERE id = 'u-e2e-test'
  AND password_hash = '$2b$12$0ztHvtq4n1HuN9u2ScrYl.KGzfqb6O50UaR3qJZ5qMawigzahcqAC';

-- u-admin (see 058).
DELETE FROM user_roles
WHERE user_id = 'u-admin'
  AND EXISTS (
    SELECT 1 FROM users u
    WHERE u.id = 'u-admin'
      AND u.password_hash = '$2b$12$EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.'
  );

DELETE FROM users
WHERE id = 'u-admin'
  AND password_hash = '$2b$12$EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.';
