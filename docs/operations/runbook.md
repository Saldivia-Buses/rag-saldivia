---
title: Operations: Runbook
audience: ai
last_reviewed: 2026-04-15
related:
  - ./incidents.md
  - ./monitoring.md
  - ./deploy.md
---

## Purpose

Step-by-step troubleshooting for the most frequent failure modes. Use during
an incident to bisect quickly. For metrics and dashboards (what to watch),
see [monitoring.md](./monitoring.md). For the response workflow (who to call,
postmortem), see [incidents.md](./incidents.md).

## Quick triage

Always start with these three:

```bash
# 1. Container state
docker compose -f deploy/docker-compose.prod.yml ps

# 2. Per-service health
for entry in 8001:auth 8002:ws 8003:chat 8004:agent 8005:notification \
             8006:platform 8007:ingest 8008:feedback 8009:traces \
             8010:search 8011:astro 8012:bigbrother 8013:erp; do
  port=${entry%%:*}; name=${entry#*:}
  code=$(curl -sf --max-time 2 http://localhost:$port/health \
    -o /dev/null -w "%{http_code}" || echo 000)
  printf "%-14s :%s -> %s\n" "$name" "$port" "$code"
done

# 3. Dependencies
docker exec $(docker ps -q -f name=postgres) pg_isready -U sda
curl -sf http://localhost:8222/healthz   # NATS
docker exec $(docker ps -q -f name=redis) redis-cli ping
```

## Failure modes

### Service won't start

```bash
docker logs sda-{service}-1 --tail 50
```

| Log signature | Cause | Fix |
|---------------|-------|-----|
| `open /run/secrets/...: no such file or directory` | Missing secret file | Create file under `deploy/secrets/dynamic/` and restart |
| `failed to connect to database` | Platform DB not ready | `docker compose ... ps postgres-platform`; wait for healthy |
| `nats: Authorization Violation` | NATS password mismatch | Check `{SVC}_NATS_PASS` matches `deploy/nats/nats-server.conf` |
| `pq: relation "X" does not exist` | Migrations not applied | `bash deploy/scripts/migrate.sh` |
| `pq: FATAL: database "X" does not exist` | Tenant DB missing | Create tenant via Platform API first |
| `module not found` (build) | Service not in `go.work` | Add service path to `go.work` |

### 401 Unauthorized on every request

JWT key mismatch between auth (signs with private key) and other services
(verify with public key).

```bash
openssl pkey -in deploy/secrets/dynamic/jwt-private.pem -pubout \
  | diff - deploy/secrets/dynamic/jwt-public.pem
# No diff = keys match
```

If keys were rotated, restart the whole stack so every service reloads the
public key:

```bash
docker compose -f deploy/docker-compose.prod.yml restart
```

### DB connection failures

```bash
docker compose -f deploy/docker-compose.prod.yml ps postgres-platform
cat deploy/secrets/dynamic/db-platform-url
# Hostname must be "postgres-platform" (Docker network), not "localhost"
```

`too many open connections` → a service is leaking pgxpool connections.
Default `PoolMaxConns=4` per tenant. Restart the offending service.

### NATS issues

| Symptom | Diagnosis |
|---------|-----------|
| `nats: no responders` | Consumer not connected — `curl localhost:8222/jsz` |
| Events not delivered | Wrong subject — must follow `tenant.{slug}.{service}.{entity}` |
| JetStream lost messages after restart | `nats_data` volume not on persistent disk |

### WebSocket connections won't upgrade

- `WS_ALLOWED_ORIGINS` must contain `https://*.{SDA_DOMAIN}`.
- Traefik router `ws` must route `PathPrefix(/ws)` on the `websecure`
  entrypoint.
- Client must send `Authorization: Bearer {jwt}` in the upgrade request.

### Cross-tenant data leak (CRITICAL)

If a user from tenant A sees data from tenant B: stop traffic, take the
service offline, audit the offending query for missing `WHERE tenant_id =
$1`. Do not restart until the query is fixed and tested. See
[incidents.md](./incidents.md) for response process.

### Emails not arriving

```bash
curl -s http://localhost:8222/jsz | jq '.streams[] | select(.name=="NOTIFICATIONS")'
# Should show consumer "notification-service" with num_pending = 0

# In dev (Mailpit)
curl -sf http://localhost:8025
```

If consumer is stuck, restart the notification service.

### Migration regret

Run `.down.sql` manually and remove the row from `schema_migrations`. See
[deploy.md](./deploy.md) — Rollback section.

## Where to look for state

| Subsystem | Tool |
|-----------|------|
| Application logs | `docker logs sda-{svc}-1 --tail 100 -f` |
| Aggregated logs | Loki via Grafana — see [monitoring.md](./monitoring.md) |
| Metrics | Prometheus / Grafana dashboards |
| Distributed traces | Tempo / Grafana |
| NATS state | `http://localhost:8222/jsz` (JetStream), `/varz` |
| Traefik routes | Dashboard disabled in prod; check `deploy/traefik/dynamic/` |
| Container restarts | `docker logs $(docker ps -q -f name=autoheal)` |
| Database | `docker exec ... psql -U sda` |
| Backups | `mc ls sda/sda-backups/` |

## Escalation

Persistent failure or suspected data integrity issue → see
[incidents.md](./incidents.md).
