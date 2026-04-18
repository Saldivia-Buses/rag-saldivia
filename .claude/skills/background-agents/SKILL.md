---
name: background-agents
description: Use when designing or implementing an always-on agent that runs detached from any user chat — mail ingestion, WhatsApp ingestion, memory curator, analytics trainers, etc. Owns Phase 3 of ADR 027 (background agents + data hoarding). These agents run as NATS consumers inside the app monolith with durable checkpoints and elevated capability scopes.
---

# background-agents

Scope: agents that run 24/7 without a user prompt. They ingest external
data (mail, WhatsApp), curate internal data (memories), train models,
or run periodic analytics. All live inside `services/app/internal/<domain>/`,
registered as goroutines in `services/app/cmd/main.go` startup.

## The three agents the product needs

### Mail agent (Phase 3)

Ingests all company mail.

- **Source:** IMAP against the company mail server (credentials in
  workstation env / Docker secret, never repo).
- **Cadence:** durable IMAP IDLE, poll fallback every 60s.
- **Storage:** `erp_mail` table (id, message-id, thread-id, from, to,
  date, subject, body_text, body_html, attachments), plus MinIO for
  attachments. Raw EML kept in MinIO for replay.
- **Dedup:** message-id unique index. Idempotent re-run.
- **Tree integration:** each mail is a tree root (subject is title,
  body is paragraphs), linked to its thread. The RAG ingest pipeline
  is reused (see `rag-pipeline`).
- **Collection:** mails get a collection based on sender/recipient
  area mapping; cross-area mail is multi-collection.

### WhatsApp agent (Phase 3)

Ingests messages from the company's internal WhatsApp numbers.

- **Source:** a WPPConnect (or similar) bridge per WhatsApp number.
  One container per number if more than one.
- **Storage:** `erp_whatsapp_messages` (id, wpp_message_id,
  chat_id, sender, sender_phone, date, text, media_path, message_type).
- **Media:** images/audio/video → MinIO + extracted transcript for
  audio (SGLang Whisper call).
- **Dedup:** wpp_message_id unique index.
- **Tree integration:** one tree per chat thread, day-grouped; chat
  messages are leaves. Retrieval surfaces conversation slices with
  date context.
- **PII note:** messages likely contain personal data of employees
  and customers. ACL by role + access log on every retrieval.

### Memory curator (Phase 2 / 3)

Reads chat transcripts, extracts high-signal facts, writes them to
`erp_memories_{global,user}`.

- **Source:** last N days of chat sessions + explicit "remember this"
  markers from users.
- **Cadence:** nightly batch + real-time for explicit markers.
- **Heuristics:** extract durable facts ("X likes reports grouped by
  week", "supplier Y always invoices on net-30"), avoid ephemeral
  ones ("the user asked about July 12").
- **Confidence:** LLM-scored; memories below threshold go to a
  pending queue for admin review.
- **Dedup:** similarity-hashed content; near-duplicates merge.
- **PII:** the curator NEVER surfaces memories across users unless
  the source confirmed it's a global fact.

## Common agent pattern

Every background agent follows this shape:

```go
type Agent struct {
    name      string
    consumer  nats.JetStreamConsumer   // durable
    store     Storer                   // persists + checkpoints
    policy    Policy                   // capability + rate limits
}

func (a *Agent) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-a.consumer.Messages():
            if err := a.handle(ctx, msg); err != nil {
                a.store.RecordFailure(ctx, msg, err)
                // ack after DLQ routing, don't redeliver forever
                continue
            }
            a.store.Checkpoint(ctx, msg.SeqNumber())
            msg.Ack()
        }
    }
}
```

Rules for every agent:

1. **Durable NATS consumer** — survives restarts; checkpoint is the
   NATS SeqNumber + an agent-owned cursor in Postgres (double safety).
2. **Idempotent handler** — same message → same DB state. Use
   content-hash or external-ID as unique key.
3. **Back-pressure** — if downstream (e.g., SGLang transcription)
   slows, don't pile up. NATS pull-consumer with small batch.
4. **Elevated capability scope** — each agent runs under a system
   role with a dedicated capability set (e.g., `agent.mail.write`).
   NOT a user role. The capability model in `agent-tools` has a
   parallel "system capabilities" registry for this.
5. **Audit every write** — agent writes are tagged `by_agent:<name>`
   for traceability.
6. **Rate-limited by policy** — a config table governs max
   writes/minute, max MinIO bytes/day. Agents that exceed pause.
7. **Observable** — metrics (processed, failed, lag),
   structured logs, trace spans wrapping each handle call.

## Where they run

Inside the app monolith (ADR 025). The startup:

```
services/app/cmd/main.go
  ├── wire HTTP router
  ├── wire NATS consumers (pre-existing: outbox, notifications, ws hub)
  └── wire background agents (new)
        └── mailAgent.Run(ctx)      as goroutine
        └── whatsappAgent.Run(ctx)  as goroutine
        └── memoryCurator.Run(ctx)  as goroutine
```

Configured on or off per deployment via `APP_AGENTS=mail,whatsapp,curator`
env. Off by default until each one is hardened.

## Checkpointing + replay

Every agent stores its position in both:

- NATS (durable consumer pointer, built-in).
- Postgres: `background_agent_state` (agent_name, cursor, last_success_at).

On start, the agent reads the lower-valued cursor (safer). On a 24h
outage + restart, the agent replays from the NATS durable cursor, which
Postgres confirms is safe (no ack lost).

## Tool use by agents

Agents can call tools too — the same tool registry as user-agents.
But:

- They dispatch with a **system user context**, not a human user.
- Their capability set is the agent's, not a human's.
- Their calls are tagged `system: true` in the audit log.
- They cannot call tools marked `user_only: true` (e.g., "ask the
  user a question").

## Failure modes + recovery

- **Dedup break** (same external ID, different content): detect via
  content hash mismatch, route to DLQ, page a human.
- **Source authentication drift** (IMAP password changed): agent
  fails on connect, logs `auth_failure`, stays paused, alert fires.
- **Storage full** (MinIO quota): agent pauses writes, NATS pull
  naturally buffers. Alert on sustained pause.
- **Hallucinated memory** (curator writes a plausible-but-wrong
  fact): admin review queue + periodic audit against source
  conversations.

## Testing

- **Integration test** per agent: start a real NATS + Postgres + a
  mock source (fake IMAP, fake WPPConnect), feed N messages, assert
  DB state + idempotency on replay.
- **Failure-injection test**: kill the agent mid-batch, restart, no
  duplicates + no losses.
- **ACL test**: messages from user A's area are NOT visible to user B
  without the capability.

## Integration with ADR 027

This skill owns the Phase 3 items: mail-agent-running, whatsapp-agent-
connected, memory-curator-scheduled, analytics-predictions. Each
agent's "done" is defined as: 24h continuous run + replay test passes
+ metrics + audit trail visible.

The memory-curator contributes to Phase 2 "memories" alongside
`prompt-layers`, which owns the retrieval / assembly side.

## Don't

- Don't run a background agent under a human user's context.
  Elevate via the system capability registry.
- Don't skip checkpointing — "it's fine, NATS remembers". When NATS
  fails or is reset, the double cursor saves you.
- Don't write unbounded logs of every message. Tag + sample; raw
  content goes to the DB, not the log stream.
- Don't grant global read across collections to the agent. Agents
  also obey per-collection ACL for any retrieval step.
- Don't add a new agent without a 24h replay test + an owner.
  Unowned background agents rot.
