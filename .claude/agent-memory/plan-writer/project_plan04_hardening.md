---
name: Plan 04 Hardening context
description: Context behind plan 04 — hardening & production readiness, 8 phases, based on security audit of 2026-04-03
type: project
---

Plan 04 was written based on a comprehensive audit that scored the system at ~40% of spec. Security scored 4/10, Infra 4/10.

Key decisions:
- Ed25519 over RS256 for JWT — smaller tokens, faster verify, recommended by NIST
- Permissions embedded in JWT claims (not DB lookup per request) — accepted 15-min staleness for v1
- MFA uses TOTP (pquerna/otp library), not WebAuthn — simpler, mobile-friendly
- Docker secrets for prod (not Vault) — intermediate step, Vault is Plan 06
- sqlc migration and repository layer refactor deliberately excluded — separate plan (05)
- Network segmentation, mTLS, CrowdSec excluded — separate plan (06)

**Why:** Enzo's top 5 priorities were RBAC enforcement, docker-compose.prod, MFA, audit logging, deploy pipeline fix.

**How to apply:** When executing this plan, follow phase order strictly (JWT first because RBAC depends on permissions in claims, MFA depends on asymmetric JWT). Phases 5 and 8 are independent and can be parallelized.
