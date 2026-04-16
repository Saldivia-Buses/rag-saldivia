---
title: Multi-tenancy Model
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/tenant.md
  - ../flows/tenant-routing.md
  - auth-jwt.md
  - database-postgres.md
---

This document defines how SDA isolates tenants. Read it before touching any
code path that selects a database, builds a NATS subject, or uses the
`X-Tenant-Slug` header — these are the invariants that prevent cross-tenant
data leakage.

## Identity of a tenant

A tenant has two identifiers, both stored in the platform DB `tenants` table
(`db/platform/migrations/001_init.up.sql:19`):

- `slug` — the URL-safe subdomain (`saldivia`, `acme`). Always lowercase,
  unique. Appears in DNS, JWT claims, NATS subjects, and storage keys.
- `id` — a UUID. Used for foreign keys (e.g. `tenant_modules.tenant_id`).

The slug is the canonical key in the application layer; the UUID exists only
for relational integrity inside the platform DB.

## Per-tenant infrastructure

Each tenant has its **own** PostgreSQL database and Redis instance. The
platform DB stores their connection URLs (encrypted at rest when
`postgres_url_enc` / `redis_url_enc` are populated, see
`db/platform/migrations/003_encrypted_credentials.up.sql`). NATS, MinIO, and
SGLang are shared across tenants; isolation is enforced via subject
prefixes, key prefixes, and per-request authorization rather than separate
infrastructure.

## Slug derivation

In production, Traefik extracts the slug from the request subdomain and
injects `X-Tenant-Slug`. The Cloudflare Tunnel routes `*.${SDA_DOMAIN}` to
Traefik (`deploy/cloudflare/config.yml.tmpl:25`). In dev, the
`dev-tenant` middleware in `deploy/traefik/dynamic/dev.yml` injects a fixed
slug. Either way, the auth middleware then **cross-validates** that the slug
in the JWT matches the header — mismatches return `403 tenant mismatch`
(`pkg/middleware/auth.go:115`).

## Resolver: slug → pools

`tenant.Resolver` (`pkg/tenant/resolver.go:32`) maps a slug to a
`*pgxpool.Pool` and `*redis.Client`, lazily creating and caching them on
first use. Key behaviours:

- Singleflight-style locking — only one pool is created per slug under
  concurrent load (`resolver.go:69`).
- 5-minute TTL cache for the platform DB lookup (`resolver.go:51`).
- `PoolMaxConns` defaults to 4 per tenant; tune before first use.
- `StartHealthCheck` periodically pings every cached pool and evicts
  unhealthy ones so the next request reconnects (`resolver.go:329`).
- `Close()` shuts down every cached connection on graceful shutdown.
- `ensureSSLMode` forces `sslmode=require` when the URL omits it
  (`resolver.go:357`).

Services that need tenant-scoped storage call
`Resolver.PostgresPool(ctx, slug)` per request; never share or reuse a pool
across tenants.

## Context propagation

`tenant.Info{ID, Slug}` is stored in `context.Context` via
`tenant.WithInfo` (`pkg/tenant/context.go:28`) and read via
`tenant.FromContext`. The `Auth` middleware injects this on every protected
request after JWT verification. Downstream code reads only the slug from
context — never from headers.

## Row-Level Security (RLS)

ERP tables in the tenant DB enable PostgreSQL RLS with a policy that filters
by `current_setting('app.tenant_id', true)` (e.g.
`db/tenant/migrations/017_erp_entities.up.sql:82`). Code activates this via
`database.SetTenantID(ctx, tx, tenantID)` (`pkg/database/pool.go:54`), which
runs `SET LOCAL app.tenant_id = $1` inside the transaction. RLS is the second
line of defence; the primary defence is using the right per-tenant pool. Non-
ERP tables rely on per-tenant pool isolation alone.

## Modules per tenant

`tenant_modules` (platform DB) records which modules each tenant has
enabled. `Resolver.ListEnabledModules` joins this table with `modules` and
unions the always-on `core` modules (`resolver.go:264`). The agent uses the
same list to decide which tool manifests to load.

## What you must never do

- Read tenant identity from a client-supplied header without the auth
  middleware running first — it strips and re-injects every identity header.
- Cache a pool keyed by anything other than slug.
- Build a NATS subject without the `tenant.{slug}.` prefix (see
  nats-events.md).
- Hardcode a tenant slug or UUID anywhere in service code.
