---
name: PR #87 Plan 08 Phase 1 review
description: 5 critical fixes (audit, Traefik, sqlc, SSL, NATS auth) -- blockers: NATS env vars not passed to container, platform missing NATS_URL
type: project
---

Reviewed PR #87 (Plan 08 Phase 1) on 2026-04-05. Status: CAMBIOS REQUERIDOS.

Blockers:
- B1: nats-server.conf uses $VAR password expansion but docker-compose.prod.yml NATS container has no environment section -- passwords not available
- B2: Platform service missing NATS_URL in docker-compose.prod.yml despite connecting to NATS in main.go and having a user defined in nats-server.conf

Must fix:
- Orphan nats_token secret still defined (dead config)
- Secrets README still documents shared NATS token pattern
- docker-compose.dev.yml still has deprecated rag service (broken Dockerfile path)
- feedback raw SQL not migrated to sqlc (C3 gap, accepted exception needed)
- Audit entries missing IP/UserAgent in ingest+platform

What was good:
- Audit logging clean integration (ingest: upload+delete, platform: 8 mutations)
- All 6 sqlc.yaml correctly centralized
- ensureSSLMode in resolver follows spec
- NATS permissions match actual publish/subscribe patterns for all 10 services
- Traefik dev+prod configs complete with correct middleware chains

**Why:** establishes review record for Plan 08 Phase 1 implementation
**How to apply:** reference when reviewing follow-up fixes or subsequent Plan 08 phases
