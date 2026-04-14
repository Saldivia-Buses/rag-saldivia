---
name: Plan 23 Phase 4 HealthWatch audit
description: Security audit of healthwatch service, notification /send endpoint, triage cron, and post-deploy workflow (Plan 23 Phase 4)
type: project
---

Audit of Plan 23 Phase 4 commits (2026-04-14). Covers healthwatch service (8014), notification /send endpoint, daily-triage.yml, post-deploy.yml.

**Verdict:** NOT APTO — 1 critical, 3 high, 4 medium.

**Critical:**
- /send endpoint under requireUserID only: any authenticated user can send arbitrary emails to any address via POST /v1/notifications/send. No platform admin or role check. Token comes from JWT via sdamw.AuthWithConfig (so JWT is verified), but no role validation exists for this endpoint.

**High:**
1. Blacklist fail-open in healthwatch requirePlatformAdmin (handler/healthwatch.go:130-135): `if err == nil && revoked` — Redis error silently allows revoked admin tokens. Platform service has same pattern. Correct behavior: fail closed on Redis error (return 503 or 401).
2. Token not masked in post-deploy.yml:34 — TOKEN=$(get-service-token.sh healthwatch) written directly without ::add-mask::, exposed in GitHub Actions logs.
3. Missing traefik.docker.network=frontend label on healthwatch in docker-compose.prod.yml. Container is on 3 networks (frontend, backend, data). Without this label, Traefik may route via the wrong network.

**Medium:**
1. ${{ steps.auth.outputs.token }} used directly in `run:` block in daily-triage.yml lines 43 and 152 — context expression injected into shell, not passed via env. Any secret that ends up in outputs should use env: to prevent script injection.
2. health_snapshots table has no check constraint on status column — INSERT accepts arbitrary strings for status/severity (no enum enforcement at DB level, only app-level).
3. HEALTHY variable in post-deploy.yml:49 echoed to log — service names are not sensitive but the pattern (echo HEALTHY unmasked) could leak if the value were a token.
4. Traefik router for healthwatch missing tls label — entrypoints=websecure is set but no certresolver or tls=true label. Other services have the same pattern so this may be system-wide, but the websecure entrypoint without explicit TLS config at label level means it relies entirely on global Traefik config.

**Why:** The /send privilege escalation is the most dangerous finding. A regular tenant user can send emails to any address by crafting a JWT-authenticated POST. The blacklist fail-open is inherited from the platform service pattern and applies to the highest-privilege endpoint in the system.

**How to apply:** When reviewing admin-only endpoints implemented with inline middleware (not pkg/middleware), always check: (1) does blacklist error fail-closed? (2) is the role check exhaustive? When reviewing CI workflows with dynamic tokens, always verify ::add-mask:: is present.
