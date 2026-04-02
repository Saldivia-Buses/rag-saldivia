# Platform Service

Core service for system administration. Manages tenants, modules, feature flags, and global configuration. Only accessible by platform admins (Saldivia team).

## What it does

- **Tenant management** — CRUD, enable/disable, plan assignment
- **Module management** — enable/disable modules per tenant, catalog
- **Feature flags** — global or per-tenant, toggle on-the-fly
- **Global config** — system-wide key/value configuration

## Architecture

```
Platform Admin (CLI / MCP / UI)
  → Traefik (verifies platform admin JWT)
    → Platform Service (port 8006)
      → Platform DB (cross-tenant metadata)
```

The Platform Service connects to the **Platform DB** — not to tenant DBs. It's the control plane that tracks all tenants, their plans, enabled modules, and system configuration.

## REST API

All endpoints require `X-Platform-Admin: true` header (set by gateway after JWT verification).

### Tenants

| Method | Path | Description |
|---|---|---|
| GET | `/v1/platform/tenants` | List all tenants (summary) |
| POST | `/v1/platform/tenants` | Create tenant |
| GET | `/v1/platform/tenants/{slug}` | Get tenant detail |
| PUT | `/v1/platform/tenants/{tenantID}` | Update tenant |
| POST | `/v1/platform/tenants/{tenantID}/disable` | Disable tenant |
| POST | `/v1/platform/tenants/{tenantID}/enable` | Re-enable tenant |

### Tenant Modules

| Method | Path | Description |
|---|---|---|
| GET | `/v1/platform/tenants/{tenantID}/modules` | List enabled modules |
| POST | `/v1/platform/tenants/{tenantID}/modules` | Enable module |
| DELETE | `/v1/platform/tenants/{tenantID}/modules/{moduleID}` | Disable module |

### Modules Catalog

| Method | Path | Description |
|---|---|---|
| GET | `/v1/platform/modules` | List all available modules |

### Feature Flags

| Method | Path | Description |
|---|---|---|
| GET | `/v1/platform/flags` | List all flags |
| PATCH | `/v1/platform/flags/{flagID}` | Toggle flag |

### Global Config

| Method | Path | Description |
|---|---|---|
| GET | `/v1/platform/config/{key}` | Get config value |
| PUT | `/v1/platform/config/{key}` | Set config value |

## Database

Uses the **Platform DB** (not tenant DBs). Tables:

- `plans` — subscription plans (Starter, Business, Professional, Enterprise)
- `tenants` — registry with connection info, plan, state
- `modules` — available module catalog
- `tenant_modules` — which modules each tenant has
- `global_config` — system-wide key/value config
- `rag_models` — RAG model registry
- `feature_flags` — global or per-tenant flags
- `deploy_log` — deployment history

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PLATFORM_PORT` | `8006` | HTTP port |
| `POSTGRES_PLATFORM_URL` | required | Platform database URL |

## Dev

```bash
make build-platform    # Build binary
make test-platform     # Run tests
```
