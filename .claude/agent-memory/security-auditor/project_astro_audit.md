---
name: Astro service security audit
description: Full audit of services/astro/ v0.1.0 -- APTO. 0 critical, 0 high, 3 medium, 6 low. Ed25519 JWT, tenant isolation correct, CGO safe.
type: project
---

Astro service v0.1.0 audited 2026-04-06. Verdict: APTO for production.

Key findings:
- M1: FailOpen:true on auth blacklist (consistent with other non-critical services)
- M2: No year range validation on /query SSE endpoint (parseRequest has it, Query does not)
- M3: No input validation on CreateContact body (name, lat/lon, etc)
- L1: contact_name reflected in error messages
- L3: LLM prompt injection possible but low impact (no tools, single-user scope)

**Why:** First new service post-rewrite, sets the bar for security posture of modular services.
**How to apply:** For future astro PRs, verify M2/M3 fixes were applied. For other new services, check same patterns (FailOpen, year/input validation, error message leakage).
