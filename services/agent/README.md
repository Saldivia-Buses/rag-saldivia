# Agent Runtime

## What it does

LLM-powered agent that orchestrates tool calls across all SDA services.
Receives user messages, runs guardrails, calls LLM with available tools,
executes tool calls against other services, publishes execution traces,
and returns responses via SSE streaming.

Version: 0.1.0 | Port: 8004

## Architecture

```
User message → guardrails → LLM (via pkg/llm) → tool calls?
  YES → execute tools (Search, Ingest, Notify, etc.) → feed results → LLM → ...
  NO  → output guardrails → response
```

The loop runs up to `max_loop_iterations` times or until the LLM returns text.

## Tool System

Tools are loaded from two sources:

1. **Core tools** (hardcoded): `search_documents`, `create_ingest_job`, `check_job_status`, `send_notification`
2. **Module tools** (YAML manifests): loaded from `modules/*/tools.yaml` for enabled modules

### Module Tool Loading

The loader (`internal/tools/loader.go`) supports two protocols:
- **gRPC**: `method: SearchVehicles` → POST to `baseURL/method`
- **HTTP**: `endpoint: POST /v1/fleet/vehicles` → verb + path parsed from endpoint field

Currently enabled modules:
- `fleet` — transport/logistics tools
- `erp` — ERP business modules
- `bigbrother` — network intelligence

### Tool Execution

The executor (`internal/tools/executor.go`) handles:
- HTTP tool calls with JWT forwarding
- gRPC calls for search (optional, falls back to HTTP)
- Confirmation flow for dangerous tools (`requires_confirmation: true`)
- 30s timeout per tool call

## Trace Publishing

Publishes execution traces to NATS via `pkg/traces/publisher.go`:
- `traces.start` — query received
- `traces.event` — each LLM call + tool call
- `traces.end` — final response with token counts + cost
- `feedback.*` — quality metrics

## Endpoints

| Path | Method | Auth | Description |
|---|---|---|---|
| `/health` | GET | No | Health check |
| `/v1/agent/query` | POST | JWT | Run a query through the agent (SSE) |

### POST /v1/agent/query

```json
{
  "message": "cuántos buses están en producción",
  "history": []
}
```

## Environment

| Variable | Required | Default | Description |
|---|---|---|---|
| `AGENT_PORT` | No | `8004` | HTTP listen port |
| `JWT_PUBLIC_KEY` | Yes | — | Ed25519 public key (base64) |
| `SGLANG_LLM_URL` | No | `http://localhost:8102` | LLM endpoint |
| `SGLANG_LLM_MODEL` | No | — | Model ID |
| `LLM_API_KEY` | No | — | API key for cloud models |
| `SEARCH_SERVICE_URL` | No | `http://localhost:8010` | Search service |
| `INGEST_SERVICE_URL` | No | `http://localhost:8007` | Ingest service |
| `NOTIFICATION_SERVICE_URL` | No | `http://localhost:8005` | Notification service |
| `SEARCH_GRPC_URL` | No | — | gRPC endpoint for search (optional) |
| `NATS_URL` | No | `nats://localhost:4222` | NATS for trace publishing |
| `MODULES_DIR` | No | `modules` | Directory with module tool manifests |
| `REDIS_URL` | No | `localhost:6379` | Redis for token blacklist |
| `SYSTEM_PROMPT` | No | — | Override default system prompt |

## Dependencies

- **Search Service** — document search tool
- **Ingest Service** — document upload tool
- **Notification Service** — notification sending tool
- **NATS** — trace publishing (optional, degrades gracefully)
- **Redis** — JWT blacklist
