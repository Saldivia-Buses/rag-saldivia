---
name: PR 85 token blacklist audit
description: Security audit of JWT token blacklist PR -- 1 critical (dead code, nothing wired), 3 high. NOT APTO.
type: project
---

PR #85 feat/plan07-token-blacklist audited 2026-04-05.

1 critical: blacklist primitives exist but are never activated (no Redis client in auth, no SetBlacklist call, no AuthWithConfig usage anywhere). Entire feature is dead code.

3 high: fail-open on Redis failure, no tenant namespace in blacklist prefix, RevokeAll not wired to password change.

**Why:** Plan 07 Phase 5 requires end-to-end wiring including integration test. PR only delivers the building blocks.

**How to apply:** When re-reviewing, verify: (a) auth main.go creates Redis client + TokenBlacklist, (b) SetBlacklist called in both single-tenant and multi-tenant paths, (c) all services use AuthWithConfig, (d) integration test exists.
