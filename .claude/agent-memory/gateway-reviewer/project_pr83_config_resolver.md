---
name: PR #83 Config Resolver review
description: Plan 07 Phase 2b -- pkg/config Resolver with scope cascade, Redis cache, slot resolution. Blockers: global cache invalidation broken, stale reads across tenants.
type: project
---

PR #83 adds `pkg/config/resolver.go` with `Get` (scope cascade: tenant > plan > global), `GetString`, `GetInt`, `ResolveSlot`, `GetActivePrompt`, `InvalidateCache`. Uses Platform DB (`agent_config`, `llm_models`, `prompt_versions` tables) with optional Redis cache (5min TTL).

Blockers:
1. No way to invalidate global config changes across all tenants. `InvalidateCache(tenantID)` only clears ONE tenant's cached values. A global key change leaves stale values in every other tenant's cache for up to 5 minutes with no override.
2. Related: cache correctness depends on invalidation working, but no NATS event or platform API triggers invalidation yet.

Must-fix:
- 3 sequential DB queries for full cascade -- should be 1 query with CTE/COALESCE
- `GetString` silently returns raw JSON for non-string values instead of erroring
- `pgx.ErrNoRows` compared with `==` instead of `errors.Is` (3 places)
- `ResolveSlot` llm_models query never cached (extra DB hit per call)
- `InvalidateCache` uses SCAN O(N) + one-by-one DEL instead of pipeline or hash

Tests only cover global scope -- no plan-scoped or tenant-override tests despite plan spec requiring cascade tests.

**Why:** This is the central config package that all services will import. Cache correctness issues here affect every service.

**How to apply:** When reviewing services that wire Resolver, verify they call InvalidateCache on config changes (NATS lifecycle events). Verify tests cover all 3 cascade levels.
