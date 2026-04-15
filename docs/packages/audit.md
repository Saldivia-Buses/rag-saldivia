---
title: Package: pkg/audit
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./tenant.md
  - ./middleware.md
---

## Purpose

Shared writer for the per-tenant `audit_log` table (created by `auth` migration
001). Every service that has a tenant DB pool can record audit entries for
security-sensitive operations: logins, MFA, PLC writes, remote command
execution, credential access. Import this when an action needs an immutable
record of who did what, when, and from where.

## Public API

Source: `pkg/audit/audit.go:10`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Logger` | interface | `Write(ctx, Entry)` — non-failing (errors logged, never returned) |
| `StrictLogger` | interface | `WriteStrict(ctx, Entry) error` — fail-closed; caller MUST abort on error |
| `Writer` | struct | Implements both interfaces (compile-time checked at `pkg/audit/audit.go:39`) |
| `NewWriter(*pgxpool.Pool)` | func | Builds a Writer for the given tenant pool |
| `Entry` | struct | TenantID, UserID, Action, Resource, Details, IP, UserAgent |

## Usage

```go
w := audit.NewWriter(tenantPool)
w.Write(ctx, audit.Entry{
    TenantID: tid, UserID: uid, Action: "user.login",
    Resource: email, IP: clientIP, UserAgent: ua,
})

// Critical operation — must succeed before proceeding
if err := w.WriteStrict(ctx, plcEntry); err != nil {
    return err // abort PLC write
}
```

## Invariants

- `Write` (Logger) must NEVER block or fail business logic — wrap audit failures
  in slog and continue.
- `WriteStrict` (StrictLogger) is fail-closed — callers MUST treat a non-nil
  error as a hard stop. Never log-and-continue.
- `Entry.Details` is marshaled as JSONB. Empty / unmarshalable maps store `{}`.
- The `audit_log` table schema is owned by `services/auth/migrations/001`.
  Other services depend on its existence in every tenant DB.
- Empty string fields are stored as SQL NULL via `nilIfEmpty` (`pkg/audit/audit.go:98`).

## Importers

`services/auth`, `chat`, `bigbrother`, `notification`, `platform`, `astro`,
`ingest`, `search`, `erp` (every internal/service file) — pervasive.
