# Gateway Review -- PR #83 Config Resolver (Plan 07 Phase 2b)

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS (2 blockers, 5 must-fix, 4 suggestions)

## Files reviewed

- `pkg/config/resolver.go` -- Resolver with scope cascade, slot resolution, prompt lookup, Redis cache
- `pkg/config/resolver_test.go` -- 8 integration tests against real Platform DB
- `pkg/go.mod` -- dependency changes (pgxpool, redis already existed)

---

## Bloqueantes

### B1. Cache reads skip plan and global cascade -- stale/wrong values returned

**File:** `pkg/config/resolver.go:50-54`

When `tenantID` is non-empty and a cache hit occurs, `Get()` returns immediately without ever checking whether the cached value came from the tenant scope, plan scope, or global scope. This is correct behavior **if** the cache was populated by a prior cascade. That part is fine.

The actual problem is the **inverse**: if a tenant-scoped override is **added** to `agent_config` after the global value was already cached, the cache will keep returning the stale global value for up to 5 minutes. This is expected TTL behavior. **However**, there is no `InvalidateCache` call wired to any event (NATS lifecycle, platform API) -- so in practice, the only way to invalidate is to wait for TTL expiry or manually call the function.

This is not a blocker by itself. The **real blocker** is this: when `tenantID` is empty, Get() **never caches and never reads from cache** (line 50: `if r.cache != nil && tenantID != ""`). This means every call from a non-tenant context (e.g., platform admin tools, background jobs) hits the database every time. But when `tenantID` is non-empty, it caches the result keyed by `tenantID:key`. Now consider: a platform admin updates a global config key. `InvalidateCache` clears `sda:config:{tenantID}:*` for ONE tenant. But every OTHER tenant still has the stale global value cached. **There is no way to invalidate global config changes across all tenants.**

**Fix:** `InvalidateCache` needs a global mode. When a global key changes, either:
- (a) Use a Redis key like `sda:config:global:{key}` and cache global results separately, invalidating them in one shot, OR
- (b) Prefix the cache with a generation counter: bump a counter on global change, include it in cache keys, stale keys auto-miss.

Option (a) is simplest. Add a `InvalidateGlobal(ctx, key)` method or change `InvalidateCache(ctx, tenantID)` to accept `""` as "invalidate global".

### B2. Cache key is tenant+key but cached value may be from any scope -- cross-tenant poisoning impossible BUT silent stale reads guaranteed

**File:** `pkg/config/resolver.go:68-69, 83-84, 100-101`

The cascade stores the resolved value in `sda:config:{tenantID}:{key}` regardless of which scope matched. If the resolution fell through to global, the next call for the same tenant+key returns the global value from cache -- even if a tenant override was added in the meantime (within TTL). This is inherent to the 5-minute TTL approach and acceptable **only if** invalidation works correctly. Per B1, invalidation for global changes is broken.

**Fix:** Same as B1 -- this becomes acceptable once global invalidation exists.

---

## Debe corregirse

### M1. Three sequential queries for full cascade -- should be one query

**File:** `pkg/config/resolver.go:62-96`

The cascade fires up to 3 separate queries:
1. `SELECT value FROM agent_config WHERE scope = 'tenant:'+tenantID AND key = $2`
2. `SELECT plan_id FROM tenants WHERE id = $1` then `SELECT value FROM agent_config WHERE scope = 'plan:'+planID AND key = $2`
3. `SELECT value FROM agent_config WHERE scope = 'global' AND key = $1`

In the worst case (global fallback), this is **3 round trips** to the database. Since the Resolver hits the **Platform DB** which is shared by all tenants and all services, this 3x amplification under load is a real concern.

**Fix:** Collapse into a single query using `COALESCE` + subqueries or a CTE:

