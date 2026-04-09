---
name: Plan 15 BigBrother v2 re-audit (post-rewrite)
description: Third security audit of Plan 15 after major rewrite incorporating 35+ findings. 0 critical, 3 high, 5 medium, 3 low. APTO CON CONDICIONES -- first BigBrother APTO. Conditions are small fixes, not architectural blockers.
type: project
---

Plan 15 BigBrother v2 re-audit 2026-04-09. Plan was fully rewritten incorporating all prior findings.

Key findings (new in this round):
- H1: `config.ReadSecretFile()` referenced but does not exist in pkg/config/. Must be created in Fase 1.
- H2: `bb_pending_writes` partial unique index omits tenant_id, contradicting plan's own mandate.
- H3: Redis authentication gap -- all services use `redis:6379` without password but Redis requires one. With FailOpen=false, BB would be DoS'd.
- M1: Dockerfile shown twice (broken then fixed). Remove broken version.
- M2: pkg/plc/ importable by any service with no enforcement (convention only). Acceptable due to network isolation.
- M5: Notification service needs new handler code for BigBrother events (different schema than natspub.Event).

All 35+ prior findings verified as correctly fixed except H2 (missed one index).

**Why:** Third-pass audit validates the plan rewrite. The plan is now architecturally sound with only small implementation gaps.

**How to apply:** Conditions (H1, H2, H3 doc fix, M1) should be addressed inline during implementation. No phase restructuring needed. C6 (ExecuteConfirmed) prerequisite is correctly gated as Fase 0.
