#!/bin/bash
# SDA Framework — Database Migration Runner
# Applies migration files from the centralized db/ directory.
# Tracks applied migrations in a schema_migrations table to prevent re-application.
#
# Usage:
#   ./deploy/scripts/migrate.sh
#   PLATFORM_DB_URL=... TENANT_DB_URL=... ./deploy/scripts/migrate.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -d "/db" ]; then
    DB_DIR="/db"
else
    DB_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)/db"
fi

PLATFORM_DB_URL="${PLATFORM_DB_URL:-postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable}"
TENANT_DB_URL="${TENANT_DB_URL:-postgres://sda:sda_dev@localhost:5432/sda_tenant_dev?sslmode=disable}"

log() { echo "[migrate] $1"; }

# Create schema_migrations tracking table if it doesn't exist
ensure_tracking() {
    local db_url="$1"
    psql "$db_url" --quiet -v ON_ERROR_STOP=1 -c "CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())" < /dev/null
}

# Apply a migration if not already applied
apply_migration() {
    local db_url="$1"
    local file="$2"
    local filename
    filename=$(basename "$file")

    # Check if already applied (use psql -v to avoid SQL injection)
    local applied
    applied=$(psql "$db_url" -t -v "mig_file=$filename" \
        -c "SELECT 1 FROM schema_migrations WHERE filename = :'mig_file'" < /dev/null 2>/dev/null | tr -d ' ')

    if [ "$applied" = "1" ]; then
        return 0
    fi

    log "applying $filename → $(echo "$db_url" | sed 's|.*@||; s|?.*||')"

    # Apply migration inside a transaction
    psql "$db_url" -v ON_ERROR_STOP=1 --quiet --single-transaction -f "$file" < /dev/null

    # Record as applied
    psql "$db_url" --quiet -v "mig_file=$filename" \
        -c "INSERT INTO schema_migrations (filename) VALUES (:'mig_file') ON CONFLICT DO NOTHING" < /dev/null
}

# ── Platform DB ──────────────────────────────────────────────────────────
log "=== Platform DB ==="
ensure_tracking "$PLATFORM_DB_URL"
for f in "$DB_DIR"/platform/migrations/*.up.sql; do
    [ -f "$f" ] && apply_migration "$PLATFORM_DB_URL" "$f"
done

# ── Tenant DB ────────────────────────────────────────────────────────────
log "=== Tenant DB ==="
ensure_tracking "$TENANT_DB_URL"
for f in "$DB_DIR"/tenant/migrations/*.up.sql; do
    [ -f "$f" ] && apply_migration "$TENANT_DB_URL" "$f"
done

log "=== Done ==="
