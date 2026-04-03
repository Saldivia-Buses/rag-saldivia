# WebSocket Hub Service

> Real-time event relay. Accepts WebSocket connections authenticated via JWT (Ed25519), subscribes clients to tenant-scoped channels, and bridges NATS events to connected browsers.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check (includes connected client count) |
| GET | `/ws` | Bearer (header) | WebSocket upgrade. JWT verified inline, not via middleware |

## WebSocket Protocol

1. Client connects with `Authorization: Bearer <token>` header
2. JWT verified (Ed25519 public key). Claims extracted: `user_id`, `email`, `tenant_id`, `slug`, `role`
3. Client auto-subscribed to channels: `notifications`, `modules`, `presence`
4. NATS bridge forwards `tenant.{slug}.*` events to matching WebSocket clients
5. Messages are JSON: `{ "type": "event", "channel": "...", "data": {...} }`

## Database

None. Stateless relay.

## NATS Events

**Published:** None directly

**Consumed (via NATSBridge):**
- `tenant.*.>` -- wildcard subscription. Bridges ALL tenant-scoped NATS events to WebSocket clients
- Subject parsed as `tenant.{slug}.{channel}` and routed to clients of that tenant subscribed to that channel

**Example subjects forwarded:**
- `tenant.saldivia.notifications` -- notification events from notification service
- `tenant.saldivia.chat.messages` -- chat updates
- `tenant.saldivia.ingest.jobs` -- ingest progress
- `tenant.saldivia.modules` -- module enable/disable events

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `WS_PORT` | No | `8002` | HTTP listen port |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server URL |
| `WS_ALLOWED_ORIGINS` | No | -- | Comma-separated origin patterns (e.g., `*.sda.app,localhost:3000`). Empty = accept all (dev mode, logs warning) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **NATS:** Consumer (bridges events to WebSocket clients)
- **pkg/jwt:** Ed25519 key loading, token verification
- **pkg/middleware:** SecureHeaders
- **coder/websocket:** WebSocket library (not gorilla)

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests
```
