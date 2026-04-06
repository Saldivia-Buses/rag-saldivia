#!/usr/bin/env bash
# SDA Framework — Automated Backup Script
#
# Backs up all PostgreSQL databases (platform + per-tenant), Redis, and NATS.
# Backups are compressed, encrypted, and uploaded to MinIO.
#
# Usage: ./backup.sh
# Cron:  0 3 * * * /path/to/deploy/scripts/backup.sh >> /var/log/sda-backup.log 2>&1
#
# Prerequisites:
#   - pg_dump, gzip, age (or gpg), mc (MinIO client) installed
#   - BACKUP_ENCRYPTION_KEY: path to age public key file
#   - MINIO_ALIAS: mc alias configured (e.g., "sda")
#   - PLATFORM_DB_URL: PostgreSQL connection URL for platform DB

set -euo pipefail

# ── Configuration ───────────────────────────────────────────────────────
BACKUP_DIR="${BACKUP_DIR:-/tmp/sda-backups}"
MINIO_ALIAS="${MINIO_ALIAS:-sda}"
MINIO_BUCKET="${MINIO_BUCKET:-sda-backups}"
PLATFORM_DB_URL="${PLATFORM_DB_URL:?PLATFORM_DB_URL is required}"
BACKUP_ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY:-}"
RETENTION_DAILY=30
RETENTION_MONTHLY=365

DATE=$(date +%Y-%m-%d)
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_PATH="${BACKUP_DIR}/${TIMESTAMP}"

mkdir -p "${BACKUP_PATH}"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# ── Encrypt helper ──────────────────────────────────────────────────────
encrypt_file() {
    local src="$1"
    if [ -n "${BACKUP_ENCRYPTION_KEY}" ] && command -v age &>/dev/null; then
        age -R "${BACKUP_ENCRYPTION_KEY}" -o "${src}.age" "${src}"
        rm "${src}"
        echo "${src}.age"
    else
        log "WARNING: no encryption key or age not found — backup is NOT encrypted"
        echo "${src}"
    fi
}

# ── PostgreSQL: Platform DB ─────────────────────────────────────────────
log "Backing up platform database..."
PLATFORM_DUMP="${BACKUP_PATH}/platform.sql.gz"
pg_dump "${PLATFORM_DB_URL}" --no-owner --no-privileges | gzip > "${PLATFORM_DUMP}"
PLATFORM_FINAL=$(encrypt_file "${PLATFORM_DUMP}")
sha256sum "${PLATFORM_FINAL}" > "${PLATFORM_FINAL}.sha256"
log "Platform DB: $(du -h "${PLATFORM_FINAL}" | cut -f1)"

# ── PostgreSQL: Tenant DBs ──────────────────────────────────────────────
log "Fetching tenant list from platform DB..."
TENANTS=$(psql "${PLATFORM_DB_URL}" -t -A -c \
    "SELECT slug, postgres_url FROM tenants WHERE enabled = true")

TENANT_COUNT=0
while IFS='|' read -r slug pg_url; do
    [ -z "${slug}" ] && continue
    log "Backing up tenant: ${slug}..."
    TENANT_DUMP="${BACKUP_PATH}/tenant_${slug}.sql.gz"
    pg_dump "${pg_url}" --no-owner --no-privileges | gzip > "${TENANT_DUMP}"
    TENANT_FINAL=$(encrypt_file "${TENANT_DUMP}")
    sha256sum "${TENANT_FINAL}" > "${TENANT_FINAL}.sha256"
    TENANT_COUNT=$((TENANT_COUNT + 1))
done <<< "${TENANTS}"
log "Backed up ${TENANT_COUNT} tenant databases"

# ── Redis ───────────────────────────────────────────────────────────────
log "Backing up Redis..."
REDIS_URL="${REDIS_URL:-localhost:6379}"
REDIS_PASS="${REDIS_PASSWORD:-}"
REDIS_AUTH=""
[ -n "${REDIS_PASS}" ] && REDIS_AUTH="-a ${REDIS_PASS}"

# Trigger background save and wait
redis-cli -h "${REDIS_URL%%:*}" -p "${REDIS_URL##*:}" ${REDIS_AUTH} BGSAVE 2>/dev/null || true
sleep 2

# Copy RDB file if accessible
REDIS_RDB="${BACKUP_PATH}/redis.rdb.gz"
if [ -f /data/dump.rdb ]; then
    gzip -c /data/dump.rdb > "${REDIS_RDB}"
    REDIS_FINAL=$(encrypt_file "${REDIS_RDB}")
    sha256sum "${REDIS_FINAL}" > "${REDIS_FINAL}.sha256"
    log "Redis: $(du -h "${REDIS_FINAL}" | cut -f1)"
else
    log "WARNING: Redis RDB not accessible at /data/dump.rdb — skipping"
fi

# ── NATS JetStream ──────────────────────────────────────────────────────
log "Backing up NATS JetStream data..."
NATS_DATA="${NATS_DATA_DIR:-/data}"
if [ -d "${NATS_DATA}/jetstream" ]; then
    NATS_BACKUP="${BACKUP_PATH}/nats_jetstream.tar.gz"
    tar -czf "${NATS_BACKUP}" -C "${NATS_DATA}" jetstream/
    NATS_FINAL=$(encrypt_file "${NATS_BACKUP}")
    sha256sum "${NATS_FINAL}" > "${NATS_FINAL}.sha256"
    log "NATS: $(du -h "${NATS_FINAL}" | cut -f1)"
else
    log "WARNING: NATS JetStream data not found at ${NATS_DATA}/jetstream — skipping"
fi

# ── Upload to MinIO ─────────────────────────────────────────────────────
log "Uploading to MinIO (${MINIO_ALIAS}/${MINIO_BUCKET})..."
mc cp --recursive "${BACKUP_PATH}/" "${MINIO_ALIAS}/${MINIO_BUCKET}/daily/${DATE}/"
log "Upload complete"

# ── Monthly backup (1st of month) ──────────────────────────────────────
DAY_OF_MONTH=$(date +%d)
if [ "${DAY_OF_MONTH}" = "01" ]; then
    log "Creating monthly backup copy..."
    MONTH=$(date +%Y-%m)
    mc cp --recursive "${MINIO_ALIAS}/${MINIO_BUCKET}/daily/${DATE}/" \
        "${MINIO_ALIAS}/${MINIO_BUCKET}/monthly/${MONTH}/"
    log "Monthly backup: ${MONTH}"
fi

# ── Retention cleanup ───────────────────────────────────────────────────
log "Cleaning up old backups..."
CUTOFF_DAILY=$(date -d "-${RETENTION_DAILY} days" +%Y-%m-%d 2>/dev/null || date -v-${RETENTION_DAILY}d +%Y-%m-%d)
mc rm --recursive --force "${MINIO_ALIAS}/${MINIO_BUCKET}/daily/" \
    --older-than "${RETENTION_DAILY}d" 2>/dev/null || true

CUTOFF_MONTHLY=$(date -d "-${RETENTION_MONTHLY} days" +%Y-%m-%d 2>/dev/null || date -v-${RETENTION_MONTHLY}d +%Y-%m-%d)
mc rm --recursive --force "${MINIO_ALIAS}/${MINIO_BUCKET}/monthly/" \
    --older-than "${RETENTION_MONTHLY}d" 2>/dev/null || true

# ── Cleanup local ───────────────────────────────────────────────────────
rm -rf "${BACKUP_PATH}"
log "Backup complete. Daily: ${RETENTION_DAILY}d retention, Monthly: ${RETENTION_MONTHLY}d retention"
