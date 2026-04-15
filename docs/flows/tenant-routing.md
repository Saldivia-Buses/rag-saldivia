---
title: Flow: Tenant Routing
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/multi-tenancy.md
  - ../packages/tenant.md
---

## Purpose

How a request is bound to a single tenant — from subdomain extraction at
the gateway to a per-tenant `pgxpool.Pool` with RLS-scoped queries. Read
this before touching `pkg/tenant`, the auth middleware, the resolver, or
any handler that calls `Resolver.PostgresPool`. The model (subdomains, RLS
schema, isolation guarantees) lives in `architecture/multi-tenancy.md`.

## Steps

1. Cloudflare/Traefik receives `https://{slug}.sda.app/...`, derives the
   tenant slug from the host, and injects `X-Tenant-Slug: {slug}` before
   forwarding to the backend service.
2. `pkg/middleware/auth.go:42` saves the trusted Traefik header value into
   `traefikSlug`, then deletes every client-controllable identity header
   (`X-User-*`, `X-Tenant-*`) so spoofed values cannot survive.
3. The middleware verifies the bearer JWT with the ed25519 public key and
   checks the blacklist; the request fails with 401 before any tenant code
   runs if either check fails.
4. After verification it cross-validates `claims.Slug == traefikSlug`
   (`pkg/middleware/auth.go:115`) and returns `403 tenant mismatch` if they
   diverge — this blocks token replay across subdomains.
5. The middleware re-injects the canonical identity headers from the JWT
   and stores `tenant.Info{ID, Slug}` in the request context via
   `tenant.WithInfo`, the single source consumed by handlers.
6. Handlers call `tenant.FromContext(ctx)` to read the slug, then
   `Resolver.PostgresPool(ctx, slug)` at `pkg/tenant/resolver.go:68`; the
   first call per slug takes the resolver mutex and goes to step 7.
7. `resolveConnInfo` (`pkg/tenant/resolver.go:102`) queries the platform DB
   `tenants` table for the tenant id + Postgres/Redis URLs, decrypts the
   `_enc` columns when an encryption key is present, and caches the result
   for `defaultCacheTTL`.
8. `createPoolLocked` (`pkg/tenant/resolver.go:169`) parses the URL,
   forces SSL, releases the mutex during pool creation, then stores the
   `*pgxpool.Pool` keyed by slug so subsequent calls hit the in-memory map.
9. Repository code begins a transaction and calls
   `database.SetTenantID(ctx, tx, slug)` (`pkg/database/pool.go:54`) to set
   `app.tenant_id`; PostgreSQL RLS policies then filter every row by
   `current_setting('app.tenant_id')`.
10. Cross-service calls forward the raw bearer token (Authorization header
    or gRPC metadata) so the next service repeats steps 2–9 with the same
    tenant context — the tenant is always derived from the JWT, never from
    a request body.

## Invariants

- The auth middleware is the only writer of `X-User-*`/`X-Tenant-*`
  headers — handlers never set them and always trust them once the
  middleware has run.
- `claims.Slug` MUST equal the Traefik-injected slug; mismatch returns 403
  and is logged for security review.
- One pool per tenant slug: `Resolver.PostgresPool` MUST be the only path
  that produces tenant DB pools; never call `pgxpool.New` directly inside
  a handler.
- Every tenant-scoped query runs inside a transaction with
  `database.SetTenantID` set, or hits a sqlc query that includes
  `tenant_id` explicitly in the WHERE clause.
- Platform-admin code paths use the platform pool only; tenant code must
  never reach into another tenant's pool, even temporarily.

## Failure modes

- `403 tenant mismatch` — JWT slug ≠ subdomain slug; check
  `pkg/middleware/auth.go:115` and confirm the client is on the right host.
- `502 tenant not available` — `Resolver.PostgresPool` returned an error;
  inspect `resolveConnInfo` logs and the platform DB `tenants` row for the
  slug (`enabled=true` required).
- `ErrTenantUnknown` — slug missing from `platform.tenants`; the resolver
  cache will retry after `defaultCacheTTL`.
- Cross-tenant data leak — usually caused by a sqlc query missing
  `tenant_id` or a handler reusing a pool variable; run
  `bash .claude/hooks/check-invariants.sh` to flag it.
- Pool exhaustion — `PoolMaxConns` defaults to 4; raise it on the resolver
  before first use if a tenant saturates connections.
- Decrypt failure on `postgres_url_enc` — wrong KEK; resolver returns the
  decrypt error, see `resolveConnInfo` line 142.
