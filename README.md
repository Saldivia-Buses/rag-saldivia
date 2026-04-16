<p align="center">
  <strong>SDA Framework</strong><br>
  <em>Enterprise AI platform — multi-tenant, inhouse-powered, modular</em>
</p>

<p align="center">
  <a href="docs/README.md">Docs</a> &middot;
  <a href="docs/architecture/overview.md">Architecture</a> &middot;
  <a href="docs/CHANGELOG.md">Changelog</a>
</p>

---

SDA Framework is a multi-tenant SaaS platform that combines AI services with
industry-specific business modules. Built as Go microservices running inhouse on
an NVIDIA RTX PRO 6000 (96GB VRAM, 256GB RAM), with Next.js frontends in the cloud.

The cost of every AI query is electricity — no per-token bills. This lets us
offer more AI for less than any cloud-based competitor, while keeping all
customer data on controlled infrastructure.

## Architecture

```
CLOUD                                    INHOUSE (workstation)
┌──────────────┐                         ┌────────────────────────────────────┐
│  Next.js     │                         │  Traefik (gateway)                 │
│  Frontend    │──── REST/WS ──────────►│    ├── Auth        (JWT, MFA, RBAC)│
│  (CDN)       │   Cloudflare Tunnel    │    ├── WebSocket Hub (real-time)   │
└──────────────┘                         │    ├── Chat        (sessions, msgs)│
                                         │    ├── Agent       (LLM + tools)  │
                                         │    ├── Search      (tree RAG)     │
                                         │    ├── Ingest      (doc pipeline) │
                                         │    ├── Notification (push + email)│
                                         │    ├── Platform    (tenant mgmt)  │
                                         │    ├── Traces      (observability)│
                                         │    ├── Feedback    (quality loop) │
                                         │    ├── Astro       (55+ astro AI) │
                                         │    └── Extractor   (Python OCR)   │
                                         │                                    │
                                         │  PostgreSQL per-tenant             │
                                         │  Redis (blacklist + cache)         │
                                         │  NATS JetStream (events)           │
                                         │  MinIO (S3 storage)                │
                                         │  SGLang (GPU model server)         │
                                         └────────────────────────────────────┘
```

## Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.25 — chi router, sqlc queries, slog logging |
| **Database** | PostgreSQL 16 per-tenant (physical isolation) |
| **Cache** | Redis 7 (token blacklist, config cache) |
| **Messaging** | NATS + JetStream (async events, per-service auth) |
| **Inter-service** | gRPC with JWT-forwarding interceptors |
| **Frontend** | Next.js + React + shadcn/ui + Tailwind + TanStack Query |
| **Gateway** | Traefik v3 (subdomain routing, rate limiting, TLS) |
| **Auth** | Ed25519 JWT + bcrypt + TOTP MFA + token blacklist |
| **RAG** | Tree-based search (no vectors) — LLM navigates doc structure |
| **LLM** | Model-agnostic via SGLang (OpenAI-compatible API) |
| **Storage** | MinIO (S3-compatible, migrable to AWS) |
| **Observability** | OpenTelemetry + Grafana (Tempo + Prometheus + Loki) |
| **Security** | CrowdSec IDS, Docker socket proxy, distroless containers |
| **CI/CD** | GitHub Actions (build, vet, test, gosec, trivy) |
| **CLI** | Go (Cobra) — `sda` binary for system management |

## Services

| Service | Port | gRPC | Purpose |
|---|---|---|---|
| **auth** | 8001 | — | JWT, RBAC, MFA, brute force protection, token blacklist |
| **ws** | 8002 | — | WebSocket Hub — real-time channel, mutation routing via gRPC |
| **chat** | 8003 | 50052 | Sessions, messages, history |
| **agent** | 8004 | — | LLM runtime + tool execution (search, ingest, notify, astro) |
| **notification** | 8005 | — | In-app + email push, NATS consumer, user preferences |
| **platform** | 8006 | — | Tenant registry, modules, feature flags, global config |
| **ingest** | 8007 | — | Document upload, extraction pipeline, tree generation |
| **feedback** | 8008 | — | Quality metrics, health scores, alerting |
| **traces** | 8009 | — | Execution traces, cost tracking, NATS consumer |
| **search** | 8010 | 50051 | Tree-based document search (3-phase: navigate → extract → cite) |
| **astro** | 8011 | — | Astrological intelligence (55+ techniques, 64 endpoints, CGO) |
| **extractor** | 8012 | — | Document extraction (Python, OCR + vision) |

## Quick start

```bash
# Development (infra in Docker, services on host)
make dev              # start PostgreSQL, Redis, NATS, Traefik, MinIO
go run ./services/auth/cmd/...    # run a service

# Full stack in Docker
make dev-full         # all services containerized

# Testing
make test             # all Go tests
make test-auth        # single service
make test-coverage    # with HTML report

# Build
make build            # all services → ./bin/
make proto            # regenerate gRPC code from protos

# Deploy
make deploy           # production deployment
make status           # service health + GPU status
```

## Project structure

```
services/                    Go microservices (10 services)
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
  extractor/                 Document extraction (Python, OCR)

pkg/                         Shared Go packages
  jwt/                       Ed25519 JWT sign/verify
  middleware/                 Auth, RBAC, rate limiting, security headers
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
  otel/                      OpenTelemetry setup

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
- **Token blacklist** via Redis in all 10 services (HTTP + gRPC)
- **MFA/TOTP** with encrypted secrets
- **RBAC** with granular permissions per handler
- **Tenant isolation** — separate PostgreSQL per tenant, JWT cross-validation with subdomain
- **Rate limiting** — login 5/min, refresh 10/min, MFA 5/min, AI 30/min
- **NATS per-service auth** — 10 users with least-privilege publish/subscribe
- **CrowdSec IDS** — intrusion detection on Traefik access logs
- **Docker socket proxy** — Traefik can't access full Docker API
- **Distroless containers** — no shell, no package manager, non-root
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
