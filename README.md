<p align="center">
  <strong>SDA Framework</strong><br>
  <em>Enterprise AI platform — single-tenant per container, inhouse-powered, all-in-one</em>
</p>

<p align="center">
  <a href="docs/README.md">Docs</a> &middot;
  <a href="docs/architecture/overview.md">Architecture</a> &middot;
  <a href="docs/CHANGELOG.md">Changelog</a>
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

| Service | Port | gRPC | Purpose |
|---|---|---|---|
| **auth** | 8001 | — | JWT, RBAC, MFA, brute force protection, token blacklist |
| **ws** | 8002 | — | WebSocket Hub — real-time channel, mutation routing via gRPC |
| **chat** | 8003 | 50052 | Sessions, messages, history |
| **agent** | 8004 | — | LLM runtime + tool execution (search, ingest, notify) |
| **notification** | 8005 | — | In-app + email push, NATS consumer, user preferences |
| **platform** | 8006 | — | Tenant registry, modules, feature flags, global config |
| **ingest** | 8007 | — | Document upload, extraction pipeline, tree generation |
| **feedback** | 8008 | — | Quality metrics, health scores, alerting |
| **traces** | 8009 | — | Execution traces, cost tracking, NATS consumer |
| **search** | 8010 | 50051 | Tree-based document search (3-phase: navigate → extract → cite) |
| **bigbrother** | 8012 | — | Network intelligence — passive device discovery on the LAN |
| **erp** | 8013 | — | ERP modules — Histrix replacement for Saldivia business domain |
| **healthwatch** | 8014 | — | Platform-level health monitoring + AI triage |
| **extractor** | 8090 | — | Document extraction (Python sidecar, OCR + vision) |

> Service count will shrink to 3–5 internal modules of a single binary as the
> consolidation from ADR 021 + ADR 024 progresses. The table above is today's
> snapshot, not the target.

## Quick start

```bash
# Development (infra in Docker, services on host)
make dev              # start PostgreSQL, Redis, NATS, Traefik, MinIO
make dev-services     # all Go services on host
make dev-frontend     # Next.js on localhost:3000

# Testing
make test             # all Go tests
make test-auth        # single service
make test-coverage    # with HTML report

# Build
make build            # all services → ./bin/
make proto            # regenerate gRPC code from protos

# Deploy (per-tenant all-in-one container — ADR 023/024)
make deploy TENANT=saldivia   # stop + rm + run with host volumes mounted
make deploy TENANT=all        # iterate over deploy/tenants/*/
make status TENANT=saldivia   # docker ps + /readyz probe
make logs TENANT=saldivia     # s6-overlay-prefixed logs
```

## Project structure

```
services/                    Go microservices (14 today, consolidating to 3-5 — ADR 021)
  auth/                      Auth Gateway + RBAC + MFA
  ws/                        WebSocket Hub + mutation routing
  chat/                      Chat sessions + messages
  agent/                     Agent Runtime (LLM + tools)
  search/                    Tree-based document search
  ingest/                    Document pipeline + extraction
  notification/              In-app + email notifications
  platform/                  Tenant management (admin only)
  traces/                    Execution traces + cost tracking
  feedback/                  Quality metrics + alerting
  bigbrother/                Network intelligence (LAN device discovery)
  erp/                       ERP modules (Histrix replacement)
  healthwatch/               Platform health monitoring + AI triage
  extractor/                 Document extraction (Python sidecar, OCR + vision)

pkg/                         Shared Go packages (24 today, target ~10 — ADR 021)
  jwt/                       Ed25519 JWT sign/verify
  middleware/                Auth, RBAC, rate limiting, security headers
  grpc/                      gRPC interceptors, server/client factories
  tenant/                    Multi-tenant context + DB resolver
  nats/                      NATS publisher + DLQ + subject validation
  audit/                     Audit log writer
  guardrails/                Input/output validation, prompt injection detection
  config/                    Config resolver with scope cascade
  security/                  Token blacklist, brute force
  llm/                       OpenAI-compatible LLM client
  storage/                   S3-compatible file storage
  pagination/                Query pagination helpers
  crypto/                    AES-256-GCM encryption
  server/                    Service bootstrap (chi, OTel, signal ctx, graceful shutdown)

gen/go/                      Generated gRPC code (7 protos, 14 files)
proto/                       Protocol Buffer definitions
apps/web/                    Next.js frontend (login lives at /login route)
modules/                     Industry vertical tool manifests (YAML)
tools/cli/                   CLI binary (sda)
tools/mcp/                   MCP Server for AI tooling
deploy/                      Docker Compose, Traefik, scripts, secrets
docs/                        Plans, decisions, artifacts, changelog
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
| [Docs index](docs/README.md) | Entry point to all modular documentation |
| [Architecture overview](docs/architecture/overview.md) | System map and tech choices |
| [Conventions](docs/conventions/) | Go, frontend, git, testing, security rules |
| [Operations](docs/operations/) | Deploy, runbook, monitoring, backups, incidents |
| [Changelog](docs/CHANGELOG.md) | Platform-level release notes |
| [CLAUDE.md](CLAUDE.md) | Agent-optimized project primer |

## License

Private repository. All rights reserved.
