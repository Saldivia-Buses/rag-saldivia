# ADR-015: Database-per-tenant isolation

**Fecha:** 2026-04-01
**Estado:** Aceptado

---

## Contexto

SDA Framework is a multi-tenant SaaS platform. Each tenant (e.g., Saldivia Buses, Empresa2) has users, chat sessions, documents, notifications, and feedback data. The system needs to guarantee that one tenant's data is never visible to another, even if there is a bug in the application layer.

The platform also needs a control plane that tracks all tenants, their plans, modules, and system-wide configuration.

## Opciones consideradas

- **Opcion A -- Shared schema with `tenant_id` column:** All tenants share one database. Every table has a `tenant_id` column. Every query must include `WHERE tenant_id = ?`. Pros: simple infrastructure (one DB), easy cross-tenant queries. Contras: one missed `WHERE` clause = data leak, schema changes affect all tenants simultaneously, noisy neighbor performance issues, harder to comply with data residency requirements.

- **Opcion B -- Schema-per-tenant (one Postgres, many schemas):** Each tenant gets a schema within one database. Pros: single Postgres instance, `SET search_path` for isolation. Contras: still one database to backup/restore (all-or-nothing), connection pool sharing, schema-level isolation is weaker than DB-level.

- **Opcion C -- Database-per-tenant:** Each tenant gets their own PostgreSQL instance (or logical database). A platform database stores the tenant registry with connection strings. Pros: strongest isolation, per-tenant backup/restore, per-tenant resource limits, per-tenant migration timing. Contras: more infrastructure to manage, cross-tenant queries require multiple connections.

## Decision

**Opcion C -- Database-per-tenant.** Implementation:

- **Platform DB** (`POSTGRES_PLATFORM_URL`): stores `tenants` table with `postgres_url` and `redis_url` per tenant, plus plans, modules, feature flags, deploy log, feedback metrics, health scores
- **Tenant DBs** (`POSTGRES_TENANT_URL` or resolved via `pkg/tenant.Resolver`): each has users, sessions, messages, notifications, feedback events, ingest jobs
- `pkg/tenant.Resolver` caches tenant connections in memory, resolves per-request from `X-Tenant-Slug` header
- Auth service operates in multi-tenant mode when `POSTGRES_PLATFORM_URL` is set (resolves tenant DB per login request)
- Other services currently connect to a single tenant DB directly (multi-tenant routing planned for production)
- Credential encryption columns added (`postgres_url_enc`, `redis_url_enc`) for AES-256-GCM at rest

## Consecuencias

**Positivas:**
- Data isolation is enforced at the database level -- no application-layer bug can leak data across tenants
- Per-tenant backup and restore: can recover one tenant without affecting others
- Per-tenant migration timing: can roll out schema changes to one tenant at a time
- Per-tenant resource limits: a heavy tenant cannot saturate the database for others
- Compliance: tenant data can be stored in different regions if required
- Tenant deletion is a simple `DROP DATABASE` -- no orphan cleanup

**Negativas / trade-offs:**
- More PostgreSQL instances to manage (mitigated by Docker Compose, one instance per tenant in dev)
- Cross-tenant analytics requires querying multiple databases (platform DB stores aggregated metrics via feedback service)
- Connection pooling per tenant adds memory overhead (mitigated by `Resolver` with pool caching and idle timeout)
- Tenant creation requires provisioning a new database and running migrations
- Connection strings stored in platform DB must be encrypted (implemented via `003_encrypted_credentials.up.sql`)

## Referencias

- `pkg/tenant/resolver.go` -- tenant DB resolution and pool caching
- `services/platform/db/migrations/001_init.up.sql` -- platform DB schema (tenants table)
- `services/platform/db/migrations/003_encrypted_credentials.up.sql` -- encrypted credential columns
- `services/auth/cmd/main.go` -- multi-tenant vs single-tenant mode switch
