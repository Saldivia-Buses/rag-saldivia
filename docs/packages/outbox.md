---
title: Package: pkg/outbox
audience: ai
last_reviewed: 2026-04-16
related:
  - ../README.md
  - ./spine.md
  - ../plans/2.0.x-plan26-spine.md
---

## Purpose

Transactional outbox for spine events. Guarantees that every committed
business write eventually produces a NATS event (at-least-once). Combined
with `pkg/spine.EnsureFirstDelivery` on the consumer side, achieves
effectively-once semantics without 2PC.

Introduced by Plan 26 Fase 3. The outbox table lives in each tenant DB
(migration `056_event_outbox`).

## Public API

Source: `pkg/outbox/outbox.go`, `pkg/outbox/worker.go`, `pkg/outbox/registry.go`

| Symbol | Kind | Description |
|--------|------|-------------|
| `PublishTx[T](ctx, tx, subject, env)` | func | Insert envelope into outbox within caller's tx |
| `Row` | struct | Single unpublished outbox entry |
| `DrainerWorker` | struct | Per-tenant background goroutine that polls + publishes |
| `NewDrainer(pool, nc, slug, opts...)` | func | Create a drainer for one tenant |
| `DrainerWorker.Run(ctx)` | method | Blocking drain loop (call as goroutine) |
| `WithPollIntervals(active, idle)` | func | Override poll timing |
| `DrainerRegistry` | struct | Manages drainers for all tenants |
| `NewRegistry(pool, nc, opts...)` | func | Create multi-tenant registry |
| `DrainerRegistry.Start(ctx, slugs)` | method | Bootstrap from known tenants (blocking) |
| `DrainerRegistry.AddTenant(ctx, slug)` | method | Hotload drainer for new tenant |
| `DrainerRegistry.RemoveTenant(slug)` | method | Stop drainer for removed tenant |
| `DrainerRegistry.ActiveTenants()` | method | List tenants with active drainers |

## Flow

```
Service handler:
  tx := pool.Begin(ctx)
  repo.InsertMessage(ctx, tx, msg)         // business write
  outbox.PublishTx(ctx, tx, subject, env)  // outbox insert (same tx)
  tx.Commit(ctx)                           // both succeed or both fail

DrainerWorker (background):
  SELECT ... FROM event_outbox WHERE published_at IS NULL FOR UPDATE SKIP LOCKED
  nc.PublishMsg(row.Subject, row.Payload)
  UPDATE event_outbox SET published_at = now()
```

## Concurrency

Multiple replicas poll the same tenant DB. `FOR UPDATE SKIP LOCKED` ensures
each row is claimed by exactly one replica. Ordering is best-effort (not
strict cross-replica). Consumers that need ordering use `Envelope.OccurredAt`.

## LISTEN/NOTIFY

The migration includes a trigger that `PERFORM pg_notify('spine_outbox_new')`
on INSERT. The DrainerWorker `LISTEN`s on the channel and wakes immediately
instead of waiting for the next poll cycle.

## Invariants

- `PublishTx` validates the subject before inserting.
- The outbox row's `id` is the envelope's `ID` (UUIDv7).
- Backoff for failed publishes: `2^attempts` seconds, capped at 60s.
- Dynamic batch size: `clamp(unpublished/100, 1, 1000)`.

## Importers

Fase 3 migrates chat, ingest, auth to use `outbox.PublishTx` + registry.
