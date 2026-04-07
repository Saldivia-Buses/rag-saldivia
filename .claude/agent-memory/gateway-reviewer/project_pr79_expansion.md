---
name: PR #79 Service Expansion P1/P2 review
description: ExtractorConsumer tenant isolation broken (single pool for all tenants), NATS type injection in platform lifecycle, DB errors swallowed in consumer
type: project
---

PR #79 feat/service-expansion-p1p2 reviewed 2026-04-05. CAMBIOS REQUERIDOS.

**Why:** Three blocker-level issues found.

1. ExtractorConsumer uses single pgxpool.Pool but subscribes to tenant.*.extractor.result.> -- writes all tenants' data to one DB. Needs tenant.Resolver per-message.
2. Platform publishLifecycleEvent passes dotted event types ("platform.tenant.created") into Notify which doesn't validate parsed.Type -- recurring NATS subject injection (same root cause as PR #52).
3. ExtractorConsumer ignores all DB errors and marks docs "ready" even on partial failure.

Also: consumer not wired in main.go, Notify panics on nil NATS conn from platform, tree gen failure still marks "ready".

**How to apply:** These patterns recur -- always check: (a) NATS consumers that wildcard on tenant must resolve per-tenant pools, (b) any string interpolated into NATS subjects needs validation, (c) DB operation errors must be checked.
