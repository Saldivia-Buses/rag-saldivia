---
title: Object Storage (MinIO/S3)
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/storage.md
  - ../services/ingest.md
  - ../services/extractor.md
  - rag-tree-search.md
  - multi-tenancy.md
---

This document describes how SDA stores and isolates files. Read it before
introducing a new file type, changing the key layout, or wiring a new
storage backend — the key prefix is the only mechanism preventing
cross-tenant file access.

## Backend

A single S3-compatible bucket is shared across tenants. In dev and
production it is hosted by **MinIO**; the same code works against AWS S3
because the client uses `aws-sdk-go-v2` with `UsePathStyle = true` and a
custom `BaseEndpoint` (`pkg/storage/s3.go:47`).

Bucket name is `sda-documents` by default
(`deploy/docker-compose.dev.yml:34`). Configuration is wired by env:

| Env var               | Notes                              |
|-----------------------|------------------------------------|
| `STORAGE_ENDPOINT`    | MinIO/S3 URL (e.g. `http://minio:9000`) |
| `STORAGE_BUCKET`      | Bucket name (default `sda-documents`)   |
| `STORAGE_ACCESS_KEY`  | Required (no default, fail-fast)        |
| `STORAGE_SECRET_KEY`  | Required (no default, fail-fast)        |

The Python extractor enforces required env at startup
(`services/extractor/main.py:88`); Go services do likewise via
`pkg/config`.

## The `Store` interface

`storage.Store` (`pkg/storage/storage.go:22`) is the only interface code
should depend on. Methods: `Put`, `Get`, `Delete`, `Exists`. The default
implementation is `S3Store` (`pkg/storage/s3.go:26`).

`PutOptions.ContentType` defaults to `application/octet-stream`. Errors
unwrap to `ErrNotFound` for missing keys (`s3.go:138`). `EnsureBucket` is
safe to call on every startup — it issues `HeadBucket` first and only
creates on a genuine "not found" so permission errors are never masked.

## Key layout

Keys are slash-separated paths shaped as:

```
{tenant_slug}/{document_id}/{role}.{ext}
```

- `tenant_slug` — the canonical isolation prefix.
- `document_id` — UUID generated when the `documents` row is created.
- `role` — currently `original` (the uploaded file). The pipeline may add
  more roles (`pages.json`, derived artefacts) under the same prefix.

The Ingest service builds the key with
`fmt.Sprintf("%s/%s/original.%s", tenant, doc.ID, fileType)`
(`services/ingest/internal/service/documents.go:89`).

A short-lived `pending-{hash}` prefix is used between row insert and the
final `Put` so the DB never holds an empty `storage_key`
(`documents.go:74`); it is overwritten with the final key on success.

## Tenant isolation rules

- **Always** prefix keys with the tenant slug — no exceptions.
- **Never** use a path with `..` or absolute prefixes — keys are validated
  to start with the slug from context (the ingest service's worker tests
  enforce this, `services/ingest/internal/service/worker_test.go:143`).
- The Extractor receives the storage key inside the NATS payload; it must
  not derive its own — see `documents.go:108`.
- Bucket-level ACLs do **not** enforce tenant isolation; the prefix and
  application-layer auth do.

## Lifecycle

Files written by the ingest pipeline are immutable once `documents.status`
becomes `ready`. Deleting a document via the ingest API removes the row and
calls `Store.Delete` for every key under its prefix. `Exists` is used by
the dedup path before re-uploading a file with a known hash.

## Failure modes

- MinIO unreachable → `Put` returns a wrapped error; the service marks the
  document `error` and returns 500 to the client.
- Hash collision (same `file_hash` already in `documents`) → the existing
  document row is returned and no new upload happens
  (`documents.go:62`).
- Partial upload → no orphan key is left because the row insert and `Put`
  share a transaction-like flow; the temp `pending-` key is replaced with
  the real one only after `Put` succeeds.
