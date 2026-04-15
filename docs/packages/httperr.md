---
title: Package: pkg/httperr
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./middleware.md
---

## Purpose

Structured error type with constructor helpers and a shared response writer.
Standardises error JSON across every HTTP handler so clients always see
`{"error":"...","code":"..."}`. Import this in every chi handler — it
replaces ad-hoc `http.Error(w, "msg", 400)` calls.

## Public API

Source: `pkg/httperr/httperr.go:3`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Code` | type | Machine-readable error code string |
| `CodeInternal`/`CodeNotFound`/`CodeInvalidInput`/`CodeUnauthorized`/`CodeConflict`/`CodeForbidden` | const | Code values |
| `Error` | struct | `StatusCode`, `Code`, `Message`, `Cause` (implements `error`, `Unwrap`) |
| `Internal(cause)` | func | 500 wrapping a cause |
| `NotFound(resource)` | func | 404 — `"<resource> not found"` |
| `InvalidInput(msg)` | func | 400 with caller's message |
| `InvalidID(id)` | func | 400 — `"invalid id: <id>"` |
| `Conflict(msg)` | func | 409 |
| `Unauthorized(msg)` | func | 401 |
| `Forbidden(msg)` | func | 403 |
| `Wrap(cause, code, msg, status)` | func | Build any combination |
| `WriteError(w, r, err)` | func | Writes JSON, logs based on severity, falls back to 500 |

## Usage

```go
user, err := repo.Get(ctx, id)
if errors.Is(err, sql.ErrNoRows) {
    httperr.WriteError(w, r, httperr.NotFound("user"))
    return
}
if err != nil {
    httperr.WriteError(w, r, httperr.Internal(err))
    return
}
```

## Invariants

- `Cause` is the underlying error and is NEVER serialized to clients
  (`pkg/httperr/httperr.go:30`). Only `Code` and `Message` go in the response
  body.
- `WriteError` logs 5xx as `slog.Error` (with `Cause`) and 4xx as `slog.Warn`
  (without `Cause`) (`pkg/httperr/httperr.go:97`).
- Unknown error types passed to `WriteError` become `internal_error` 500 —
  same code path, plain text never leaks.
- This is mandated by Invariant #7 in `.claude/CLAUDE.md` — every API error
  must be JSON.

## Importers

19+ files across services: `auth`, `agent`, `astro`, `bigbrother`, `chat`,
`feedback`, `healthwatch`, `ingest`, `notification`, `platform`, `search`,
`traces`, `ws`.
