---
title: Package: pkg/spine
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ../architecture/nats-events.md
  - ../plans/2.0.x-plan26-spine.md
  - ./nats.md
---

## Purpose

Event spine framework on top of `pkg/nats`. Wraps typed payloads in a canonical
`Envelope[T]` with UUIDv7 ID, schema version, trace context, and correlation
metadata. Consumers decode into a known T or `PeekHeader` to route by Type
before dispatching.

Introduced by Plan 26. Over time, `pkg/nats.Publisher` becomes the underlying
wire layer and callers interact via `pkg/spine`. See
`docs/plans/2.0.x-plan26-spine.md` for rollout phases.

## Public API

Source: `pkg/spine/envelope.go`, `pkg/spine/subject.go`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Envelope[T any]` | struct | Typed event wrapper (ID, TenantID, Type, SchemaVersion, OccurredAt, RecordedAt, Trace/CorrelationID, Payload) |
| `Header` | struct | Metadata-only subset, decodable without knowing T |
| `ErrSchemaVersion` | struct | Schema mismatch error with Type, Expected, Got |
| `New[T](tenantID, type, version, payload)` | func | Builds envelope with UUIDv7 ID + now() timestamps |
| `Encode[T](env)` | func | JSON serialize for NATS publish |
| `Decode[T](data)` | func | JSON deserialize + required-field validation |
| `PeekHeader(data)` | func | Route-by-Type without payload type commitment |
| `CheckSchemaVersion[T](env, expected)` | func | Fail-fast on schema mismatch |
| `BuildSubject(template, replacements)` | func | `{key}` interpolation + per-segment validation |
| `ValidateSubject(s)` | func | Canonical dot-separated subject check |

## Envelope shape

```go
type Envelope[T any] struct {
    ID            uuid.UUID
    TenantID      string
    Type          string
    SchemaVersion uint8
    OccurredAt    time.Time
    RecordedAt    time.Time
    TraceID       string     // omitempty
    SpanID        string     // omitempty
    CorrelationID string     // omitempty
    CausationID   *uuid.UUID // omitempty
    ActorUserID   *string    // omitempty
    Payload       T
}
```

`OccurredAt` is when the event happened (caller may override after `New`).
`RecordedAt` is when the envelope was built. Consumers ordering by OccurredAt
tolerate minor drainer reordering (see plan 26 § concurrency).

## Schema version policy

Monotonic `uint8`. Additive changes keep the version. Any breaking change bumps
`version + 1` and the old spec stays in `pkg/events/spec/<type>_v1.cue` until
`spine_consume_total{schema_version="1"} == 0` for ≥14 days. The CI lint
rejects a same-version diff of a spec. See `docs/conventions/cue.md`.

## Usage

```go
// Publish side
env, err := spine.New("saldivia", "chat.new_message", 1, payload)
if err != nil { return err }
data, _ := spine.Encode(env)
_ = nc.Publish("tenant.saldivia.notify.chat.new_message", data)

// Consume side (route by header, then decode)
h, _ := spine.PeekHeader(msg.Data)
if h.Type != "chat.new_message" { /* skip */ }
env, err := spine.Decode[ChatNewMessage](msg.Data)
if err != nil { return err }
if err := spine.CheckSchemaVersion(env, 1); err != nil { /* DLQ */ }
```

## Invariants

- `New` refuses empty `tenantID`, empty `eventType`, or `schemaVersion == 0`.
- `Decode` refuses missing `id`, `tenant_id`, `type`, or `schema_version == 0`
  to catch silent decode-failed-into-zero-value bugs.
- `BuildSubject` validates every placeholder value with
  `natspub.IsValidSubjectToken` — no dots or spaces smuggle in via substitution.
- IDs are UUIDv7 (time-ordered) so `event_outbox` and `processed_events` rely
  on B-tree scans by ID for ordering and TTL cleanup.

## Importers

Empty on merge (Fase 1 introduces the package). Fase 2 adds the consumer
framework (`Consume`, middleware, DLQ push). Fase 3 migrates chat/ingest/auth
to publish via `outbox.PublishTx` with a spine envelope. `pkg/nats.Publisher`
stays the wire layer and gains an `Envelope[Event]` internal wrapper.
