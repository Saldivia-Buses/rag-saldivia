---
title: Service: platform
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/multi-tenancy.md
  - ../packages/featureflags.md
  - ../flows/tenant-routing.md
  - ./auth.md
---

## Purpose

Control plane for tenants, modules, feature flags, global config, and deploy
log. Source of truth for which tenants exist, which DB/Redis URLs they map
to, and which modules are enabled per tenant. The auth service's
`tenant.Resolver` reads from the same platform DB this service writes to.
Read this when adding tenant-scoped settings, lifecycle events, feature
flags, or the deploy registry.

## Endpoints

All `/v1/platform/*` routes require platform-admin role
(`requirePlatformAdmin`,
`services/platform/internal/handler/platform.go:620`).

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/nats/redis |
| GET | `/v1/platform/tenants` | platform admin | List tenants (paginated) |
| POST | `/v1/platform/tenants` | platform admin | Create tenant + provision DB/Redis URLs |
| GET | `/v1/platform/tenants/by-id/{tenantID}` | platform admin | Get by UUID |
| GET | `/v1/platform/tenants/{slug}` | platform admin | Get by slug |
| PUT | `/v1/platform/tenants/{tenantID}` | platform admin | Update tenant fields |
| POST | `/v1/platform/tenants/{tenantID}/disable` | platform admin | Mark disabled |
| POST | `/v1/platform/tenants/{tenantID}/enable` | platform admin | Re-enable |
| GET | `/v1/platform/tenants/{tenantID}/modules` | platform admin | List enabled modules |
| POST | `/v1/platform/tenants/{tenantID}/modules` | platform admin | Enable module |
| DELETE | `/v1/platform/tenants/{tenantID}/modules/{moduleID}` | platform admin | Disable module |
| GET | `/v1/platform/modules` | platform admin | Module catalog |
| GET | `/v1/platform/flags` | platform admin | List feature flags |
| POST | `/v1/platform/flags` | platform admin | Create flag |
| PUT | `/v1/platform/flags/{flagID}` | platform admin | Update flag |
| PATCH | `/v1/platform/flags/{flagID}` | platform admin | Toggle flag |
| DELETE | `/v1/platform/flags/{flagID}/kill` | platform admin | Kill switch |
| GET | `/v1/platform/config/{key}` | platform admin | Read global config |
| PUT | `/v1/platform/config/{key}` | platform admin | Set global config |
| POST | `/v1/platform/deploys` | platform admin | Record a deploy |
| GET | `/v1/platform/deploys` | platform admin | List deploys |
| GET | `/v1/flags/evaluate` | JWT | Per-tenant flag evaluation for any service/UI |

Routes wired in `services/platform/internal/handler/platform.go:65` and
`/v1/flags` separately at `services/platform/cmd/main.go:78`.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.{slug}.notify.platform_tenant_created` | pub | New tenant (`services/platform/internal/service/platform.go:200`) |
| `tenant.platform.notify.platform_flag_created` | pub | Feature flag created |
| `tenant.platform.notify.platform_flag_updated` | pub | Flag edited or toggled |
| `tenant.platform.notify.platform_flag_killed` | pub | Kill switch tripped |
| `tenant.platform.notify.platform_deploy_created` | pub | Deploy logged |

Event types are `platform_` + dot-replaced-with-underscore so the subject
stays at four segments matching the `tenant.*.notify.*` permission grant
(`services/platform/internal/service/platform.go:65`). Platform does not
subscribe.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `PLATFORM_PORT` | no | `8006` | HTTP listener port |
| `POSTGRES_PLATFORM_URL` | yes | — | Platform DB |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Lifecycle events |
| `REDIS_URL` | yes | `localhost:6379` | Token blacklist (mandatory; `services/platform/cmd/main.go:51`) |
| `PLATFORM_TENANT_SLUG` | no | `platform` | Required for admin role check |

## Dependencies

- **PostgreSQL platform** — `tenants`, `tenant_modules`, `feature_flags`,
  `global_config`, `deploys`.
- **Redis** — token blacklist (mandatory).
- **NATS** — lifecycle event publishing.
- No outbound application calls.

## Permissions used

No `RequirePermission` middleware. Authorization is `requirePlatformAdmin`
on every `/v1/platform/*` route plus a standard `Auth(publicKey)` on
`/v1/flags/evaluate` (any authenticated user can evaluate flags for their
own tenant).
