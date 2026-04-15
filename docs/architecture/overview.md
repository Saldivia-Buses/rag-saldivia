---
title: Architecture Overview
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - multi-tenancy.md
  - auth-jwt.md
  - nats-events.md
  - websocket-hub.md
  - rag-tree-search.md
  - llm-sglang.md
  - observability.md
  - gateway-traefik.md
  - storage-minio.md
  - database-postgres.md
---

This document is the entry point for the SDA Framework architecture. Read it
first to map services, packages, and the deployment topology — every other
architecture doc deepens one slice of what is summarised here.

## Shape

SDA is a multi-tenant SaaS composed of **15 microservices** (14 Go, 1 Python)
and **27 shared Go packages** (`pkg/`). The frontend is a Next.js app served
from the cloud. The backend runs **inhouse** on a workstation that hosts the
GPU (RTX PRO 6000, 96 GB VRAM), the SGLang model servers, PostgreSQL, Redis,
NATS + JetStream, and MinIO. A Cloudflare Tunnel forwards public HTTPS to the
local Traefik gateway — no inbound ports are opened on the workstation.

## Deployment topology

```
CLOUD: Next.js -> Cloudflare Tunnel
INHOUSE: Traefik :80
  +-> Go services (auth 8001, ws 8002, chat 8003, agent 8004,
      notification 8005, platform 8006, ingest 8007, feedback 8008,
      traces 8009, search 8010, astro 8011, bigbrother 8012, erp 8013,
      healthwatch 8014) + Python extractor (NATS-only)
INFRA: SGLang (GPU) | Postgres platform + per-tenant
       Redis platform + per-tenant | NATS+JetStream | MinIO
```

Authoritative routing: `deploy/traefik/dynamic/dev.yml` (dev) and Docker
labels on each service in `deploy/docker-compose.prod.yml` (prod).

## Service roles

- **Edge / identity:** `auth` issues JWTs (see auth-jwt.md), Traefik routes by
  subdomain (see gateway-traefik.md).
- **Real-time bus:** `ws` bridges NATS subjects to WebSocket clients (see
  websocket-hub.md, nats-events.md).
- **Conversational core:** `chat` stores sessions/messages, `agent` runs the
  LLM tool loop, `search` walks document trees, `ingest` + `extractor` build
  those trees (see rag-tree-search.md, llm-sglang.md).
- **Platform plane:** `platform` manages tenants/modules/flags, `traces`
  records execution traces, `feedback` aggregates quality, `notification`
  dispatches in-app and email, `healthwatch` triages system health.
- **Verticals:** `astro`, `bigbrother`, `erp` host industry-specific
  capabilities and expose tools to the agent.

## Shared packages (`pkg/`)

Every service composes the same primitives: `jwt`, `tenant`, `database`,
`middleware`, `nats`, `llm`, `storage`, `otel`, `health`, `server`,
`security`, `metrics`, `cache`, `config`, `crypto`, `audit`, `guardrails`,
`pagination`, `httperr`, `traces`, `featureflags`, `approval`, `plc`,
`remote`, `grpc`, `build`, `export`. See `../packages/` for per-package
contracts.

## Cross-cutting invariants

- Tenant isolation is enforced top-to-bottom (see multi-tenancy.md).
- JWT is the single source of identity (see auth-jwt.md).
- All NATS subjects start with `tenant.{slug}.` (see nats-events.md).
- Every write publishes a NATS event so the WS Hub can fan out.
- Migrations come in `.up.sql` / `.down.sql` pairs (see database-postgres.md).
- Storage keys are `{tenant}/{docID}/...` (see storage-minio.md).

For end-to-end request traces (login, chat, ingestion), jump to
`../flows/`. For per-service contracts, jump to `../services/`.