```sql
WITH tenant_val AS (
    SELECT value FROM agent_config WHERE scope = $1 AND key = $2
),
plan_val AS (
    SELECT ac.value FROM agent_config ac
    JOIN tenants t ON ac.scope = 'plan:' || t.plan_id
    WHERE t.id = $3 AND ac.key = $2
),
global_val AS (
    SELECT value FROM agent_config WHERE scope = 'global' AND key = $2
)
SELECT COALESCE(
    (SELECT value FROM tenant_val),
    (SELECT value FROM plan_val),
    (SELECT value FROM global_val)
) AS value;
```

Parameters: `$1 = 'tenant:'+tenantID, $2 = key, $3 = tenantID`. This is 1 round trip and the database optimizer can handle it efficiently since `(scope, key)` has a UNIQUE index.

### M2. `GetString` silently returns raw JSON when unmarshal fails

**File:** `pkg/config/resolver.go:113-114`

```go
if err := json.Unmarshal(raw, &s); err != nil {
    return string(raw), nil // return raw if not a quoted string
}
```

If the stored value is `10000` (number), `42.5` (float), `true` (bool), `["a","b"]` (array), or `{"x":1}` (object), `GetString` returns the raw JSON text with **no error**. A caller doing `GetString(ctx, tid, "guardrails.input_max_length")` gets `"10000"` (the string) instead of an error telling them to use `GetInt`.

This hides type mismatches. The caller thinks they got a valid string but it is actually a JSON number literal.

**Fix:** Return an error when the value is not a JSON string:

```go
if err := json.Unmarshal(raw, &s); err != nil {
    return "", fmt.Errorf("config %q is not a string: %w", key, err)
}
```

If the intent is to support both quoted strings and bare values, document it explicitly and rename to `GetStringCoerce`.

### M3. `pgx.ErrNoRows` compared with `==` instead of `errors.Is`

**File:** `pkg/config/resolver.go:94, 149, 169`

Three places compare `err == pgx.ErrNoRows`. While this works today with pgx v5, the idiomatic Go pattern is `errors.Is(err, pgx.ErrNoRows)` to handle wrapped errors. If pgx ever wraps the sentinel, or if middleware wraps the error, this comparison silently fails and the error propagates as a generic query failure instead of "not found".

Lines:
- 94: `if err == pgx.ErrNoRows {`
- 149: `if err == pgx.ErrNoRows {` (in ResolveSlot)
- 169: `if err == pgx.ErrNoRows {` (in GetActivePrompt)

**Fix:** Replace all three with `errors.Is(err, pgx.ErrNoRows)` and add `"errors"` to imports.

### M4. `ResolveSlot` does not cache model lookups

**File:** `pkg/config/resolver.go:134-158`

`Get` (called via `GetString`) caches the config value (the model ID string). But the subsequent `llm_models` query on line 145-148 is **never cached**. Since model data changes extremely rarely (only when a platform admin adds/updates a model), this is a missed optimization and adds a DB round trip on every slot resolution even on cache hit.

**Fix:** Cache the full `ModelConfig` result, or at minimum cache the `llm_models` row separately with a longer TTL (e.g., 15 minutes). The cache key could be `sda:model:{modelID}`.

### M5. `InvalidateCache` uses SCAN which is O(N) on the total Redis keyspace

**File:** `pkg/config/resolver.go:178-188`

`SCAN` with pattern `sda:config:{tenantID}:*` iterates the **entire** keyspace in pages of 100. With many tenants and many config keys cached, this becomes expensive. Worse, it calls `Del` one key at a time inside the loop instead of batching.

**Fix:** Two options:
1. **Pipeline the deletes:** Collect all keys from SCAN into a slice, then `r.cache.Del(ctx, keys...)` in one call.
2. **Use a Redis Hash per tenant:** Store all config for a tenant in `sda:config:{tenantID}` as a hash. Invalidation becomes a single `DEL sda:config:{tenantID}`. This also avoids SCAN entirely.

Option 2 is cleaner and more scalable but changes the caching layer. Option 1 is a minimal fix.

