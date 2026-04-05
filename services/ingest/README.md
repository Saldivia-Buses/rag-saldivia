# Ingest Service

> Async document upload and processing pipeline. Stages files to disk, tracks jobs in tenant DB, and processes via NATS JetStream worker that forwards to the NVIDIA RAG Blueprint for vectorization.

## Endpoints

All authenticated routes require Bearer auth with `ingest.write` permission.

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| GET | `/health` | No | -- | Health check |
| POST | `/v1/ingest/upload` | Bearer | `ingest.write` | Multipart document upload (returns 202 Accepted) |
| GET | `/v1/ingest/jobs` | Bearer | `ingest.write` | List jobs for user (`?limit=50`) |
| GET | `/v1/ingest/jobs/{jobID}` | Bearer | `ingest.write` | Get job status |
| DELETE | `/v1/ingest/jobs/{jobID}` | Bearer | `ingest.write` | Delete job record |

## Upload

```bash
curl -X POST http://localhost:8007/v1/ingest/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@document.pdf" \
  -F "collection=contratos"
```

**Max upload size:** 100MB. **Allowed extensions:** `.pdf`, `.docx`, `.doc`, `.txt`, `.md`, `.csv`, `.xlsx`, `.pptx`, `.html`, `.json`, `.xml`

Returns 202 with job object. Poll `GET /v1/ingest/jobs/{id}` or subscribe to WebSocket channel `ingest.jobs` for real-time status.

## Processing Pipeline

```
Client -> POST /v1/ingest/upload -> Handler
                                      |
                                 1. Stage file to disk
                                 2. Create job (status: pending)
                                 3. Publish to NATS JetStream
                                 4. Return 202 Accepted
                                      |
                                 NATS JetStream
                                 (tenant.*.ingest.process)
                                      |
                                 Worker (JetStream consumer)
                                 1. Validate tenant from NATS subject matches payload
                                 2. Update job -> processing
                                 3. Forward to Blueprint POST /v1/documents
                                 4. On success: job -> completed, clean up staged file
                                 5. On failure: retry (max 3), then job -> failed
                                 6. Publish notification + WS broadcast
```

Collection namespacing: files are forwarded to Blueprint with collection name `{tenant_slug}-{collection}` for tenant isolation.

## Database

**Instance:** Tenant DB

**Tables:**
- `ingest_jobs` -- job tracking (status: pending/processing/completed/failed, file info, error)
- `connectors` -- external source configs (future: Google Drive, OneDrive, S3, local)

**Migrations:** `db/migrations/000_deps.up.sql`, `001_init.up.sql`

## NATS Events

**Internal (JetStream):**

| Subject | Stream | Durable | Description |
|---------|--------|---------|-------------|
| `tenant.*.ingest.process` | `INGEST` | `ingest-worker` | Job queue (handler -> worker). Max 3 deliveries, 24h retention |

**Published (outbound):**
- `tenant.{slug}.notify.ingest.completed` -- notification event (consumed by notification service)
- `tenant.{slug}.ingest.jobs` -- WS broadcast for real-time UI progress

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `INGEST_PORT` | No | `8007` | HTTP listen port |
| `POSTGRES_TENANT_URL` | Yes | -- | Tenant DB connection string |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server URL |
| `RAG_SERVER_URL` | No | `http://localhost:8081` | NVIDIA RAG Blueprint URL |
| `INGEST_STAGING_DIR` | No | `/tmp/ingest-staging` | Local staging directory for uploads |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **PostgreSQL:** Tenant DB (job tracking)
- **NATS:** Publisher + JetStream consumer (job queue, notifications, WS broadcast)
- **NVIDIA RAG Blueprint:** Document vectorization endpoint (`/v1/documents`)
- **pkg/jwt:** Ed25519 key loading
- **pkg/middleware:** Auth middleware, RequirePermission, SecureHeaders
- **pkg/nats:** Typed event publishing

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests
```
