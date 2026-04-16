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
| `Consume[T](ctx, nc, js, pool, cfg, handler)` | func | JetStream pull consumer with full middleware chain |
| `TypedHandler[T]` | type | `func(ctx, pgx.Tx, Envelope[T]) error` — runs in handler tx |
| `ConsumerConfig` | struct | Durable, Stream, Subject, MaxDeliver, backoff params |
| `TenantPool` | interface | `PostgresPool(ctx, slug)` — tenant-scoped pool resolver |
| `SubjectSlug(subject)` | func | Extract tenant slug from subject |
| `ValidateTenantMatch(subject, header)` | func | Subject slug == envelope TenantID |
| `EnsureFirstDelivery(ctx, tx, eventID, consumer)` | func | Atomic dedup in handler tx |
| `ExtractTraceContext(ctx, header, natsHeaders)` | func | OTel from envelope or headers |
| `Backoff(attempt, base, max)` | func | Exponential 2^n * base, capped at max |
| `PushDLQ(ctx, nc, entry)` | func | Structured DLQ publish |
| `DLQEntry` | struct | Original envelope + metadata for healthwatch |
| `PublishTotal`, `ConsumeTotal`, ... | metrics | OTel counters + histograms |

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

## Consumer middleware chain (Fase 2)

`Consume[T]` creates a JetStream durable pull consumer that runs these
steps per message (in order):

1. Panic recovery → Nak + backoff, slog.Error with stack trace
2. MaxDeliver check → PushDLQ + Term if exceeded
3. PeekHeader → extract tenant slug + Type
4. ValidateTenantMatch → Term on mismatch
5. Decode envelope as `Envelope[T]`
6. ExtractTraceContext from envelope TraceID+SpanID (or NATS headers)
7. Get tenant pool → Nak on error (transient)
8. Begin tx → EnsureFirstDelivery → Ack+skip if duplicate
9. Run handler in tx
10. Commit → Ack on success, Nak on error

## Importers

Fase 2: framework only (no service uses it yet — notification retrofit
deferred to when producers emit envelopes in Fase 3). Fase 3 migrates
chat/ingest/auth to `outbox.PublishTx` and switches notification consumer
to `spine.Consume`.
