# SDA Framework — Production Deployment Runbook

Source of truth: `deploy/docker-compose.prod.yml`, `deploy/.env.example`, `deploy/secrets/README.md`.

---

## Pre-deployment Checklist

### Server requirements

- Docker Engine 24+ and Docker Compose v2
- Ports 80 and 443 open to the internet (Traefik)
- `git`, `openssl`, `psql` (PostgreSQL client) available on the host
- If running BigBrother with active scanning: `NET_RAW` capability available (granted per-container in compose)

### Docker secrets

Each file must exist in `deploy/secrets/dynamic/` before starting.

| File | Used by | How to generate |
|------|---------|-----------------|
| `jwt-private.pem` | auth only | `bash deploy/scripts/gen-jwt-keys.sh` |
| `jwt-public.pem` | all Go services | same script |
| `db-platform-url` | platform, feedback, traces | write PostgreSQL URL string |
| `db-tenant-template-url` | auth, chat, search, ingest, notification, astro, erp, bigbrother | write PostgreSQL URL template |
| `redis-password` | redis | write password string |
| `encryption-master-key` | auth | write 32-byte AES-256 key (base64) |
| `db-platform-password` | postgres-platform container | write PostgreSQL password string |
| `bb-kek` | bigbrother | write KEK for envelope encryption |

All files in `deploy/secrets/dynamic/` are gitignored. Never commit them.

### Environment variables (`.env`)

Copy `deploy/.env.example` to `deploy/.env` and set all REQUIRED values.

| Variable | Required | Description |
|----------|----------|-------------|
| `SDA_DOMAIN` | yes | Base domain (e.g., `saldivia.com.ar`) — controls Traefik routing and CORS |
| `SDA_ACME_EMAIL` | yes | Let's Encrypt email for TLS certificates |
| `SDA_TENANT_SLUG` | yes | Active tenant slug for NATS subjects and DB naming |
| `POSTGRES_PASSWORD` | yes | PostgreSQL platform DB password |
| `AUTH_NATS_PASS` | yes | Per-service NATS password (must match `nats-server.conf`) |
| `WS_NATS_PASS` | yes | — |
| `CHAT_NATS_PASS` | yes | — |
| `AGENT_NATS_PASS` | yes | — |
| `SEARCH_NATS_PASS` | yes | — |
| `TRACES_NATS_PASS` | yes | — |
| `NOTIF_NATS_PASS` | yes | — |
| `PLATFORM_NATS_PASS` | yes | — |
| `INGEST_NATS_PASS` | yes | — |
| `FEEDBACK_NATS_PASS` | yes | — |
| `ASTRO_NATS_PASS` | yes | — |
| `BIGBROTHER_NATS_PASS` | yes | — |
| `ERP_NATS_PASS` | yes | — |
| `EXTRACTOR_NATS_PASS` | yes | — |
| `SMTP_HOST` | yes | SMTP server for notification service |
| `SMTP_PORT` | no | Default: `587` |
| `SMTP_FROM` | no | Default: `noreply@{SDA_DOMAIN}` |
| `SGLANG_LLM_URL` | no | SGLang LLM server (default: `http://host.docker.internal:8102`) |
| `SGLANG_LLM_MODEL` | no | Model name — empty = server default |
| `SGLANG_OCR_URL` | no | PaddleOCR-VL server (default: `http://host.docker.internal:8100`) |
| `SGLANG_VISION_URL` | no | Qwen vision server (default: `http://host.docker.internal:8101`) |
| `EPHE_DATA_PATH` | yes (astro) | Host path to Swiss Ephemeris data files |
| `LAN_INTERFACE` | yes (bigbrother) | Host network interface for macvlan (e.g., `eth0`) |
| `LAN_SUBNET` | yes (bigbrother) | LAN subnet for macvlan (e.g., `192.168.1.0/24`) |
| `GRAFANA_ADMIN_PASSWORD` | no | Default: `sda-grafana-dev` — change in production |
| `GIT_SHA` | no | Injected by CI; set manually for traceability |
| `BUILD_TIME` | no | Injected by CI |

