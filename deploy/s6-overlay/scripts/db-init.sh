#!/usr/bin/env bash
# db-init: create app databases and apply migrations before the app starts.
#
# Runs as a oneshot after the postgres longrun comes up. Idempotent: if the
# databases already exist migrate.sh only applies what is missing (it tracks
# applied files in schema_migrations). Unix-socket-only per ADR 023 — the
# container is our isolation boundary, so there is no password.

set -euo pipefail

PGHOST="${PGHOST:-/var/run/postgresql}"
PGUSER="${PGUSER:-postgres}"
PLATFORM_DB="${PLATFORM_DB:-sda_platform}"
TENANT_DB="${TENANT_DB:-sda_tenant_dev}"

log() { echo "[db-init] $1"; }

# Wait up to 30 s for Postgres to accept connections. On a cold-start /data
# volume the postgres longrun needs ~20 s for initdb; on a warm volume it
# is sub-second. Either is covered.
tries=0
until pg_isready -h "$PGHOST" -U "$PGUSER" -q; do
    tries=$((tries + 1))
    if [[ $tries -ge 60 ]]; then
        log "postgres not ready after 30 s" >&2
        exit 1
    fi
    sleep 0.5
done

for db in "$PLATFORM_DB" "$TENANT_DB"; do
    exists=$(psql -h "$PGHOST" -U "$PGUSER" -tAc \
        "SELECT 1 FROM pg_database WHERE datname='$db'")
    if [[ "$exists" != "1" ]]; then
        log "creating database $db"
        createdb -h "$PGHOST" -U "$PGUSER" "$db"
    fi
done

export PLATFORM_DB_URL="postgres:///${PLATFORM_DB}?host=${PGHOST}&user=${PGUSER}&sslmode=disable"
export TENANT_DB_URL="postgres:///${TENANT_DB}?host=${PGHOST}&user=${PGUSER}&sslmode=disable"

log "applying migrations"
exec /etc/s6-overlay/scripts/migrate.sh
