---
title: Service: healthwatch
audience: ai
last_reviewed: 2026-04-15
related:
  - ../flows/self-healing-triage.md
  - ./platform.md
  - ../packages/health.md
---

## Purpose

System health monitor and AI-triage entry point. Pulls per-service status
from Prometheus, container state from a Docker socket proxy, and Go-level
liveness from the in-cluster `pkg/health` checks; persists snapshots to the
platform DB; serves a platform-admin dashboard; and exposes the
triage-history endpoint that the daily AI-triage GitHub Action reads to
build incident reports. Read this when adjusting collectors, the retention
policy, alert wiring, or the cron triage feedback loop. Companion to
`docs/flows/self-healing-triage.md`.

## Endpoints

All routes require platform-admin role (verified by
`requirePlatformAdmin` middleware,
`services/healthwatch/internal/handler/healthwatch.go:116`).

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres ping |
| GET | `/v1/healthwatch/summary` | JWT (platform admin) | Aggregate status snapshot |
| GET | `/v1/healthwatch/services` | JWT (platform admin) | Per-service status list |
| GET | `/v1/healthwatch/alerts` | JWT (platform admin) | Active alerts |
| POST | `/v1/healthwatch/check` | JWT (platform admin) | Force a fresh probe (10s cooldown, 429 with `Retry-After`) |
| GET | `/v1/healthwatch/triage` | JWT (platform admin) | Last 50 triage records |

Routes assembled in `services/healthwatch/internal/handler/healthwatch.go:46`.

## NATS events

Healthwatch is currently HTTP-only — it does not connect to NATS, so neither
publishes nor subscribes. Alerting and triage propagation happen via the
GitHub Actions workflow `.github/workflows/daily-triage.yml`.

| Subject | Direction | Trigger |
|---|---|---|
| (none) | — | — |

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `HEALTHWATCH_PORT` | no | `8014` | HTTP listener port |
| `POSTGRES_PLATFORM_URL` | yes | — | Snapshots + triage records (also accepted as `/run/secrets/db_platform_url`) |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `REDIS_URL` | yes | `localhost:6379` | Token blacklist — service refuses to start without it (`services/healthwatch/cmd/main.go:34`) |
| `PROMETHEUS_URL` | no | `http://prometheus:9090` | Prometheus base URL |
| `DOCKER_PROXY_URL` | no | `http://docker-socket-proxy:2375` | Docker socket proxy |
| `PLATFORM_TENANT_SLUG` | no | `platform` | Required for admin role check |

## Dependencies

- **PostgreSQL platform** — `health_snapshots`, `triage_records` tables, plus
  retention cleanup via `service.StartRetentionCleanup`
  (`services/healthwatch/cmd/main.go:59`).
- **Redis** — token blacklist (mandatory).
- **Prometheus** — `collector.NewPrometheus` queries metrics
  (`services/healthwatch/cmd/main.go:47`).
- **Docker socket proxy** — `collector.NewDocker` reads container state.
- **In-process `collector.NewService`** — fan-out to each Go service's
  `/health` endpoint.
- **No NATS.** No outbound calls to other application services.

## Permissions used

No `RequirePermission` middleware. Authorization is handled by
`requirePlatformAdmin`, which:
1. Strips spoofed `X-User-*` headers,
2. Validates JWT against the public key + Redis blacklist,
3. Confirms the user belongs to the platform tenant (`PLATFORM_TENANT_SLUG`)
   with the admin role.
