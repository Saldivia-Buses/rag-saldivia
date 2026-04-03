# Chat Service

> Manages chat sessions and messages. RBAC-enforced endpoints for CRUD operations. Publishes events to NATS for notifications and real-time updates.

## Endpoints

All routes mounted at `/v1/chat/sessions` and require Bearer auth.

| Method | Path | Auth | Permission | Description |
|--------|------|------|------------|-------------|
| GET | `/health` | No | -- | Health check |
| GET | `/v1/chat/sessions` | Bearer | `chat.read` | List user's sessions |
| POST | `/v1/chat/sessions` | Bearer | `chat.read` | Create session (optional: `collection` for RAG) |
| GET | `/v1/chat/sessions/{sessionID}` | Bearer | `chat.read` | Get session detail |
| PATCH | `/v1/chat/sessions/{sessionID}` | Bearer | `chat.write` | Rename session |
| DELETE | `/v1/chat/sessions/{sessionID}` | Bearer | `chat.write` | Delete session + messages |
| GET | `/v1/chat/sessions/{sessionID}/messages` | Bearer | `chat.read` | Get messages for session |
| POST | `/v1/chat/sessions/{sessionID}/messages` | Bearer | `chat.read` | Add message (role: user/assistant/system) |

## Database

**Instance:** Tenant DB

**Tables:**
- `sessions` -- chat sessions with optional RAG collection binding
- `messages` -- messages with role, content, thinking (extended reasoning), sources (RAG citations), metadata
- `tags` -- session tags (unique per session)
- `chat_feedback` -- thumbs up/down per message per user

**Migrations:** `db/migrations/000_deps.up.sql`, `001_init.up.sql`, `002_add_thinking.up.sql`

## NATS Events

**Published:**
- `tenant.{slug}.notify.*` -- notification events via `pkg/nats` Publisher
- `tenant.{slug}.chat.*` -- broadcast to WS Hub for real-time session list updates

**Consumed:** None

## Configuration

| Env var | Required | Default | Description |
|---------|----------|---------|-------------|
| `CHAT_PORT` | No | `8003` | HTTP listen port |
| `POSTGRES_TENANT_URL` | Yes | -- | Tenant DB connection string |
| `JWT_PUBLIC_KEY` | Yes | -- | Base64-encoded Ed25519 public key (PEM) |
| `NATS_URL` | No | `nats://localhost:4222` | NATS server URL |
| `TENANT_SLUG` | No | `dev` | Tenant slug for NATS subject namespacing |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **PostgreSQL:** Tenant DB (sessions, messages, tags, feedback tables)
- **NATS:** Publisher (session/message events)
- **pkg/jwt:** Ed25519 key loading
- **pkg/middleware:** Auth middleware, RequirePermission, SecureHeaders
- **pkg/nats:** Typed event publishing

## Development

```bash
go run ./cmd/...    # run locally
go test ./...       # run tests
```
