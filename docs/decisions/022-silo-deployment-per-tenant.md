# ADR 022 — Silo deployment per tenant (replaces pool-tenant model)

**Status:** accepted
**Date:** 2026-04-16
**Deciders:** Enzo Saldivia
**Supersedes:** the implicit pool-tenant stance in ADR 015 (db-per-tenant)

## Context

The codebase was designed as a **pool-tenant** system: a single deployment that
hosts many tenants, with code-level isolation enforced everywhere (`WHERE
tenant_id = $1`, tenant-namespaced NATS subjects, pool-per-tenant DB pools,
tenant-aware middleware in every service). This matches a SaaS with many small
customers (pool model).

Two facts invalidate that posture for this project:

1. **The workstation is a vertical-scale box** (Threadripper Pro 9975WX,
   256 GB RAM ECC, 96 GB VRAM, 8 TB NVMe). There is no cluster, no horizontal
   scale. One machine hosts everything.
2. **The customer profile is B2B enterprise**, not many-small-SMBs. Saldivia is
   a large organization with sensitive data (ERP, documents, payroll). If the
   product ever sells, the next customers look similar: few, large, sensitive.

Under those constraints, pool-tenancy buys nothing and costs a lot: every query
carries a tenant predicate, every service carries tenant-aware middleware, every
NATS subject is namespaced, every shared package has a tenant parameter. The
entire blast-radius of a missed `tenant_id` is cross-tenant data leakage.

The industry term for the opposite pattern is the **silo model** (AWS
terminology): one deployment per tenant, code is single-tenant, isolation is
physical/containerised. GitLab self-hosted, Sentry self-hosted, Supabase
self-hosted, Linear Sync, and Mattermost all use this model for enterprise
customers.

## Decision

**Adopt the silo model. The code is single-tenant; the tenant is the deployment.**

Operational shape:

- **One stack per tenant.** Each customer gets its own `docker compose` bundle
  with its own Postgres, NATS, Redis, MinIO, services, frontend, Traefik route.
- **Tenant identity is an env var.** `SDA_TENANT=saldivia` (or similar) is injected
  at container start. Services read it for branding, feature flags, storage paths,
  log context — but **never as a query predicate**, because there is no cross-tenant
  data in the same DB.
- **Deploy layout:** `deploy/tenants/<slug>/` holds the compose file, `.env`, and
  any tenant-specific config overrides.
- **Upgrade workflow:** `make deploy TENANT=<slug>` (or `TENANT=all` to iterate).

Code-level consequences (to be executed in subsequent sessions):

- **Remove `pkg/tenant/` and `pkg/middleware/tenantmw`.**
- **Drop `tenant_id` columns** from tables whose data is per-deployment anyway.
- **Drop JWT `tenant_slug` claim** (identity is implicit in the deployment).
- **Flatten NATS subjects** — `chat.message.created` replaces `tenant.{slug}.chat.message.created`.
- **Collapse the two migration trees** (platform + tenant) into a single tree.
- **Invariants reduce from 7 to 4** (see CLAUDE.md).

## Consequences

**Positive**

- Enormous code simplification. Estimated 20-30% LOC drop from removing
  tenant-aware scaffolding alone.
- Bug-by-construction immunity to cross-tenant leakage — there is no cross-tenant.
- Compliance question "where does customer X's data live?" has a trivial answer:
  in its own container/DB/volume.
- Customization per tenant without branches — env vars toggle features, branding, limits.
- Per-tenant resource scaling becomes a `docker compose` knob, not a code concern.
- Operations collapse: logs, metrics, dumps, restores, migrations are all scoped
  to one tenant by the shape of the deploy.

**Negative**

- Upgrades multiply by tenant count. Mitigated by `make deploy TENANT=all` iterating
  over `deploy/tenants/*/`. For today's N=1 this is a no-op; at N=10 it's a few
  minutes of sequential `docker compose pull && up`.
- Resource overhead per tenant (each has its own Postgres, NATS, Redis, service
  containers). Baseline per tenant ~2-4 GB RAM. The box fits dozens comfortably.
- Cross-tenant analytics require downstream aggregation. Not a near-term need.

**Trap to avoid**

**Silo on the outside, pool on the inside.** Keeping the tenant-aware code "just
in case" while each deployment only ever runs with one tenant is the worst of
both worlds — the complexity tax is paid, the benefit isn't. The silo commitment
means the tenant-aware code gets deleted.

## Alternatives considered

1. **Keep the pool model.**
   Rejected. The pool model exists to amortise shared infrastructure across many
   tenants. There is no cluster to share. The cost (code complexity, bug surface)
   is paid; the benefit is theoretical.

2. **Hybrid: shared control plane + per-tenant data plane.**
   Rejected. Attempts to keep some pool machinery (for admin, billing, signup)
   while siloing the real work. For a B2B-enterprise product with few customers,
   the control plane is also small enough that a per-tenant deploy covers it.
   Adds complexity without closing a real gap.

3. **Single-tenant only (one customer forever).**
   Rejected as a framing. Silo model is single-tenant *per deployment*, plural
   deployments. That is the path that accommodates additional customers without
   reintroducing pool complexity.

4. **Keep pool-tenant scaffolding "for optionality".**
   Rejected (see the trap above). Optionality has a real carrying cost and
   dilutes the clarity of the codebase. If the market ever demands pool-tenant,
   it can be rebuilt then, against evidence of need rather than a hypothesis.
