---
name: Plan 07 — Consolidation review
description: Plan 07 is a pure refactoring plan covering shared schema, pkg/config, pkg/llm dedup, NATS standardization, token blacklist wiring, cmd/main.go dedup. Reviewed 2026-04-05 with conditional approval — 3 blockers found.
type: project
---

Plan 07 reviewed 2026-04-05 with CONDITIONAL APPROVAL.

**Why:** Post-Plan-06 audit found 10 copies of env(), 9 copies of loadPublicKey(), 3 LLM clients, 3 NATS validation patterns, blacklist not wired, config package empty.

**Blockers found:**
1. Phase 7 (multimodal wiring) is a feature, not refactoring — needs split
2. Migration renumbering without state tracking risks double-apply on existing DBs
3. `pkg/services/` cleanup is a false positive — directory doesn't exist

**How to apply:** When executing Plan 07, address blockers first. Phase 3 can be parallelized by splitting into 3a (extract pkg/llm) and 3b (add NewFromSlot). Tests must be added to every phase.
