#!/usr/bin/env bash
# SQLite online backup with 7-day rotation.
#
# Usage: ./scripts/backup-sqlite.sh [destination_dir]
# Cron:  0 */6 * * * /path/to/scripts/backup-sqlite.sh /backups/
#
# Requires: sqlite3 CLI installed on the host.
# Safe with WAL mode (uses .backup command for atomic copy).

set -euo pipefail

DB_PATH="${DATABASE_PATH:-./data/app.db}"
DEST="${1:-./backups}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="${DEST}/app-${TIMESTAMP}.db"

if [ ! -f "$DB_PATH" ]; then
  echo "Error: database not found at ${DB_PATH}" >&2
  exit 1
fi

mkdir -p "$DEST"
sqlite3 "$DB_PATH" ".backup '${BACKUP_FILE}'"

# Retain last 7 days (28 backups at 6h intervals)
find "$DEST" -name "app-*.db" -mtime +7 -delete 2>/dev/null || true

echo "Backup: ${BACKUP_FILE} ($(du -h "$BACKUP_FILE" | cut -f1))"
