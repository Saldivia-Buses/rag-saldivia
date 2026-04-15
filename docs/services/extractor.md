---
title: Service: extractor
audience: ai
last_reviewed: 2026-04-15
related:
  - ../flows/document-ingestion.md
  - ../architecture/llm-sglang.md
  - ../architecture/storage-minio.md
  - ./ingest.md
---

## Purpose

Document extraction pipeline. **The only Python service in the stack** —
conventions differ from the Go services. Subscribes to a NATS JetStream queue
of jobs from `ingest`, downloads PDFs from MinIO, runs OCR through
PaddleOCR-VL (SGLang), runs image analysis through Qwen3.5-9B (SGLang), and
publishes structured `ExtractionResult` events back. Read this when changing
the extraction pipeline, model selection, MinIO access, or NATS contract with
ingest.

## Endpoints

The extractor is **NATS-driven**. Its only HTTP surface is a health endpoint
served by the stdlib `http.server` (`services/extractor/main.py:55`):

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness probe (returns `{"status":"ok"}`) |

No JWT, no chi, no per-tenant routing — the worker trusts the NATS subject
for tenant attribution.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.*.extractor.job` | sub | Ingest enqueues a new extraction job (durable consumer `extractor-worker`, `services/extractor/main.py:148`) |
| `tenant.{slug}.extractor.result.{document_id}` | pub | Extraction completed — payload is `ExtractionResult` JSON |

The JetStream stream `EXTRACTOR` is created on startup if missing
(`services/extractor/main.py:108`). Tenant slug and document ID come from
the validated `ExtractionJob` payload (allowlisted with `[a-zA-Z0-9_-]+`)
and are echoed into the result subject — never trust client-side fields for
routing decisions.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `NATS_URL` | no | `nats://localhost:4222` | NATS connection |
| `SGLANG_OCR_URL` | no | `http://localhost:8100` | PaddleOCR-VL endpoint |
| `SGLANG_VISION_URL` | no | `http://localhost:8101` | Qwen3.5-9B vision endpoint |
| `STORAGE_ENDPOINT` | yes | — | MinIO/S3 endpoint (fail fast if missing) |
| `STORAGE_BUCKET` | yes | — | Bucket holding source PDFs |
| `STORAGE_ACCESS_KEY` | yes | — | MinIO/S3 access key |
| `STORAGE_SECRET_KEY` | yes | — | MinIO/S3 secret key |
| `HEALTH_PORT` | no | `8090` | Health endpoint port |

`_require_env` (`services/extractor/main.py:43`) hard-fails on missing
storage credentials so the worker never silently runs without persistence.

## Dependencies

- **NATS JetStream** — single durable consumer with `max_deliver=3` and
  `ack_wait=300` for large PDFs.
- **MinIO / S3** — `extractor.storage.StorageClient` downloads source PDFs.
- **SGLang OCR** — `extractor.ocr.OCRClient`.
- **SGLang Vision** — `extractor.vision.VisionClient`.
- **No PostgreSQL.** The worker is stateless.
- **No outbound HTTP to other Go services** — results travel by NATS.

## Permissions used

None — the extractor is internal-only and does not authenticate callers.
Authorization is upstream: only `ingest` is authorized to publish to
`tenant.*.extractor.job`, gated by NATS-server permissions.
