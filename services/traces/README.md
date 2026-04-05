# Traces Service

Observability for the intelligence layer. Receives trace events via NATS,
persists to Platform DB, and exposes REST API for trace listing, detail,
and cost tracking.

## Architecture

- NATS subscriber: `traces.start.*`, `traces.end.*`, `traces.event.*`
- Platform DB: `execution_traces` + `trace_events`
- REST API for querying traces and costs

## Endpoints

| Path | Method | Description |
|---|---|---|
| `/health` | GET | Health check |
| `/v1/traces` | GET | List traces (query: tenant_id, limit, offset) |
| `/v1/traces/{traceID}` | GET | Trace detail with all events |
| `/v1/traces/costs/{tenantID}` | GET | Cost summary (query: from, to) |

## NATS Subjects

| Subject | Direction | Payload |
|---|---|---|
| `traces.start.{tenant}` | Subscribe | TraceStartEvent |
| `traces.end.{tenant}` | Subscribe | TraceEndEvent |
| `traces.event.{tenant}` | Subscribe | TraceEvent |

## Environment

| Variable | Required | Description |
|---|---|---|
| `TRACES_PORT` | No | Default: `8009` |
| `POSTGRES_PLATFORM_URL` | Yes | Platform DB connection |
| `NATS_URL` | No | Default: `nats://localhost:4222` |
| `JWT_PUBLIC_KEY` | Yes | Ed25519 public key (base64) |
