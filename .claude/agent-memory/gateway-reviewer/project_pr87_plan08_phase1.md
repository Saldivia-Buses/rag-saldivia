---
name: PR #87 Plan 08 Phase 1 review
description: PR review of 5 critical backend hardening fixes -- NATS env vars missing from container, platform NATS_URL missing, event type dots break NATS permissions
type: project
---

Reviewed PR #87 (Plan 08 Phase 1) on 2026-04-05. Status: CAMBIOS REQUERIDOS.

3 blockers:
1. NATS container has no `environment:` block -- `$*_NATS_PASS` vars resolve to empty strings, all users have empty passwords
2. Platform service missing `NATS_URL` in docker-compose.prod.yml -- defaults to localhost:4222 which fails in Docker
3. Platform `publishLifecycleEvent` creates 5-segment subjects (type has dots) that NATS `tenant.*.notify.*` permission denies

5 corrections:
- Dev compose still has stale `rag` service (non-existent Dockerfile)
- Orphan `nats_token` secret declared but never used
- Secrets README still documents single shared NATS token
- Platform migration 005 missing action format CHECK constraint (tenant has it)
- `ensureSSLMode` silently swallows URL parse errors

Pattern found: NATS event types with dots create multi-segment subjects. This is a recurring issue (previously found in PR #82). The `IsValidEventType` regex allows dots, but dots are NATS segment separators. Services using `Notify()` with dotted types (like `platform_tenant.created`) will be denied by single-token `*` permissions.

**Why:** tracks blockers for Plan 08 Phase 1 merge
**How to apply:** verify these are fixed before approving the PR. The NATS dots-in-types issue may affect other services too -- check all `Notify()` callers.
