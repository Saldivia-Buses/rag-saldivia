---
title: Service: auth
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/auth-jwt.md
  - ../flows/login-jwt.md
  - ../packages/jwt.md
  - ../packages/middleware.md
  - ../packages/security.md
---

## Purpose

Auth Gateway: login, JWT issuance + refresh, token revocation, MFA (TOTP), and
service-account token minting for machine-to-machine calls. Read this when
touching login UX, JWT signing keys, MFA flows, RBAC permission checks, or the
multi-tenant `tenant.Resolver` integration. Runs in two modes: single-tenant
(legacy/dev) and multi-tenant (production, resolves DB per request).

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + nats/redis/postgres pings |
| POST | `/v1/auth/login` | none (5/min IP) | Username/password → JWT pair (or MFA challenge) |
| POST | `/v1/auth/refresh` | none (10/min IP) | Refresh token → new access token |
| POST | `/v1/auth/logout` | none | Revoke current access token (Redis blacklist) |
| POST | `/v1/auth/service-token` | shared key (5/min IP) | Issue service-account JWT |
| POST | `/v1/auth/mfa/verify` | mfa_token (5/min IP) | Complete MFA login challenge |
| GET | `/v1/auth/me` | JWT | Current user profile |
| PATCH | `/v1/auth/me` | JWT | Update own profile |
| GET | `/v1/auth/users` | JWT + `users.read` | List users in tenant |
| GET | `/v1/modules/enabled` | JWT | Tenant's enabled modules (used by frontend) |
| POST | `/v1/auth/mfa/setup` | JWT | Generate TOTP secret + QR |
| POST | `/v1/auth/mfa/verify-setup` | JWT | Confirm TOTP enrollment |
| POST | `/v1/auth/mfa/disable` | JWT | Disable TOTP |

Routing wired in `services/auth/cmd/main.go:138`. Handlers in
`services/auth/internal/handler/auth.go`.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.{slug}.notify.auth.login_success` | pub | Successful login (`services/auth/internal/service/auth.go:617`) |
| `tenant.{slug}.notify.auth.account_locked` | pub | Brute-force lockout |

Auth never subscribes; only publishes via `pkg/nats.Publisher.Notify`.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `AUTH_PORT` | no | `8001` | HTTP listener port |
| `POSTGRES_PLATFORM_URL` | yes (multi-tenant) | — | Platform DB for tenant resolution |
| `POSTGRES_TENANT_URL` | yes (single-tenant) | — | Direct tenant DB (dev mode) |
| `JWT_PRIVATE_KEY` | yes | — | Ed25519 private key (base64 PEM) |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key (base64 PEM) |
| `NATS_URL` | no | `nats://localhost:4222` | NATS for notify events |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist + rate-limit store |
| `SERVICE_ACCOUNT_KEY` | no | — | Shared secret enabling `/v1/auth/service-token` |
| `PLATFORM_TENANT_ID` | no | `platform` | Tenant ID stamped into service JWTs |
| `PLATFORM_TENANT_SLUG` | no | `platform` | Tenant slug stamped into service JWTs |
| `TENANT_ID` | no | `dev` | Single-tenant mode only |
| `TENANT_SLUG` | no | `dev` | Single-tenant mode only |

`SERVICE_ACCOUNT_KEY` may also be loaded from `/run/secrets/service_account_key`
(`services/auth/cmd/main.go:116`).

## Dependencies

- **PostgreSQL platform** — `tenant.Resolver` looks up tenant DB URLs
  (multi-tenant mode only).
- **PostgreSQL tenant** — users, sessions, MFA secrets. Pool acquired via
  `pkg/tenant.Resolver` per request.
- **Redis** — `pkg/security.TokenBlacklist` for revocation + per-IP rate
  limiting on login/refresh/MFA/service-token.
- **NATS** — publishes login/lockout events for the notification service to
  fan out as in-app notifications.
- **No outbound service calls.** All other services validate JWTs locally with
  the public key — auth is the only writer of tokens.

## Permissions used

- `users.read` (gates `GET /v1/auth/users` —
  `services/auth/cmd/main.go:149`).

All other authenticated endpoints rely solely on the access-token check
(`pkg/middleware.AuthWithConfig`); RBAC happens in services that read
permission claims from the JWT, not here.
