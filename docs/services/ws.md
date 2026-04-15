---
title: Service: ws
audience: ai
last_reviewed: 2026-04-15
related:
  - ../architecture/websocket-hub.md
  - ../flows/websocket-realtime.md
  - ../packages/nats.md
  - ./chat.md
---

## Purpose

WebSocket Hub: holds long-lived browser connections, bridges NATS events into
per-tenant per-channel fan-out, and proxies user-typed messages straight to the
chat service over gRPC. Read this before changing the wire protocol, channel
naming, subscription limits, the NATS-to-WS bridge, or the mutation pipeline
from browser back to data services.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + nats/redis check + active client count |
| GET | `/ws` | JWT (Authorization header or `?token=`) | WebSocket upgrade |

`/ws` runs with no HTTP write timeout (`services/ws/cmd/main.go:80`) because
connections are long-lived. Authentication happens during the upgrade
(`services/ws/internal/handler/ws.go:92`); after upgrade, identity is pinned
on the `Client` struct.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.*.>` | sub | All tenant events — bridge fans out to subscribed clients (`services/ws/internal/hub/nats.go:33`) |

The hub does **not** publish to NATS. Mutations go via gRPC (`chat`).
The wire protocol exposes three default channels per client: `notifications`,
`modules`, `presence` (`services/ws/internal/handler/ws.go:81`); additional
channels (e.g. `chat.messages:{sessionID}`) are subscribed on demand via the
`subscribe` message type defined in `services/ws/internal/hub/protocol.go`.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `WS_PORT` | no | `8002` | HTTP/WS listener port |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key for upgrade auth |
| `NATS_URL` | no | `nats://localhost:4222` | NATS connection |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |
| `CHAT_GRPC_URL` | no | — | If set, browser-originated mutations are forwarded to chat |

## Dependencies

- **NATS** — `tenant.*.>` wildcard subscription powers all real-time fan-out.
- **Redis** — token blacklist for revoked JWTs (rejected at upgrade).
- **gRPC outbound — chat** — `hub.MutationHandler` proxies browser
  `chat.send` messages to `ChatService.AddMessage`
  (`services/ws/cmd/main.go:53`).
- **No PostgreSQL.** The hub is stateless; client state lives in memory.

## Permissions used

None. The hub trusts the JWT's tenant claim for routing and lets producers
control what they publish per tenant. Per-channel authorization is enforced
upstream by the publishing service (`chat`, `notification`, etc.).
