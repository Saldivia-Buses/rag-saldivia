---
name: PR #100 Blacklist Wiring Review
description: Plan 10 Phase 1 — blacklist wiring across all services, blockers in auth multi-tenant path and Redis client leak
type: project
---

PR #100 — CAMBIOS REQUERIDOS

Blockers:
1. `services/auth/cmd/main.go:107` — multi-tenant auth path calls `NewMultiTenantAuth` without passing `blacklist`. Single-tenant path wires it via `SetBlacklist`. Logout in production (multi-tenant) never actually blocks revoked tokens.
2. `services/auth/cmd/main.go:71-77` — failed Redis Ping leaves `redis.NewClient` pool open (no Close on error path). Auth should use `security.InitBlacklist` instead of inline init.

Must-fix:
3. `platform.requirePlatformAdmin` only re-injects `X-User-ID` after verify — should inject full 5-header set for future-proofing.
4. `feedback` `/v1/platform/feedback` route uses `FailOpen: true` — platform-facing routes should be `FailOpen: false`.

Suggestion: `REDIS_URL` treated as bare `host:port` in `InitBlacklist` — will silently fail if value is `redis://` URL scheme. Document or parse properly.

**Why:** Auth multi-tenant is the production path. Blacklist being absent there means token revocation (logout) is ineffective for all real tenants.
**How to apply:** When reviewing auth, always check both the single-tenant and multi-tenant branches. They have separate wiring paths.
