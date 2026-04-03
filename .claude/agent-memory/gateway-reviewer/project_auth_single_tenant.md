---
name: Auth service multi-tenant mode in progress (PR #54)
description: Auth now supports dual-mode (single/multi-tenant) but multi-tenant has blockers -- slug-as-tenantID, missing Traefik header injection for login
type: project
---

As of PR #54 (`feat/multi-tenant-openrouter`), the auth service supports two modes:
- **Single-tenant:** `POSTGRES_TENANT_URL` env var, creates one `service.Auth` at startup (unchanged from before).
- **Multi-tenant:** `POSTGRES_PLATFORM_URL` env var, uses `pkg/tenant.Resolver` to resolve per-request via `X-Tenant-Slug` header.

**Blockers identified in review:**
1. Login/Refresh/Logout routes are public (no Auth middleware), so `X-Tenant-Slug` must come from Traefik -- but the dev config doesn't inject it for auth routes.
2. `tenantID := slug` puts the slug into JWT `tid` claim instead of the real tenant UUID, which will break all downstream services that use `X-Tenant-ID` for SQL queries.

**Why:** Multi-tenant auth is necessary for production SaaS. The dual-mode pattern is correct but needs these gaps fixed before merging.

**How to apply:** When reviewing future auth changes, verify that tenant ID is a UUID (not slug) and that public routes receive tenant context from Traefik/gateway, not from the Auth middleware.
