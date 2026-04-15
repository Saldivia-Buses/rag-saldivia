---
title: Package: pkg/traces
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./nats.md
  - ./otel.md
---

## Purpose

Shared NATS-based publisher for execution-trace events. Used by `agent` and
`astro` (and by extension every domain that runs through the agent runtime)
to record reasoning traces — start, end, per-step events, feedback,
notifications, and broadcasts. The Traces Service consumes these and persists
them; the Notification Service and WS Hub consume the notify/broadcast
flavours. Import this in any service that produces structured execution
events. Distinct from `pkg/otel` which handles distributed tracing for
HTTP/gRPC spans.

## Public API

Source: `pkg/traces/publisher.go:3`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Publisher` | struct | Wraps `*nats.Conn`; if nil, all methods are no-ops |
| `NewPublisher(nc)` | func | Constructor |
| `Publisher.Start(slug, sessionID, userID, query)` | method | Returns new `traceID`; publishes `tenant.{slug}.traces.start` |
| `Publisher.End(slug, traceID, status, modelsUsed, durationMS, in, out, toolCalls, costUSD)` | method | Publishes `tenant.{slug}.traces.end` |
| `Publisher.Event(slug, traceID, eventType, seq, durationMS, data)` | method | Per-step event on `tenant.{slug}.traces.event` |
| `Publisher.Feedback(slug, category, data)` | method | Publishes `tenant.{slug}.feedback.{category}` |
| `Publisher.Notify(slug, eventType, data)` | method | Publishes `tenant.{slug}.notify.{eventType}` |
| `Publisher.Broadcast(slug, channel, data)` | method | Publishes `tenant.{slug}.{channel}` |
| `ValidateToken(s)` | func | Allowlist regex `^[a-zA-Z0-9_-]+$` for subject tokens |

## Usage

```go
p := traces.NewPublisher(nc) // nil nc = no-op publisher
traceID := p.Start(slug, sessionID, userID, query)
p.Event(slug, traceID, "llm_call", 1, 320, map[string]any{
    "model": "qwen2.5", "tokens_in": 1200, "tokens_out": 480,
})
// ...later
p.End(slug, traceID, "ok", []string{"qwen2.5"}, 1850, 1200, 480, 0, 0.012)
```

## Invariants

- `Publisher` returns immediately when `p.nc == nil` — services can construct
  one with no NATS connection during tests without writing `if publisher != nil`.
- All methods validate slug/event-type/channel tokens before publishing
  (`pkg/traces/publisher.go:30`). Invalid tokens are silently dropped (the
  goal is "don't crash production"; observability gaps are acceptable).
- Trace IDs are UUIDs generated client-side (`pkg/traces/publisher.go:29`),
  so the publisher remains useful even when NATS is down.
- Subject naming follows the SDA NATS convention (Invariant #3) —
  `tenant.{slug}.{kind}` — and overlaps with `pkg/nats.Publisher`. This
  package is the trace-shaped variant; the typed `Event` payloads of
  `pkg/nats` are for the Notification Service.
- Marshal/publish errors are logged via `slog.Error` and never returned —
  trace publishing must never fail business logic.

## Importers

`services/agent/internal/service/agent.go`,
`services/agent/internal/service/traces.go`,
`services/astro/internal/handler/astro.go`,
`services/astro/cmd/main.go`,
`services/erp/cmd/main.go` and most `services/erp/internal/service/*.go`
(audit-heavy domains call traces alongside audit).
