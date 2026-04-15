---
title: Service: ingest
audience: ai
last_reviewed: 2026-04-15
related:
  - ../flows/document-ingestion.md
  - ./extractor.md
  - ../architecture/rag-tree-search.md
  - ../architecture/storage-minio.md
---

## Purpose

Document ingestion pipeline: accepts uploads, persists job state, hands work
to the Python extractor over NATS, consumes results, and produces the tree
representations used by the search service. Read this when changing upload
limits, the extractor contract, collection lifecycle, or the result-handler
that writes pages and tree nodes back to the tenant DB.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/nats/redis check |
| POST | `/v1/ingest/upload` | JWT + `ingest.write` | Multipart upload — creates a document + job, kicks off processing |
| GET | `/v1/ingest/jobs` | JWT + `ingest.write` | List ingestion jobs (paginated) |
| GET | `/v1/ingest/jobs/{jobID}` | JWT + `ingest.write` | Get job detail (status, error, page count) |
| DELETE | `/v1/ingest/jobs/{jobID}` | JWT + `ingest.write` | Cancel/delete a job |
| GET | `/v1/ingest/collections` | JWT + `collections.read` | List document collections |
| POST | `/v1/ingest/collections` | JWT + `collections.write` | Create a new collection |

Routes registered in `services/ingest/internal/handler/ingest.go:52`. All
routes use `FailOpen=false` to avoid creating ghost uploads during a Redis
outage.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.{slug}.ingest.process` | pub | New job ready for the worker (`services/ingest/internal/service/ingest.go:154`) |
| `tenant.{slug}.extractor.job` | pub | Send PDF to extractor (`services/ingest/internal/service/documents.go:114`) |
| `tenant.*.extractor.result.>` | sub | Extractor completed — durable consumer in `services/ingest/internal/service/extractor_consumer.go:67` |

The result consumer runs from the `EXTRACTOR_RESULTS` stream (created on
startup) and parses tenant slug from the subject, never the payload.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `INGEST_PORT` | no | `8007` | HTTP listener port |
| `POSTGRES_TENANT_URL` | yes | — | Documents, pages, jobs, collections, trees |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Job + result transport |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |
| `RAG_SERVER_URL` | no | `http://localhost:8081` | Legacy NVIDIA Blueprint server (used for tree generation hooks) |
| `INGEST_STAGING_DIR` | no | `/tmp/ingest-staging` | Local staging dir for uploads before MinIO push |

## Dependencies

- **PostgreSQL tenant** — `documents`, `document_pages`, `document_trees`,
  `ingest_jobs`, `collections`.
- **NATS JetStream** — outbound jobs, inbound results.
- **Redis** — token blacklist.
- **Extractor service** — over NATS only (no HTTP coupling).
- **Local staging dir** — used by the worker before persistence.
- **Optional Blueprint server** at `RAG_SERVER_URL` — referenced by service
  config but not the primary path now that tree search is in-house.

## Permissions used

- `ingest.write` — upload, list/get/delete jobs.
- `collections.read` — list collections.
- `collections.write` — create collection.

Note: there is no `ingest.read` — `GET /v1/ingest/jobs` is gated by
`ingest.write` (`services/ingest/internal/handler/ingest.go:55`). Treat job
listings as a write-side debug surface.
