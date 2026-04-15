---
title: Service: erp
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/audit.md
  - ../packages/traces.md
  - ../packages/pagination.md
  - ../architecture/multi-tenancy.md
---

## Purpose

ERP backend covering 19 functional modules: suggestions, catalogs, entities
(employees/customers/suppliers), stock, accounting, treasury, purchasing,
sales, invoicing, current accounts, production, HR, quality, maintenance,
manufacturing, admin, analytics, safety, and workshop. Every write hits the
audit writer and broadcasts a real-time NATS event so the frontend never
polls. Read this when adding/changing an ERP module, wiring new permissions,
or extending the audit/event payload.

## Endpoints

Every module is mounted at `/v1/erp/{module}` in
`services/erp/cmd/main.go:121`, using the same pattern: read group with
`erp.{module}.read`, write group additionally guarded by `authWrite`
(FailOpen=false) and `erp.{module}.write`.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/nats/redis check |
| GET | `/v1/erp/suggestions[/...]` | JWT + `erp.suggestions.read` | List/get suggestions, unread count |
| POST/PATCH | `/v1/erp/suggestions[/...]` | JWT + `erp.suggestions.write` | Create, respond, mark read |
| GET/POST/PUT/DELETE | `/v1/erp/entities[/...]` | JWT + `erp.entities.{read,write}` | Employees / customers / suppliers + contacts/notes/documents |
| GET/POST | `/v1/erp/stock[/...]` | JWT + `erp.stock.{read,write}` | Items, movements, adjustments |
| GET/POST | `/v1/erp/accounting[/...]` | JWT + `erp.accounting.{read,write,post,close}` | Accounts, journal entries, period close |
| GET/POST | `/v1/erp/treasury[/...]` | JWT + `erp.treasury.{read,write}` | Bank accounts, payments, reconciliation |
| GET/POST | `/v1/erp/purchasing[/...]` | JWT + `erp.purchasing.{read,write,approve,inspect}` | POs, approvals, receipt inspection |
| GET/POST | `/v1/erp/sales[/...]` | JWT + `erp.sales.{read,write}` | Sales orders, quotes |
| GET/POST | `/v1/erp/invoicing[/...]` | JWT + `erp.invoicing.{read,write}` | Invoices, AFIP integration hooks |
| GET/POST | `/v1/erp/accounts[/...]` | JWT + `erp.accounts.{read,write}` | Customer/supplier current accounts |
| GET/POST | `/v1/erp/production[/...]` | JWT + `erp.production.{read,write}` | Production orders |
| GET/POST | `/v1/erp/hr[/...]` | JWT + `erp.hr.{read,write}` | HR payroll/staffing |
| GET/POST | `/v1/erp/quality[/...]` | JWT + `erp.quality.{read,write}` | QA checks, NCRs |
| GET/POST | `/v1/erp/maintenance[/...]` | JWT + `erp.maintenance.{read,write}` | Assets, work orders |
| GET/POST | `/v1/erp/manufacturing[/...]` | JWT + `erp.manufacturing.{read,write,control,certify}` | Chassis assembly + certification |
| GET/POST | `/v1/erp/admin[/...]` | JWT + `erp.admin.{read,write}` | Internal communications, configs |
| GET | `/v1/erp/analytics[/...]` | JWT + `erp.analytics.read` | Cross-module analytics |
| GET/POST | `/v1/erp/safety[/...]` | JWT + `erp.safety.{read,write}` | Incidents, safety reports |
| GET/POST | `/v1/erp/workshop[/...]` | JWT + `erp.maintenance.{read,write}` | Shop floor maintenance |

Per-module routes are defined in `services/erp/internal/handler/{module}.go`
(see e.g. `services/erp/internal/handler/suggestions.go:81`).

## NATS events

Every module write calls `traces.Publisher.Broadcast(tenantID, channel, data)`
which lands on `tenant.{tenantID}.{channel}`. **Note:** the ERP code passes
the tenant **UUID** (not slug) as the first argument, so subjects look like
`tenant.{uuid}.erp_entities`, not `tenant.saldivia.erp_entities` (see
`services/erp/internal/service/entities.go:163`).

| Channel | Trigger |
|---|---|
| `erp_suggestions` | Suggestion created / responded |
| `erp_entities` | Entity created / updated / deleted |
| `erp_catalogs` | Catalog item changed |
| `erp_stock` | Stock movement / adjustment |
| `erp_accounting` | Journal entry / period close |
| `erp_treasury` | Payment / reconciliation |
| `erp_purchasing` | PO created / approved / received |
| `erp_sales` | Sales order changed |
| `erp_invoicing` | Invoice issued / cancelled |
| `erp_production` | Production order updated |
| `erp_maintenance` | Asset / work-order changes |
| `erp_manufacturing` | Chassis production events |
| `erp_certification` | Manufacturing certification |
| `erp_admin` | Admin communication |

ERP does not subscribe.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `ERP_PORT` | no | `8013` | HTTP listener port |
| `POSTGRES_TENANT_URL` | yes | — | Tenant DB (all ERP tables) |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Broadcast events |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |

## Dependencies

- **PostgreSQL tenant** — every ERP module owns its own tables
  (`erp_suggestions`, `erp_entities`, `erp_stock_*`, etc.) generated via
  sqlc.
- **Redis** — token blacklist.
- **NATS** — fan-out to WebSocket Hub for live ERP updates.
- **`pkg/audit`** — every write goes through `audit.Writer`.
- No outbound service calls.

## Permissions used

`erp.{module}.read` and `erp.{module}.write` for every module above. Some
modules add finer grants: `erp.accounting.post`, `erp.accounting.close`,
`erp.purchasing.approve`, `erp.purchasing.inspect`,
`erp.manufacturing.control`, `erp.manufacturing.certify`. See each module's
`Routes(authWrite)` in `services/erp/internal/handler/`.
