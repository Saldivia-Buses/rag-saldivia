---
name: PR #78 Service Wiring Expansion review
description: Agent TracePublisher has hardcoded "default" tenant (isolation broken), no NATS validation, dead TraceEvent; Chat guardrails allow system role unfiltered
type: project
---

PR #78 addresses P0/P1 from a service analysis audit. Three blocker findings:

1. Agent TracePublisher hardcodes tenant slug as "default" -- all tenants' traces merge into one bucket, breaking tenant isolation and cost attribution.
2. TracePublisher has no NATS subject injection validation (unlike pkg/nats/publisher.go which has isValidSubjectToken).
3. TraceEnd not published on pending_confirmation path -- orphaned "running" traces.

Must-fix: TraceEvent is dead code (never called from agent loop), chat guardrails let `system` role through without validation (prompt injection vector), guardrails config is hardcoded inline.

**Why:** Tenant isolation is the #1 security invariant; hardcoding "default" bypasses it entirely. The NATS injection pattern has been flagged in PRs #66, #68, #70, #71 already.

**How to apply:** When agent trace publishing is next touched, verify tenant slug comes from context (not hardcoded), validate it before building NATS subjects, and wire TraceEvent calls into the agent loop.
