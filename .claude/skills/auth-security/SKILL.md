---
name: auth-security
description: Use when touching JWT verification, RBAC, auth middleware, rate limiting, or anything that affects who can read or write what. Tenant isolation is handled by the silo deployment model (ADR 022), not by code — so this skill focuses on identity (JWT), permissions (RBAC), and the security rules that still apply at the code boundary.
---

# auth-security

Scope: `pkg/jwt/`, `pkg/middleware/`, `services/app/internal/core/auth/`,
`services/erp/internal/handler/`, any handler that reads/writes data,
any NATS publisher or subscriber.

## Tenant isolation is not a code concern

Since ADR 022 the system is **silo-deployed**: one stack per tenant, with its own
DB, NATS, Redis. The code is single-tenant by construction. That means:

- **No `WHERE tenant_id = $1`** — the DB for this deployment is the tenant's DB.
- **No `pkg/tenant/` in handlers.** If you see it still being referenced, it is dead code
  scheduled for removal under the consolidation bias.
- **No tenant claim in the JWT.** Identity is `user_id` + `roles`, nothing else.
- **NATS subjects are flat** (`chat.message.created`, not `tenant.{slug}.chat.message.created`).

Any code that re-introduces tenant plumbing at the app layer is a **blocking**
finding — it reintroduces complexity the silo model exists to eliminate.

## The load-bearing invariant

### JWT is the only source of identity

- Verified **locally** in every service via ed25519. Public key from config.
- `pkg/middleware/jwtmw.Verify` decodes, validates, puts claims in `ctx`.
- Handlers read identity with `jwtmw.Claims(ctx)`. Never from headers, query params, or the body.
- Claims that matter: `sub` (user ID), `roles` (array), `exp`. No `tenant_slug`.
- Refresh tokens are opaque; short-lived access tokens carry all identity.

```go
// Correct
claims := jwtmw.Claims(ctx)
if !claims.HasRole("admin") { ... }

// Wrong — trusts the caller
if r.Header.Get("X-Is-Admin") == "true" { ... }
```

### NATS subjects are flat

```go
// Correct — flat subject, this deployment's NATS is already scoped to this tenant
subject := "chat.message.created"

// Wrong — reintroduces pool-tenant shape that no longer applies
subject := fmt.Sprintf("tenant.%s.chat.message.created", slug)
```

- The NATS cluster for each deployment sees only its tenant's traffic by construction.
- Don't build subjects with string concatenation. Use constants or a small builder.

## RBAC

- Roles are strings in the JWT: `admin`, `user`, `operator`, `service:<name>`.
- Permission checks live in the service layer, not handlers:
  `service.Create(ctx, input)` returns `httperr.Forbidden` if claims don't allow.
- No database table of permissions. The JWT is the permission.

## Rate limiting

- `pkg/middleware/ratelimit` with Redis.
- Keyed on `tenant_id + user_id + route`, not on IP.
- Defaults in each service's config. Override per route if needed.

## Secrets

- Never logged. Never in commits (`.env*` is gitignored).
- JWT private key: only in `services/auth`. All other services have the public key.
- Rotation: `tools/cli/sda keys rotate` (pending, docs in decisions).

## Before merge (security checklist)

When the diff touches this scope, you must be able to answer **yes** to all:

- [ ] No new code reads identity from headers/body — only from JWT claims via `ctx`.
- [ ] No tenant plumbing reintroduced (no `tenant_id`, no `pkg/tenant`, no tenant-namespaced subjects).
- [ ] No secret is logged (grep the diff for `slog`, `log.`, `fmt.Print`).
- [ ] New endpoints have a role check (or are explicitly public).

When in doubt, dispatch the `parallel-research` flow to audit: one Explore agent for
call-site coverage, one for tests, one for event subjects.
