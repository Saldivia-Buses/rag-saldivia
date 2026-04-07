---
name: PR #90 Plan 08 Phase 4a NATS Standardization
description: Mechanical refactor migrating all services to natspub.Connect() + Drain() + consumer context fix. APPROVED with minor observation.
type: project
---

PR #90 — Plan 08 Phase 4a, NATS standardization. APPROVED with minor observation (no blockers).

**Changes verified:**
- All 9 services use `natspub.Connect()` — confirmed line by line
- 6 services use `defer nc.Drain()` — auth, ws, chat, ingest, feedback, notification, traces; agent and platform use conditional Drain (correct, since NATS is optional for those two)
- `notification/service/consumer.go` and `feedback/service/consumer.go` use `c.ctx` struct field instead of `context.Background()` — enables clean shutdown
- `feedback/service/aggregator.go` uses `a.ctx` struct field — same benefit
- `agent.Query()` now accepts `userID` parameter — handler reads from `X-User-ID` header (post-middleware, safe). Only one caller exists.
- `nats.go` import removed from agent and platform (only those two; ws, notification, auth, chat, ingest, traces, feedback legitimately retain it for `nats.DefaultURL` or `nats.StreamConfig`)

**Observation (not blocking):**
`platform/cmd/main.go` calls `natspub.New(nc)` even when `nc` may be nil (NATS failure is Warn, not Exit). `natspub.New` stores nil nc without panicking, but any actual `Publish` call would panic at `p.nc.Publish()`. This was pre-existing before PR #90. Future hardening: add nil guard in Publisher methods or no-op Publisher pattern.

**Why:** `natspub.Connect()` centralizes reconnect options (RetryOnFailedConnect, MaxReconnects(-1), ReconnectWait 2s, disconnect/reconnect slog handlers). Previously each service called `nats.Connect()` with no options — silent disconnects would not retry.
