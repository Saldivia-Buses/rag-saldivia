---
name: Plan 06 Phase 2 review
description: PR #66 Extractor service (Python) review -- NATS subject injection, convention mismatch, no partial failure handling, no health check
type: project
---

PR #66 Extractor service (Python, first non-Go service) reviewed 2026-04-05. Result: CAMBIOS REQUERIDOS.

Key blockers:
- NATS subject injection: tenant_slug interpolated into subject without validation (same bug class as PRs #52, #55 in Go)
- NATS subject naming uses `extractor.>` instead of `tenant.{slug}.extractor.>` convention used by all Go services
- No MaxDeliver on JetStream consumer (infinite retries if SGLang is down)
- No partial failure handling: one bad page kills entire document extraction

**Why:** Python services need the same subject validation Go gets from `isValidSubjectToken()`. Convention mismatch makes cross-service filtering impossible.

**How to apply:** When reviewing future Python services (or any non-Go service), check: NATS subject validation, subject naming convention alignment with Go services, retry budget, and partial failure handling.
