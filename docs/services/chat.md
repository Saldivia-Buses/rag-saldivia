---
title: Service: chat
audience: ai
last_reviewed: 2026-04-15
related:
  - ../flows/chat-agent-pipeline.md
  - ../architecture/websocket-hub.md
  - ../packages/database.md
  - ../packages/grpc.md
---

## Purpose

Chat session and message store. Serves the conversation list, message history,
session CRUD, and exposes a gRPC surface so the WebSocket Hub can persist
client-originated messages without round-tripping through HTTP. Read this when
adding session metadata, changing message schemas, wiring new
WebSocket → DB mutations, or adjusting `chat.new_message` event publishing.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/nats/redis check |
| GET | `/v1/chat/sessions` | JWT + `chat.read` | List user's chat sessions |
| POST | `/v1/chat/sessions` | JWT + `chat.write` | Create new session |
| GET | `/v1/chat/sessions/{sessionID}` | JWT + `chat.read` | Get one session |
| PATCH | `/v1/chat/sessions/{sessionID}` | JWT + `chat.write` | Rename session |
| DELETE | `/v1/chat/sessions/{sessionID}` | JWT + `chat.write` | Delete session + messages |
| GET | `/v1/chat/sessions/{sessionID}/messages` | JWT + `chat.read` | Paginated message history |
| POST | `/v1/chat/sessions/{sessionID}/messages` | JWT + `chat.write` | Append message (user/assistant/system) |

Routes registered in `services/chat/internal/handler/chat.go:42`. gRPC server
on a separate port for `ws` mutations: `services/chat/cmd/main.go:88`.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.{slug}.notify.chat.new_message` | pub | New `role=user` message persisted (`services/chat/internal/service/chat.go:192`) |

Assistant/system messages are not published — agent and downstream tools own
their own trace/event surface.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `CHAT_PORT` | no | `8003` | HTTP listener port |
| `CHAT_GRPC_PORT` | no | `50052` | gRPC listener port |
| `POSTGRES_TENANT_URL` | yes | — | Tenant DB connection string |
| `TENANT_SLUG` | no | `dev` | NATS subject prefix for events |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | NATS connection |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |

## Dependencies

- **PostgreSQL tenant** — `chat_sessions`, `chat_messages` tables.
- **Redis** — token blacklist (`pkg/security`).
- **NATS** — publishes `chat.new_message` for the WebSocket Hub.
- **gRPC clients (incoming)** — `ws` calls `ChatService.AddMessage` to persist
  messages typed in the browser without a second HTTP hop. Contract in
  `proto/chat/v1` and registered at `services/chat/cmd/main.go:89`.
- No outbound calls; chat is a leaf data service.

## Permissions used

- `chat.read` — list/read sessions and messages.
- `chat.write` — create/update/delete sessions and append messages.

Both checks are enforced inline at the route definition
(`services/chat/internal/handler/chat.go:47`).
