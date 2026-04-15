---
title: Flow: Document Ingestion
audience: ai
last_reviewed: 2026-04-15
related:
  - ../services/ingest.md
  - ../services/extractor.md
---

## Purpose

How an uploaded file becomes a searchable document with extracted pages
and a tree summary. Read this before changing the upload handler, the
NATS message shape, the bulk page insert, the extractor consumer, or the
tree generator. Architecture (RAG storage, tree shape, MinIO layout) is
in `services/ingest.md` — this file owns the end-to-end sequence.

## Steps (modern Extractor path)

1. User posts `multipart/form-data` to `POST /v1/ingest/upload`; handler
   `services/ingest/internal/handler/ingest.go:71` enforces a 100MB cap,
   reads `X-User-ID`/`X-Tenant-Slug`, and validates the file extension
   against `allowedExts`.
2. Handler calls into `service.DocumentService.UploadDocument`
   (`services/ingest/internal/service/documents.go:50`); a `TeeReader`
   streams the file through SHA-256 while buffering bytes for upload.
3. `GetDocumentByHash` is consulted first — duplicates short-circuit and
   return the existing document row, so re-uploads are idempotent.
4. A new row is inserted into `documents` with a temporary storage key,
   the bytes are streamed to MinIO at `{tenant}/{doc_id}/original.{ext}`
   via `pkg/storage`, then the storage key is updated to the real path.
5. The service publishes `ExtractionJobMessage{document_id, tenant_slug,
   storage_key, file_name, file_type}` on `tenant.{slug}.extractor.job`
   and updates the document status to `extracting`
   (`documents.go:114`); publish failure marks the document `error`.
6. `services/extractor/main.py` consumes the JetStream subject, downloads
   from MinIO, runs OCR/vision (PaddleOCR-VL via SGLang), and emits the
   result on `tenant.{slug}.extractor.result.>` as `ExtractionResult`.
7. `service.ExtractorConsumer.handleResult`
   (`services/ingest/internal/service/extractor_consumer.go:100`) parses
   the message, sets document status to `indexing`, and updates
   `total_pages` on the row.
8. `BulkInsertPages` (`services/ingest/internal/service/bulk_pages.go:23`)
   runs inside a transaction: deletes existing pages for the doc id and
   inserts every page via `tx.CopyFrom` for a single round-trip; failure
   marks the document `error` via `setDocError`.
9. `tree.Generator.Generate` builds the PageIndex-style tree from the
   page text, and `InsertDocumentTree` stores the JSON tree, doc
   description, model used, and node count
   (`extractor_consumer.go:171`); failure flags the document but keeps
   the pages.
10. The consumer marks the document `ready` (or `error` on partial
    failure) and acks the message; the WS hub picks up downstream NATS
    notifications so the UI flips the job from `processing` to
    `completed`.

## Legacy Blueprint path (deprecated)

`Ingest.Submit` (`services/ingest/internal/service/ingest.go:116`) stages
the file to `/tmp/ingest-staging/{tenant}/`, creates an `ingest_jobs` row,
and publishes on `tenant.{slug}.ingest.process` for the JetStream worker
(`worker.go:121`) which forwards the bytes to the NVIDIA RAG Blueprint at
`/v1/documents`. Retains 3 deliveries before marking the job `failed`. New
work MUST use the modern Extractor path; the Blueprint route is kept only
for `/v1/ingest/jobs` listings until rip-out.

## Invariants

- File hashes are computed once in a streaming pass and stored in
  `documents.file_hash`; dedup MUST consult `GetDocumentByHash` before any
  MinIO write.
- NATS subjects are tenant-namespaced (`tenant.{slug}.extractor.*`); the
  tenant slug is validated against the regex `^[a-zA-Z0-9_-]+$` in the
  `DocumentService` constructor.
- `BulkInsertPages` runs inside a transaction and replaces all pages for
  the document; partial inserts MUST be rolled back via `tx.Rollback`.
- Document state machine: `pending → extracting → indexing → ready` (or
  `error` at any step); the consumer is the only writer of `ready`.
- Tree generation failure does NOT discard pages — the document stays in
  `error` so the operator can re-run the tree without re-extracting.
- Extractor results may redeliver up to 3 times (`MaxDeliver=3`);
  consumers MUST be idempotent because `BulkInsertPages` deletes existing
  rows first.

## Failure modes

- `400 unsupported file type` — extension not in `allowedExts`; add the
  type or convert before upload.
- `400 invalid multipart form` / `413` — body exceeded `MaxUploadSize`;
  raise the cap or split the upload.
- Document stuck `extracting` — Extractor service is offline or its
  consumer is unsubscribed; check `tenant.*.extractor.job` JetStream
  metrics.
- `bulk insert pages failed` — schema mismatch between
  `repository.CreateMessageParams` and `document_pages`; rerun `make
  sqlc` and inspect the consumer logs.
- `tree generation failed` — LLM endpoint down or content too large;
  document remains in `error`, retry once SGLang recovers.
- `failed to publish extraction job` — NATS down; the document is set to
  `error` immediately so the user sees the failure
  (`documents.go:117`).
- Legacy job stuck `processing` — JetStream worker `ingest-worker` is
  redelivering; check `worker.go:121` logs and the `INGEST` stream
  consumer state.
