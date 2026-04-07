---
name: Plan 08 post-hardening audit
description: POST-implementation audit of plan 08 (52 fixes verified). First APTO verdict. 0 critical, 3 high, 6 medium, 4 low.
type: project
---

Plan 08 post-implementation audit completed 2026-04-05. All 52 hardening fixes verified as implemented.

**Verdict: APTO for controlled production** (first APTO in project history).

Key findings remaining:
- H1: Non-auth services do not check token blacklist (15-min window for revoked tokens)
- H2: Platform requirePlatformAdmin does not strip spoofed X-User-ID before setting it
- H3: Dev Traefik injects hardcoded tenant headers (dev-only risk)
- M1: NATS falls back to unauthenticated localhost if env var missing
- M2: ListMessages SQL lacks user_id (handler checks, SQL does not)
- M3: Document queries lack tenant_id column (per-tenant-DB makes this safe)
- M4: Feedback queries not user-scoped or role-gated
- M5: ensureSSLMode silently proceeds on parse failure
- M6: TOTP MFA secrets stored unencrypted (pkg/crypto exists, wiring missing)

What was fixed and verified:
- Ed25519 JWT (not HS256), JTI auto-gen, blacklist FailOpen:false in Auth
- Rate limiting: login 5/min, refresh 10/min, MFA 5/min, AI 30/min
- NATS per-service auth with 10 users + permission matrix
- SSL enforcement, ReadHeaderTimeout on all services
- Header spoofing stripped in auth middleware
- Refresh token rotation + SHA-256 hashing + account lockout
- MFA token replay prevention, secure cookies, RBAC permissions
- Docker network segmentation, secrets, CrowdSec, CORS restriction

**Why:** This audit proves the system crossed the production-readiness threshold.
**How to apply:** H1 (blacklist wiring) should be addressed before opening to external users. For 3-user controlled deployment, current state is acceptable.
