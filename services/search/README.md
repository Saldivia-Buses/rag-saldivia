# Search Service

## What it does

Tree-based document search, inspired by PageIndex. Navigates hierarchical
document trees using LLM reasoning, then extracts relevant pages. No vectors
— the LLM reads structural metadata to find information.

Version: 0.1.0 | Port: 8010 | gRPC: 50051

## Architecture

3-phase search pipeline:
1. **Navigate** — LLM reads tree TOC (titles + summaries only), selects nodes
2. **Extract** — Code maps nodes → page ranges, reads from `document_pages`
3. **Return** — Assembles selections with citations (document, pages, sections)

## Endpoints

### HTTP

| Path | Method | Auth | Description |
|---|---|---|---|
| `/health` | GET | No | Health check |
| `/v1/search/query` | POST | JWT | Search documents |

### gRPC (port 50051)

| Service | Method | Description |
|---|---|---|
| `SearchService` | `Search` | Same as HTTP but via gRPC (used by Agent Runtime) |

### POST /v1/search/query

```json
{
  "query": "medida del disco de freno delantero",
  "collection_id": "optional-collection-id",
  "max_nodes": 5
}
```

## Environment

| Variable | Required | Default | Description |
|---|---|---|---|
| `SEARCH_PORT` | No | `8010` | HTTP listen port |
| `SEARCH_GRPC_PORT` | No | `50051` | gRPC listen port |
| `POSTGRES_TENANT_URL` | Yes | — | Tenant DB connection |
| `JWT_PUBLIC_KEY` | Yes | — | Ed25519 public key (base64) |
| `SGLANG_LLM_URL` | No | `http://localhost:8102` | LLM endpoint |
| `SGLANG_LLM_MODEL` | No | — | Model ID for tree navigation |
| `REDIS_URL` | No | `localhost:6379` | Redis for JWT blacklist |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry |

## Dependencies

- **PostgreSQL** — document trees and pages (per-tenant)
- **LLM** — tree navigation reasoning
- **Redis** — JWT blacklist
