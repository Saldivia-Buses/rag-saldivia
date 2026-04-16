---
title: NATS Events & JetStream
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/nats.md
  - ../conventions/error-handling.md
  - websocket-hub.md
  - multi-tenancy.md
  - observability.md
---

This document describes how SDA services communicate asynchronously over
NATS. Read it before publishing a new event, creating a new JetStream
consumer, or routing anything new through the WebSocket Hub — subject
shape, durability, and tenant isolation are all enforced here.

## Subject conventions

Every subject published by SDA starts with `tenant.{slug}.`. The full
canonical shape is:

```
tenant.{slug}.{service}.{entity}[.{action}]
```

Examples observed in the codebase:

- `tenant.{slug}.notify.{eventType}` — events for the Notification Service.
- `tenant.{slug}.extractor.job` — extraction jobs for the Python extractor
  (`services/ingest/internal/service/documents.go:114`).
- `tenant.{slug}.traces.>` — agent / astro execution traces consumed by the
  traces service (`services/traces/internal/service/consumer.go:16`).
- `tenant.{slug}.{channel}` — UI broadcasts the WS Hub forwards to
  subscribed clients (`pkg/nats/publisher.go:154`).

The slug, channel, and event type are validated against the regex
`^[a-zA-Z0-9_-]+$` (`pkg/nats/publisher.go:19`) before being concatenated
into a subject — wildcards, dots, and whitespace are rejected to prevent
subject injection.

## Connection helper

`natspub.Connect(url)` (`pkg/nats/publisher.go:24`) is the canonical way to
open a NATS connection. It enables retry-on-failed-connect, infinite
reconnect, and disconnect/reconnect logging so service startup never blocks
on a flaky NATS. Per-service NATS credentials are passed via
`{SERVICE}_NATS_PASS` env vars in production
(`deploy/docker-compose.prod.yml:193`).

## Publisher

`natspub.Publisher` (`pkg/nats/publisher.go:90`) exposes two entry points:

- `Notify(ctx, slug, evt)` — wraps an `Event` (or any struct with a `type`
  field) and publishes to `tenant.{slug}.notify.{type}`.
- `Broadcast(ctx, slug, channel, data)` — sends a raw payload to
  `tenant.{slug}.{channel}`. The WS Hub forwards these directly to clients
  subscribed to that channel.

Both methods have `…Ctx` variants that inject the active OpenTelemetry trace
context into NATS message headers (`InjectTraceContext`). Use the context
variants from new code so traces span the publish/consume boundary.

## JetStream — durable streams

Two long-lived JetStream streams exist today:

| Stream         | Subjects               | Storage | Retention | Owner          |
|----------------|------------------------|---------|-----------|----------------|
| `NOTIFICATIONS`| `tenant.*.notify.>`    | File    | 7 days    | notification   |
| `TRACES`       | `tenant.*.traces.>`    | File    | (default) | traces         |

Both consumers are durable, use `AckExplicitPolicy`, and cap redelivery at
5 attempts before a message goes to the DLQ (`pkg/nats/dlq.go`). See
`services/notification/internal/service/consumer.go:67` and
`services/traces/internal/service/consumer.go:47`.

## Non-JetStream (fire-and-forget)

The WS Hub bridge subscribes to the wildcard `tenant.*.>` with a plain core
NATS subscription (`services/ws/internal/hub/nats.go:33`). Messages that
arrive while a client is offline are lost — this is intentional. Anything
that needs durability must go through JetStream, not Broadcast.

## Tenant isolation

Subjects are the only mechanism preventing cross-tenant event leakage:

- The publisher refuses to build a subject with an invalid slug
  (`publisher.go:108`).
- The WS bridge derives the tenant from `subject.split('.')[1]` and only
  forwards to clients whose `Slug` matches (`hub/nats.go:60`,
  `hub.BroadcastToTenant`).
- Consumers should always include `tenant.*.` in their `FilterSubject`.
  Never subscribe to bare `>` from inside an SDA service.

## Spine (Plan 26)

`pkg/spine` wraps NATS with typed `Envelope[T]`, schema versioning, and
consumer guarantees. New events should use the spine framework:

- **Publishing:** `outbox.PublishTx` inside a DB tx (guaranteed delivery).
- **Consuming:** `spine.Consume[T]` with built-in panic recovery, idempotency,
  OTel trace context, and DLQ.
- **Event types:** defined in `pkg/events/spec/*.cue`, generated via
  `make events-gen`. See `docs/conventions/cue.md`.
- **Architecture:** `docs/architecture/spine.md`.
- **DLQ:** dead events visible at `/admin/dlq`, replay/drop via healthwatch.

Legacy `pkg/nats.Publisher.Notify/Broadcast` still works for non-migrated
Types. Migration is per-Type, atomic (producer + consumer in same PR).

## When to publish what

Every state-changing endpoint must publish a NATS event so the WS Hub can
push the change to live clients (frontend has zero polling). Reads do not
publish. Use `outbox.PublishTx` for new events (guaranteed delivery). Legacy
`Notify` for notification badges, `Broadcast` for live UI updates.
