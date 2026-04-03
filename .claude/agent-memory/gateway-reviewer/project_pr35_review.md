---
name: PR #35 review patterns
description: Second gateway review -- shared packages (middleware, NATS publisher) and event wiring in auth+chat services. Key issues found.
type: project
---

PR #35 reviewed 2026-04-02. Shared packages + NATS event publishing.

Key findings to watch for in future PRs:
- Chat handler has ownership gaps: GetMessages and AddMessage do not verify session ownership (B3). Pattern to check in all CRUD handlers.
- NATS subject injection risk via tenant slug string concatenation. Need slug validation before it hits subject construction.
- `RetryOnFailedConnect(true)` makes `nats.Connect()` never return errors -- error branch is dead code in auth/chat main.go. WS Hub handles this correctly with disconnect/reconnect handlers.
- `truncate()` in chat slices bytes not runes -- will corrupt multi-byte UTF-8.
- Chat publishes `user_id: ""` in notification events, which causes Notification consumer to Term() the message immediately.

**Why:** These patterns (ownership checks, NATS subject safety, UTF-8 handling) will recur as more services are added.
**How to apply:** Check every new handler for ownership verification, every NATS subject for injection, every string operation for UTF-8 safety.
