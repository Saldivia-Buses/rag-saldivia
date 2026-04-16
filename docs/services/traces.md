---
title: Service: traces
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/traces.md
  - ../architecture/observability.md
  - ./agent.md
  - ./astro.md
---

## Purpose

Execution-trace store and cost ledger. Subscribes to every
`tenant.*.traces.>` event published by the agent and astro services,
persists them to the platform DB, and serves a REST surface for the trace
detail UI and per-tenant cost rollups. Read this when changing the trace
schema, cost calculation, or adding a new producer that must hit this
consumer.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/nats/redis check |
| GET | `/v1/traces/` | JWT | List traces for the caller's tenant (paginated) |
| GET | `/v1/traces/costs` | JWT | Cost rollup over a date range (`from`, `to` query params, `2006-01-02` format) |
| GET | `/v1/traces/{traceID}` | JWT | Trace detail with steps, model usage, tokens, USD cost |

Routes registered in `services/traces/internal/handler/traces.go:36`. Uses
`FailOpen=true` so dashboards survive a Redis outage.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.*.traces.>` | sub | Stream consumer with `FilterSubject: tenant.*.traces.>` (`services/traces/internal/service/consumer.go:59`) — handles `traces.start`, `traces.end`, `traces.event` |

Action is parsed from the subject (`tenant.{slug}.traces.{action}`).
Tenant slug is taken from the subject, never the payload. Traces does not
publish.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `TRACES_PORT` | no | `8009` | HTTP listener port |
| `POSTGRES_PLATFORM_URL` | yes | — | Trace storage (platform DB so all tenants share the rollup tables) |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Subscriber |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |

## Dependencies

- **PostgreSQL platform** — `traces`, `trace_events`, cost rollups.
- **NATS JetStream** — single durable consumer on `tenant.*.traces.>`.
- **Redis** — token blacklist.
- No outbound calls to other services.

## Permissions used

No `RequirePermission` — any authenticated user can read traces for their
own tenant. Tenant scoping is enforced inside the handler/repo by the
`X-Tenant-Slug` header (set by middleware from the JWT). Cross-tenant reads
are not exposed.
