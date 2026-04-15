---
name: CreateFeatureFlag NULL scan bug
description: CreateFeatureFlag scans tenant_id into plain string — breaks for global flags (NULL tenant_id)
type: project
---

`services/platform/internal/service/platform.go` — `CreateFeatureFlag` function has a scan bug.

The query `RETURNING id, name, tenant_id, enabled, rollout_pct` scans `tenant_id` into `&f.TenantID` where `f.TenantID` is `string` (not `*string`). When creating a global flag (no tenant scope), `tenant_id` is NULL in the DB and pgx cannot scan NULL into a plain `string`.

Error: `can't scan into dest[2] (col: tenant_id): cannot scan NULL into *string`

**Why:** `FeatureFlag.TenantID` field is `string` but should be `*string` for nullable column. Or the scan should use a local `var tid *string` then assign.

**How to apply:** When writing tests that call `CreateFeatureFlag` for global flags (no TenantID), use direct INSERT instead. Mark with TDD-ANCHOR comment. Fix in production: change `Scan(&f.ID, &f.Name, &f.TenantID, ...)` to use `var tid *string; Scan(..., &tid, ...)` then `if tid != nil { f.TenantID = *tid }`.

Fix location: `services/platform/internal/service/platform.go:361`
