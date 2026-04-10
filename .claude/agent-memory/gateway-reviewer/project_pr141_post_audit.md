---
name: PR #141 Post-audit ERP fixes
description: Post-audit fixes for Plan 17 ERP -- GetCheck, cash count diff, encrypted field removal, StrictLogger, audit completeness
type: project
---

PR #141 reviewed 2026-04-10. Post-audit fixes addressing findings from 4-auditor comprehensive ERP review.

**Why:** Audit of complete Plan 17 ERP found ~10 issues including missing GetCheck query, hardcoded cash count diff, encrypted fields leaking, missing audit on sub-entity operations, missing StrictLogger for Allocate.

**How to apply:** APPROVED with 1 medium finding (float64 precision for cash count difference calculation). All PR#138 blockers (trigger crash, UpdateMovementBalance :exec) are confirmed fixed. Encrypted fields removed from all query SELECT/INSERT/UPDATE/RETURNING. StrictLogger correctly wired for CurrentAccounts and Treasury.
