# RAG Service

> Proxy to the NVIDIA RAG Blueprint. Authenticates requests, enforces RBAC, and streams SSE responses from the Blueprint back to the client. Stateless -- no database.

## Endpoints

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| GET | `/health` | No | -- | Health check (also checks Blueprint reachability) |
| POST | `/v1/rag/generate` | Bearer | `collections.read` | Send messages to RAG, stream SSE response |
| GET | `/v1/rag/collections` | Bearer | `collections.read` | List available collections for the tenant |

## SSE Streaming

`POST /v1/rag/generate` proxies the request to the Blueprint and streams the response as Server-Sent Events directly to the client. `WriteTimeout` is set to 0 (no limit) to support long-running streams.

Request body:
```json
{
  "messages": [{"role": "user", "content": "..."}],
  "collection": "contratos",
  "model": "optional-model-override"
}
```

Response: SSE stream with `Content-Type` from Blueprint (typically `text/event-stream`).

## Database

None. Stateless proxy.

## NATS Events

None. The RAG service does not publish or consume NATS events.

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `RAG_PORT` | No | `8004` | HTTP listen port |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `RAG_SERVER_URL` | No | `http://localhost:8081` | NVIDIA RAG Blueprint URL |
| `RAG_TIMEOUT_MS` | No | `120000` | Timeout for Blueprint requests (ms) |
| `RAG_API_KEY` | No | -- | API key for Blueprint (if required) |
| `RAG_MODEL` | No | -- | Default model override |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **NVIDIA RAG Blueprint:** Upstream service for retrieval-augmented generation
- **pkg/jwt:** Ed25519 key loading
- **pkg/middleware:** Auth middleware, RequirePermission, SecureHeaders

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests
```
