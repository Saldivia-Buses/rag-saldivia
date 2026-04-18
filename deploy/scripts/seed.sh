#!/bin/bash
# SDA Framework — Dev Seed Data
# Creates test users, roles, and a platform tenant for local development.
#
# Usage: ./deploy/scripts/seed.sh
# Requires: migrate.sh has been run first

set -euo pipefail

PLATFORM_DB_URL="${PLATFORM_DB_URL:-postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable}"
TENANT_DB_URL="${TENANT_DB_URL:-postgres://sda:sda_dev@localhost:5432/sda_tenant_dev?sslmode=disable}"

log() { echo "[seed] $1"; }

# ── Platform DB — register the dev tenant ────────────────────────────────
log "seeding platform db..."
# DEV ONLY — production must use secrets manager for connection strings
psql "$PLATFORM_DB_URL" -v ON_ERROR_STOP=1 --quiet <<'SQL'
-- Dev tenant (matches docker-compose dev config)
INSERT INTO tenants (id, slug, name, plan_id, postgres_url, redis_url, settings)
VALUES (
    'dev-tenant-id',
    'dev',
    'Dev Tenant',
    'business',
    'postgres://sda:sda_dev@postgres:5432/sda_tenant_dev?sslmode=disable',
    'redis://redis:6379/1',
    '{"theme":"default"}'
) ON CONFLICT (slug) DO NOTHING;

-- Enable core modules for dev tenant
INSERT INTO tenant_modules (tenant_id, module_id, config, enabled_by)
VALUES
    ('dev-tenant-id', 'chat', '{}', 'seed'),
    ('dev-tenant-id', 'auth', '{}', 'seed'),
    ('dev-tenant-id', 'notifications', '{}', 'seed'),
    ('dev-tenant-id', 'docs', '{}', 'seed')
ON CONFLICT DO NOTHING;

-- Global feature flag for dev
INSERT INTO feature_flags (id, name, description, enabled)
VALUES ('ff-dev-mode', 'dev_mode', 'Enables dev tools and verbose logging', true)
ON CONFLICT DO NOTHING;
SQL

# ── Tenant DB — create dev users and roles ───────────────────────────────
log "seeding tenant db..."

# Password: "admin123" for admin/user, "testpassword123" for e2e-test — bcrypt cost 12.
# Passed via psql -v to avoid shell expansion of $ in bcrypt hashes (belt + suspenders:
# the heredoc is single-quoted too, so shell expansion is already disabled).
ADMIN_HASH='$2b$12$EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.'
USER_HASH='$2b$12$EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.'
# E2E_TEST_HASH migrated from db/tenant/migrations/057_e2e_test_user.up.sql — see
# ADR 027 F-02. Kept out of migrations so production silos never even briefly
# hold a publicly-documented admin credential.
E2E_TEST_HASH='$2b$12$0ztHvtq4n1HuN9u2ScrYl.KGzfqb6O50UaR3qJZ5qMawigzahcqAC'

psql "$TENANT_DB_URL" -v ON_ERROR_STOP=1 --quiet \
    -v admin_hash="$ADMIN_HASH" \
    -v user_hash="$USER_HASH" \
    -v e2e_test_hash="$E2E_TEST_HASH" <<'SQL'
-- Admin user (Enzo). Previously seeded by 058_admin_user.up.sql, moved here
-- to keep the hash out of production silos on cold-start.
INSERT INTO users (id, email, name, password_hash)
VALUES ('u-admin', 'admin@sda.local', 'Enzo Saldivia', :'admin_hash')
ON CONFLICT (email) DO NOTHING;

-- Regular user
INSERT INTO users (id, email, name, password_hash)
VALUES ('u-user', 'user@sda.local', 'Usuario Test', :'user_hash')
ON CONFLICT (email) DO NOTHING;

-- E2E test user (apps/web/e2e/api/, apps/web/e2e/workstation/). Previously
-- seeded by 057_e2e_test_user.up.sql, moved here for the same reason.
INSERT INTO users (id, email, name, password_hash, is_active)
VALUES ('u-e2e-test', 'e2e-test@saldivia.local', 'E2E Test User', :'e2e_test_hash', true)
ON CONFLICT (email) DO NOTHING;

-- Assign roles
INSERT INTO user_roles (user_id, role_id)
VALUES ('u-admin', 'role-admin')
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
VALUES ('u-user', 'role-user')
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
VALUES ('u-e2e-test', 'role-admin')
ON CONFLICT DO NOTHING;
SQL

log "=== Done ==="
log ""
log "Dev credentials:"
log "  admin@sda.local          / admin123         (role: admin)"
log "  user@sda.local           / admin123         (role: user)"
log "  e2e-test@saldivia.local  / testpassword123  (role: admin, e2e suite)"
