---
title: Flow: WebSocket Real-time
audience: ai
last_reviewed: 2026-04-15
related:
  - ../services/ws.md
  - ../architecture/websocket-hub.md
---

## Purpose

How a single browser session goes from `/ws` upgrade to receiving live
events from any service. Read this before changing the upgrade handler,
the hub event loop, the NATS bridge, or the mutation router. The model
(channels, message shapes, why no polling) lives in
`architecture/websocket-hub.md`; this file documents the runtime sequence.

## Steps

1. Client opens `GET /ws` with `Authorization: Bearer <jwt>`; handler
   `services/ws/internal/handler/ws.go:34` `Upgrade` extracts the bearer
   from the header (never the query string, to avoid log leakage).
2. The handler verifies the JWT with the ed25519 public key and
   short-circuits with 401 on failure; if a `*security.TokenBlacklist` is
   wired, revoked JTIs are rejected before the upgrade completes.
3. `websocket.Accept` is called with the configured `OriginPatterns` from
   `WS_ALLOWED_ORIGINS`; if unset, dev mode logs a warning and accepts
   any origin (`InsecureSkipVerify=true`).
4. `hub.NewClientWithIdentity` creates a `Client` with `UserID`, `Email`,
   `TenantID`, `Slug`, `Role`, and the raw JWT (forwarded to gRPC tools)
   then `Hub.Register` enqueues it on the `register` channel.
5. The hub event loop (`services/ws/internal/hub/hub.go:38`) enforces
   `MaxClients` and `MaxClientsPerTenant`; over-capacity connections get
   an `Error` message and are immediately closed via `markClosed`.
6. The handler auto-subscribes the client to `notifications`, `modules`,
   and `presence`, then launches `client.WritePump(ctx)` in a goroutine
   and runs `client.ReadPump(ctx)` in the request goroutine until close.
7. `ReadPump` (`services/ws/internal/hub/client.go:124`) decodes inbound
   `Message` envelopes and dispatches them via `Hub.handleMessage`
   (hub.go:148): `Subscribe`/`Unsubscribe` mutate the client's channel
   set; `Mutation` goes to step 8.
8. `MutationHandler.Handle`
   (`services/ws/internal/hub/mutations.go:52`) runs the mutation in a
   goroutine and dispatches via gRPC to `chat` (and any other configured
   service); errors are mapped to clean codes (`token_expired`,
   `permission denied`, etc.).
9. Backend services publish events on `tenant.{slug}.{service}.{entity}`;
   `NATSBridge.handleNATSMessage`
   (`services/ws/internal/hub/nats.go:52`) parses the subject, derives
   `tenantSlug` and `channel`, and wraps the payload in a `Message{Event}`.
10. The bridge calls `Hub.BroadcastToTenant(slug, channel, msg)` which
    walks the client map, picks every subscribed connection of that
    tenant, marshals once, and pushes through `Client.TrySend`; the
    `WritePump` then writes to the socket with a `writeTimeout`.

## Invariants

- Tokens MUST come from the `Authorization` header; query-string tokens
  are never read because they leak into access logs.
- Every event is tenant-namespaced — the bridge requires three
  dot-segments (`tenant.{slug}.{rest}`) and drops anything else.
- A client subscribes to at most `maxSubscriptions` (64) channels;
  `Subscribe` returns false when the limit is hit, which the hub turns
  into an error response.
- `Client.send` is closed exactly once via `markClosed`; the
  `closed atomic.Bool` is the only safe gate before pushing into the
  channel.
- Mutations go through gRPC, not direct DB access, so RBAC and tenant
  resolution stay centralized in the target service.

## Failure modes

- `401 missing authorization token` / `401 invalid token` — caller didn't
  send a bearer or sent an expired one; `handler/ws.go:38`.
- `Error: server at capacity` / `tenant connection limit reached` — hit
  `MaxClients`/`MaxClientsPerTenant`; raise the limits in
  `cmd/main.go` wiring or shed connections.
- `client send buffer full, dropping message` (slog warn) — slow consumer;
  the hub drops the message rather than blocking the broadcast loop.
- Client never receives an event — wrong NATS subject prefix; verify the
  publisher uses `tenant.{slug}.{channel}` and the client subscribed to
  the matching channel string.
- `mutations not available` — `MutationHandler` is `nil` because
  `chatGRPCTarget` was empty at startup; check the WS service config.
- WS close immediately after upgrade — `markClosed` fired during
  registration; check capacity logs in `hub.Run`.
