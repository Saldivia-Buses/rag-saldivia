---
title: Package: pkg/storage
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ../architecture/storage-minio.md
---

## Purpose

Pluggable file storage interface plus an S3-compatible default implementation
(MinIO locally, AWS S3 in production). All services that store or retrieve
files use this — the interface lets tests substitute a fake without pulling
the AWS SDK. Import this for any document, attachment, or generated artifact
that needs persistent blob storage.

## Public API

Sources: `pkg/storage/storage.go`, `pkg/storage/s3.go`

| Symbol | Kind | Description |
|--------|------|-------------|
| `ErrNotFound` | var | Returned by `Get` when key is missing |
| `PutOptions` | struct | `ContentType` (defaults to `application/octet-stream`) |
| `Store` | interface | `Put`, `Get`, `Delete`, `Exists` |
| `S3Config` | struct | `Endpoint`, `Bucket`, `AccessKey`, `SecretKey`, `Region` (defaults `us-east-1`) |
| `S3Store` | struct | Implements `Store` against S3-compatible APIs |
| `NewS3Store(ctx, cfg)` | func | Constructor (does NOT create the bucket) |
| `S3Store.EnsureBucket(ctx)` | method | Idempotent bucket-create — safe to call on every startup |
| `S3Store.Put / Get / Delete / Exists` | method | `Store` interface impl |

## Usage

```go
s, _ := storage.NewS3Store(ctx, storage.S3Config{
    Endpoint: "http://minio:9000", Bucket: "sda-documents",
    AccessKey: ak, SecretKey: sk,
})
_ = s.EnsureBucket(ctx)

key := tenantSlug + "/" + docID + "/original.pdf"
err := s.Put(ctx, key, file, &storage.PutOptions{ContentType: "application/pdf"})

rc, err := s.Get(ctx, key)
defer rc.Close()
```

## Invariants

- Keys are slash-separated paths. Convention: `{tenant}/{doc_id}/{filename}`
  so all of a tenant's blobs share a prefix and can be enumerated/deleted
  together (`pkg/storage/storage.go:21`).
- `Get` returns `ErrNotFound` (wrapped) when the object is missing — callers
  use `errors.Is(err, storage.ErrNotFound)`.
- Caller MUST close the `io.ReadCloser` returned by `Get`.
- `Delete` returns nil if the key didn't exist (idempotent semantically; AWS
  itself returns success for missing keys).
- `EnsureBucket` only creates on a real `NotFound`. Permission errors and
  network failures propagate so misconfiguration isn't masked
  (`pkg/storage/s3.go:67`).
- MinIO requires `UsePathStyle: true`, set automatically when an `Endpoint` is
  provided (`pkg/storage/s3.go:50`).

## Importers

`services/ingest/internal/service/documents.go`,
`services/ingest/internal/service/documents_test.go` — currently the only
production user.
