---
name: PR #84 pkg/llm unified client
description: Plan 07 Phase 3 review -- pkg/llm extracts 3 duplicated LLM clients into one. Blockers: ingest go.mod missing pkg dep (Docker build fail), plan verification criterion not met (internal/llm shims still exist).
type: project
---

PR #84 creates `pkg/llm/client.go` as the single OpenAI-compatible HTTP client for all services. Agent and ingest use re-export shims; search imports `pkg/llm` directly.

Blockers:
1. `services/ingest/go.mod` missing `replace` and `require` for `pkg` -- compiles via go.work locally but will fail in Docker/CI.
2. Plan verification says `grep -r "internal/llm" services/` should return 0 results, but 4 files still import internal/llm through shims. Plan intent is full migration, not shims.

Must-fix: SimplePrompt hardcodes max_tokens=4096 (wasteful for search, should be a parameter), dead comment at search.go:328, response body not drained after decode (hurts connection reuse for cloud providers).

Design note: Plan specified functional options (`ChatOption`) but PR uses positional params. Acceptable for now but will need breaking change when adding tool_choice/response_format/stream.

**Why:** This is the foundation for all LLM calls in the system. Getting the API surface right now prevents churn later.

**How to apply:** When reviewing future LLM-related PRs, verify services import `pkg/llm` directly (not through internal shims). Verify ingest go.mod has the pkg dependency.
