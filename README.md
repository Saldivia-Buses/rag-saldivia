<p align="center">
  <strong>SDA Framework</strong><br>
  <em>Enterprise AI platform — single-tenant per container, inhouse-powered, all-in-one</em>
</p>

<p align="center">
  <a href="CLAUDE.md">Primer</a> &middot;
  <a href="docs/decisions/">ADRs</a> &middot;
  <a href="docs/CHANGELOG.md">Changelog</a> &middot;
  <a href="https://github.com/Saldivia-Buses/rag-saldivia/releases">Releases</a>
</p>

---

SDA Framework is an enterprise AI platform delivered as **one container per
tenant**. Each tenant container carries the entire product — Go backend,
Next.js frontend, Postgres, NATS, Redis, MinIO — supervised by `s6-overlay`.
State lives on the host volume; the container is reproducible and replaceable.
See ADR 022, 023, 024.

Runs inhouse on an NVIDIA RTX PRO 6000 (96 GB VRAM, 256 GB RAM). The cost of
every AI query is electricity — no per-token bills. All customer data stays on
controlled infrastructure, with physical isolation per tenant (separate
container, separate Postgres, separate everything).

## Architecture

```
HOST WORKSTATION
┌─────────────────────────────────────────────────────────────────┐
│  Traefik edge (shared) — routes {tenant}.sda.local → container  │
│                                                                 │
│  ┌──────────────────────────── sda:<version> ─────────────────┐ │
│  │  Tenant container (s6-overlay supervises)                  │ │
│  │                                                            │ │
│  │    :80  Go app — chi + reverse-proxy for non-API paths     │ │
│  │         ├── /v1/*   → in-process handlers                  │ │
│  │         ├── /ws/*   → in-process handlers                  │ │
│  │         └── /*      → proxy to Next.js (:3000)             │ │
│  │                                                            │ │
│  │    :3000  Next.js standalone (SSR + RSC)                   │ │
│  │                                                            │ │
│  │    Unix sockets / localhost:                               │ │
│  │         Postgres   NATS JetStream   Redis   MinIO          │ │
│  │                                                            │ │
│  │    Volumes (host) — /data/<tenant>/{postgres,nats,…}       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                 │
│  SGLang model servers (GPU, shared across tenants)              │
└─────────────────────────────────────────────────────────────────┘
```

## Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.25 — chi router, sqlc queries, slog logging |
| **Database** | PostgreSQL 16 per-tenant (physical isolation) |
| **Cache** | Redis 7 (token blacklist, config cache) |
| **Messaging** | NATS + JetStream (async events, per-service auth) |
| **Inter-service** | in-process (after consolidation into the monolith — ADR 021); gRPC legacy for cross-service where it still exists |
| **Frontend** | Next.js + React + shadcn/ui + Tailwind + TanStack Query (bundled inside the container — ADR 024) |
| **Gateway** | Traefik v3 on the host (subdomain-per-tenant); inside the container, the Go app handles `:80` routing |
| **Auth** | Ed25519 JWT + bcrypt + TOTP MFA + token blacklist |
| **RAG** | Tree-based search (no vectors) — LLM navigates doc structure |
| **LLM** | Model-agnostic via SGLang (OpenAI-compatible API) |
| **Storage** | MinIO (S3-compatible, migrable to AWS) |
| **Observability** | OpenTelemetry + Grafana (Tempo + Prometheus + Loki) |
| **Security** | CrowdSec IDS on host, per-tenant container isolation, separate Postgres per tenant |
| **CI/CD** | GitHub Actions (build, vet, test, gosec, trivy) |
| **CLI** | Go (Cobra) — `sda` binary for system management |

## Services

After ADR 021 + ADR 025 the 14-service microservice set is folded into
**3 deployable services** (down from 14):

| Service | Purpose | Internal modules |
|---|---|---|
| **app** | Go monolith — auth + platform + feedback + RAG (ingest/search/agent) + realtime (chat/ws/notification) + ops (bigbrother/healthwatch/traces) | `internal/core`, `internal/rag`, `internal/realtime`, `internal/ops`, plus `internal/events`, `internal/llm`, `internal/outbox`, `internal/spine`, `internal/guardrails`, `internal/httperr` |
| **erp** | ERP (Histrix replacement for Saldivia business domain). Standalone for now — 38k LOC, fusion into `app` deferred until the ERP shape stabilises | — |
| **extractor** | Document extraction (Python sidecar, OCR + vision via SGLang) | — |

The consolidated `app` binary owns all former HTTP and gRPC endpoints
in-process. Only the ERP fusion is left as an open consolidation item
(see ADR 025 + `docs/decisions/025-consolidation-shape-target.md`).

## Quick start

