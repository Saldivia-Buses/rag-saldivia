# SDA Framework

Multi-tenant enterprise AI platform. Go microservices backend running inhouse
on NVIDIA RTX PRO 6000 (96GB VRAM), Next.js frontend in the cloud. RAG, agents,
document processing, and industry-specific modules — all AI-enhanced.

## Architecture

```
Cloud (CDN)                          Inhouse (workstation)
┌──────────────┐                     ┌────────────────────────────────┐
│  Next.js     │──── REST/WS ──────►│  Traefik → Go microservices   │
│  Frontend    │   Cloudflare       │  PostgreSQL/Redis per-tenant  │
└──────────────┘   Tunnel           │  NATS + Milvus + NIMs         │
                                     └────────────────────────────────┘
```

## Stack

| Component | Technology |
|---|---|
| Backend | Go (chi + sqlc + slog) |
| Database | PostgreSQL per-tenant |
| Cache/Queue | Redis + NATS JetStream |
| Frontend | Next.js + React + shadcn/ui + TanStack Query |
| Gateway | Traefik (subdomain routing per tenant) |
| RAG | NVIDIA RAG Blueprint v2.5.0 |
| LLM | Nemotron-Super-49B (NVIDIA API) |
| Agents | NeMo Agent Toolkit |
| Observability | OpenTelemetry + Grafana |

## Services

| Service | Purpose |
|---|---|
| **auth** | JWT, RBAC, MFA, brute force protection |
| **ws** | WebSocket Hub — real-time data channel |
| **chat** | Sessions, messages, history |
| **rag** | Proxy to NVIDIA Blueprint |
| **notification** | In-app + email push |
| **platform** | Tenant management (platform admins only) |
| **ingest** | Document upload + processing pipeline |

## Quick start

```bash
make dev          # start all services
make test         # run tests
make build        # build all services
make status       # check service health + GPU
```

## Repo structure

```
services/         Go microservices (auth, ws, chat, rag, ...)
pkg/              Shared Go packages (jwt, tenant, middleware, ...)
proto/            gRPC protobuf definitions
apps/web/         Next.js frontend
apps/login/       Isolated login page
ai/               Agent configs, guardrails, model profiles
modules/          Industry vertical modules
tools/            CLI (sda) + MCP Server
deploy/           Docker Compose, Traefik, deploy scripts
config/           NVIDIA Blueprint configs
docs/             Documentation, ADRs, plans
_archive/         1.0.x code (reference)
```

## Docs

- **[Spec](docs/plans/2.0.x-plan01-sda-framework.md)** — complete system specification
- **[Bible](docs/bible.md)** — permanent work rules and conventions
- **[Architecture decisions](docs/decisions/)** — ADRs
