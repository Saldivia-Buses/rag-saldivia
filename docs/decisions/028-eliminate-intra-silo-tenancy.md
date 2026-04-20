# ADR 028 — Eliminate intra-silo tenancy

**Status:** Accepted (2026-04-20)
**Supersedes:** N/A
**Related:** ADR 022 (one container per tenant), ADR 026 (SDA replaces Histrix)

## Context

ADR 022 declared the **silo deployment model**: one container per
tenant, the tenant **is** the container. Infrastructure-level isolation
was the chosen mechanism for separating tenants.

In parallel, the application code grew **a second tenancy layer** over
the years:

- Every ERP table carries a `tenant_id` text column.
- Every sqlc query filters with `WHERE tenant_id = $1`.
- The JWT carries `tid` (UUID id) and `slug` claims.
- `pkg/middleware/auth.go` cross-validates `claims.Slug` against the
  Traefik-injected `X-Tenant-Slug` header (anti-token-replay check).
- `platform.tenants` registry holds id/slug/postgres_url per tenant.
- Frontend reads `NEXT_PUBLIC_TENANT_SLUG` (build-time arg) and
  `getTenantSlug()` from the hostname.

The 2026-04-20 prod cutover (tenant `dev` → `saldivia`) surfaced 7
bugs in cascade, all caused by the **second** tenancy layer:

1. `POSTGRES_TENANT_URL` hardcoded in compose.dev.yml.
2. Traefik dev.yml hardcoded `X-Tenant-Slug: "dev"`.
3. Cookie `sda_refresh` with `Secure` flag broken on HTTP.
4. `NEXT_PUBLIC_TENANT_SLUG` is build-arg → cutover requires rebuild.
5. Migrator wrote `tenant_id="saldivia_bench"`; backend filters by
   slug `saldivia` → all rows invisible.
6. `sidebar6.tsx` hardcoded — `MODULE_REGISTRY` is not the source of
   truth.
7. `platform.tenants.id="saldivia-tenant-id"`, JWT `tid=saldivia`,
   data `tenant_id=saldivia_bench` → **four distinct values for
   the same tenant**.

The root cause is not any of the seven bugs individually. It is that
**the silo already separates tenants**; the in-code tenancy is
redundant overhead that introduces N opportunities for ID drift, with
N proportional to the number of code paths.

## Decision

**Eliminate intra-silo tenancy from the application code.** The silo
container is the single tenant boundary. No code reads, writes, or
checks tenant identity below the container boundary.

Concretely:

- **SQL schema**: drop `tenant_id` columns from all tables. Indexes
  prefixed with `tenant_id` get rebuilt without it.
- **sqlc queries**: drop the `WHERE tenant_id = $1` clause and the
  parameter from every query. Generated code regenerates without
  the field.
- **JWT claims**: keep `uid`, `email`, `name`, `role`, `perms`,
  `iat`, `exp`. Drop `tid` and `slug`.
- **`pkg/middleware/auth.go`**: drop the cross-validation block that
  compares `claims.Slug` against `X-Tenant-Slug`. Drop the
  `tenant.WithInfo` injection.
- **Traefik config**: drop the `X-Tenant-Slug` header injection from
  `deploy/traefik/dynamic/{dev,prod}.yml`. The middleware that does
  it gets removed.
- **`platform.tenants` table**: kept (operational metadata for the
  fleet of silos), but **never read by application code**. Only
  the deploy orchestrator and observability use it.
- **Frontend**: drop `NEXT_PUBLIC_TENANT_SLUG`, `getTenantSlug()`,
  `auth.user.tenantId`, `auth.user.tenantSlug`. The frontend talks
  to its silo's backend; it doesn't know about other tenants.

## Consequences

### Positive

- **Cutover becomes a one-liner.** `git pull && docker compose up -d`.
  No env juggling, no ID rewrites, no header gymnastics.
- **Schema is simpler.** Every ERP query loses one parameter. Smaller
  indexes (no tenant_id prefix). Less storage.
- **Migrator is simpler.** It no longer has to invent or map a
  tenant_id field; the data lands as-is.
- **Bugs from this class are extinct.** All seven 2026-04-20 bugs
  cease to exist by construction.
- **Aligns with ADR 022.** The silo was always the boundary; the
  code now reflects it.

### Negative

- **Schema migration touches every ERP table** (~676 tables).
  `ALTER TABLE … DROP COLUMN tenant_id` per table. Done in a single
  transaction inside the silo (no cross-silo concerns).
- **Every sqlc query file gets touched.** `WHERE tenant_id = $1` and
  the parameter get removed; sqlc regenerates Go bindings.
- **JWT shape changes.** Existing tokens stop working at cutover
  (the backend ignores `tid/slug`, but old tests / clients that
  asserted on those claims must be updated).
- **`platform.tenants` becomes write-only from the deploy side.**
  Code doesn't read it. Documented in ADR 022 amendment if needed.

### Rollout — big bang allowed

There are **no real users on prod** (the saldivia silo is dev-grade,
no operational data being entered yet). That removes the need for a
phased migration with coexistence/backward-compat. The refactor is a
single bigger PR that:

- Drops every `WHERE tenant_id` from queries and regenerates sqlc.
- Drops every `tenant_id` column from the schema (single migration).
- Drops `tid`/`slug` from the JWT.
- Drops the cross-validation in `pkg/middleware/auth.go`.
- Drops `X-Tenant-Slug` from Traefik dynamic configs.
- Drops `NEXT_PUBLIC_TENANT_SLUG` / `getTenantSlug()` / tenant
  fields from `AuthUser` and the auth store.

Tests get updated in the same PR. CI verifies. Merge → done.

If something breaks in prod after merge, we fix forward without
downtime concerns (no users to disrupt). This is the **explicit
trade-off** of the dev-grade-prod state — leverage it now while it
holds, document it in this ADR, and stop using "phased rollout"
language in plans that don't need it.

### Priority

**Maximum.** Listed as the next-cycle (2.0.22) headline work item
once 2.0.21 closes. Blocks any future tenant cutover until done.

## Out of scope

This ADR does **not** address:

- The fleet model (how many silos exist, how they're orchestrated,
  whether `platform.tenants` becomes a separate service). That
  remains under ADR 022.
- Multi-tenant SaaS deployments. SDA is silo-only by ADR 022; this
  decision reinforces that.
- Audit log cross-tenant aggregation (currently handled out-of-band
  via the platform DB, not via in-silo tenant_id columns).
