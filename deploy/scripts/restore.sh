#!/usr/bin/env bash
# SDA Framework — Restore Script
#
# Restores a backup created by backup.sh.
# Run manually from the workstation — NOT deployed to containers.
#
# Usage: ./restore.sh <backup-date> [tenant-slug]
#   ./restore.sh 2026-04-05           # restore all (platform + all tenants)
#   ./restore.sh 2026-04-05 saldivia  # restore single tenant
#
# Prerequisites:
#   - psql, age (if backups encrypted), mc (MinIO client)
#   - BACKUP_ENCRYPTION_KEY: path to age identity (private key) file
#   - MINIO_ALIAS: mc alias configured

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <backup-date> [tenant-slug]"
    echo "  backup-date: YYYY-MM-DD (e.g., 2026-04-05)"
    echo "  tenant-slug: optional, restore single tenant only"
    exit 1
fi

BACKUP_DATE="$1"
TENANT_FILTER="${2:-}"
MINIO_ALIAS="${MINIO_ALIAS:-sda}"
MINIO_BUCKET="${MINIO_BUCKET:-sda-backups}"
BACKUP_ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY:-}"
PLATFORM_DB_URL="${PLATFORM_DB_URL:?PLATFORM_DB_URL is required}"
RESTORE_DIR="/tmp/sda-restore-${BACKUP_DATE}"

mkdir -p "${RESTORE_DIR}"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# ── Download from MinIO ─────────────────────────────────────────────────
log "Downloading backup from ${BACKUP_DATE}..."
mc cp --recursive "${MINIO_ALIAS}/${MINIO_BUCKET}/daily/${BACKUP_DATE}/" "${RESTORE_DIR}/"

# ── Decrypt helper ──────────────────────────────────────────────────────
decrypt_file() {
    local src="$1"
    if [[ "${src}" == *.age ]]; then
        if [ -z "${BACKUP_ENCRYPTION_KEY}" ]; then
            log "ERROR: encrypted backup but BACKUP_ENCRYPTION_KEY not set"
            exit 1
        fi
        local dst="${src%.age}"
        age -d -i "${BACKUP_ENCRYPTION_KEY}" -o "${dst}" "${src}"
        echo "${dst}"
    else
        echo "${src}"
    fi
}

# ── Verify checksum ─────────────────────────────────────────────────────
verify_checksum() {
    local file="$1"
    local checksum_file="${file}.sha256"
    if [ -f "${checksum_file}" ]; then
        if sha256sum --check --quiet "${checksum_file}" 2>/dev/null; then
            log "Checksum OK: $(basename "${file}")"
        else
            log "ERROR: checksum mismatch for $(basename "${file}")"
            exit 1
        fi
    else
        log "WARNING: no checksum file for $(basename "${file}")"
    fi
}

# ── Restore Platform DB ────────────────────────────────────────────────
if [ -z "${TENANT_FILTER}" ]; then
    PLATFORM_FILE=$(find "${RESTORE_DIR}" -name "platform.sql.gz*" -not -name "*.sha256" | head -1)
    if [ -n "${PLATFORM_FILE}" ]; then
        verify_checksum "${PLATFORM_FILE}"
        PLATFORM_FILE=$(decrypt_file "${PLATFORM_FILE}")
        log "Restoring platform database..."
        gunzip -c "${PLATFORM_FILE}" | psql "${PLATFORM_DB_URL}" --single-transaction -q
        log "Platform DB restored"
    else
        log "WARNING: platform backup not found"
    fi
fi

# ── Restore Tenant DBs ─────────────────────────────────────────────────
for tenant_file in "${RESTORE_DIR}"/tenant_*.sql.gz*; do
    [ -f "${tenant_file}" ] || continue
    [[ "${tenant_file}" == *.sha256 ]] && continue

    slug=$(basename "${tenant_file}" | sed 's/tenant_//;s/\.sql\.gz.*$//')

    if [ -n "${TENANT_FILTER}" ] && [ "${slug}" != "${TENANT_FILTER}" ]; then
        continue
    fi

    verify_checksum "${tenant_file}"
    tenant_file=$(decrypt_file "${tenant_file}")

    # Get tenant's DB URL from platform DB
    pg_url=$(psql "${PLATFORM_DB_URL}" -t -A -c \
        "SELECT postgres_url FROM tenants WHERE slug = '${slug}'")

    if [ -z "${pg_url}" ]; then
        log "WARNING: tenant ${slug} not found in platform DB — skipping"
        continue
    fi

    log "Restoring tenant: ${slug}..."
    gunzip -c "${tenant_file}" | psql "${pg_url}" --single-transaction -q
    log "Tenant ${slug} restored"
done

# ── Cleanup ─────────────────────────────────────────────────────────────
rm -rf "${RESTORE_DIR}"
log "Restore complete"
