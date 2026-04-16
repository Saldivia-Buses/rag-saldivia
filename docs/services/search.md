---
title: Service: search
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/rag-tree-search.md
  - ../architecture/llm-sglang.md
  - ../packages/grpc.md
  - ./agent.md
---

## Purpose

PageIndex-inspired tree-RAG search over ingested documents. The agent calls
this for `search_documents`. There are no vectors — the LLM navigates a
hierarchical document tree and selects relevant sections, returning text +
citations. Exposes both an HTTP API (for tests, debugging, and
human-callable surfaces) and a gRPC server (`SEARCH_GRPC_PORT`) that the
agent prefers when configured. Read this when changing the search algorithm,
the LLM tree-walk strategy, or the contract with the agent.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/redis check |
| POST | `/v1/search/query` | JWT + `chat.read` (30/min user) | Tree search; returns selected nodes + cited text |

Routes registered in `services/search/internal/handler/search.go:29`. HTTP
uses `FailOpen=true` so the agent can still query during a Redis blip.

A gRPC service (`searchv1.SearchService`) is registered at
`services/search/cmd/main.go:72` on `SEARCH_GRPC_PORT` (default `50051`).
The agent prefers gRPC if `SEARCH_GRPC_URL` is set
(`services/agent/cmd/main.go:104`).

## NATS events

Search neither publishes nor subscribes to NATS — it is a stateless query
service. Audit events are written to the platform/tenant DB through
`pkg/audit`.

| Subject | Direction | Trigger |
|---|---|---|
| (none) | — | — |

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `SEARCH_PORT` | no | `8010` | HTTP listener port |
| `SEARCH_GRPC_PORT` | no | `50051` | gRPC listener port |
| `POSTGRES_TENANT_URL` | yes | — | Reads `documents`, `document_pages`, `document_trees` |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | not used | — | (no NATS dependency) |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |
| `SGLANG_LLM_URL` | no | `http://localhost:8102` | LLM endpoint for tree navigation |
| `SGLANG_LLM_MODEL` | no | — | Model name (empty = server default) |

## Dependencies

- **PostgreSQL tenant** — read-only access to documents, pages, and trees
  produced by `ingest`.
- **Redis** — token blacklist.
- **LLM** (SGLang) — drives node selection during a tree walk.
- **Agent service (incoming)** — primary HTTP/gRPC client.
- **`pkg/audit`** — every query is recorded.

## Permissions used

- `chat.read` — same permission as the chat surface, since search backs the
  agent which is consumed by chat. Enforced inline at
  `services/search/internal/handler/search.go:32`.
