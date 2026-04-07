---
name: Plan 06 Phase 7 review
description: PR #71 Agent Runtime review -- blockers: history injection (prompt injection bypass), port collision 8004, tool error leak to LLM context, no tool param validation, empty output guardrails
type: project
---

PR #71 Agent Runtime (services/agent/) reviewed 2026-04-05. Result: BLOQUEADO.

5 blockers:
1. History injection: client can inject system/tool role messages via `history` field, bypassing all guardrails
2. Port collision: agent defaults to 8004 (same as RAG), needs 8011
3. Tool error bodies (up to 1MB) leak to LLM context and then to user
4. `guardrails.ValidateToolParams()` exists but is never called -- LLM-generated params go unchecked
5. Output guardrails pass empty config, so ValidateOutput is a no-op -- system prompt can leak

**Why:** This is the most security-critical service (handles user input -> LLM -> tools). History injection is the worst because it completely bypasses input guardrails.

**How to apply:** When reviewing fix PR, verify (1) history messages are filtered to user/assistant only OR loaded server-side, (2) port changed to 8011, (3) error bodies truncated/redacted before LLM context, (4) ValidateToolParams called in executor, (5) system prompt fragments passed to OutputConfig.
