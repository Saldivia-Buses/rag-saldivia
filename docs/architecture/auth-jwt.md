---
title: JWT Authentication
audience: ai
last_reviewed: 2026-04-15
related:
  - ../services/auth.md
  - ../packages/jwt.md
  - ../flows/login-jwt.md
  - multi-tenancy.md
---

This document describes the JWT trust model used by every SDA service. Read
it before changing the auth middleware, the JWT claim shape, the refresh
flow, or the token blacklist — these touch the security boundary of the
entire platform.

## Trust model

SDA uses **Ed25519 asymmetric signing**:

- The Auth Service (and only the Auth Service) holds the private key. It
  signs both access and refresh tokens.
- Every other service holds only the public key and verifies tokens
  locally — no service-to-Auth round trip per request.
- A compromised non-auth service cannot forge tokens.

Keys are loaded from base64-encoded PEM env vars (`JWT_PRIVATE_KEY`,
`JWT_PUBLIC_KEY`) by `jwt.ParsePrivateKeyEnv` / `jwt.ParsePublicKeyEnv`
(`pkg/jwt/jwt.go:187`). Every Go service calls
`jwt.MustLoadPublicKey("JWT_PUBLIC_KEY")` at startup
(`pkg/jwt/jwt.go:206`).

## Token lifetimes

| Token        | Default expiry | Where it lives             |
|--------------|----------------|----------------------------|
| Access       | 15 minutes     | `Authorization: Bearer …`  |
| Refresh      | 7 days         | Hashed in tenant DB        |

Defaults come from `jwt.DefaultConfig` (`pkg/jwt/jwt.go:52`). Override via
`JWT_ACCESS_EXPIRY` / `JWT_REFRESH_EXPIRY`. Refresh tokens are stored as
bcrypt hashes in the tenant DB `refresh_tokens` table
(`db/tenant/migrations/001_auth_init.up.sql:56`); the raw token never hits
disk.

## Claim shape

`jwt.Claims` (`pkg/jwt/jwt.go:31`) carries the standard registered claims
plus:

| Claim   | JSON  | Meaning                         |
|---------|-------|---------------------------------|
| UserID  | `uid` | User UUID in the tenant DB      |
| Email   | `email` | User email                    |
| Name    | `name` | Display name                   |
| TenantID| `tid` | Tenant UUID from the platform DB|
| Slug    | `slug`| Tenant subdomain slug           |
| Role    | `role`| Primary role name               |
| Permissions | `perms` | RBAC permission strings    |
| ID (JTI)| `jti` | UUID — populated automatically when missing |

`Verify` rejects a token whose `UserID`, `TenantID`, or `Slug` is empty.

## Verification middleware

`middleware.AuthWithConfig` (`pkg/middleware/auth.go:38`) wraps every
protected handler. On each request it:

1. Strips any client-supplied identity headers (`X-User-*`, `X-Tenant-*`).
2. Skips `/health`.
3. Extracts the bearer token, calls `jwt.Verify`, returns 401 on failure.
4. Checks the token JTI against the Redis blacklist; on Redis error returns
   503 unless `FailOpen` is set.
5. Rejects tokens whose role is `mfa_pending` (those are valid only for
   `/v1/auth/mfa/verify`).
6. Re-injects identity headers from the verified claims.
7. Stores `tenant.Info`, role, permissions, user id, and email in the
   request context for downstream handlers.
8. **Cross-validates** the JWT slug against the Traefik-injected slug — a
   mismatch returns `403 tenant mismatch` (line 115).

## Revocation (blacklist)

`security.TokenBlacklist` (`pkg/security/blacklist.go:17`) stores revoked
JTIs in shared Redis under the prefix `sda:token:blacklist:`. The TTL on
each entry equals the token's remaining lifetime so the set self-prunes.
Logout, password change, and rotation paths call `Revoke` /
`RevokeAll`. Auth, ws, ingest, traces, and healthwatch all wire the
blacklist into their middleware (e.g. `services/auth/cmd/main.go:146`).

## Refresh and rotation

`POST /v1/auth/refresh` accepts a refresh token, validates the bcrypt hash
in `refresh_tokens`, marks the row revoked, and issues a fresh access +
refresh pair. Rate limiting on `/v1/auth/refresh` is 10 requests per minute
per IP (`services/auth/cmd/main.go:133`); login is 5 per minute. The login
flow with full sequence is documented in `../flows/login-jwt.md`.

## Service-to-service tokens

For machine-to-machine calls, Auth exposes
`POST /v1/auth/service-token` keyed by `SERVICE_ACCOUNT_KEY` (a Docker
secret). It mints an access token scoped to the platform tenant. Disable by
unsetting the key (`services/auth/cmd/main.go:117`).
