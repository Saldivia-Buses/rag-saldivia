---
name: Plan 15 BigBrother post-fix audit (second pass)
description: Second security audit of Plan 15 (BigBrother network intelligence service) after first-round fixes applied. 3 critical, 5 high, 9 medium, 4 low findings. NOT APTO -- critical fixes needed before implementation.
type: project
---

Plan 15 BigBrother second-pass audit 2026-04-09 after first-round fixes (C1-C4, B1-B3, H1-H5, D1-D7, M1-M8, S1-S7, L3).

Key findings:
- C5: Docker Compose network config contradicts Security section (3 networks in compose vs 4 in spec, missing `frontend`, Traefik label references missing network)
- C6: ExecuteConfirmed() bug in agent is NOT just a BigBrother issue -- it affects ALL confirmed tools platform-wide, and plan only "notes" it instead of blocking on it
- C7: bb_pending_writes has no STATUS column and no concurrent request guard -- two admins can race to create pending writes for the same register simultaneously

Other critical gaps:
- 5 HIGH: tmpfs noexec missing, WinRM no cert validation, SFTP symlink traversal, in-memory rate limit for decrypt, audit_log missing tenant_id filter
- 9 MEDIUM: OPC-UA cert not persisted, libcap installed twice, nmap CVE pinning not enforced at runtime, Modbus TCP no auth, etc.

**Why:** Second-pass audit catches gaps in the FIXES themselves. The first audit found design holes; this one validates the patches.

**How to apply:** C5-C7 must be fixed in the plan text before implementation. All HIGH findings should have implementation notes added. MEDIUM can be addressed during implementation.
