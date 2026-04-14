# Platform Service

> Control plane for system administration. Manages tenants, modules, feature flags, and global configuration. Only accessible by platform admins (verified via JWT slug + admin role).

## Endpoints

All routes under `/v1/platform/` require platform admin JWT (admin role + platform tenant slug).

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |

### Tenants

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/platform/tenants` | Platform admin | List all tenants |
| POST | `/v1/platform/tenants` | Platform admin | Create tenant (slug, name, plan, postgres_url, redis_url) |
| GET | `/v1/platform/tenants/{slug}` | Platform admin | Get tenant detail |
| PUT | `/v1/platform/tenants/{tenantID}` | Platform admin | Update tenant (name, plan, settings) |
| POST | `/v1/platform/tenants/{tenantID}/disable` | Platform admin | Disable tenant |
| POST | `/v1/platform/tenants/{tenantID}/enable` | Platform admin | Re-enable tenant |

### Tenant Modules

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/platform/tenants/{tenantID}/modules` | Platform admin | List enabled modules for tenant |
| POST | `/v1/platform/tenants/{tenantID}/modules` | Platform admin | Enable module for tenant |
| DELETE | `/v1/platform/tenants/{tenantID}/modules/{moduleID}` | Platform admin | Disable module for tenant |

### Module Catalog

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/platform/modules` | Platform admin | List all available modules |

### Feature Flags (Admin)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/platform/flags` | Platform admin | List all feature flags (with rollout_pct) |
| POST | `/v1/platform/flags` | Platform admin | Create flag (default enabled=false, rollout_pct=0) |
| PUT | `/v1/platform/flags/{flagID}` | Platform admin | Update flag (enabled, rollout_pct) |
| PATCH | `/v1/platform/flags/{flagID}` | Platform admin | Toggle feature flag |
| DELETE | `/v1/platform/flags/{flagID}/kill` | Platform admin | Kill switch — disable immediately |

### Feature Flags (Evaluate)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/flags/evaluate` | Any JWT | Evaluate flags for caller (tenant/user from JWT, deterministic rollout) |

Response: `{"flags": {"flag_name": true, ...}}` — booleans only, no metadata (DS7).
Rollout: `fnv32(flagID + ":" + userID) % 100 < rollout_pct` (deterministic per user).

### Deploy Log

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/platform/deploys` | Platform admin | Record a deployment |
| GET | `/v1/platform/deploys` | Platform admin | List recent deployments |

### Global Config

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/platform/config/{key}` | Platform admin | Get config value |
| PUT | `/v1/platform/config/{key}` | Platform admin | Set config value |

## Auth Model

The `requirePlatformAdmin` middleware verifies:
1. Valid JWT (Ed25519 signature)
2. `claims.Role == "admin"`
3. `claims.Slug == PLATFORM_TENANT_SLUG` (distinguishes platform admins from tenant admins)

## Database

**Instance:** Platform DB (single, cross-tenant)

**Tables:**
- `plans` -- subscription plans (Starter $49, Business $299, Professional $999, Enterprise custom)
- `tenants` -- registry with connection info, plan, enabled state, settings
- `modules` -- module catalog (core, platform, vertical, ai_service categories)
- `tenant_modules` -- per-tenant module enablement with config
- `global_config` -- system-wide key/value configuration
- `rag_models` -- RAG model registry (provider, endpoint, VRAM, category)
- `feature_flags` -- global or per-tenant flags
- `deploy_log` -- deployment history
- `feedback_metrics` -- hourly aggregated feedback (added by `002_feedback_metrics.up.sql`)
- `tenant_health_scores` -- composite health scores per tenant
- `feedback_alerts` -- active alerts for tenant issues

**Migrations:** `db/migrations/001_init.up.sql`, `002_feedback_metrics.up.sql`, `003_encrypted_credentials.up.sql`

## NATS Events

**Published (lifecycle events):**
- `tenant.platform.notify.platform_tenant_created` — new tenant registered
- `tenant.platform.notify.platform_deploy_created` — deploy recorded
- `tenant.platform.notify.platform_flag_created` — flag created
- `tenant.platform.notify.platform_flag_updated` — flag updated/toggled
- `tenant.platform.notify.platform_flag_killed` — flag killed

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `PLATFORM_PORT` | No | `8006` | HTTP listen port |
| `POSTGRES_PLATFORM_URL` | Yes | -- | Platform DB connection string |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `PLATFORM_TENANT_SLUG` | No | `platform` | Slug that identifies platform admin tenant |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **PostgreSQL:** Platform DB (control plane metadata)
- **NATS:** Lifecycle event publishing (tenant, deploy, flag changes)
- **Redis:** Token blacklist for admin endpoints
- **pkg/jwt:** Ed25519 key loading, token verification
- **pkg/middleware:** Auth, SecureHeaders
- **pkg/audit:** Immutable audit log

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests
```