NATS passwords must match the values in `deploy/nats/nats-server.conf`. Generate unique passwords per service:

```bash
for svc in AUTH CHAT WS NOTIF INGEST FEEDBACK AGENT TRACES PLATFORM EXTRACTOR ASTRO BIGBROTHER ERP; do
  echo "${svc}_NATS_PASS=$(openssl rand -base64 24)"
done
```

### DNS and TLS

Traefik uses Cloudflare DNS challenge for Let's Encrypt (configured in `deploy/traefik/traefik.prod.yml.tmpl`). Required:

- DNS A/CNAME record pointing `{SDA_DOMAIN}` and `platform.{SDA_DOMAIN}` to the host IP
- `CLOUDFLARE_DNS_API_TOKEN` environment variable available to Traefik (or configured in `traefik.prod.yml`)
- `SDA_ACME_EMAIL` set to a valid address

HTTP (port 80) is redirected to HTTPS automatically. Traefik dashboard is disabled in production.

---

## Step-by-Step Deployment

### 1. Clone the repository

```bash
git clone https://github.com/Camionerou/rag-saldivia.git
cd rag-saldivia
```

### 2. Generate JWT keys

```bash
bash deploy/scripts/gen-jwt-keys.sh
# Output: deploy/secrets/dynamic/jwt-private.pem + jwt-public.pem
```

### 3. Populate secrets

```bash
cd deploy/secrets/dynamic

# PostgreSQL
echo "postgres://sda:STRONG_PASSWORD@postgres-platform:5432/sda_platform?sslmode=require" > db-platform-url
echo "postgres://sda:STRONG_PASSWORD@postgres-platform:5432/sda_tenant_{slug}?sslmode=require" > db-tenant-template-url
echo "STRONG_PASSWORD" > db-platform-password

# Redis
echo "$(openssl rand -base64 32)" > redis-password

# AES-256 master key for envelope encryption (auth service + BigBrother KEK)
echo "$(openssl rand -base64 32)" > encryption-master-key
echo "$(openssl rand -base64 32)" > bb-kek
```

### 4. Configure environment

```bash
cp deploy/.env.example deploy/.env
# Edit deploy/.env — set all REQUIRED variables
```

### 5. Run database migrations

The migration script requires `psql` on the host and direct access to the PostgreSQL container. Run after the database container is healthy.

```bash
# Start only the database first
docker compose -f deploy/docker-compose.prod.yml up -d postgres-platform
docker compose -f deploy/docker-compose.prod.yml exec postgres-platform pg_isready -U sda

# Run migrations
PLATFORM_DB_URL="postgres://sda:STRONG_PASSWORD@localhost:5432/sda_platform?sslmode=disable" \
TENANT_DB_URL="postgres://sda:STRONG_PASSWORD@localhost:5432/sda_tenant_dev?sslmode=disable" \
bash deploy/scripts/migrate.sh
```

The script applies `.up.sql` files from `db/platform/migrations/` and `db/tenant/migrations/` in filename order. Applied migrations are tracked in `schema_migrations` — re-running is safe.

### 6. Start all services

```bash
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env up -d
```

### 7. Verify health

```bash
# All services should show healthy within ~60s (autoheal grace period)
docker compose -f deploy/docker-compose.prod.yml ps

# Individual health checks
curl -sf https://{SDA_DOMAIN}/v1/auth/health
curl -sf https://{SDA_DOMAIN}/v1/chat/health
curl -sf https://{SDA_DOMAIN}/v1/agent/health
curl -sf https://{SDA_DOMAIN}/v1/search/health
curl -sf https://{SDA_DOMAIN}/v1/ingest/health
curl -sf https://{SDA_DOMAIN}/v1/traces/health
curl -sf https://{SDA_DOMAIN}/v1/notifications/health
curl -sf https://{SDA_DOMAIN}/ws/health
curl -sf https://{SDA_DOMAIN}/v1/platform/health
curl -sf https://{SDA_DOMAIN}/v1/feedback/health
curl -sf https://{SDA_DOMAIN}/v1/astro/health
curl -sf https://{SDA_DOMAIN}/v1/bigbrother/health
curl -sf https://{SDA_DOMAIN}/v1/erp/health
```

