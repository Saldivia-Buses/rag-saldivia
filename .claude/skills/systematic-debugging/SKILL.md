---
name: systematic-debugging
description: Use when anything is unexpected — a test fails, a service misbehaves, a deploy breaks, output doesn't match. Enforces root cause FIRST, fix SECOND. No proposing fixes without understanding why. Covers the Go + Docker + NATS + Postgres debugging loop, how to reproduce deterministically, and how to prove the fix actually fixes it.
---

# systematic-debugging

Scope: any unexpected behavior — failing test, panic, 500 response, stuck NATS
consumer, missing event, slow query, container that won't start.

## The rule

**Root cause first. Fix second.**

Not "I'll try this and see". Not "probably the retry logic". Understand what is
happening before changing anything. The cost of a wrong fix is a silent recurrence.

## The loop

1. **Reproduce** deterministically.
   - If it happens once in ten runs, find the condition that makes it happen every time.
   - If you can't reproduce, you don't understand it, and your fix is a guess.
2. **Observe** the minimum signal needed.
   - Logs at the boundary (entry + exit of the failing function).
   - DB state before and after.
   - NATS subject actually published (not what you think was published).
3. **Form one hypothesis** that explains **all** the signals.
   - Not "the DB is slow"; that's a category. Be specific: "the query for X hits a
     seq scan because the index was dropped in migration 0041".
4. **Verify** the hypothesis before fixing.
   - If the hypothesis is "X is nil", prove it with a log line or a test.
5. **Fix** only what the hypothesis identifies.
   - No drive-by cleanups during debugging. They hide what the fix actually did.
6. **Prove** the fix works.
   - The reproducer now passes.
   - Add it as a regression test.

## Go failures

- `go test -v -run TestX ./...` — reproduce. No `-race`? Add it.
- Panic with stack? Read the **first** frame inside your code, not the deepest.
  The deepest is usually runtime.
- "unexpected EOF", "connection refused" — which side closed first? Add a log on
  the other side.
- "context deadline exceeded" — which ctx? The request's, or a derived one? Log
  the deadline at entry.

## Docker / container failures

- Container exits immediately: `docker logs <container>` — first 50 lines, not last.
- Container won't start: `docker inspect` → find `ExitCode` and `Error` fields.
- Networking: exec into the container, `wget -O- http://<service>:<port>/healthz`.
- Volume mount: `docker exec <c> ls <mountpoint>` — is the file even there?

## NATS issues

- "Event published but no consumer reacted":
  - Subscribe live to the subject from a shell: `nats sub 'tenant.*.>'`. Do you see it?
  - If yes, the consumer isn't bound. Check subject string exactly — typos kill silently.
  - If no, the publish didn't happen. Log the publisher.
- "Consumer stuck": check `nats consumer info <stream> <consumer>`. Pending messages?
  Ack wait expired?

## Postgres issues

- "query is slow": `EXPLAIN (ANALYZE, BUFFERS) <query>`. Seq scan on a tenant table =
  missing index on `tenant_id + ...`.
- "deadlock detected": log both transactions' `pg_stat_activity` at the moment.
  Usually two updates in opposite order.
- "connection refused": is the pool exhausted? `SELECT count(*) FROM pg_stat_activity;`.

## When you are stuck

If 30 minutes pass without a hypothesis that explains all the signals:

1. **Write down** what you know and what you don't. Literally write it.
2. Dispatch the `parallel-research` flow — one Explore for call sites, one for
   tests, one for recent commits that touched the area.
3. Check `git log --oneline -20 -- <file>` — did someone change this recently?
4. Ask.

## Don't

- Don't add retries to mask a bug.
- Don't add `time.Sleep` to paper over a race.
- Don't catch-all errors during debugging; you need them loud.
- Don't "clean up" unrelated code in the same commit as the fix.
- Don't declare it fixed without a regression test.

## Proof of fix

When you claim the bug is fixed, the commit must contain:

- The minimal change that addresses the root cause.
- A test that fails without the change and passes with it.
- A commit message that states **what was broken**, **why**, **how the fix works**.

## Known patterns in this repo (check here first)

When a Go service misbehaves, these are the recurring root-cause shapes the
audit found. Ruling them in or out early saves time.

### `context.Background()` inside a handler path

If a handler calls `context.Background()` instead of propagating the request's
`ctx`, shutdown signals and deadlines silently die at that call site.

```bash
rg -n "context\.Background\(\)" services/*/internal/ -g '!*_test.go'
```

Audited offenders: 8 files in production code. Symptom: services hang on
shutdown, timeouts don't fire, traces are missing a span.

### Goroutine without stop mechanism

Fire-and-forget `go func()` leaks on shutdown and on reconfig. Symptom: RSS
climbs over days, old versions of a ticker keep firing after a restart,
goroutine count visible in pprof never goes down.

```bash
rg -nB1 "go func\(" services/ -g '!*_test.go'
```

Audited offenders: `services/healthwatch/internal/service/healthwatch.go:280`,
`services/feedback/internal/service/aggregator.go:43`,
`services/chat/cmd/main.go:98`,
`services/ingest/internal/service/extractor_consumer.go:83`,
`services/erp/internal/handler/analytics.go:868`.

### `panic()` in a request path

A handler or service method panicking takes the whole request down (and with
the recoverer middleware, produces a 500 with no useful error body). The fix is
always the same: return an `httperr` typed error.

```bash
rg -n "panic\(" services/ -g '!*_test.go'
```

Audited offenders: `services/ingest/internal/service/documents.go:38`,
`services/bigbrother/internal/scanner/stub.go:34`.

### Tenant-residual as a debugging clue

If a service behaves inconsistently between "feels like it works locally" and
"breaks in the deploy", double-check whether it still reads `tenant_slug` or
`tenant_id` from somewhere (env, JWT claim, middleware). After ADR 022 that
plumbing no longer has a live source of truth — it silently reads zero values.

```bash
rg -n "tenant_slug\|TenantSlug\|tenant_id" services/ pkg/
```

2,480+ hits at audit time. Any of them can produce a stealthy "X is empty / zero"
bug that looks like a config error but is an ADR-022 violation.

### NATS subject mismatch

Publisher uses flat subject, subscriber still listens on old `tenant.{slug}.*`
(or vice versa). Symptom: event is published, nothing reacts, no error.
Diagnostic: subscribe live with `nats sub '>'` and watch what actually lands.
Fix: align both sides on the flat form (ADR 022).
