---
name: Plan 06 Phase 3 review
description: PR #67 pkg/guardrails + Tenant/Platform DB migrations + config seed. Blockers: UTF-8 truncation, case-sensitive leak detection, incomplete JSON schema validation. Must-fix: missing CHECK constraints, missing seed keys, no ingest_jobs data migration.
type: project
---

PR #67 reviewed 2026-04-05. Plan 06 Phase 3: pkg/guardrails + DB migrations + config seed.

**Result:** CAMBIOS REQUERIDOS (3 blockers, 6 must-fix, 6 suggestions)

**Blockers:**
- ValidateInput truncates by bytes, not runes -- breaks multi-byte UTF-8
- ValidateOutput system prompt leak detection is case-sensitive exact-match only
- ValidateToolParams only checks required fields, no type/enum validation (plan promises full JSON schema validation)

**Must-fix:**
- Platform DB tables missing CHECK constraints (execution_traces.status, tool_registry.type/protocol, llm_models.location)
- Seed missing tools.enabled, slot temperatures, fallback_chain, classifier_prompt keys
- CompilePatterns exported but unused/untested -- remove or integrate
- No data migration from ingest_jobs to documents (plan says "migrar si hay data")
- Platform down migration uses overly broad WHERE clause (updated_by = 'system')
- tool_calls.status missing CHECK constraint

**Why:** This is the DB foundation for all subsequent Plan 06 phases. Schema gaps (missing CHECKs) will be painful to fix later with data in place. Guardrails are imported by Agent Runtime, Search, and Ingest -- getting the API right now prevents breaking changes.

**How to apply:** When reviewing Phase 4+ PRs, verify they use the guardrails properly (rune-based truncation fix landed, JSON schema validation works against real tool_registry schemas). Verify config resolution works against the seeded keys.
