---
title: WebSocket Hub
audience: ai
last_reviewed: 2026-04-15
related:
  - ../services/ws.md
  - ../flows/websocket-realtime.md
  - nats-events.md
  - auth-jwt.md
---

This document describes the WebSocket Hub: SDA's real-time backbone. Read
it before adding a new client-pushed event, a new mutation channel, or
changing how WebSocket auth and subscriptions work — every live UI update
in the frontend depends on this layer.

## Role

The Hub is the single WebSocket endpoint exposed to clients (`/ws`,
`services/ws/cmd/main.go:64`). It does **not** own business state; it
multiplexes:

- A persistent connection per browser tab.
- Per-connection subscriptions to logical channels (e.g.
  `chat.messages:session-123`, `notifications`).
- A NATS bridge that fans out tenant-scoped events to subscribed clients.
- An optional gRPC mutation handler that lets a client invoke chat actions
  over the same socket instead of opening a new HTTP request.

## Connection lifecycle

1. Client opens `wss://{slug}.{domain}/ws` with a bearer token in either
   the `Authorization` header or a `?token=` query param.
2. The handler verifies the JWT (Ed25519, blacklist check) before
   upgrading.
3. On success the Hub creates a `Client` carrying `UserID`, `Email`,
   `TenantID`, `Slug`, `Role`, and the raw JWT for forwarding to gRPC
   (`services/ws/internal/hub/client.go:39`).
4. The Hub enforces caps: 1000 total clients, 300 per tenant
   (`services/ws/internal/hub/hub.go:10`). Excess connections receive an
   `Error` envelope and are closed.

## Wire protocol

All frames use a single envelope (`hub/protocol.go:25`):

| Field    | Notes                                              |
|----------|----------------------------------------------------|
| `type`   | `subscribe` / `unsubscribe` / `mutation` / `event` / `error` |
| `channel`| logical channel name (e.g. `chat.messages:session-id`) |
| `action` | only for mutations (`create_session`, `send_message`)   |
| `id`     | correlation id for request/response matching       |
| `data`   | opaque JSON payload                                |
| `error`  | populated only on `error` frames                   |

Common channel prefixes are constants — `sessions`, `chat.messages`,
`notifications`, `admin.stats`, `ingest.jobs`, `presence`, `collections`,
`modules`, `fleet.vehicles`, `fleet.maintenance`
(`hub/protocol.go:36`). A client may hold up to 64 active subscriptions
(`hub/client.go:17`).

## Server → client (NATS bridge)

`hub.NATSBridge` subscribes to the wildcard `tenant.*.>`
(`services/ws/internal/hub/nats.go:33`). For each NATS message it:

1. Splits the subject into `["tenant", slug, channel...]`.
2. Tries to unmarshal the payload as a `Message`; otherwise wraps the raw
   bytes as an `event` envelope on the derived channel.
3. Calls `Hub.BroadcastToTenant(slug, channel, msg)` — only clients of that
   tenant subscribed to that channel receive the frame.

This is where the tenant prefix from `nats-events.md` becomes
isolation: the slug is taken straight from the subject the publisher built.

## Client → server (mutations)

When `MutationHandler` is wired (gRPC URL via `CHAT_GRPC_URL`,
`services/ws/cmd/main.go:53`), the Hub accepts `mutation` frames and
forwards them to the chat gRPC service. The user's JWT travels in the
metadata so the downstream service performs its own auth and tenant checks.
Without `CHAT_GRPC_URL` set, mutations are silently disabled — clients
should still POST to the REST endpoints.

## Health

The `/health` endpoint reports NATS connectivity, optional Redis
(blacklist), and the live `clients` count
(`services/ws/cmd/main.go:67`). The Hub uses an internal `Run` loop
goroutine for register/unregister; per-client send loops own their write
deadlines (`writeTimeout = 10s`) and `markClosed` is atomic to prevent
write-after-close panics.

## What you must never do

- Subscribe to a NATS subject without the tenant wildcard — see
  nats-events.md.
- Push to a channel from a service handler — publish a NATS event instead;
  the bridge will fan it out.
- Trust client-supplied identity in a mutation; the gRPC backend re-runs
  middleware on the forwarded JWT.
