---
title: Operations: Backup & Restore
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/database-postgres.md
  - ./deploy.md
---

## Purpose

How backups are produced, where they live, how they are encrypted, and the
exact procedure to restore a single tenant or the platform DB. Read before
running a restore — every step is destructive in some way.

## What gets backed up

| Source | Producer | Frequency |
|--------|----------|-----------|
| Platform PostgreSQL | `pg_dump sda_platform` | Daily |
| Each tenant PostgreSQL | `pg_dump sda_tenant_{slug}` per tenant | Daily |
| NATS JetStream state (`nats_data`) | tar of volume | Daily |
| Uploaded files in MinIO | MinIO mirror to `sda-backups/files/` | Continuous |

The producer is `deploy/scripts/backup.sh` (cron job on the workstation).
It iterates every tenant DB present in the platform `tenants` table.

## Storage layout (MinIO bucket `sda-backups/`)

```
sda-backups/
  daily/{YYYY-MM-DD}/
    platform.sql.gz.age
    platform.sql.gz.age.sha256
    tenant_{slug}.sql.gz.age
    tenant_{slug}.sql.gz.age.sha256
    nats_jetstream.tar.gz.age
  monthly/{YYYY-MM}/
    (first daily snapshot of the month, retained 12 months)
  files/
    (continuous mirror of uploaded files)
```

Daily backups are retained 30 days. Monthly backups are retained 12 months.

## Encryption

Each `.sql.gz` is encrypted with `age` using the recipient public key in
`deploy/scripts/backup.sh`. The matching identity (`~/.age/identity.txt`)
lives only on the workstation and on Enzo's offline backup. Never commit
either key.

Each backup file has an adjacent `.sha256` checksum — verify before
decrypting.

## Restore

### 0. Choose a backup and verify integrity

```bash
mc ls sda/sda-backups/daily/
mc cp sda/sda-backups/daily/2026-04-12/platform.sql.gz.age ./restore/
mc cp sda/sda-backups/daily/2026-04-12/platform.sql.gz.age.sha256 ./restore/
cd ./restore
sha256sum -c platform.sql.gz.age.sha256   # must say OK
```

### 1. Decrypt and decompress

```bash
age -d -i ~/.age/identity.txt -o platform.sql.gz platform.sql.gz.age
gunzip platform.sql.gz
```

### 2. Restore the platform DB

```bash
# Stop services that write to platform DB
docker compose -f deploy/docker-compose.prod.yml stop platform feedback traces

# Drop and recreate (DESTRUCTIVE — read twice)
docker compose -f deploy/docker-compose.prod.yml exec postgres-platform \
  dropdb -U sda sda_platform
docker compose -f deploy/docker-compose.prod.yml exec postgres-platform \
  createdb -U sda sda_platform

# Restore
PLATFORM_DB_URL="postgres://sda:PASS@localhost:5432/sda_platform?sslmode=disable"
psql "$PLATFORM_DB_URL" < platform.sql

# Restart services
docker compose -f deploy/docker-compose.prod.yml start platform feedback traces
```

### 3. Restore a single tenant DB

```bash
mc cp sda/sda-backups/daily/2026-04-12/tenant_saldivia.sql.gz.age ./restore/
age -d -i ~/.age/identity.txt -o tenant_saldivia.sql.gz tenant_saldivia.sql.gz.age
gunzip tenant_saldivia.sql.gz

TENANT_DB_URL="postgres://sda:PASS@localhost:5432/sda_tenant_saldivia?sslmode=disable"

# Stop tenant-facing services to avoid mid-restore writes
docker compose -f deploy/docker-compose.prod.yml stop \
  auth chat agent search ingest notification astro bigbrother erp

docker compose -f deploy/docker-compose.prod.yml exec postgres-platform \
  dropdb -U sda sda_tenant_saldivia
docker compose -f deploy/docker-compose.prod.yml exec postgres-platform \
  createdb -U sda sda_tenant_saldivia

psql "$TENANT_DB_URL" < tenant_saldivia.sql

docker compose -f deploy/docker-compose.prod.yml start \
  auth chat agent search ingest notification astro bigbrother erp
```

### 4. Verify

- `make versions` — all services running.
- Log in as a known user — JWT issued, session creates.
- Spot-check a recent record (e.g., last chat session) — present in DB.
- NATS `jsz` shows healthy streams.

## Recovery Time / Recovery Point

- **RPO**: 24h for DB (daily snapshot), continuous for files.
- **RTO**: ~10 minutes per tenant DB on the workstation hardware.

## Test restores

Run a restore drill against a scratch tenant DB at least once per quarter.
Procedure: copy a recent backup → restore into `sda_tenant_drill_YYYYMM` →
diff schema with current tenant template → drop the drill DB. Logs of the
last drill go in `docs/artifacts/restore-drill-{date}.md`.

## Never do

- Restore on top of a live DB without stopping writers — silent corruption.
- Skip the SHA-256 verification.
- Commit `~/.age/identity.txt`, the encryption pubkey file, or a decrypted
  `.sql` file.
