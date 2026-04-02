# Ingest Service

Async document upload and processing pipeline. Stages documents to disk,
tracks jobs in the tenant database, and processes via NATS JetStream worker
that forwards to the NVIDIA RAG Blueprint for vectorization.

## Architecture

```
Client ─── POST /v1/ingest/upload ───► Handler
                                         │
                                    1. Stage file to disk
                                    2. Create job (status: pending)
                                    3. Publish to NATS
                                    4. Return 202 Accepted
                                         │
                                    NATS JetStream
                                    (tenant.*.ingest.process)
                                         │
                                    Worker (consumer)
                                    1. Update job → processing
                                    2. Forward to Blueprint /v1/documents
                                    3. Update job → completed/failed
                                    4. Clean up staged file
                                    5. Publish notification + WS event
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/ingest/upload` | Multipart document upload (returns 202) |
| GET | `/v1/ingest/jobs` | List jobs for user |
| GET | `/v1/ingest/jobs/{jobID}` | Get job status |
| DELETE | `/v1/ingest/jobs/{jobID}` | Delete job record |
| GET | `/health` | Health check |

## Upload

```bash
curl -X POST http://localhost:8007/v1/ingest/upload \
  -H "X-User-ID: <uuid>" \
  -H "X-Tenant-Slug: saldivia" \
  -F "file=@document.pdf" \
  -F "collection=contratos"
```

Returns 202 with job object. Poll `GET /v1/ingest/jobs/{id}` or subscribe to
WebSocket channel `ingest.jobs` for real-time status.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `INGEST_PORT` | `8007` | HTTP port |
| `POSTGRES_TENANT_URL` | (required) | Tenant database connection |
| `NATS_URL` | `nats://localhost:4222` | NATS server (required) |
| `RAG_SERVER_URL` | `http://localhost:8081` | Blueprint URL |
| `INGEST_STAGING_DIR` | `/tmp/ingest-staging` | Staging directory for uploads |

## NATS

| Subject | Direction | Description |
|---------|-----------|-------------|
| `tenant.{slug}.ingest.process` | Internal | Job queue (handler → worker) |
| `tenant.{slug}.notify.ingest.completed` | Outbound | Notification event |
| `tenant.{slug}.ingest.jobs` | Outbound | WS broadcast for real-time UI |

JetStream stream: `INGEST`, durable consumer: `ingest-worker`, max 3 delivery
attempts, 24h retention.

## Database tables

- `ingest_jobs` — upload job tracking (status, file, collection)
- `connectors` — external source configs (future: Google Drive, OneDrive, S3)
