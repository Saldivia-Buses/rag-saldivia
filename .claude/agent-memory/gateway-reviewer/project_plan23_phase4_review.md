---
name: Plan 23 Phase 4 review
description: HealthWatch service + self-healing loop + /send endpoint, 6 commits reviewed
type: project
---

HealthWatch service (port 8014) approved with several must-fix items.

Blockers found:
- testify declared in go.sum but NOT in go.mod (tests will fail: `go: cannot find module providing github.com/stretchr/testify`)
- POST /v1/notifications/send is accessible by any authenticated user — no role check. Allows any tenant user to send email to any address.
- `get-service-token.sh` calls `/v1/auth/service-token` which does not exist in auth service (404 at runtime, CI workflows will fail).
- dev.yml Traefik file missing healthwatch router/service entry.

Must-fix:
- persistSnapshots called inside Summary() without decoupling — DB errors in snapshot persist can cause warnings flood but not failure; acceptable, but check log verbosity.
- `autoheal` uses `AUTOHEAL_CONTAINER_LABEL: all` — restarts ALL containers including infra. Should use a specific label.
- Send endpoint missing tenant isolation: the `in_app` path calls Create() with `To` as userID, which is cross-tenant if To is user from another tenant (no validation).

**Why:** Approved with must-fixes. No compile-blocking issues in main service code.
**How to apply:** Future reviews of notification changes must re-check /send authorization.
