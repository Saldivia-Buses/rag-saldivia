#!/usr/bin/env bash
# postgres-init: first-boot scaffolding for the per-tenant Postgres.
#
# Runs once per container start, before the `postgres` longrun. Idempotent:
# if $PGDATA/PG_VERSION exists, only touches perms + socket dir and exits.
#
# Per ADR 023, Postgres in the all-in-one container listens on a Unix socket
# only — the rest of the app talks to it via /var/run/postgresql.

set -euo pipefail

PGDATA="${PGDATA:-/data/postgres}"
PGBIN="/usr/lib/postgresql/16/bin"
PGSOCK_DIR="/var/run/postgresql"
PGUID="$(id -u postgres)"
PGGID="$(id -g postgres)"

mkdir -p "$PGSOCK_DIR"
chown "$PGUID:$PGGID" "$PGSOCK_DIR"
chmod 2775 "$PGSOCK_DIR"

mkdir -p "$(dirname "$PGDATA")"

if [[ ! -s "$PGDATA/PG_VERSION" ]]; then
    echo "postgres-init: initialising cluster at $PGDATA"
    mkdir -p "$PGDATA"
    chown "$PGUID:$PGGID" "$PGDATA"
    chmod 700 "$PGDATA"
    # Explicit auth methods: local socket connections trust the uid (the
    # container is our isolation boundary — ADR 023), but any TCP listener
    # that someone might re-enable later requires a real password. Be
    # defensive: someone else editing postgresql.conf shouldn't also have
    # to remember to tighten pg_hba.conf.
    s6-setuidgid postgres "$PGBIN/initdb" \
        --encoding=UTF8 \
        --locale=C.UTF-8 \
        --username=postgres \
        -D "$PGDATA" \
        --auth-local=trust \
        --auth-host=scram-sha-256

    # Single-tenant, local-only hardening. The container is the isolation
    # boundary; Postgres listens on the Unix socket only.
    cat >> "$PGDATA/postgresql.conf" <<'EOF'

# --- all-in-one tenant overrides (ADR 023) ---
listen_addresses = ''
unix_socket_directories = '/var/run/postgresql'
shared_buffers = 256MB
max_connections = 100
log_timezone = 'UTC'
timezone = 'UTC'
EOF
else
    echo "postgres-init: cluster already present at $PGDATA"
    chown -R "$PGUID:$PGGID" "$PGDATA"
fi
