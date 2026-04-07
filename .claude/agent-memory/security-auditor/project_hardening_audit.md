---
name: Hardening PR audit
description: Security audit of feat/hardening PR -- 2 critical (MFA token reuse, NATS auth disconnect), 3 high. NOT APTO.
type: project
---

Hardening PR audit completed 2026-04-03. 17 phases of security hardening reviewed.

**CRITICAL:**
- C1: MFA temp token reusable within 5min window -- no JTI tracking after use. services/auth/internal/service/auth.go:246-253
- C2: NATS auth token defined in docker-compose.prod.yml but no Go service passes it in nats.Connect(). All cmd/main.go files missing nats.Token() option.

**HIGH:**
- H1: mfa_secret stored plaintext in tenant DB (users.mfa_secret) despite pkg/crypto existing
- H2: Redis password from ${REDIS_PASSWORD} env var instead of Docker secret
- H3: Platform admin slug check skipped when platformSlug is empty string

**MEDIUM:**
- M1: Slug cross-validation after header injection (code clarity, not exploitable)
- M2: Audit write synchronous in request path despite "non-blocking" doc
- M3: Auth resolver uses nil encryption key (encrypted credentials not read)
- M4: JWT permissions immutable for 15min access token lifetime

**Why:** The hardening PR adds significant security (Ed25519, RBAC, MFA, AES-GCM, audit) but has integration gaps where features are defined but not wired together.

**How to apply:** C1+C2 must be fixed before any production deployment. H1-H3 should be fixed in same PR before merge.
