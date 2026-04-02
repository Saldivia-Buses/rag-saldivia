# Notification Service

Core service that handles in-app notifications, email delivery, and real-time push via WebSocket Hub.

## What it does

- **Consumes NATS events** from other services (auth, chat, ingest, etc.)
- **Persists notifications** in the tenant database
- **Sends emails** via SMTP (Mailpit in dev, real SMTP in prod)
- **Pushes real-time** to the browser via NATS -> WS Hub -> WebSocket

## Architecture

```
Other services ──NATS──► Notification Service
                              ├─► Tenant DB (persist)
                              ├─► SMTP (email)
                              └─► NATS ──► WS Hub ──► Browser
```

## NATS subjects

| Subject | Direction | Description |
|---|---|---|
| `tenant.*.notify.>` | Consumed | Events from other services that trigger notifications |
| `tenant.{slug}.notifications` | Published | Real-time push to WS Hub for browser delivery |

### Publishing a notification (for other services)

Publish to `tenant.{slug}.notify.{type}` with this payload:

```json
{
  "tenant_slug": "saldivia",
  "user_id": "user-123",
  "type": "chat.new_message",
  "title": "Nuevo mensaje",
  "body": "Recibiste un mensaje en la sesion 'Contratos'",
  "data": {"session_id": "sess-456", "email": "user@example.com"},
  "channel": "both"
}
```

Channels: `in_app` (default), `email`, `both`.

## REST API

| Method | Path | Description |
|---|---|---|
| GET | `/v1/notifications` | List notifications (`?unread=true&limit=50`) |
| GET | `/v1/notifications/count` | Unread count |
| PATCH | `/v1/notifications/{id}/read` | Mark one as read |
| POST | `/v1/notifications/read-all` | Mark all as read |
| GET | `/v1/notifications/preferences` | Get user preferences |
| PUT | `/v1/notifications/preferences` | Update user preferences |

All endpoints require `X-User-ID` header (injected by gateway).

## Database tables

- `notifications` — persisted notifications per user
- `notification_preferences` — per-user settings (email on/off, quiet hours, muted types)

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `NOTIFICATION_PORT` | `8005` | HTTP port |
| `POSTGRES_TENANT_URL` | required | Tenant database URL |
| `NATS_URL` | `nats://localhost:4222` | NATS server |
| `SMTP_HOST` | `localhost` | SMTP server host |
| `SMTP_PORT` | `1025` | SMTP server port |
| `SMTP_FROM` | `noreply@sda.local` | From address for emails |

## Dev

```bash
make build-notification    # Build binary
make test-notification     # Run tests
```

Emails in dev go to Mailpit (web UI at `localhost:8025`).
