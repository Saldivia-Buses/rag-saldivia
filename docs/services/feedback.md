---
title: Service: feedback
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/nats.md
  - ../architecture/multi-tenancy.md
  - ./traces.md
  - ./platform.md
---

## Purpose

Feedback collection and aggregation across services. Subscribes to all
`tenant.*.feedback.>` events (NPS, response quality, error reports, usage,
performance), persists them per tenant, runs an aggregator that computes
hourly summaries into the platform DB, and fires an alerter that pushes
critical alerts to platform admins via the WebSocket Hub. Exposes both
tenant-scoped and platform-admin REST surfaces. Read this when changing
feedback schemas, alert thresholds, or aggregation cadence.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + tenant/platform postgres + nats + redis |
| GET | `/v1/feedback/summary` | JWT | Tenant feedback summary over a period |
| GET | `/v1/feedback/quality` | JWT | Response quality metrics |
| GET | `/v1/feedback/errors` | JWT | Recent error reports (filterable) |
| GET | `/v1/feedback/usage` | JWT | Per-user usage breakdown |
| GET | `/v1/feedback/health-score` | JWT | Composite tenant health score |
| GET | `/v1/platform/feedback/tenants` | JWT (platform admin) | Cross-tenant overview |
| GET | `/v1/platform/feedback/alerts` | JWT (platform admin) | Active platform alerts |
| GET | `/v1/platform/feedback/quality` | JWT (platform admin) | Cross-tenant quality view |

Tenant routes mounted at `/v1/feedback` in
`services/feedback/cmd/main.go:126`. Platform routes mounted at
`/v1/platform/feedback` (`services/feedback/cmd/main.go:133`) with
`FailOpen=false`. Tenant routes use `FailOpen=true` to keep dashboards alive
during a Redis blip.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.*.feedback.>` | sub | Stream `FEEDBACK`, durable consumer (`services/feedback/internal/service/consumer.go:54`) |
| `tenant.{slug}.feedback.alerts` | pub | Critical alert broadcast for the WS Hub (`services/feedback/internal/service/alerter.go:197`) — platform admins subscribe to this channel |

Tenant slug is derived **only** from `msg.Subject()`, never from payload —
guards documented in `services/feedback/internal/service/consumer.go` test
file.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `FEEDBACK_PORT` | no | `8008` | HTTP listener port |
| `POSTGRES_TENANT_URL` | yes | — | Tenant DB for raw feedback |
| `POSTGRES_PLATFORM_URL` | yes | — | Platform DB for aggregations + alerts |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Subscriber + alert publisher |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |
| `AGGREGATION_INTERVAL` | no | `1h` | Set to `1m` to speed up tests |
| `TENANT_ID` | no | `dev` | Aggregator's target tenant ID |
| `TENANT_SLUG` | no | `dev` | Aggregator's target tenant slug |

## Dependencies

- **PostgreSQL tenant** — raw feedback events.
- **PostgreSQL platform** — hourly aggregations + alert table.
- **NATS** — JetStream subscription + alert broadcast.
- **Redis** — token blacklist.
- No outbound calls to other services.

## Permissions used

Tenant routes do not enforce specific permissions — any authenticated user
can view their tenant's feedback. Platform routes verify `X-User-Role` /
`X-Tenant-Slug` headers (set by middleware) to gate platform-admin access
inside the handler (`services/feedback/internal/handler/platform_feedback.go:35`).
