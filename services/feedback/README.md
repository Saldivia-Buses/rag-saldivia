# Feedback Service

Central nervous system of the SDA platform. Consumes events from all services
via NATS JetStream, stores granular data in tenant DB, computes hourly
aggregates and health scores in platform DB, and alerts when thresholds
are crossed.

## Port

`:8008`

## Architecture

```
All Services ──NATS (tenant.*.feedback.>)──► Consumer
                                               │
                            ┌──────────────────┼────────────────┐
                            ▼                  ▼                ▼
                      Tenant DB           Platform DB       Notification
                      (granular)          (aggregated)      Service
                                                            (alerts)
```

## NATS Subjects Consumed

| Subject | Category | Producer |
|---|---|---|
| `tenant.{slug}.feedback.response_quality` | RAG response ratings | Chat |
| `tenant.{slug}.feedback.agent_quality` | Agent action success | Agent |
| `tenant.{slug}.feedback.extraction` | Document extraction corrections | DocAI |
| `tenant.{slug}.feedback.detection` | Vision detection confirmations | Vision |
| `tenant.{slug}.feedback.error_report` | Error reports | Any |
| `tenant.{slug}.feedback.feature_request` | Feature requests | Frontend |
| `tenant.{slug}.feedback.nps` | NPS surveys | Frontend |
| `tenant.{slug}.feedback.usage` | Feature usage events | Frontend |
| `tenant.{slug}.feedback.performance` | Latency/performance data | Services |
| `tenant.{slug}.feedback.security` | Security event summaries | Auth |

## REST Endpoints (all read-only)

### Tenant (require auth)

| Method | Path | Description |
|---|---|---|
| GET | `/v1/feedback/summary` | AI quality, errors, features, NPS overview |
| GET | `/v1/feedback/quality` | Paginated AI quality feedback |
| GET | `/v1/feedback/errors` | Error reports by status |
| GET | `/v1/feedback/usage` | Usage analytics by module |
| GET | `/v1/feedback/health-score` | Composite health score |

### Platform (require admin JWT)

| Method | Path | Description |
|---|---|---|
| GET | `/v1/platform/feedback/tenants` | Health scores for all tenants |
| GET | `/v1/platform/feedback/alerts` | Cross-tenant alerts |
| GET | `/v1/platform/feedback/quality` | Cross-tenant AI quality comparison |

## Health Score Algorithm

Composite 0-100, updated hourly:

```
overall = ai_quality * 0.30 + errors * 0.25 + performance * 0.20 + security * 0.15 + usage * 0.10
```

## Environment Variables

| Var | Required | Default | Description |
|---|---|---|---|
| `FEEDBACK_PORT` | No | `8008` | HTTP port |
| `POSTGRES_TENANT_URL` | Yes | - | Tenant DB connection |
| `POSTGRES_PLATFORM_URL` | Yes | - | Platform DB connection |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server |
| `JWT_SECRET` | Yes | - | JWT verification |
| `TENANT_ID` | No | `dev` | Tenant ID for aggregation |
| `AGGREGATION_INTERVAL` | No | `1h` | Aggregation cycle |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OTel collector |
