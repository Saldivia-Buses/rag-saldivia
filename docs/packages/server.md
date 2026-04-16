---
title: Package: pkg/server
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./build.md
  - ./otel.md
  - ./middleware.md
  - ./config.md
---

## Purpose

Bootstrap helper that eliminates ~30 lines of identical boilerplate from every
`cmd/main.go`: structured JSON logger, signal-aware context (SIGINT/SIGTERM),
OpenTelemetry init, chi router pre-wired with `RequestID`, `RealIP`,
`Recoverer`, `SecureHeaders`, `Timeout`, the `/v1/info` endpoint, and an HTTP
server with standard timeouts plus graceful shutdown. Import this in every
new service. DB, NATS, Redis, gRPC remain in `cmd/main.go` because they vary.

## Public API

Source: `pkg/server/server.go:1`

| Symbol | Kind | Description |
|--------|------|-------------|
| `App` | struct | Holds name, port, router, signal context, cleanup callbacks |
| `Option` | type | `func(*App)` |
| `WithPort(envVar, defaultPort)` | func | Read port from env var |
| `WithTimeout(d)` | func | Override the 30s request timeout middleware (use 0 for WebSocket/SSE) |
| `New(name, opts...)` | func | Build App: logger + signal ctx + OTel + chi router with base middleware |
| `App.Context()` | method | Signal-aware ctx (cancelled on SIGINT/SIGTERM) |
| `App.Router()` | method | The chi.Router |
| `App.Port()` / `App.Name()` | method | Accessors |
| `App.OnShutdown(fn)` | method | Register cleanup callback (LIFO order) |
| `App.Run()` | method | Listens, blocks until signal, graceful shutdown (30s) |
| `App.RunWithWriteTimeout(d)` | method | Run with custom WriteTimeout (use 0 for WS) |

## Usage

```go
app := server.New("sda-auth", server.WithPort("AUTH_PORT", "8001"))
ctx := app.Context()
r := app.Router()

pool, _ := database.NewPool(ctx, dbURL)
app.OnShutdown(func() { pool.Close() })

r.Get("/health", hc.Handler())
r.Post("/v1/auth/login", h.Login)

app.Run()
```

## Invariants

- Base middleware order: `RequestID` → `RealIP` → `Recoverer` →
  `SecureHeaders` → `Timeout` (`pkg/server/server.go:107`). Auth middleware is
  added by the service after this.
- `/v1/info` is wired automatically (`pkg/server/server.go:117`) — every
  service exposes its build info.
- Default request timeout is 30s. Pass `WithTimeout(0)` to disable for
  WebSocket and SSE handlers.
- Default WriteTimeout is 30s, IdleTimeout 60s. `RunWithWriteTimeout(0)`
  switches IdleTimeout to 120s — required for long-lived WS connections.
- Cleanup callbacks run in REVERSE registration order so dependencies are
  released before their consumers (`pkg/server/server.go:194`).
- Graceful shutdown deadline is 30s (`pkg/server/server.go:152`). Beyond that,
  the process exits with active connections dropped.

## Importers

`services/auth`, `astro`, `agent`, `bigbrother`, `chat`, `erp`, `feedback`,
`healthwatch`, `ingest`, `notification`, `platform`, `search`, `traces`, `ws`,
`.scaffold` — every service.
