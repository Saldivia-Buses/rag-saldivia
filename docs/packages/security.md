---
title: Package: pkg/security
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./jwt.md
  - ./middleware.md
---

## Purpose

JWT revocation primitives. `TokenBlacklist` stores revoked JWT IDs (JTI) in
Redis with a TTL matching each token's remaining lifetime. The auth middleware
checks `IsRevoked()` on every request, so logout, password change, and admin
revocation effects are immediate. Import this in `cmd/main.go` to wire the
blacklist for the auth middleware.

## Public API

Sources: `pkg/security/blacklist.go`, `pkg/security/init.go`

| Symbol | Kind | Description |
|--------|------|-------------|
| `TokenBlacklist` | struct | Redis-backed JTI revocation store, prefix `sda:token:blacklist:` |
| `NewTokenBlacklist(rdb)` | func | Constructor |
| `TokenBlacklist.Revoke(ctx, jti, expiresAt)` | method | Adds JTI with TTL = `time.Until(expiresAt)` |
| `TokenBlacklist.RevokeAll(ctx, jtis, expiresAt)` | method | Pipeline-revoke many tokens (e.g., on password change) |
| `TokenBlacklist.IsRevoked(ctx, jti)` | method | Existence check |
| `TokenBlacklist.Ping(ctx)` | method | Used by health checks |
| `InitBlacklist(ctx, redisURL)` | func | Connects to Redis and returns blacklist; `nil` if Redis unavailable |

## Usage

```go
// In every cmd/main.go after Redis is up
bl := security.InitBlacklist(ctx, redisURL) // may return nil
r.Use(middleware.AuthWithConfig(pubKey, middleware.AuthConfig{
    Blacklist: bl, FailOpen: false,
}))

// On logout
err := bl.Revoke(ctx, claims.ID, claims.ExpiresAt.Time)
```

## Invariants

- TTL is computed from `time.Until(expiresAt)` — if the token is already
  expired, `Revoke` is a no-op (`pkg/security/blacklist.go:36`). The blacklist
  size stays bounded.
- `InitBlacklist` returns `nil` when `redisURL == ""` or when Redis is
  unreachable; the auth middleware then runs without revocation checking
  (degraded but functional).
- Setting `AuthConfig.FailOpen=false` (the default) means any Redis error
  during `IsRevoked` returns 503 — we refuse to authenticate without
  revocation visibility.
- Keys use the prefix `sda:token:blacklist:` — never reuse this prefix in
  Redis for anything else.
- `RevokeAll` uses a Redis pipeline (`pkg/security/blacklist.go:57`), so even
  password-change revocation of many sessions is one round trip.

## Importers

All services that authenticate HTTP requests: `auth`, `astro`, `agent`,
`bigbrother`, `chat`, `erp`, `feedback`, `healthwatch`, `ingest`,
`notification`, `platform`, `search`, `traces`, `ws`.
