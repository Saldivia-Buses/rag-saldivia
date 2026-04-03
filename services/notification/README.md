# Notification Service

> Consumes NATS events from all services, persists in-app notifications, sends emails via SMTP, and pushes real-time events to browsers via WS Hub.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| GET | `/v1/notifications` | Bearer | List notifications (`?unread=true&limit=50`) |
| GET | `/v1/notifications/count` | Bearer | Unread count |
| PATCH | `/v1/notifications/{notificationID}/read` | Bearer | Mark one as read |
| POST | `/v1/notifications/read-all` | Bearer | Mark all as read |
| GET | `/v1/notifications/preferences` | Bearer | Get user notification preferences |
| PUT | `/v1/notifications/preferences` | Bearer | Update preferences (email, in-app, quiet hours, muted types) |

All authenticated endpoints require `X-User-ID` header (injected by auth middleware).

## Architecture

```
Other services ──NATS──> Notification Service
                              |-> Tenant DB (persist notification)
                              |-> SMTP (email delivery)
                              |-> NATS -> WS Hub -> Browser (real-time push)
```

## Database

**Instance:** Tenant DB

**Tables:**
- `notifications` -- persisted notifications per user (type, title, body, data, channel, read status)
- `notification_preferences` -- per-user settings (email on/off, in-app on/off, quiet hours, muted types)

**Migrations:** `db/migrations/000_deps.up.sql`, `001_init.up.sql`

## NATS Events

**Consumed (JetStream durable consumer):**

| Subject | Stream | Durable | Description |
|---------|--------|---------|-------------|
| `tenant.*.notify.>` | `NOTIFICATIONS` | `notification-service` | Events from other services that trigger notifications |

JetStream config: file storage, 7 days retention, explicit ack.

**Published:**
- `tenant.{slug}.notifications` -- real-time push to WS Hub for browser delivery

### Publishing a notification (for other services)

Use `pkg/nats` Publisher:
```go
publisher.Notify("saldivia", natspub.Event{
    UserID:  "user-123",
    Type:    "ingest.completed",
    Title:   "Documento procesado",
    Body:    "contratos.pdf ingresado",
    Channel: "both",  // "in_app" (default), "email", "both"
})
```

The consumer checks user preferences (muted types, quiet hours) before persisting or sending email.

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `NOTIFICATION_PORT` | No | `8005` | HTTP listen port |
| `POSTGRES_TENANT_URL` | Yes | -- | Tenant DB connection string |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server URL |
| `SMTP_HOST` | No | `localhost` | SMTP server host |
| `SMTP_PORT` | No | `1025` | SMTP server port (Mailpit in dev) |
| `SMTP_FROM` | No | `noreply@sda.local` | From address for emails |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **PostgreSQL:** Tenant DB (notification + preferences tables)
- **NATS:** Consumer (JetStream durable) + Publisher (WS Hub forwarding)
- **SMTP:** Email delivery (Mailpit in dev, real SMTP in prod)
- **pkg/jwt:** Ed25519 key loading
- **pkg/middleware:** Auth middleware, SecureHeaders

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests
```

Emails in dev go to Mailpit (web UI at `localhost:8025`).
