---
title: Convention: Security Practices
audience: ai
last_reviewed: 2026-04-15
related:
  - ./error-handling.md
  - ./sqlc.md
  - ./logging.md
  - ../architecture/multi-tenancy.md
  - ../architecture/auth-jwt.md
  - ../packages/security.md
  - ../packages/jwt.md
  - ../packages/crypto.md
---

Security is a constraint, not a feature. Every rule here is enforced either by middleware, the pre-commit invariant hook, or the security-auditor agent on every PR.

## JWT is the only identity source

DO derive `UserID`, `TenantID`, `Slug`, `Role`, and `Permissions` from the JWT claims provided by `pkg/middleware`. The middleware verifies the Ed25519 signature locally with the public key; no service calls Auth Service per request. See `pkg/jwt/jwt.go:31`.

DO strip identity headers from incoming requests in middleware before passing to handlers — clients cannot supply `X-User-Id`. The Traefik gateway also enforces this, but defense in depth.

DON'T trust any header containing identity unless `pkg/middleware/auth.go` set it. DO validate JWT slug equals the Traefik-injected slug to prevent a token from one tenant being used against another.

## Tenant isolation checklist

For any new flow or endpoint, every box below must be checked.

- [ ] JWT slug cross-validated against Traefik slug (`pkg/middleware/auth.go:115`)
- [ ] Identity headers stripped before processing (`pkg/middleware/auth.go:45-49`)
- [ ] NATS subject includes tenant slug; consumer validates it
- [ ] Database queries scoped to tenant (per-tenant DB or RLS)
- [ ] MinIO/S3 keys prefixed with tenant slug
- [ ] Redis keys namespaced per tenant
- [ ] WebSocket broadcasts filtered by tenant slug
- [ ] Tool calls forward user JWT (not service account)
- [ ] Collections namespaced: `{slug}-{collection}`
- [ ] Slug validated against `^[a-zA-Z0-9_-]+$` before any NATS publish

DO use `pkg/tenant.Resolver` to obtain a per-tenant DB pool. Never share a pool across tenants. See [sqlc](./sqlc.md) for query-level rules.

## Input validation at boundaries

DO validate request bodies at the handler. Length checks, format checks, presence checks happen before reaching the service layer. See `services/auth/internal/handler/auth.go:122-129` for the canonical pattern.

DO wrap every POST/PUT/PATCH body with `http.MaxBytesReader(w, r.Body, N)` before decoding JSON. 1MB (`1<<20`) is the default; use less if the endpoint expects only small payloads. See `services/auth/internal/handler/auth.go:114`.

DO validate that user-supplied identifiers belong to the requesting tenant before acting on them.

DON'T accept slugs, IDs, or paths from request bodies and pass them to filesystem, shell, or SQL without validating against an allowlist or strict regex.

## Rate limiting and brute force

DO use `pkg/security` for token blacklists and brute-force counters. Auth service tracks `failed_logins` and `locked_until` per user and locks the account after N attempts (`services/auth/db/queries/auth.sql:62`).

DO add per-IP and per-tenant rate limits at handlers exposed to unauthenticated traffic (login, refresh, password reset). Limits live in `pkg/security/init.go`.

## Encryption of sensitive data at rest

DO encrypt fields containing TOTP secrets, API keys, OAuth tokens, and any third-party credential before storing in PostgreSQL.

DO use `pkg/crypto` envelope encryption (KEK/DEK + AAD) for tenant-scoped secrets — never AES directly with a single key. The KEK rotates; the DEK is per-record. AAD must include `tenant_id` so a ciphertext from one tenant cannot be decrypted in another's context.

See `pkg/crypto/envelope.go` for the API.

## Secrets handling

DON'T hardcode secrets, API keys, JWT private keys, or DB passwords in code. The pre-commit invariant scans for common patterns.

DO load secrets from Docker secrets, environment variables (set via the orchestrator), or a future Vault integration. See `pkg/config`.

DON'T log secrets, even at Debug level. See [logging](./logging.md) PII section.

DON'T commit `.env` files with real values. The repo's `.gitignore` excludes them; review carefully when adding new env files.

## Every tenant write is event-sourced

DO publish a NATS event after every successful tenant write so the WebSocket Hub can push real-time updates. Subject format: `tenant.{slug}.{service}.{entity}[.{action}]`. Without the event, the frontend stays stale and falls back to polling — which is forbidden.

DO validate the slug against `^[a-zA-Z0-9_-]+$` before constructing the subject. A bad slug poisons the subject namespace.

## RBAC at the handler

DO check permissions at the handler entry, before resolving the service. Permissions come from JWT claims (`Permissions []string`) populated from `role_permissions`.

DO add the new permission row in the same migration that introduces the endpoint (see [migrations](./migrations.md) RBAC section).

## Audit log

DO write an audit row via `pkg/audit` for every state-changing action: who, when, what resource, what changed. The table is append-only and replicated.

DON'T conflate audit entries with operational logs. See [logging](./logging.md).
