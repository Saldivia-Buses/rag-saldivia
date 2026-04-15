---
title: Operations: Monitoring
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/observability.md
  - ./runbook.md
  - ./incidents.md
---

## Purpose

What to watch in production: dashboards, log streams, metrics, traces, and
the always-on watch list for the first 24h after a deploy. For incident
response (how to act on alerts), see [incidents.md](./incidents.md). For
fixing specific failures, see [runbook.md](./runbook.md).

## Stack

| Signal | System | Source |
|--------|--------|--------|
| Logs | Loki | All Go services log JSON via `slog`; Promtail tails Docker logs |
| Metrics | Prometheus | OTel exporter on each Go service + Traefik + Postgres exporter |
| Traces | Tempo | OTel SDK in every Go service; trace ID propagated via `traceparent` header |
| Dashboards | Grafana | `http://localhost:3000` — admin password from `GRAFANA_ADMIN_PASSWORD` |
| Network bans | CrowdSec | Reads Traefik access logs from `traefik_logs` volume |

## Health endpoints

Every Go service exposes `GET /health` (returns `{"status":"ok","version":"..."}`)
and `GET /v1/info` (returns `{"git_sha","build_time","version"}`). The
extractor (Python) returns `{"status":"ok"}`.

```bash
# Per-service health and SHA
for entry in 8001:auth 8002:ws 8003:chat 8004:agent 8005:notification \
             8006:platform 8007:ingest 8008:feedback 8009:traces \
             8010:search 8011:astro 8012:bigbrother 8013:erp; do
  port=${entry%%:*}; name=${entry#*:}
  health=$(curl -sf --max-time 2 http://localhost:$port/health \
    -o /dev/null -w "%{http_code}" || echo 000)
  sha=$(curl -sf --max-time 2 http://localhost:$port/v1/info \
    | jq -r '.git_sha // "-"')
  printf "%-14s :%s health=%s sha=%s\n" "$name" "$port" "$health" "$sha"
done
```

`make versions` does the same and flags `MATCH` vs `STALE` against source.

## Key Grafana dashboards

| Dashboard | Watch for |
|-----------|-----------|
| Service overview | Per-service request rate, p95 latency, error rate |
| Auth | Login attempts, JWT verify failures, brute-force lockouts |
| NATS / JetStream | Stream lag, consumer pending, redeliveries |
| PostgreSQL | Connections, slow queries, replication lag |
| Redis | Memory usage, evictions (`maxmemory-policy allkeys-lru`) |
| GPU | SGLang VRAM, inference latency, queue depth |
| CrowdSec | Active bans, alerts/hour |
| Traefik | 4xx/5xx rate per router, TLS cert expiry |

## Logs

```bash
# One service
docker logs sda-auth-1 --tail 100 -f

# Whole stack
docker compose -f deploy/docker-compose.prod.yml logs -f --tail 50

# Traefik access log (volume)
docker compose -f deploy/docker-compose.prod.yml exec traefik \
  tail -f /var/log/traefik/access.log
```

JSON log fields used everywhere: `level`, `msg`, `service`, `tenant_id`,
`trace_id`, `request_id`. In Loki, filter by `{service="auth"} | json`.

## CrowdSec

```bash
# Alerts (detected attacks)
docker compose -f deploy/docker-compose.prod.yml exec crowdsec cscli alerts list

# Active decisions (bans)
docker compose -f deploy/docker-compose.prod.yml exec crowdsec cscli decisions list
```

Active collections: `crowdsecurity/traefik`, `crowdsecurity/http-cve`.

## Autoheal

`autoheal` restarts containers Docker marks `unhealthy` (the standard
restart policy doesn't do this). Check interval 30s, startup grace 60s.

```bash
docker logs $(docker ps -q -f name=autoheal) --tail 20
```

If a container restart-loops, autoheal log shows the cycle — check that
service's log next.

## First-24h watch list (post-deploy)

| Signal | Threshold to investigate |
|--------|--------------------------|
| Auth: `verify token` errors | Any sustained rate (key mismatch) |
| NATS: `auth_error` in container log | Any (wrong password) |
| Postgres: `too many clients` | Any (pool exhaustion) |
| Redis: evictions/min | > 0 sustained (memory pressure) |
| Traefik: 502 Bad Gateway | > a few/min (service unhealthy) |
| BigBrother: `NET_RAW` errors | Any (capability not granted on host) |
| Autoheal: restart events | More than 1 restart per service in an hour |

When a watch-list signal trips, jump to [runbook.md](./runbook.md) for the
matching failure mode and to [incidents.md](./incidents.md) if it impacts
users.
