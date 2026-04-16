---
title: Architecture: Event Spine
audience: ai
last_reviewed: 2026-04-16
related:
  - ./nats-events.md
  - ./multi-tenancy.md
  - ../packages/spine.md
  - ../packages/outbox.md
  - ../conventions/cue.md
---

## Overview

The spine is the typed event bus built on NATS JetStream. It replaces raw
NATS publishing with envelope-wrapped, schema-versioned events that carry
trace context and support effectively-once delivery (at-least-once publish
via outbox + idempotent consume via `processed_events`).

## Components

| Component | Package | Purpose |
|-----------|---------|---------|
| Envelope | `pkg/spine` | `Envelope[T]` — typed wrapper with ID, tenant, trace, schema version |
| Publisher | `pkg/spine` | `Publish[T]` — serialize + OTel inject + PublishMsg |
| Consumer | `pkg/spine` | `Consume[T]` — JetStream pull + panic recovery + idempotency + DLQ |
| Outbox | `pkg/outbox` | Transactional outbox — atomic write + event in same DB tx |
| Drainer | `pkg/outbox` | Background worker polls outbox, publishes to NATS |
| Registry | `pkg/outbox` | Multi-tenant drainer manager with hotload |
| Codegen | `tools/eventsgen` | CUE specs → Go/TS/docs |
| DLQ Supervisor | `services/healthwatch` | Persists dead events, admin replay/drop UI |

## Data flow

```
Service handler
  └─ tx.Begin()
  └─ repo.InsertMessage(tx)      ← business write
  └─ outbox.PublishTx(tx, env)   ← outbox insert (same tx)
  └─ tx.Commit()

DrainerWorker (background)
  └─ SELECT ... FOR UPDATE SKIP LOCKED
  └─ nc.PublishMsg(subject, envelope)
  └─ UPDATE published_at

spine.Consume[T] (consumer)
  └─ Fetch batch from JetStream
  └─ PeekHeader → validate tenant
  └─ Decode[T] → ExtractTraceContext
  └─ tx.Begin()
  └─ EnsureFirstDelivery(tx)     ← idempotency
  └─ handler(tx, env)            ← business logic
  └─ tx.Commit() → Ack

On MaxDeliver exceeded:
  └─ PushDLQ → dlq.{stream}.{subject}
  └─ healthwatch persists to dead_events
  └─ operator replays or drops via /admin/dlq
```

## Guarantees

- **At-least-once publish:** outbox + drainer. NATS down → outbox accumulates.
- **Effectively-once consume:** `processed_events` INSERT in same tx as handler.
- **Panic safety:** consumer recovers, Naks with backoff, loop continues.
- **Schema evolution:** monotonic uint8 version, coexistence via codegen.

## Key decisions

See `docs/plans/spine/decisions.md` for the 8 technical decisions cemented
in Fase 0, including CUE choice, UUID v7, JSON wire, tenant-DB outbox.