Service ports (internal, not exposed externally):

| Service | Port |
|---------|------|
| auth | 8001 |
| ws | 8002 |
| chat | 8003 |
| agent | 8004 |
| notification | 8005 |
| platform | 8006 |
| ingest | 8007 |
| feedback | 8008 |
| traces | 8009 |
| search | 8010 |
| astro | 8011 |
| bigbrother | 8012 |
| erp | 8013 |
| extractor | 8090 |

### 8. Create first tenant and admin user

Use the `platform` service API (accessible only on `platform.{SDA_DOMAIN}`):

```bash
# Create tenant
curl -X POST https://platform.{SDA_DOMAIN}/v1/platform/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {PLATFORM_ADMIN_TOKEN}" \
  -d '{"slug":"saldivia","name":"Saldivia Buses"}'

# Then create admin user via auth service
curl -X POST https://{SDA_DOMAIN}/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@saldivia.com.ar","password":"...","tenant_slug":"saldivia"}'
```

---

## Post-Deployment Verification

### Expected health check responses

All Go services return HTTP 200 with JSON body:

```json
{"status":"ok","version":"2.0.x"}
```

The `extractor` (Python) returns HTTP 200 with `{"status":"ok"}`.

### NATS connectivity

```bash
docker compose -f deploy/docker-compose.prod.yml exec nats \
  wget -qO- http://localhost:8222/healthz
# Expected: {"status":"ok"}
```

### Redis connectivity

```bash
docker compose -f deploy/docker-compose.prod.yml exec redis \
  sh -c 'redis-cli -a $(cat /run/secrets/redis_password) ping'
# Expected: PONG
```

### Tenant creation smoke test

1. Register a user in the new tenant
2. Log in — verify JWT is returned
3. Open a WebSocket connection to `wss://{SDA_DOMAIN}/ws` with the JWT
4. Send a chat message — verify it appears in real time

---

## Rollback Procedure

### Roll back to a previous image version

```bash
# Edit deploy/.env — change service version vars
# e.g., AUTH_VERSION=2.0.4

docker compose -f deploy/docker-compose.prod.yml up -d auth
# Docker Compose pulls the specified image and restarts only the changed service
```

### Run down migrations (manual)

The migration runner (`migrate.sh`) applies `.up.sql` files only. There is no automatic rollback. To roll back a migration:

```bash
# Find the corresponding .down.sql
ls db/platform/migrations/
# e.g., 0042_add_sessions.up.sql → 0042_add_sessions.down.sql

PLATFORM_DB_URL="postgres://sda:PASS@localhost:5432/sda_platform?sslmode=disable"

# Apply the down migration
psql "$PLATFORM_DB_URL" -f db/platform/migrations/0042_add_sessions.down.sql

# Remove it from the tracking table so it can be re-applied
psql "$PLATFORM_DB_URL" -c "DELETE FROM schema_migrations WHERE filename = '0042_add_sessions.up.sql'"
```

Repeat for the tenant DB if the migration touches `db/tenant/migrations/`.

### Restore from backup

Backups are stored in MinIO under `sda-backups/daily/{YYYY-MM-DD}/` and `sda-backups/monthly/{YYYY-MM}/`.

```bash
# List available backups
mc ls sda/sda-backups/daily/

# Download and decrypt
mc cp sda/sda-backups/daily/2026-04-12/platform.sql.gz.age ./restore/
age -d -i ~/.age/identity.txt -o platform.sql.gz platform.sql.gz.age
gunzip platform.sql.gz

# Restore platform DB
psql "$PLATFORM_DB_URL" < platform.sql

# Restore a specific tenant DB
mc cp sda/sda-backups/daily/2026-04-12/tenant_saldivia.sql.gz.age ./restore/
age -d -i ~/.age/identity.txt -o tenant_saldivia.sql.gz tenant_saldivia.sql.gz.age
gunzip tenant_saldivia.sql.gz
psql "$TENANT_DB_URL" < tenant_saldivia.sql
```