---

## Sugerencias

### S1. `NewResolver` should accept a TTL option

The 5-minute TTL is hardcoded. Different deployments may want different TTLs (longer for production stability, shorter for dev iteration). Consider:

```go
func NewResolver(pool *pgxpool.Pool, cache *redis.Client, opts ...ResolverOption) *Resolver
```

Or at minimum a `WithTTL(d time.Duration)` setter.

### S2. Add `GetActivePrompt` caching

`GetActivePrompt` hits the database every time. Prompts change even less frequently than config. A simple Redis cache with a longer TTL (e.g., 30 minutes) would reduce Platform DB load from all services resolving prompts on every request.

### S3. Resolver is thread-safe but has no documentation about it

The `Resolver` struct has no mutex and all its methods just call `pool.QueryRow` (which is safe) and `cache.Get/Set` (which is safe). But this should be explicitly documented in the struct comment:

```go
// Resolver is safe for concurrent use.
type Resolver struct { ... }
```

This is important because every service will share a single `Resolver` instance.

### S4. Test coverage gaps

The 8 tests only cover global scope and error cases. Missing:
- **Plan-scoped resolution:** A test where tenant X has plan "starter" and a plan:starter config exists but no tenant-scoped override. Verify plan value is returned.
- **Tenant-scoped override:** A test where both global and tenant-scoped values exist. Verify tenant value wins.
- **Cache invalidation:** A test with a Redis client (currently all tests pass `nil` for cache). Verify that after `InvalidateCache`, the next `Get` re-queries the DB.
- **Concurrent access:** A test that calls `Get` from multiple goroutines simultaneously (verifies no races, use `go test -race`).
- **`ResolveSlot` with disabled model:** A model exists but `enabled = false`. Verify the "not found or disabled" error.

The plan spec explicitly requires: "Resolver.Get() con cascada de scopes (tenant > plan > global)". The current tests only verify global scope.

---

## Lo que esta bien

1. **SQL injection: clean.** All 6 queries use `$1/$2` parameterized placeholders. No string interpolation in SQL. The scope string is built with Go string concatenation (`"tenant:"+tenantID`) but this goes into a `$1` parameter, not into the query text. Safe.

2. **Null/empty model ID handling in ResolveSlot.** Line 139 checks `modelID == "" || modelID == "null"`, which correctly handles the seeded `'null'` JSONB values for unconfigured slots.

3. **Graceful degradation when cache is nil.** `NewResolver` accepts `nil` for cache, and every method checks `r.cache != nil` before using it. Services that don't need caching can skip Redis.

4. **Clean separation from `env.go`.** The package correctly separates infra config (env vars) from business config (DB resolver). The two files have no coupling.

5. **Package import graph is reasonable.** `pkg/config` importing `pgxpool` and `redis` is appropriate -- it is a config resolution layer that inherently needs database and cache access. The existing `pkg/tenant/resolver.go` already imports both of these from `pkg/`, so the precedent is established.

6. **Error messages are descriptive without leaking internals.** Errors like `"config key not found: %q"` and `"model %q not found or disabled"` are useful for debugging without exposing SQL or schema details.

7. **ModelConfig struct matches the DB schema.** The fields align with `llm_models` table columns. `COALESCE(api_key, '')` correctly handles NULL api_key.

---

## Summary

The implementation is solid for a first pass. The SQL is clean, the cascade logic is correct, and the API surface matches what the plan spec calls for in Phase 2b. The two blockers are both about cache invalidation correctness -- the resolver caches results but has no way to invalidate global config changes across all tenants, which means a platform admin changing a global setting has no way to force all services to pick up the change except waiting 5 minutes. The must-fix items are about performance (3 queries should be 1), type safety (GetString silently coercing), and idiomatic Go (errors.Is). None of the issues are security vulnerabilities -- there is no SQL injection risk and no cross-tenant data leakage path.
