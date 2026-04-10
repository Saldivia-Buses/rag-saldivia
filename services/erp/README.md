# ERP Service

Replaces the legacy Histrix intranet ERP. Each legacy module (99 total)
becomes a sub-module under `/v1/erp/`. Starts with suggestions, grows
with each business module migrated.

## Modules

| Module | Legacy | Endpoints | Status |
|---|---|---|---|
| Suggestions | SUGERENCIAS + SUGERESP (3 XML forms) | 6 | Done |

## Endpoints — Suggestions

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/v1/erp/suggestions` | `erp.read` | List (paginated) |
| GET | `/v1/erp/suggestions/unread` | `erp.read` | Count unread |
| GET | `/v1/erp/suggestions/{id}` | `erp.read` | Detail + responses |
| POST | `/v1/erp/suggestions` | `erp.write` | Create |
| POST | `/v1/erp/suggestions/{id}/respond` | `erp.write` | Add response |
| PATCH | `/v1/erp/suggestions/{id}/read` | `erp.write` | Mark as read |

## NATS events

| Produces | Subject |
|---|---|
| New suggestion | `tenant.{slug}.notify.new_suggestion` |
| Real-time update | `tenant.{slug}.erp_suggestions` |

## Stack used

- pkg/server (bootstrap)
- pkg/database (pool)
- pkg/audit (audit logging)
- pkg/traces (NATS publisher for notify + broadcast)
- pkg/pagination (API pagination)
- pkg/middleware (auth + RBAC)
- pkg/health (health checks)
- sqlc queries (repository layer)

## Development

```bash
cd services/erp
go run ./cmd/...
```
