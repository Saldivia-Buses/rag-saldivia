---
title: SDA Framework Documentation Index
audience: ai
last_reviewed: 2026-04-15
related:
  - ./glossary.md
  - ../CLAUDE.md
---

Entry point for all SDA Framework documentation. Audience is AI agents working on this codebase. Every doc is ≤200 lines, English, one aspect per file. Read this first, follow relevant links.

## Quick start

- New to the repo? Read [architecture/overview.md](./architecture/overview.md).
- Unfamiliar term? Check [glossary.md](./glossary.md).
- About to edit code? Find the relevant service or package below, read the related flow, then edit.
- About to open a PR? Read [conventions/git.md](./conventions/git.md).

## Architecture

| Doc | Topic |
|---|---|
| [overview](./architecture/overview.md) | System map: services, packages, deployment |
| [multi-tenancy](./architecture/multi-tenancy.md) | Tenant isolation, slug routing, per-tenant DB |
| [auth-jwt](./architecture/auth-jwt.md) | JWT with ed25519, refresh, MFA, RBAC |
| [nats-events](./architecture/nats-events.md) | Subjects, JetStream, tenant namespacing |
| [websocket-hub](./architecture/websocket-hub.md) | WS Hub, subscriptions, frontend bridge |
| [rag-tree-search](./architecture/rag-tree-search.md) | PageIndex-style reasoning, no vectors |
| [llm-sglang](./architecture/llm-sglang.md) | Model server, slot-per-step |
| [observability](./architecture/observability.md) | OTel + Tempo + Prom + Loki |
| [gateway-traefik](./architecture/gateway-traefik.md) | Gateway routing, TLS, Cloudflare Tunnel |
| [storage-minio](./architecture/storage-minio.md) | S3-compat object storage |
| [database-postgres](./architecture/database-postgres.md) | Per-tenant PostgreSQL, sqlc, migrations |

## Services

One doc per Go microservice. Covers: purpose, endpoints, NATS events published/consumed, dependencies, permissions.

[agent](./services/agent.md) · [astro](./services/astro.md) · [auth](./services/auth.md) · [bigbrother](./services/bigbrother.md) · [chat](./services/chat.md) · [erp](./services/erp.md) · [extractor](./services/extractor.md) · [feedback](./services/feedback.md) · [healthwatch](./services/healthwatch.md) · [ingest](./services/ingest.md) · [notification](./services/notification.md) · [platform](./services/platform.md) · [search](./services/search.md) · [traces](./services/traces.md) · [ws](./services/ws.md)

## Packages

One doc per shared Go package in `pkg/`. Covers: public API, invariants, usage examples.

[approval](./packages/approval.md) · [audit](./packages/audit.md) · [build](./packages/build.md) · [cache](./packages/cache.md) · [config](./packages/config.md) · [crypto](./packages/crypto.md) · [database](./packages/database.md) · [export](./packages/export.md) · [featureflags](./packages/featureflags.md) · [grpc](./packages/grpc.md) · [guardrails](./packages/guardrails.md) · [health](./packages/health.md) · [httperr](./packages/httperr.md) · [jwt](./packages/jwt.md) · [llm](./packages/llm.md) · [metrics](./packages/metrics.md) · [middleware](./packages/middleware.md) · [nats](./packages/nats.md) · [otel](./packages/otel.md) · [pagination](./packages/pagination.md) · [plc](./packages/plc.md) · [remote](./packages/remote.md) · [security](./packages/security.md) · [server](./packages/server.md) · [storage](./packages/storage.md) · [tenant](./packages/tenant.md) · [traces](./packages/traces.md)

## Flows

End-to-end request flows that span multiple services. Read before editing any code in the flow.

| Flow | Files involved |
|---|---|
| [login-jwt](./flows/login-jwt.md) | `services/auth/internal/` |
| [tenant-routing](./flows/tenant-routing.md) | `pkg/tenant/`, `pkg/middleware/`, `pkg/database/` |
| [chat-agent-pipeline](./flows/chat-agent-pipeline.md) | `services/chat/`, `services/agent/` |
| [document-ingestion](./flows/document-ingestion.md) | `services/ingest/`, `services/extractor/` |
| [websocket-realtime](./flows/websocket-realtime.md) | `services/ws/`, `pkg/nats/` |
| [deploy-pipeline](./flows/deploy-pipeline.md) | `deploy/`, `.github/workflows/` |
| [self-healing-triage](./flows/self-healing-triage.md) | `services/healthwatch/` |

## Conventions

Rules for writing code. Linted and enforced by invariants where possible.

[go](./conventions/go.md) · [frontend](./conventions/frontend.md) · [git](./conventions/git.md) · [migrations](./conventions/migrations.md) · [testing](./conventions/testing.md) · [sqlc](./conventions/sqlc.md) · [error-handling](./conventions/error-handling.md) · [logging](./conventions/logging.md) · [security](./conventions/security.md)

## Operations

For deploy, incidents, and running the system.

[deploy](./operations/deploy.md) · [runbook](./operations/runbook.md) · [monitoring](./operations/monitoring.md) · [backup-restore](./operations/backup-restore.md) · [incidents](./operations/incidents.md)

## AI tooling

Documentation of the Claude Code integration: agents, skills, hooks, invariants, MCP servers, memory.

[agents](./ai/agents.md) · [skills](./ai/skills.md) · [hooks](./ai/hooks.md) · [invariants](./ai/invariants.md) · [mcp-servers](./ai/mcp-servers.md) · [memory-system](./ai/memory-system.md)

## Plans

Historical record of planned and executed work. Live in [plans/](./plans/). Not maintained after completion.

## Rules for this documentation

1. Every doc ≤200 lines including frontmatter. If it doesn't fit, rewrite tighter.
2. Every doc has frontmatter with `title`, `audience: ai`, `last_reviewed`, `related`.
3. English only.
4. No redundancy — each fact lives in one doc, others link.
5. Code references as `path/to/file.go:42`, not embedded source blocks.
6. First paragraph after frontmatter = "what this is + when to read it".
7. A `doc-sync` agent keeps these in sync with code on every commit; a daily audit detects drift.
