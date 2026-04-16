---
title: Flow: Login to JWT
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/auth-jwt.md
  - ../services/auth.md
---

## Purpose

End-to-end sequence for issuing, refreshing, and revoking SDA access/refresh
tokens. Read this when changing any login, refresh, logout, or middleware
verification path. Architecture (claims shape, key material, RBAC) lives in
`architecture/auth-jwt.md` — this file documents the runtime sequence only.

## Steps

1. Client sends `POST /v1/auth/login` with `{email, password}`; handler
   `services/auth/internal/handler/auth.go:112` caps body at 1MB and decodes
   the `loginRequest`.
2. Handler resolves the per-tenant `service.Auth` via `resolveService(r)`
   (subdomain → `tenant.Resolver` pool) before any DB access.
3. `Auth.Login` at `services/auth/internal/service/auth.go:123` normalizes
   the email, calls `GetUserByEmail`, and runs bcrypt with a dummy hash on
   miss to keep response time constant.
4. On success it checks `is_active`/`locked_until`, records the attempt via
   `recordSuccessfulLogin`, and consults `CheckMFARequired`; an MFA-pending
   short-lived JWT is returned instead of real tokens when MFA is enabled.
5. Otherwise `pkg/jwt/jwt.go:73` `CreateAccess` mints an EdDSA access token
   (15 min, embeds `UserID`, `TenantID`, `Slug`, `Role`, random `JTI`) and
   `CreateRefresh` issues the refresh token.
6. Service revokes prior refresh tokens for the user and stores the new one
   in `auth.refresh_tokens` (rotation policy); handler then sets the
   HttpOnly `sda_refresh` cookie via `setRefreshCookie` and returns the
   access token in the JSON body.
7. On every protected request `pkg/middleware/auth.go:38` `AuthWithConfig`
   strips client-supplied identity headers, verifies the bearer with the
   ed25519 public key, and rejects the `mfa_pending` role.
8. The middleware checks `cfg.Blacklist.IsRevoked(ctx, claims.ID)` (Redis
   set populated on logout/password change) and fails closed when Redis is
   unreachable unless `FailOpen` is set.
9. `POST /v1/auth/refresh` reads the refresh token from cookie or body,
   `Auth.Refresh` (auth.go:357) verifies it, deletes the row, and rotates a
   new access+refresh pair atomically; reuse of a deleted token returns
   `ErrInvalidRefreshToken`.
10. `POST /v1/auth/logout` (handler line 218) extracts the access JTI,
    invokes `svc.Logout` to delete the refresh row and push the JTI into the
    blacklist with TTL = remaining access lifetime, then clears the cookie.

## Invariants

- Access tokens are EdDSA only (`SigningMethodEdDSA`); HMAC tokens MUST be
  rejected by `pkg/jwt`.
- Every access JWT carries a non-empty `JTI`; the middleware refuses tokens
  without one when a blacklist is configured.
- Refresh tokens are single-use — `Refresh` deletes the row before issuing
  the next pair; replay of a consumed token returns 401.
- The `sda_refresh` cookie is HttpOnly + `Secure` + `SameSite=Strict` and
  scoped to `/v1/auth`; non-browser clients pass the token in the body.
- The `mfa_pending` role is only valid against `/v1/auth/mfa/verify`; any
  other route MUST return 401 (auth middleware line 90).

## Failure modes

- `401 invalid email or password` — wrong creds, disabled user, or unknown
  user; check `services/auth/internal/service/auth.go:127` and the bcrypt
  branches around line 145.
- `429 too many attempts` (from `ErrAccountLocked`) — `auth.users.locked_until`
  is in the future; clear it manually or wait for `recordFailedLogin` to
  expire.
- `401 mfa verification required` mid-flow — caller used the MFA challenge
  token on a non-MFA endpoint; finish `/v1/auth/mfa/verify` first.
- `503 auth check unavailable` — Redis blacklist is down and `FailOpen=false`;
  inspect `pkg/security` blacklist init logs in the consuming service.
- `401 invalid token` after rotation — caller cached an old refresh token;
  client must re-login. Logs land in `services/auth/internal/service/auth.go`
  around `Refresh` (line 358) and the middleware verify call.
- `403 tenant mismatch` — JWT slug ≠ subdomain slug; see `tenant-routing.md`.