```bash
# Development (infra in Docker, services on host)
make dev              # infra only (Postgres, Redis, NATS, Traefik, MinIO)
make dev-services     # Go services (app + erp) on host
make dev-frontend     # Next.js on localhost:3000
make dev-all          # infra + services + frontend

# Testing
make test             # Go tests across pkg + services/app + services/erp
make lint             # golangci-lint across the workspace
make build            # build all Go binaries

# Database
make migrate          # run platform + tenant migrations
make sqlc             # regenerate sqlc after query changes

# Status
make status           # service health + GPU
make versions         # running image versions vs available
make check-prod-drift # Phase 0 gate: workstation SHA == origin/main

# Deploy (per-tenant all-in-one container — ADR 023/024)
make deploy           # production deploy
```

## Project structure

```
services/                    3 deployable services (down from 14 — ADR 025)
  app/                       Go monolith (5 internal domain modules)
    internal/core/           auth + platform + feedback
    internal/rag/            ingest + search + agent
    internal/realtime/       chat + ws + notification
    internal/ops/            bigbrother + healthwatch + traces
    internal/{events,llm,outbox,spine,guardrails,httperr}
                             supporting infra (codegen, LLM client,
                             transactional outbox, NATS event spine, …)
  erp/                       ERP standalone (38k LOC, fusion deferred)
  extractor/                 Document extraction (Python sidecar)

pkg/                         13 shared Go packages (down from 24 — ADR 021/025)
  audit/                     Audit log writer
  config/                    Config resolver with scope cascade
  crypto/                    AES-256-GCM encryption
  database/                  pgx/pgxpool helpers, health probes
  health/                    Health/readiness probes, HTTP handlers
  jwt/                       Ed25519 JWT sign/verify
  middleware/                Auth, RBAC, rate limiting, security headers
  nats/                      NATS publisher + DLQ + subject validation
  pagination/                Query pagination helpers
  security/                  Token blacklist, brute force protection
  server/                    Service bootstrap (chi, OTel, signal ctx,
                             graceful shutdown, distroless healthcheck)
  tenant/                    Multi-tenant context + DB resolver
  traces/                    OTel tracing helpers

apps/web/                    Next.js 16 frontend (bundled inside the
                             per-tenant container — ADR 024)
modules/                     Industry vertical tool manifests (YAML) —
                             bigbrother, chat, construction, erp,
                             feedback, fleet, ingest, notifications, rag
tools/cli/                   `sda` CLI (migrations, admin, deploy)
tools/eventsgen/             CUE → Go/TS/Markdown event codegen
tools/mcp/                   MCP server for AI tooling
deploy/                      Docker Compose (dev/prod), Dockerfiles,
                             Traefik, s6-overlay scripts, secrets
docs/                        Plans, decisions (ADRs), parity (Phase 1),
                             CHANGELOG, next-session brief
```

## Security

The system passed its first security audit with an **APTO** (fit for production)
verdict after implementing Plan 08 (52 hardening findings).

Key security features:
- **Ed25519 JWT** with 15-min access tokens + 7-day refresh rotation
- **Token blacklist** via Redis inside the tenant container
- **MFA/TOTP** with encrypted secrets
- **RBAC** with granular permissions per handler
- **Tenant isolation** — one container per tenant, separate Postgres/NATS/Redis/MinIO
  per tenant (physical isolation — ADR 022, 023, 024)
- **Rate limiting** — login 5/min, refresh 10/min, MFA 5/min, AI 30/min
- **CrowdSec IDS** on the host — intrusion detection on Traefik access logs
- **Encrypted backups** — age encryption, SHA-256 checksums, MinIO upload

## Modules

The platform supports industry-specific modules activated per tenant:

| Category | Examples |
|---|---|
| **Core** (always active) | Chat + RAG, Auth, Notifications |
| **Platform** | Documents, Knowledge Base, Tasks, CRM, Workflows |
| **Vertical** | Fleet/Logistics, Construction, Professional Services |
| **AI Services** | Agents, Document AI, Speech, Vision, Optimization |

Each module has a YAML manifest (`modules/*/tools.yaml`) defining tools that
the Agent Runtime loads dynamically.

## Documentation

| Document | Purpose |
|---|---|
| [CLAUDE.md](CLAUDE.md) | Repo primer — agent harness, phases (ADR 026/027), commands, invariants |
| [ADR index](docs/decisions/) | Architectural decision records |
| [ADR 026](docs/decisions/026-sda-replaces-histrix.md) | The north star — SDA replaces Histrix |
| [ADR 027](docs/decisions/027-mvp-success-criteria.md) | Phased MVP checklist (Phase 0 → Phase 4) |
| [Changelog](docs/CHANGELOG.md) | Per-cycle release index, pointers to GitHub Releases |
| [Parity docs](docs/parity/) | Histrix → SDA coverage + waivers (Phase 1) |
| [Next-session brief](docs/next-session.md) | What the next cycle picks up |

## License

Private repository. All rights reserved.
