---
name: Plan 06 Intelligence Layer Audit
description: Full audit of agent/search/traces/extractor services -- 2 critical (search no tenant isolation, traces NATS payload trust), 5 high. NOT APTO for Plan 06 services.
type: project
---

Plan 06 security audit completed 2026-04-05. Result: NOT APTO for Plan 06 services.

**Why:** 2 critical tenant isolation failures, 5 high-severity gaps in the 4 new services.

**How to apply:**
- C1 (search tenant isolation) and C2 (traces NATS payload trust) are deploy blockers -- no exceptions
- H3 (missing prod compose) must be resolved before deploy is even possible
- Core services (auth, ws, chat, notification, platform, ingest, feedback) passed -- APTO independently
- Full report: docs/artifacts/plan06-security-audit.md
