---
name: Plan 06 Phase 4 review
description: PR #68 tree generation pipeline + document service. Blockers: dead code in Put, storage_key never written to DB, unbounded io.ReadAll, NATS subject injection, context-unaware semaphore. Must-fix: dedup cross-user leak, ignored DB errors, UTF-8 truncation (same class as Phase 3).
type: project
---

PR #68 reviewed 2026-04-05. Plan 06 Phase 4: tree generation + document service in Ingest.

**Result:** CAMBIOS REQUERIDOS (5 blockers, 10 must-fix, 8 suggestions)

**Blockers:**
- documents.go has dead code: spurious Put with nil body before real Put
- storage_key never written back to DB after document creation (empty string persisted)
- Unbounded io.ReadAll loads entire file (up to 100MB) into memory for hashing
- NATS subject injection: tenant slug not validated (same class as PRs #52, #55, #66)
- generateSummaries semaphore does not check ctx.Done() (goroutines hang on cancelled context)

**Key patterns:**
- NATS subject injection keeps recurring -- any code using raw `nats.Conn.Publish` instead of `pkg/nats.Publisher` is vulnerable. Need to export `IsValidSubjectToken` for use outside the publisher.
- UTF-8 byte-boundary truncation keeps recurring (Phase 3 guardrails, Phase 4 summary generation). Need a shared `pkg/strings` utility.
- Pipeline config (max_depth, max_nodes_per_level, chunk_size_pages) from plan is all hardcoded. Acceptable for v1 but needs wiring before Phase 5.

**Why:** Tree generation is the core of the RAG pipeline. Correctness in the tree structure, robustness against LLM garbage, and proper file storage are all critical for Phase 5 (Search Service) to work.

**How to apply:** When reviewing Phase 5, verify: (1) tree navigation handles flat fallback trees gracefully, (2) page extraction uses storage_key from DB (needs B2 fix), (3) config values come from pipeline.indexing, not hardcoded.
