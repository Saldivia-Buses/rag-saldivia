---
title: Runbook: Spine Event Bus
audience: ai
last_reviewed: 2026-04-16
related:
  - ../architecture/spine.md
  - ../packages/spine.md
  - ../packages/outbox.md
---

## Scenario 1: NATS down — events accumulating in outbox

**Symptoms:** outbox rows accumulating (check via
`SELECT count(*) FROM event_outbox WHERE published_at IS NULL` per tenant DB),
services healthy, NATS unreachable.

**Steps:**
1. Check NATS: `docker ps | grep nats` or health endpoint.
2. If NATS crashed: `docker restart nats`. Drainers reconnect automatically
   (pkg/nats.Connect uses `MaxReconnects(-1)`).
3. If network: check Traefik/firewall rules between services and NATS.
4. Monitor outbox drain: `SELECT count(*) FROM event_outbox WHERE published_at IS NULL`
   per tenant DB. Should decrease after NATS returns.
5. If backlog >10k: drainer scales batch dynamically. No manual action needed.

**Resolution:** NATS recovery → drainer publishes backlog → consumers process
with idempotency. Zero manual intervention needed.

## Scenario 2: Consumer panics repeatedly

**Symptoms:** `SpineHandlerPanic` alert firing, `spine_consume_total{result="panic"}`
increasing.

**Steps:**
1. Check logs: `docker logs sda-<service> | grep "spine: handler panic"`.
2. Stack trace is logged with the panic — identify the offending handler.
3. Fix the panic (nil pointer, index OOB, etc.) and deploy.
4. While the fix is deploying: messages are Nak'd with backoff. After
   `MaxDeliver` attempts they go to DLQ. No data loss.
5. After deploy: any messages in DLQ from the panic can be replayed via
   `/admin/dlq`.

## Scenario 3: Dead event in DLQ — replay or drop

**Symptoms:** Event visible in `/admin/dlq`.

**Steps:**
1. Open `/admin/dlq`, find the event. Check the error reason.
2. If transient (NATS timeout, DB lock): click **Replay**. The event is
   re-published with a new event_id. The consumer processes it against
   current state.
3. If permanent (schema mismatch, invalid data): click **Drop**. The event
   is marked `dropped_at` and hidden from the list.
4. If unsure: check the full envelope JSON in the detail page. Compare
   with the current schema in `pkg/events/spec/*.cue`.

**Replay semantics:**
- Replay = re-execute against current state. NOT time-travel.
- New event_id is generated. Causation_id links to original.
- If replayed 3+ times, the UI shows a warning.

## Scenario 4: Schema version mismatch

**Symptoms:** `spine_consume_total{result="decode_error"}` increasing.
Messages are Term'd (not retried).

**Steps:**
1. Check which consumer + event type is affected (from metric labels).
2. Verify the CUE spec version matches the producer's schema_version.
3. If producer was updated without bumping version: **this is a bug**.
   Fix the producer to use the correct version.
4. If a new version was deployed: ensure the consumer is also updated
   to handle the new version. Deploy consumer first, then producer.

## Scenario 5: Outbox drainer stalled for one tenant

**Symptoms:** One tenant's events not arriving, others fine.

**Steps:**
1. Check drainer logs: `docker logs <service> | grep "outbox-drainer"`.
2. If pool error: tenant DB may be down. Check tenant DB connectivity.
3. If the drainer goroutine died silently (no logs after "drainer started"):
   restart the service. The drainer restarts automatically.
4. In multi-tenant auth: check `registry.ActiveTenants()` via health
   endpoint. If the tenant is missing, it wasn't in the bootstrap list
   and no `platform.lifecycle.tenant_created` event was received.
