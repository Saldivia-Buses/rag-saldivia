---
title: Package: pkg/tenant
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ../architecture/multi-tenancy.md
  - ./database.md
  - ./crypto.md
  - ./middleware.md
---

## Purpose

The heart of multi-tenancy. Two responsibilities: (1) propagate tenant
identity through `context.Context` and (2) resolve the correct per-tenant
PostgreSQL pool and Redis client by slug. The `Resolver` is the only
sanctioned way to obtain a tenant's DB connection — see
`architecture/multi-tenancy.md` for the full model. Import this in every
HTTP handler and any service that touches tenant data. **Critical to
Invariant #1 (tenant isolation at every layer).**

## Public API

Sources: `pkg/tenant/context.go`, `pkg/tenant/resolver.go`

### Context helpers

| Symbol | Kind | Description |
|--------|------|-------------|
| `Info` | struct | `ID` (UUID), `Slug` (subdomain) |
| `WithInfo(ctx, info)` | func | Stores in context |
| `FromContext(ctx)` | func | Retrieves; returns `ErrNoTenant` if absent |
| `SlugFromContext(ctx)` | func | Convenience; PANICS if absent (only call after middleware) |
| `ErrNoTenant`/`ErrTenantUnknown` | var | Sentinel errors |

### Resolver

| Symbol | Kind | Description |
|--------|------|-------------|
| `ConnInfo` | struct | `TenantID`, `PostgresURL`, `RedisURL` (potentially decrypted) |
| `Resolver` | struct | Cached-pool registry; mutex + per-slug singleflight |
| `NewResolver(platformDB, encryptionKey)` | func | `encryptionKey` may be nil (no-encryption mode) |
| `Resolver.PostgresPool(ctx, slug)` | method | Returns pool; creates if missing (max 4 conns by default) |
| `Resolver.RedisClient(ctx, slug)` | method | Returns client; creates + pings if missing |
| `Resolver.TenantID(ctx, slug)` | method | UUID lookup (cached) |
| `Resolver.PoolMaxConns` | field | Override before first use (default 4) |
| `EnabledModule` | struct | `ID`, `Name`, `Category` |
| `Resolver.ListEnabledModules(ctx, tenantID)` | method | Tenant's enabled modules + always-on core (chat, auth, notifications) |
| `Resolver.Close()` | method | Closes all cached pools and clients; subsequent calls return errors |
| `Resolver.StartHealthCheck(ctx, interval)` | method | Background goroutine: ping pools, evict unhealthy ones |

## Usage

```go
// Auth middleware sets the context
ctx = tenant.WithInfo(ctx, tenant.Info{ID: claims.TenantID, Slug: claims.Slug})

// Handler retrieves it
slug := tenant.SlugFromContext(r.Context())
pool, err := resolver.PostgresPool(ctx, slug) // per-tenant pool, cached
rows, err := q.WithTx(pool).ListThings(ctx, ...)
```

## Invariants

- The `Resolver` is the ONLY sanctioned source of tenant DB pools. NEVER share
  a pool across tenants — that would break tenant isolation (Invariant #1).
- Pool/client creation is per-slug singleflight: the mutex is RELEASED during
  the network call (`pkg/tenant/resolver.go:184`) so other tenants don't
  block, then a double-check on re-acquisition collapses concurrent winners.
- `ensureSSLMode` (`pkg/tenant/resolver.go:357`) appends `sslmode=require` to
  any Postgres URL that doesn't specify one, enforcing TLS in production
  without breaking dev URLs that use `sslmode=disable`.
- When `encryptionKey != nil` and `postgres_url_enc`/`redis_url_enc` columns
  are populated, the encrypted values are decrypted via `pkg/crypto.Decrypt`
  (`pkg/tenant/resolver.go:142`). The plaintext columns are a fallback for
  pre-migration tenants.
- After `Close()`, every method returns an error — services should call this
  in their shutdown path.
- `StartHealthCheck` evicts unhealthy pools so the next request creates a fresh
  one — auto-recovery on transient DB failures (`pkg/tenant/resolver.go:329`).
- `coreModules()` (chat, auth, notifications) is the failover when the modules
  query errors out — services keep working even if `tenant_modules` is
  unreachable.

## Importers

`services/auth`, `astro`, `healthwatch`, `platform`, `search`, `traces`,
`agent`, plus tests across services. Pervasive — every handler that returns
tenant data.
