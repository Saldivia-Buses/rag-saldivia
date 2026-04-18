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

# Apply a migration if not already applied. psql's :'var' client-side
# substitution does not expand reliably under `-c` (confirmed against
# psql 16 inside the all-in-one image — the colon was forwarded verbatim
# and the server raised a syntax error). Shell-interpolate the filename
# into a SQL literal instead; migration filenames are globbed from
# db/*/migrations/*.up.sql so we own the value, but defend anyway by
# escaping single quotes.
apply_migration() {
    local db_url="$1"
    local file="$2"
    local filename
    filename=$(basename "$file")
    local sql_filename="${filename//\'/\'\'}"

    local applied
    applied=$(psql "$db_url" -tA \
        -c "SELECT 1 FROM schema_migrations WHERE filename = '$sql_filename'" \
        < /dev/null 2>/dev/null)

    if [ "$applied" = "1" ]; then
        return 0
    fi

    log "applying $filename → $(echo "$db_url" | sed 's|.*@||; s|?.*||')"

    psql "$db_url" -v ON_ERROR_STOP=1 --quiet --single-transaction -f "$file" < /dev/null

    psql "$db_url" --quiet \
        -c "INSERT INTO schema_migrations (filename) VALUES ('$sql_filename') ON CONFLICT DO NOTHING" \
        < /dev/null
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
