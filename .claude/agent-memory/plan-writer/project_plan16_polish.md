---
name: Plan 16 backend polish
description: Backend polish plan for v2.0.5 — 25 audit findings (6C/4H/6M/9L), 6 phases, zero new features, written 2026-04-09
type: project
---

Plan 16 written 2026-04-09 targeting version 2.0.5. Addresses 25 findings from an exhaustive codebase audit.

**Why:** 6 critical production blockers (Docker builds failing, missing NATS users, missing env vars) prevent docker-compose.prod.yml from working end-to-end. Additional 19 findings cover test gaps, code duplication, and cleanup.

**Key decisions:**
- Fase 1: all 6 criticals at once (production blockers can't be incremental)
- Fase 3: new pkg/server package to eliminate ~50 lines of boilerplate per main.go
- pkg/build enhanced to read VERSION files as fallback (no more hardcoded version strings)
- pkg/audit full integration deferred to future plan (requires audit tables in each tenant DB)
- L6 (tools/ tests) out of scope (tools are early stage, would be churn)
- Zero new features — purely polish/fix/test

**How to apply:** When executing this plan, phase 1 must come first (everything else depends on Docker builds working). Phases 2-5 can be reordered if needed. Phase 6 cleanup comes last.
