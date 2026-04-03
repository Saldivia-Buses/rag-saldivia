# Feedback Service

> Central nervous system of the SDA platform. Consumes quality/error/usage events from all services via NATS JetStream, stores granular data in tenant DB, computes hourly aggregates and health scores in platform DB, and creates alerts when thresholds are crossed.

## Endpoints

### Tenant-scoped (require Bearer auth)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| GET | `/v1/feedback/summary` | Bearer | AI quality, errors, features, NPS overview (`?period=30d`) |
| GET | `/v1/feedback/quality` | Bearer | Paginated AI quality feedback (`?period=30d&module=chat&limit=50`) |
| GET | `/v1/feedback/errors` | Bearer | Error reports by status (`?status=open&limit=50`) |
| GET | `/v1/feedback/usage` | Bearer | Usage analytics by module (`?period=30d`) |
| GET | `/v1/feedback/health-score` | Bearer | Composite health score (stub -- needs platform DB integration) |

### Platform-scoped (require admin JWT + platform slug)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/platform/feedback/tenants` | Platform admin | Health scores for all tenants (worst first) |
| GET | `/v1/platform/feedback/alerts` | Platform admin | Cross-tenant alerts with summary counts (`?status=active&limit=50`) |
| GET | `/v1/platform/feedback/quality` | Platform admin | Cross-tenant AI quality comparison (`?period=30d`) |

## Architecture

```
All Services --NATS (tenant.*.feedback.>)--> Consumer
                                               |
                            +------------------+----------------+
                            v                  v                v
                      Tenant DB           Platform DB       Notification
                      (granular            (aggregated       Service
                       events)              metrics +        (alerts via
                                            health scores)    NATS)
```

**Aggregator** runs hourly (configurable): computes `feedback_metrics` and `tenant_health_scores` in platform DB, triggers alerts via **Alerter**.

## Database

**Tenant DB tables:**
- `feedback_events` -- all feedback types in one table, discriminated by `category` (quality, errors, NPS, usage, performance, security)

**Platform DB tables:**
- `feedback_metrics` -- hourly aggregated metrics per tenant/module/category
- `tenant_health_scores` -- composite 0-100 scores per tenant (ai_quality, error_rate, usage, performance, security, NPS)
- `feedback_alerts` -- active alerts with severity, threshold, current value

**Migrations:**
- Tenant: `db/migrations/001_init.up.sql`
- Platform: `services/platform/db/migrations/002_feedback_metrics.up.sql`

## NATS Events

**Consumed (JetStream durable consumer):**

| Subject | Stream | Durable | Categories |
|---------|--------|---------|------------|
| `tenant.*.feedback.>` | `FEEDBACK` | `feedback-service` | See below |

Max 5 delivery attempts, 7 days retention, file storage.

**Feedback categories** (last segment of NATS subject):

| Category | Typical producer | Inferred module |
|----------|-----------------|-----------------|
| `response_quality` | Chat service | `chat` |
| `agent_quality` | Agent framework | `agent` |
| `extraction` | Document AI | `docai` |
| `detection` | Vision service | `vision` |
| `error_report` | Any service | `platform` |
| `feature_request` | Frontend | `platform` |
| `nps` | Frontend | `platform` |
| `usage` | Frontend | `platform` |
| `performance` | Services | `system` |
| `security` | Auth service | `auth` |

**Published (via Alerter):**
- `tenant.{slug}.notify.*` -- alert notifications when health thresholds are crossed

## Health Score Algorithm

Composite 0-100, updated hourly by the aggregator:

```
overall = ai_quality * 0.30 + errors * 0.25 + performance * 0.20 + security * 0.15 + usage * 0.10
```

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `FEEDBACK_PORT` | No | `8008` | HTTP listen port |
| `POSTGRES_TENANT_URL` | Yes | -- | Tenant DB connection string |
| `POSTGRES_PLATFORM_URL` | Yes | -- | Platform DB connection string |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server URL |
| `TENANT_ID` | No | `dev` | Tenant ID for aggregation targeting |
| `TENANT_SLUG` | No | `dev` | Tenant slug for aggregation targeting |
| `AGGREGATION_INTERVAL` | No | `1h` | Aggregation cycle (`1m` for testing) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **PostgreSQL:** Tenant DB (feedback events) + Platform DB (metrics, health scores, alerts)
- **NATS:** JetStream consumer (feedback events) + Publisher (alert notifications)
- **pkg/jwt:** Ed25519 key loading
- **pkg/middleware:** Auth middleware, SecureHeaders
- **pkg/nats:** Typed event publishing

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests
```
