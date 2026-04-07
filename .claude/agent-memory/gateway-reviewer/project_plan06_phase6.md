---
name: Plan 06 Phase 6 review
description: PR #70 Traces Service, blockers: status CHECK constraint mismatch, tenant isolation broken in 3 places (handler reads from query param/URL instead of JWT context, GetTraceDetail has no tenant filter), NATS callbacks use shutdown context, NATS subject convention diverges from tenant.{slug}.* pattern
type: project
---

PR #70 Traces Service review found 5 blockers and 8 must-fix items.

**Critical:** Tenant isolation broken in handler layer -- tenant_id read from query param and URL path instead of JWT-injected headers/context. GetTraceDetail fetches by trace ID alone with no tenant filter, allowing cross-tenant data access.

**Critical:** RecordTraceStart inserts status='running' but the DB CHECK constraint only allows ('completed','failed','cancelled','timeout'). Service is non-functional as shipped.

**Why:** The traces service was built with a platform-admin mindset (arbitrary tenant_id as parameter) but is mounted behind normal Auth middleware, so any authenticated user can query any tenant's data.

**How to apply:** When reviewing traces fixes, verify: (1) all handlers read tenant from X-Tenant-ID header, (2) all SQL queries include tenant_id in WHERE, (3) NATS subjects follow tenant.{slug}.traces.{type} convention, (4) migration adds 'running' to status CHECK.
