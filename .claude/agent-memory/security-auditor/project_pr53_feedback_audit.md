---
name: PR 53 Feedback Service audit
description: Security audit of Feedback Service + OTel instrumentation. 2 critical (NATS no auth, LIMIT concat), 3 high. NOT APTO.
type: project
---

PR #53 (feat/wiring-polish-feedback) audited 2026-04-03. Feedback service adds NATS consumer, aggregator, alerter, HTTP handlers. OTel instrumented across all 8 services.

**Why:** Feedback service handles cross-tenant data flow (tenant DB -> platform DB aggregation) and trusts NATS subject routing for tenant identification.

**How to apply:**
- NATS has zero auth anywhere in the repo (deploy/docker-compose.dev.yml, no nats.conf). This is the most systemic risk -- affects all services that use NATS, not just feedback.
- The LIMIT concatenation pattern in feedback.go:137 should be checked if it appears in other services.
- Alert notifications silently fail because aggregator passes empty slug (aggregator.go:112).
- Full report at docs/artifacts/pr53-security-audit.md.
