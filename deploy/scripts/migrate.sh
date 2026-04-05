#!/bin/bash
# SDA Framework — Database Migration Runner
# Applies all migration files in order to the correct databases.
#
# Usage:
#   ./deploy/scripts/migrate.sh                    # defaults to dev
#   PLATFORM_DB_URL=... TENANT_DB_URL=... ./deploy/scripts/migrate.sh
#
# Platform DB gets: platform migrations
# Tenant DB gets: auth → chat → notification → ingest (order matters for FK deps)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# Support both host (../../services) and container (/services) paths
if [ -d "/services" ]; then
    ROOT_DIR=""
    SERVICES_DIR="/services"
else
    ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
    SERVICES_DIR="$ROOT_DIR/services"
fi

# Database URLs (defaults for dev)
PLATFORM_DB_URL="${PLATFORM_DB_URL:-postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable}"
TENANT_DB_URL="${TENANT_DB_URL:-postgres://sda:sda_dev@localhost:5432/sda_tenant_dev?sslmode=disable}"

log() { echo "[migrate] $1"; }

run_sql() {
    local db_url="$1"
    local file="$2"
    local name
    name=$(basename "$(dirname "$(dirname "$(dirname "$file")")")")
    log "applying $name/$(basename "$file") → $(echo "$db_url" | sed 's|.*@||; s|?.*||')"
    psql "$db_url" -f "$file" -v ON_ERROR_STOP=1 --quiet
}

# ── Platform DB ──────────────────────────────────────────────────────────
log "=== Platform DB ==="
for f in "$SERVICES_DIR"/platform/db/migrations/*.up.sql; do
    [ -f "$f" ] && run_sql "$PLATFORM_DB_URL" "$f"
done

# ── Tenant DB ────────────────────────────────────────────────────────────
# Order matters: auth first (creates users table), then services with FK deps
TENANT_MIGRATION_ORDER=(auth chat notification ingest)

log "=== Tenant DB ==="
for svc in "${TENANT_MIGRATION_ORDER[@]}"; do
    for f in "$SERVICES_DIR"/"$svc"/db/migrations/*.up.sql; do
        [ -f "$f" ] && run_sql "$TENANT_DB_URL" "$f"
    done
done

log "=== Done ==="
