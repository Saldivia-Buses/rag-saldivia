---
name: Plan 04 Hardening Review
description: PR #77 hardening fixes review -- C2 tenant validation incomplete (traces.end/event missing), search SQL has no tenant_id filtering
type: project
---

PR #77 hardening fixes for security + observability audit findings. Reviewed 2026-04-05.

**Blocker:** C2 NATS tenant validation only applied to `traces.start` subscriber, not to `traces.end` or `traces.event`. `TraceEndEvent` and `TraceEvent` lack `TenantID` field. `RecordTraceEnd` UPDATE has no `tenant_id` in WHERE clause.

**Must-fix:** Search service C1 fix reads tenant from context but never passes it to SQL queries. No `tenant_id` filtering in `loadTrees` or `extractPages`. Also missing `RequirePermission` middleware on search handler.

**Why:** Traces service uses platform DB (shared across tenants), so SQL-level tenant filtering is defense-in-depth. Search service relies on per-tenant DB pool, but has no SQL-level guard.

**How to apply:** When reviewing traces/search services in future PRs, verify these gaps were closed.
