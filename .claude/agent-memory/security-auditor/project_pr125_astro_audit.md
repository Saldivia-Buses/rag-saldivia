---
name: PR 125 Astro service audit
description: Security audit of services/astro/ — CGO ephemeris, contacts CRUD, LLM narration. 1 critical (FailOpen), 3 high, 5 medium. APTO CON CONDICIONES.
type: project
---

PR #125 introduces `services/astro/` — astrological calculation engine with CGO (Swiss Ephemeris).
Audited 2026-04-07.

**Findings:** 1 critical (FailOpen bypass), 3 high (prompt injection, single tenant DB pool, no RBAC), 5 medium, 3 low, 2 info.

**Verdict:** APTO CON CONDICIONES — Critical and Highs must be addressed before production deploy.

**Why:** The critical finding is a pattern consistent across all services (FailOpen: true with blacklist),
not unique to astro. The astro-specific issues (prompt injection, single DB pool) are high but
the service currently has limited exposure (astrological queries, not core business data).

**How to apply:** Track fix status for FailOpen (framework-wide), prompt injection (astro-specific),
and RBAC permissions (astro-specific). DB pool architecture needs a decision for the whole platform.
