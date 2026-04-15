---
title: Package: pkg/middleware
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./jwt.md
  - ./tenant.md
  - ./security.md
  - ../architecture/auth-jwt.md
---

## Purpose

The shared chi HTTP middleware kit. Provides JWT verification with optional
blacklist, identity context helpers, RBAC permission and module gates,
per-key rate limiting, secure-headers, and an enriched logger that auto-tags
every log line with `tenant_id`, `user_id`, `request_id`, and `trace_id`.
Import this in every service that exposes HTTP endpoints.

## Public API

| Symbol | Source | Purpose |
|--------|--------|---------|
| `AuthConfig` | `auth.go:17` | `Blacklist`, `FailOpen` |
| `Auth(publicKey)` | `auth.go:33` | Verify JWT, inject identity (no blacklist) |
| `AuthWithConfig(publicKey, cfg)` | `auth.go:38` | Same with blacklist + fail-open toggle |
| `EnrichLogger` | `logging.go:29` | Adds tenant/user/request/trace IDs to slog context |
| `LoggerFromCtx(ctx)` | `logging.go:58` | Retrieve enriched logger |
| `RateLimitConfig` | `ratelimit.go:13` | `Requests`, `Window`, `KeyFunc` |
| `ByIP(r)` / `ByUser(r)` | `ratelimit.go:24` | Standard key extractors |
| `RateLimit(cfg)` | `ratelimit.go:47` | Token-bucket per-key rate limiter |
| `SecureHeaders()` | `security_headers.go:9` | X-Content-Type-Options, X-Frame-Options, HSTS, etc. |
| `WithRole`/`RoleFromContext` | `rbac.go:29` | Role context helpers |
| `WithUserID`/`UserIDFromContext` | `rbac.go:40` | User ID context helpers |
| `WithUserEmail`/`UserEmailFromContext` | `rbac.go:51` | Email context helpers |
| `WithPermissions`/`PermissionsFromContext` | `rbac.go:18` | Perm slice context helpers |
| `WithEnabledModules`/`EnabledModulesFromContext` | `rbac.go:137` | Per-tenant enabled modules |
| `RequirePermission(perm)` | `rbac.go:75` | 403 unless caller has the permission (admin bypasses; supports `prefix.*` and bare `*`) |
| `RequireModule(moduleID)` | `rbac.go:102` | 403 unless module is enabled for the tenant |

## Usage

```go
r := chi.NewRouter()
r.Use(middleware.SecureHeaders())
r.Use(middleware.Auth(publicKey))
r.Use(middleware.EnrichLogger)
r.Use(middleware.RateLimit(middleware.RateLimitConfig{
    Requests: 60, Window: time.Minute, KeyFunc: middleware.ByUser,
}))

r.With(middleware.RequirePermission("erp.accounting.write")).
    Post("/v1/entries", h.CreateEntry)
```

## Invariants

- The Auth middleware STRIPS client-spoofed `X-User-*`/`X-Tenant-*` headers
  before processing (`pkg/middleware/auth.go:44`). Never trust those headers
  upstream of this middleware.
- `/health` is the ONLY path skipped by Auth (`pkg/middleware/auth.go:53`).
- When the JWT slug doesn't match the Traefik-injected `X-Tenant-Slug` from
  the subdomain, the request is rejected with 403 (`pkg/middleware/auth.go:115`).
- Permission wildcards: `prefix.*` matches `prefix.<anything>`; bare `*`
  matches any permission (`pkg/middleware/rbac.go:119`).
- Role `admin` bypasses all `RequirePermission` checks (`pkg/middleware/rbac.go:81`).
- `RateLimit` uses an in-memory token bucket. Multi-node deployments need a
  Redis-backed limiter — TODO when scaling out.
- Stale rate-limit entries are purged every 10 minutes
  (`pkg/middleware/ratelimit.go:105`).

## Importers

46+ files. Used in every HTTP handler and `cmd/main.go` across `auth`, `agent`,
`astro`, `bigbrother`, `chat`, `erp`, `feedback`, `healthwatch`, `ingest`,
`notification`, `platform`, `search`, `traces`.