SHA-256 checksums are stored alongside each backup file (`.sha256` suffix) — verify before restoring.

---

## Monitoring

### Logs

```bash
# Specific service
docker logs sda-auth-1 --tail 100 -f
docker logs sda-chat-1 --tail 100 -f

# All services at once
docker compose -f deploy/docker-compose.prod.yml logs -f --tail 50

# Traefik access log (buffered, in volume)
docker compose -f deploy/docker-compose.prod.yml exec traefik \
  tail -f /var/log/traefik/access.log
```

All Go services log via `slog` in JSON format. Key fields: `level`, `msg`, `service`, `tenant_id`, `trace_id`.

### First 24h watch list

- Auth service: watch for JWT validation errors (`level=ERROR msg="verify token"`)
- NATS: watch for `auth_error` in nats container logs — means a service is using the wrong password
- PostgreSQL: watch for connection pool exhaustion (`too many clients`)
- Redis: memory usage should stay under 512MB limit — watch `maxmemory-policy allkeys-lru` evictions
- Traefik: 502 Bad Gateway = service is unhealthy; autoheal will restart it within 30s
- BigBrother: `NET_RAW` errors indicate the host does not allow the capability

### CrowdSec

CrowdSec reads Traefik access logs from the shared `traefik_logs` volume. Collections active: `crowdsecurity/traefik` and `crowdsecurity/http-cve`.

```bash
# Check CrowdSec alerts
docker compose -f deploy/docker-compose.prod.yml exec crowdsec cscli alerts list

# Check active decisions (bans)
docker compose -f deploy/docker-compose.prod.yml exec crowdsec cscli decisions list
```

### Autoheal

`autoheal` monitors Docker health checks and restarts UNHEALTHY containers (which Docker's own restart policy does not handle). Check interval: 30s, startup grace period: 60s.

```bash
docker logs $(docker ps -q -f name=autoheal) --tail 20
```

---

## Troubleshooting

### Service won't start

```bash
docker logs sda-{service}-1 --tail 50
```

Common causes:
- Missing secret file → `open /run/secrets/jwt_public_key: no such file or directory`
- DB not ready → `failed to connect to database` — check `postgres-platform` health
- NATS auth failure → `nats: Authorization Violation` — verify `{SERVICE}_NATS_PASS` matches `nats-server.conf`

### Auth failing (401 on all requests)

JWT key mismatch between the auth service (signs with private key) and other services (verify with public key).

```bash
# Verify both files are the keypair
openssl pkey -in deploy/secrets/dynamic/jwt-private.pem -pubout | \
  diff - deploy/secrets/dynamic/jwt-public.pem
# No diff = keys match
```

If keys were regenerated after services were deployed, restart all services (not just auth):

```bash
docker compose -f deploy/docker-compose.prod.yml restart
```

### DB connection error

```bash
# Verify postgres-platform is healthy
docker compose -f deploy/docker-compose.prod.yml ps postgres-platform

# Check the URL in the secret file
cat deploy/secrets/dynamic/db-platform-url
# Hostname must be "postgres-platform" (Docker network), not "localhost"

# Test connection from inside the network
docker compose -f deploy/docker-compose.prod.yml exec auth \
  wget -qO- http://localhost:8001/health
```

### WebSocket connections not upgrading

- Verify `WS_ALLOWED_ORIGINS` contains `https://*.{SDA_DOMAIN}` — mismatched domain blocks upgrades
- Check Traefik router for `ws` service is routing `PathPrefix(/ws)` on `websecure` entrypoint

### NATS JetStream not persisting

The `nats_data` volume must be on a persistent disk. If NATS restarts and loses stream state, consumers (ws, ingest) will reconnect and re-subscribe — data already consumed is gone. Backups cover JetStream data via `deploy/scripts/backup.sh`.
