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

# Password: "admin123" — bcrypt hash (cost 12)
ADMIN_HASH='$2b$12$EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.'
USER_HASH='$2b$12$EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.'

psql "$TENANT_DB_URL" -v ON_ERROR_STOP=1 --quiet <<SQL
-- Admin user (Enzo)
INSERT INTO users (id, email, name, password_hash)
VALUES ('u-admin', 'admin@sda.local', 'Enzo Saldivia', '$ADMIN_HASH')
ON CONFLICT (email) DO NOTHING;

-- Regular user
INSERT INTO users (id, email, name, password_hash)
VALUES ('u-user', 'user@sda.local', 'Usuario Test', '$USER_HASH')
ON CONFLICT (email) DO NOTHING;

-- Assign roles
INSERT INTO user_roles (user_id, role_id)
VALUES ('u-admin', 'role-admin')
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
VALUES ('u-user', 'role-user')
ON CONFLICT DO NOTHING;
SQL

log "=== Done ==="
log ""
log "Dev credentials:"
log "  admin@sda.local / admin123  (role: admin)"
log "  user@sda.local  / admin123  (role: user)"
