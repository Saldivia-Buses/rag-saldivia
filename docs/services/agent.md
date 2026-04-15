---
title: Service: agent
audience: ai
last_reviewed: 2026-04-15
related:
  - ../flows/chat-agent-pipeline.md
  - ../architecture/llm-sglang.md
  - ../packages/llm.md
  - ../packages/guardrails.md
  - ../packages/traces.md
---

## Purpose

Agent Runtime: LLM-driven loop that interprets user queries, calls registered
tools (search, ingest, module-specific), enforces guardrails, and streams
results back. Replaces the deprecated `services/rag/` (legacy NVIDIA Blueprint
wrapper). Read this when changing tool wiring, the LLM reasoning loop,
multi-turn tool-call limits, trace publishing, or how module manifests
(`modules/*/tools.yaml`) extend the agent's capabilities.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + nats/redis check |
| POST | `/v1/agent/query` | JWT + `chat.read` (30/min user) | Run agent loop on a user message |
| POST | `/v1/agent/confirm` | JWT + `chat.read` (30/min user) | Approve a deferred tool call (action tools with `RequiresConfirmation`) |

Routes mounted at `/v1/agent` in `services/agent/cmd/main.go:161`. Handler
factory at `services/agent/internal/handler/agent.go:27`.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.{slug}.traces.start` | pub | Agent loop start (`services/agent/internal/service/agent.go:91`) |
| `tenant.{slug}.traces.end` | pub | Agent loop end with cost/tokens summary |
| `tenant.{slug}.feedback.usage` | pub | Per-query token + cost telemetry |
| `tenant.{slug}.feedback.error_report` | pub | Loop error / guardrail trip |

All publishing goes through `pkg/traces.Publisher`. Agent does not subscribe.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `AGENT_PORT` | no | `8004` | HTTP listener port |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key for JWT verification |
| `NATS_URL` | no | `nats://localhost:4222` | NATS for traces/feedback |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |
| `SGLANG_LLM_URL` | no | `http://localhost:8102` | OpenAI-compatible LLM endpoint |
| `SGLANG_LLM_MODEL` | no | — | Model name passed to LLM (empty = server default) |
| `LLM_API_KEY` | no | — | Bearer token for cloud LLM (unused for SGLang) |
| `SEARCH_SERVICE_URL` | no | `http://localhost:8010` | Search HTTP endpoint for tool wiring |
| `SEARCH_GRPC_URL` | no | — | Optional gRPC target — overrides HTTP for `search_documents` |
| `INGEST_SERVICE_URL` | no | `http://localhost:8007` | Ingest HTTP endpoint |
| `NOTIFICATION_SERVICE_URL` | no | `http://localhost:8005` | Reserved (notification is NATS-driven now) |
| `ASTRO_SERVICE_URL` | no | `http://localhost:8011` | Used by `modules/astro` tool manifest |
| `BIGBROTHER_SERVICE_URL` | no | `http://localhost:8012` | Used by `modules/bigbrother` tool manifest |
| `ERP_SERVICE_URL` | no | `http://localhost:8013` | Used by `modules/erp` tool manifest |
| `MODULES_DIR` | no | `modules` | Where to load `*/tools.yaml` from |
| `ENABLED_MODULES` | no | `""` (= all) | Comma-separated module IDs, `"none"` to disable all |
| `SYSTEM_PROMPT` | no | (Spanish default) | System prompt for the LLM |

See `services/agent/cmd/main.go:48` for full env wiring.

## Dependencies

- **No PostgreSQL.** Agent is stateless — chat history lives in `chat`,
  documents in `ingest`/tenant DB, tracing in `traces`.
- **Redis** — token blacklist for revoked JWTs.
- **NATS** — publishes traces and feedback events.
- **Search service** (HTTP or gRPC) — `search_documents` tool.
- **Ingest service** (HTTP) — `create_ingest_job`, `check_job_status` tools.
- **Per-module services** — invoked by tools loaded from
  `modules/*/tools.yaml` via `services/agent/internal/tools.LoadModuleTools`.
- **LLM** (SGLang or OpenAI-compatible) — chat completions + tool calling.

## Permissions used

- `chat.read` — gates both `/v1/agent/query` and `/v1/agent/confirm`
  (`services/agent/internal/handler/agent.go:30`).

Tool authorization (per-module) is enforced downstream by each target service;
agent forwards the caller's bearer token unchanged.
