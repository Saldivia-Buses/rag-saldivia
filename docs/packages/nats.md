---
title: Package: pkg/nats
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ../architecture/nats-events.md
  - ./traces.md
---

## Purpose

NATS publishing primitives shared across SDA services. Provides a single
`Connect` with project-standard retry options, a typed `Publisher` for the
two universal subject patterns (notifications and broadcasts), OpenTelemetry
trace propagation through NATS headers, and a DLQ helper. See
`architecture/nats-events.md` for the full subject naming convention. Import
this whenever a service publishes events.

The package import alias is `natspub` (matches `package natspub`).

## Subject conventions (Invariant #3)

- `tenant.{slug}.notify.{eventType}` — consumed by Notification Service.
- `tenant.{slug}.{channel}` — broadcast directly to WS Hub channel.
- `dlq.{stream}.{originalSubject}` — dead letter queue.

Slug and event-type tokens are validated against allowlist regexes.

## Public API

Sources: `pkg/nats/publisher.go`, `pkg/nats/dlq.go`, `pkg/nats/otel.go`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Connect(url)` | func | `nats.Connect` with `MaxReconnects(-1)`, retry-on-connect, slog handlers |
| `IsValidSubjectToken(s)` | func | Allowlist regex `^[a-zA-Z0-9_-]+$` for single token |
| `IsValidEventType(s)` | func | Allows dots: `^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$` |
| `Event` | struct | Notification payload: `UserID`, `Type`, `Title`, `Body`, `Data`, `Channel` |
| `EventPublisher` | interface | `Notify`, `Broadcast` |
| `ContextAwarePublisher` | interface | Adds `NotifyCtx`, `BroadcastCtx` (trace propagation) |
| `Publisher` | struct | Implements both interfaces (`pkg/nats/publisher.go:86`) |
| `New(nc)` | func | Constructor |
| `Publisher.Notify(slug, evt)` | method | Publishes to `tenant.{slug}.notify.{type}` |
| `Publisher.NotifyCtx(ctx, slug, evt)` | method | Same with OTel trace context in headers |
| `Publisher.Broadcast(slug, channel, data)` | method | Publishes to `tenant.{slug}.{channel}` |
| `Publisher.BroadcastCtx(ctx, slug, channel, data)` | method | Same with trace propagation |
| `LogDLQ(nc, stream, subject, data, err)` | func | Logs + publishes to `dlq.{stream}.{subject}` |
| `InjectTraceContext(ctx, msg)` | func | OTel propagator → NATS headers |
| `ExtractTraceContext(ctx, msg)` | func | NATS headers → OTel context |

## Usage

```go
nc, _ := natspub.Connect("nats://nats:4222")
p := natspub.New(nc)
err := p.NotifyCtx(ctx, slug, natspub.Event{
    UserID: targetUserID, Type: "chat.new_message",
    Title: "Nuevo mensaje", Body: previewText,
})

// Consumer side: re-attach trace context
sub.Subscribe("tenant.*.notify.>", func(msg *nats.Msg) {
    ctx := natspub.ExtractTraceContext(context.Background(), msg)
    handle(ctx, msg)
})
```

## Invariants

- ALL slugs and event types MUST be validated before subject interpolation.
  `Notify`/`Broadcast` reject invalid input with `fmt.Errorf` instead of
  publishing (`pkg/nats/publisher.go:108`).
- Event type extraction avoids double JSON marshal: it pattern-matches on
  `Event`/`*Event`/`map[string]any` (`pkg/nats/publisher.go:113`).
- The Python extractor mirrors `_SAFE_SUBJECT_RE` from `IsValidSubjectToken`
  (`pkg/nats/publisher.go:18`) — keep both in sync.
- `LogDLQ` is independent of the Notification Service so it works even when
  notification is the failing consumer (`pkg/nats/dlq.go:13`).
- The Connect helper sets `MaxReconnects(-1)` (infinite). Services rely on
  reconnection logging via slog.

## Importers

`services/auth`, `agent`, `astro`, `bigbrother`, `chat`, `erp`, `feedback`,
`ingest`, `notification`, `platform`, `traces`, `ws` — every publisher and
the notification consumer.
