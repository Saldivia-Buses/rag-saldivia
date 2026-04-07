---
name: PR #89 DB & Query Hardening
description: Plan 08 Phase 3 review -- pagination pkg, LATERAL JOIN, migration 009, blockers: int32 overflow, stale sqlc models, unused SetHeaders
type: project
---

PR #89 (feat/plan08-phase-3) reviewed 2026-04-05. Result: CAMBIOS REQUERIDOS.

Blockers:
1. pagination.Offset() int32 cast overflows for large page numbers -> negative OFFSET error
2. sqlc models stale in auth/chat/ingest (missing LatencyMs field from migration 009)

Must-fix:
- SetHeaders() defined but never called (plan required X-Total-Count)
- ListMessages uses pagination.Parse but ignores page param (plan wanted cursor-based)
- PerformancePercentiles still does JSONB extraction instead of using new latency_ms generated column
- Generated column unsafe for non-numeric latency_ms values

Not implemented from plan Fase 3: M10 (batch insert), L5 (DeleteJobByID ownership), L10 (category constants).

**Why:** DB hardening for multi-user scale. Pagination + query optimization.
**How to apply:** Check follow-up PR for these fixes. Verify sqlc regen across all services after migration changes.
