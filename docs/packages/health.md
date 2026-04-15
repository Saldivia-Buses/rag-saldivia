---
title: Package: pkg/health
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./server.md
---

## Purpose

Shared `/health` endpoint builder. Each service registers named dependency
checks (DB, Redis, NATS, downstream gRPC services) and the handler runs them
in parallel with a 3-second context timeout. Returns HTTP 200 with per-dep
latency when all pass, HTTP 503 when any fail. Import this in every service
that needs a Docker/k8s-friendly health probe.

## Public API

Source: `pkg/health/health.go:6`

| Symbol | Kind | Description |
|--------|------|-------------|
| `CheckFunc` | type | `func(ctx) error` — one dependency probe |
| `ExtraFunc` | type | `func() (string, any)` — adds key-value to response |
| `Checker` | struct | Holds registered checks for a service |
| `New(service)` | func | Constructor |
| `Checker.Add(name, fn)` | method | Register a dependency check |
| `Checker.AddExtra(fn)` | method | Register an extra metric (e.g., active connections) |
| `Checker.Handler()` | method | Returns `http.HandlerFunc` |
| `DependencyStatus` | struct | `Status` (`up`/`down`), `LatencyMs`, `Error` |
| `Response` | struct | `Status`, `Service`, `Dependencies`, `Extra` |

## Usage

```go
hc := health.New("sda-auth")
hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
hc.Add("redis", func(ctx context.Context) error { return rdb.Ping(ctx).Err() })
hc.AddExtra(func() (string, any) { return "active_sessions", sessionCount() })
r.Get("/health", hc.Handler())
```

## Invariants

- Checks run in parallel with `sync.WaitGroup`. Each is bounded by a 3s
  context timeout (`pkg/health/health.go:67`).
- Per-check error messages are sanitized to `"unavailable"`
  (`pkg/health/health.go:94`) so probes don't leak internal detail. The full
  error is logged via `slog.Warn`.
- HTTP status: 200 if all up, 503 if any down. Body is always JSON.
- `Cache-Control: no-store` is set so reverse proxies never cache health.

## Importers

All Go services: `auth`, `astro`, `agent`, `bigbrother`, `chat`, `erp`,
`feedback`, `healthwatch`, `ingest`, `notification`, `platform`, `search`,
`traces`, `ws`.
