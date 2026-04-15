---
title: Package: pkg/featureflags
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./jwt.md
---

## Purpose

HTTP client that asks the Platform service to evaluate feature flags for a
given JWT and caches the result for 30 seconds. Each cache key is the
SHA-256 of the JWT (raw tokens are never stored in process memory). Import
this when a service needs to gate a feature behind a flag without hitting the
Platform service on every request.

## Public API

Source: `pkg/featureflags/client.go:3`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Client` | struct | HTTP client + per-JWT cache |
| `New(platformURL)` | func | Builds a Client; default 30s TTL, 5s HTTP timeout |
| `Client.IsEnabled(ctx, flag, jwt)` | method | Returns true iff the flag is enabled for the JWT's caller |

Internal: `cacheEntry` (5min TTL evict), `evaluateResponse` (`{flags: {...}}`),
`maxCacheEntries = 10000`.

## Usage

```go
ff := featureflags.New("http://platform:8006")
if ff.IsEnabled(ctx, "erp.workshop.beta", jwtFromAuthHeader) {
    // serve beta route
}
```

## Invariants

- The JWT is forwarded as `Authorization: Bearer <jwt>` — the Platform service
  evaluates flags using its identity (tenant, role).
- Cache key is the first 16 bytes of `sha256(jwt)` (`pkg/featureflags/client.go:62`)
  — never the raw token. This protects against memory dumps.
- On any error (network, non-200, decode), `IsEnabled` returns false — flags
  fail closed.
- When the cache hits `maxCacheEntries` (10000) the entire map is wiped — a
  simple bound, not LRU. Spike protection only.

## Importers

None in production code yet. New flag-gated features should adopt this client.
