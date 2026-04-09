---
name: Plan 15 v2 BigBrother Review
description: Re-review of Plan 15 after 35+ findings incorporated. New blockers found: ReadSecretFile ghost function, partial unique index missing tenant_id, Dockerfile contradictions, dev Traefik missing.
type: project
---

Plan 15 v2 reviewed 2026-04-09. Verdict: APTO CON CONDICIONES.

Key new findings:
- `config.ReadSecretFile()` referenced but does not exist in codebase (B-NEW1)
- `idx_bb_pending_active` partial unique index omits tenant_id, contradicts H-GW3 fix (H-NEW1)
- `bb_pending_writes` cleanup query omits tenant_id WHERE clause (H-NEW2)
- Dockerfile has BOTH wrong AND correct version, confusing for implementer (M-NEW1)
- No dev Traefik route for BigBrother port 8012 (M-NEW2)
- No `go.work` entry for new pkg/* since they live inside single `./pkg` module (info — not a bug)
- ExecuteConfirmed bug still unfixed in codebase (pre-existing, plan correctly marks as Phase 0 prerequisite)

**Why:** These are genuine gaps that survived the 35+ finding audit integration.
**How to apply:** Block implementation until ReadSecretFile is either created or the plan switches to the existing pattern (env var). The tenant_id omissions in indexes/queries must be fixed in the plan text.
