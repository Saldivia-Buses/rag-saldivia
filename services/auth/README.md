# Auth Service

> Authentication gateway: login, JWT issuance (Ed25519), refresh tokens, RBAC, MFA (TOTP), audit logging. Supports single-tenant and multi-tenant modes.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/v1/auth/login` | No | Email/password login. Returns tokens or MFA challenge |
| POST | `/v1/auth/refresh` | No | Rotate access + refresh tokens (cookie or body) |
| POST | `/v1/auth/logout` | No | Revoke refresh token, clear cookie |
| GET | `/v1/auth/me` | Bearer | Current user profile + roles + permissions |
| GET | `/v1/modules/enabled` | Bearer | Enabled modules for the tenant (currently returns core defaults) |
| POST | `/v1/auth/mfa/setup` | Bearer | Generate TOTP secret + URI for QR enrollment |
| POST | `/v1/auth/mfa/verify-setup` | Bearer | Activate MFA by proving valid TOTP code |
| POST | `/v1/auth/mfa/verify` | No | Complete MFA-gated login with `mfa_token` + TOTP code |
| POST | `/v1/auth/mfa/disable` | Bearer | Disable MFA (requires valid TOTP code as confirmation) |

## Database

**Instance:** Tenant DB (one per tenant)

**Tables:**
- `users` -- user accounts with MFA fields, brute-force lockout
- `roles` -- system roles (admin, manager, user, viewer)
- `permissions` -- granular permissions (14 seeded: users.*, roles.*, collections.*, chat.*, config.*, docs.*, ingest.*, audit.*)
- `role_permissions` -- role-to-permission mapping
- `user_roles` -- user-to-role mapping
- `refresh_tokens` -- hashed refresh tokens with expiry + revocation
- `audit_log` -- immutable audit trail (action format: `service.entity.verb`)

**Migrations:** `db/migrations/001_init.up.sql`, `002_audit_actions.up.sql`

## NATS Events

**Published:**
- `tenant.{slug}.notify.{type}` -- login events, security alerts (via `pkg/nats` Publisher)
- `tenant.{slug}.{channel}` -- broadcast to WS Hub for real-time notifications

**Consumed:** None

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `AUTH_PORT` | No | `8001` | HTTP listen port |
| `POSTGRES_TENANT_URL` | One of tenant/platform | -- | Direct tenant DB connection (single-tenant mode) |
| `POSTGRES_PLATFORM_URL` | One of tenant/platform | -- | Platform DB for tenant resolution (multi-tenant mode) |
| `JWT_PRIVATE_KEY` | Yes | -- | Base64-encoded Ed25519 private key (PEM) |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server URL |
| `TENANT_ID` | No | `dev` | Tenant ID (single-tenant mode only) |
| `TENANT_SLUG` | No | `dev` | Tenant slug (single-tenant mode only) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **PostgreSQL:** Tenant DB (user/role/token/audit tables) or Platform DB (tenant resolution)
- **NATS:** Publisher (login events, security notifications)
- **pkg/jwt:** Ed25519 key loading, token signing/verification
- **pkg/middleware:** Auth middleware, SecureHeaders
- **pkg/nats:** Typed event publishing
- **pkg/tenant:** Resolver for multi-tenant DB lookups

## Modes

- **Multi-tenant:** When `POSTGRES_PLATFORM_URL` is set. Resolves tenant DB per-request via `X-Tenant-Slug` header.
- **Single-tenant:** When only `POSTGRES_TENANT_URL` is set. Direct DB connection, dev mode.

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests (includes integration tests)
```
