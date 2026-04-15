---
title: Package: pkg/config
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./tenant.md
  - ./crypto.md
---

## Purpose

Two-layer configuration for SDA services. The simple `Env`/`MustEnv` helpers
replace the copy-pasted `env()` function in every `cmd/main.go`. The
`Resolver` reads dynamic, business-level config from the Platform DB with
scope cascade (tenant > plan > global) and Redis caching. Import this for any
configuration access — environment variables, model slot resolution, or
versioned prompts.

## Public API

Sources: `pkg/config/env.go:6`, `pkg/config/resolver.go:1`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Env(key, fallback)` | func | Reads env var or returns fallback |
| `MustEnv(key)` | func | Reads env var or panics — for required secrets/URLs |
| `RedactURL(rawURL)` | func | Strips credentials from a URL for safe logging |
| `Resolver` | struct | Cascade config reader (pgxpool + optional Redis) |
| `NewResolver(pool, cache)` | func | Default 5min TTL; cache may be nil |
| `Resolver.Get(ctx, tenantID, key)` | method | Cascade lookup, returns `json.RawMessage` |
| `Resolver.GetString(ctx, tenantID, key)` | method | Decode as JSON string |
| `Resolver.GetInt(ctx, tenantID, key)` | method | Decode as JSON int |
| `Resolver.ResolveSlot(ctx, tenantID, slot)` | method | Slot key → `*ModelConfig` from `llm_models` |
| `Resolver.GetActivePrompt(ctx, promptKey)` | method | Latest active prompt content from `prompt_versions` |
| `Resolver.InvalidateCache(ctx, tenantID)` | method | Clear cached config for one tenant |
| `Resolver.InvalidateGlobal(ctx)` | method | Clear all `sda:config:*` and `sda:model:*` |
| `ModelConfig` | struct | Endpoint, ModelID, APIKey, Location, cost rates |

## Usage

```go
r := config.NewResolver(platformPool, redisClient)
// Cascade: tenant:{id} → plan:{plan} → global, in one CTE query
mc, err := r.ResolveSlot(ctx, tenantID, "slot.chat")
client := llm.NewClient(mc.Endpoint, mc.ModelID, mc.APIKey)
```

## Invariants

- API keys are NEVER cached in Redis (`pkg/config/resolver.go:48`). Cached
  `cachedModelConfig` strips the key; `ResolveSlot` re-fetches it from
  `llm_models` on every call.
- `Get` returns `fmt.Errorf("config key not found: %q")` on miss — distinguish
  from other errors with string match if needed.
- `Resolver` is thread-safe (pgxpool + go-redis are both concurrent-safe).
- `cascadeQuery` (`pkg/config/resolver.go:61`) resolves all 3 scopes in one
  round trip via UNION + ORDER BY priority.

## Importers

`services/auth`, `chat`, `agent`, `astro`, `feedback`, `ingest`, `notification`,
`platform`, `traces`, `ws`, `bigbrother`, `erp`, `healthwatch`, `search` — all
service `cmd/main.go` files use `Env`/`MustEnv`.
