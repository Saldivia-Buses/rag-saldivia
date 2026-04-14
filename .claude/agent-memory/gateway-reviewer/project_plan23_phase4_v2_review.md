---
name: Plan 23 Phase 4 Second Pass Review
description: Second-pass review after 5 critical/high + 8 medium/low fixes; new findings around service-token endpoint missing, shell injection in issue title, rows.Err() missing, and platform context inconsistency
type: project
---

Second pass of HealthWatch + notification + CI changes after fixes in 27bcc4fe and 956c2d22.

**Result:** CAMBIOS REQUERIDOS — 2 new blockers found

**Blocker 1: `/v1/auth/service-token` endpoint does not exist**
- `get-service-token.sh` calls `POST /v1/auth/service-token` 
- Auth service has no such route — daily-triage and post-deploy CI workflows will always 401/404
- `service_account_key` Docker secret also not declared in docker-compose.prod.yml

**Blocker 2: Shell injection in issue title (daily-triage.yml)**
- `--title "$TITLE"` is direct shell interpolation of AI output
- Body is correctly written to file; title is not — partial DS6 compliance only

**Must-fix: rows.Err() missing in ListTriageRecords**
- `services/healthwatch/internal/service/healthwatch.go` ~line 228
- pgx rows iteration requires `rows.Err()` check after loop to catch mid-stream errors

**Platform context vs healthwatch context inconsistency**
- Healthwatch `requirePlatformAdmin` populates Go context (tenant.WithInfo, sdamw.WithRole, etc.)
- Platform `requirePlatformAdmin` sets only headers, not Go context
- Low severity: platform service doesn't use tenant.FromContext(), but inconsistency will cause bugs if code is refactored

**Why:** Service-token endpoint was in the plan spec but not implemented in auth service routes (cmd/main.go has no such handler). This breaks the entire CI triage/auto-close authentication chain.
