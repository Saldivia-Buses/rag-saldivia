---
name: Plan 06 Phase 5 review
description: PR #69 Search Service tree-based retrieval, blockers: prompt injection (no guardrails wired), io.ReadAll OOM, no MaxBytesReader, broken Dockerfile
type: project
---

PR #69 reviewed 2026-04-05. Search Service at :8010.

Blockers:
- User query injected raw into LLM prompt without guardrails.ValidateInput (plan explicitly requires it)
- io.ReadAll without limit on LLM error body (repeat of PR #68 finding)
- No http.MaxBytesReader on handler (every other service has it)
- Dockerfile broken: COPY ../../pkg invalid, scratch base (needs distroless for CA certs)

Must fix:
- No middleware.Timeout
- No OpenTelemetry
- No query length limit
- maxNodes unbounded from user input
- go.mod missing golang.org/x/crypto indirect
- Plan says gRPC but impl is REST, missing 3 of 4 endpoints
- Handler does not log errors
- No Makefile targets

**Why:** Prompt injection is the primary security concern -- user text goes straight into LLM navigation prompt.
**How to apply:** When reviewing Phase 7 (Agent Runtime), verify guardrails are wired for all LLM-facing inputs, not just Search.
