---
title: Package: pkg/grpc
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./jwt.md
  - ./middleware.md
  - ./tenant.md
---

## Purpose

gRPC server/client factories with the standard SDA defaults: 4MB message size,
keepalive, and an auth interceptor that reaches full parity with the HTTP
auth middleware (JWT verify + blacklist check + MFA-pending rejection +
tenant/role/permissions context injection). Used for inter-service RPCs over
the internal Docker network. Import this when adding a gRPC server or
dialing another SDA gRPC service.

## Public API

Sources: `pkg/grpc/server.go`, `pkg/grpc/client.go`, `pkg/grpc/interceptors.go`

Package import alias is `sdagrpc` (matches `package sdagrpc`).

| Symbol | Kind | Description |
|--------|------|-------------|
| `DefaultMaxRecvMsgSize` | const | 4 MiB |
| `NewServer(cfg, opts...)` | func | gRPC server with auth + keepalive + chained interceptors |
| `RegisterHealthServer(s)` | func | Registers `grpc.health.v1.Health/Check` (always SERVING) |
| `Dial(target, opts...)` | func | Insecure (internal-network) client with keepalive + recv-size |
| `ForwardJWT(ctx, jwt)` | func | Adds `authorization: Bearer <jwt>` to outgoing metadata |
| `InterceptorConfig` | struct | `PublicKey`, `Blacklist`, `FailOpen` |
| `AuthUnaryInterceptor(cfg)` | func | Verifies JWT from metadata, injects identity |
| `AuthStreamInterceptor(cfg)` | func | Streaming variant |
| `JWTFromIncomingContext(ctx)` | func | Extracts `authorization` bearer token from metadata |

## Usage

```go
srv := sdagrpc.NewServer(sdagrpc.InterceptorConfig{
    PublicKey: pubKey, Blacklist: bl, FailOpen: false,
})
sdagrpc.RegisterHealthServer(srv)
pb.RegisterChatServer(srv, chatHandler)

// Client side
conn, _ := sdagrpc.Dial("search:50051")
client := pb.NewSearchClient(conn)
ctx = sdagrpc.ForwardJWT(ctx, callerJWT)
resp, err := client.Query(ctx, req)
```

## Invariants

- Additional `grpc.ServerOption`s passed to `NewServer` MUST use
  `ChainUnaryInterceptor`, NOT `UnaryInterceptor` — the latter silently
  replaces the auth chain (`pkg/grpc/server.go:17`).
- The auth interceptor SKIPS `/grpc.health.v1.Health/Check` so health probes
  work without JWT (`pkg/grpc/interceptors.go:36`).
- `Dial` uses `insecure.NewCredentials` — only use over internal Docker network.
  Never expose externally.
- The auth interceptor calls the same `pkg/jwt.Verify` and the same blacklist
  used by `pkg/middleware/auth.go`, so HTTP and gRPC behave identically.

## Importers

`services/chat/cmd/main.go` (server), `services/search/cmd/main.go` (server),
`services/agent/internal/tools/grpc_search.go` (client),
`services/ws/internal/hub/mutations.go` (client).
