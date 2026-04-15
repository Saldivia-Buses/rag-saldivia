---
title: Package: pkg/cache
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./tenant.md
---

## Purpose

Thin Redis JSON cache wrapper with graceful degradation: when constructed with
a nil client, every operation becomes a no-op so callers don't need
`if cache != nil` checks scattered through their code. Import this when a
service wants to cache JSON-serializable values keyed by string with a TTL,
and tolerate Redis being unavailable.

## Public API

Source: `pkg/cache/redis.go:5`

| Symbol | Kind | Description |
|--------|------|-------------|
| `RedisClient` | interface | Minimal Redis surface: `Get`, `Set(key, value, ttl)`, `Del(keys...)` |
| `JSONCache` | struct | Wraps a `RedisClient` |
| `NewJSONCache(client)` | func | Constructor — pass `nil` to disable caching |
| `JSONCache.Available()` | method | True if a client was provided |
| `JSONCache.Get(ctx, key, dest)` | method | JSON-decodes into `dest`, returns false on miss/error |
| `JSONCache.Set(ctx, key, value, ttl)` | method | JSON-encodes and stores |
| `JSONCache.Del(ctx, keys...)` | method | Deletes one or more keys |

## Usage

```go
c := cache.NewJSONCache(redisClient) // or nil for no-op
var hit MyValue
if !c.Get(ctx, "sda:thing:"+id, &hit) {
    hit = expensiveLookup()
    c.Set(ctx, "sda:thing:"+id, hit, 5*time.Minute)
}
```

## Invariants

- All methods are safe to call with a nil-backed cache (graceful degradation).
- `Get` returns `false` on miss, on Redis error, OR on JSON unmarshal failure.
  Callers cannot distinguish — design accordingly.
- The package does not implement any eviction policy beyond Redis TTL.
- `RedisClient` is intentionally minimal so test doubles can satisfy it without
  pulling go-redis.

## Importers

None in production code yet. `pkg/config/resolver.go:13` and other services
use `*redis.Client` directly. Use this wrapper for new code that wants
graceful degradation.
