# Gateway Review -- PR #54 Multi-Tenant Auth + OpenRouter RAG

**Fecha:** 2026-04-03
**Branch:** `feat/multi-tenant-openrouter`
**Resultado:** CAMBIOS REQUERIDOS

**Archivos revisados:**
- `services/auth/cmd/main.go`
- `services/auth/internal/handler/auth.go`
- `services/auth/internal/service/auth.go`
- `services/auth/internal/handler/auth_test.go`
- `services/auth/internal/service/auth_integration_test.go`
- `services/rag/cmd/main.go`
- `services/rag/internal/service/rag.go`
- `services/rag/internal/service/json.go`
- `services/rag/internal/handler/rag.go`
- `services/rag/internal/handler/rag_test.go`
- `pkg/middleware/auth.go`
- `pkg/jwt/jwt.go`
- `pkg/tenant/resolver.go`
- `pkg/tenant/context.go`
- `pkg/nats/publisher.go`
- `deploy/traefik/dynamic/dev.yml`

---

## Bloqueantes

### B1. Auth login has no tenant slug source in multi-tenant mode

**[services/auth/internal/handler/auth.go:63]**

`resolveService()` reads `X-Tenant-Slug` from the request header:

```go
slug := r.Header.Get("X-Tenant-Slug")
if slug == "" {
    return nil, errors.New("missing tenant context")
}
```

But login, refresh, and logout are **public routes** -- they run without the Auth middleware that normally injects `X-Tenant-Slug` from JWT claims. The question is: who sets this header?

In the Traefik dev config (`deploy/traefik/dynamic/dev.yml:18`), the auth router explicitly does **not** use the `dev-tenant` middleware -- the YAML comment says "no tenant context needed for login". In production, Traefik would need to extract the slug from the subdomain, but that routing middleware is not yet implemented for auth routes.

**Impact:** In multi-tenant mode, `Login`, `Refresh`, and `Logout` will always fail with 502 "tenant not available" because `X-Tenant-Slug` will be empty. The feature is non-functional.

**Fix (two parts, both required):**

1. **Dev:** Add `dev-tenant` middleware to the auth Traefik route and update the comment:
```yaml
auth:
  rule: "PathPrefix(`/v1/auth`) || PathPrefix(`/v1/modules`)"
  service: auth
  entryPoints: [web]
  middlewares: [dev-tenant]  # multi-tenant mode needs slug
```

2. **Prod (when created):** Ensure the production Traefik config extracts tenant slug from subdomain for auth routes as well. Also must strip any client-spoofed `X-Tenant-Slug` header before injecting the one derived from the subdomain -- otherwise a malicious client could set `X-Tenant-Slug: victim-tenant` and authenticate against another tenant's user database. The Auth middleware (`pkg/middleware/auth.go:28-31`) strips spoofed headers, but it does not run on public auth routes.

### B2. `tenantID := slug` puts slug in JWT `tid` claim -- downstream services will break

**[services/auth/internal/handler/auth.go:74]**

```go
tenantID := slug // simplified: use slug as ID for now
return service.NewAuth(pool, h.jwtCfg, tenantID, slug, h.publisher), nil
```

This slug ends up as `a.tenant.ID` in `service.Auth` (service/auth.go:59), which is then embedded in JWT claims as `tid` (jwt.go:27). The Auth middleware injects this as `X-Tenant-ID` header (middleware/auth.go:56). Downstream services use `X-Tenant-ID` for SQL queries against tables where `tenant_id` is a UUID.

If a JWT minted by multi-tenant auth carries `tid: "saldivia"` (a slug) and a service does `WHERE tenant_id = 'saldivia'` against a table with UUID tenant IDs, the query silently returns zero rows. This is not a data leak, but it **breaks all downstream functionality silently**.

**Fix:** The Resolver already queries the `tenants` table in `resolveConnInfo`. Extend it to also return the tenant UUID:

```go
// In ConnInfo, add:
type ConnInfo struct {
    TenantID    string // tenant UUID
    PostgresURL string
    RedisURL    string
}

// In resolveConnInfo, change the query:
err := r.platformDB.QueryRow(ctx,
    `SELECT id, postgres_url, redis_url FROM tenants WHERE slug = $1 AND enabled = true`,
    slug,
).Scan(&info.TenantID, &info.PostgresURL, &info.RedisURL)
```

Then add a `TenantID(ctx, slug)` method on Resolver, or expose it via the cached `ConnInfo`. In the handler:

```go
connInfo, err := h.resolver.ConnInfo(r.Context(), slug)
pool, err := h.resolver.PostgresPool(r.Context(), slug)
return service.NewAuth(pool, h.jwtCfg, connInfo.TenantID, slug, h.publisher), nil
```

---

## Debe corregirse

### D1. No slug validation in `resolveService()`

**[services/auth/internal/handler/auth.go:63-68]**

The slug from `X-Tenant-Slug` is passed directly to `resolver.PostgresPool()` without format validation. While the Resolver will fail with `ErrTenantUnknown` for invalid slugs (SQL returns no rows), basic validation avoids unnecessary platform DB queries for malformed slugs and protects logs from control characters.

**Fix:**
```go
slug := r.Header.Get("X-Tenant-Slug")
if slug == "" || !isValidSlug(slug) {
    return nil, errors.New("invalid or missing tenant context")
}
```

Where `isValidSlug` matches the expected pattern (lowercase alphanumeric + hyphens, max 63 chars).

### D2. `ListCollections` will error in OpenRouter mode

**[services/rag/internal/service/rag.go:116-157]**

When `APIKey` is set (OpenRouter mode), `ListCollections` still calls `BlueprintURL+"/v1/collections"` -- an endpoint that does not exist on OpenRouter. This will return a confusing 404 or connection error that the handler surfaces as 502 "rag server unavailable".

**Fix:** Short-circuit in OpenRouter mode:
```go
func (r *RAG) ListCollections(ctx context.Context, tenantSlug string) ([]string, error) {
    if r.cfg.APIKey != "" {
        return []string{}, nil // external providers don't support collection listing
    }
    // ... existing Blueprint logic
}
```

### D3. Upstream error body logged unbounded -- potential key fragment exposure

**[services/rag/internal/service/rag.go:107-109]**

```go
respBody, _ := io.ReadAll(resp.Body)
return nil, "", fmt.Errorf("blueprint returned %d: %s", resp.StatusCode, string(respBody))
```

The error propagates to the handler which logs it. The client gets a generic 502, so no user-facing leak. But the log entry includes the full upstream response body, which:
- Is unbounded (could be very large)
- May contain error messages that reference the API key (e.g., "Invalid key: sk-or-v1-ab...")
- Could contain characters that corrupt structured JSON logs

**Fix:** Use `io.LimitReader` and truncate:
```go
respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
resp.Body.Close()
return nil, "", fmt.Errorf("upstream returned %d: %s", resp.StatusCode, string(respBody))
```

### D4. No handler tests for multi-tenant code path

**[services/auth/internal/handler/auth_test.go]**

All 14 existing tests use `NewAuth(mock)` (single-tenant). Zero tests exercise `NewMultiTenantAuth` or `resolveService` when `h.authSvc == nil`. Coverage gaps:

- `resolveService()` with missing `X-Tenant-Slug` header: untested
- `resolveService()` with resolver error: untested
- The 502 response for tenant resolution failure: untested
- Multi-tenant Login/Refresh/Logout/Me flows: untested

**Fix:** At minimum, add:
1. `TestLogin_MultiTenant_MissingSlug_Returns502`
2. `TestLogin_MultiTenant_ResolverError_Returns502`
3. `TestResolveService_SingleTenant_ReturnsStatic` (verifies backward compat)

To test multi-tenant mode, you need a Resolver mock. Since `Resolver` is a concrete struct (not an interface), consider either:
- Extracting a `TenantResolver` interface for testability
- Or testing at the integration level with testcontainers

### D5. Platform DB not pinged on multi-tenant startup

**[services/auth/cmd/main.go:85-93]**

In single-tenant mode, the pool is pinged at startup (line 106). In multi-tenant mode, the platform pool is created but never pinged. If the platform DB is down, the service starts successfully and then fails on every request.

**Fix:**
```go
platformPool, err := pgxpool.New(ctx, platformDBURL)
if err != nil { ... }
if err := platformPool.Ping(ctx); err != nil {
    slog.Error("failed to ping platform database", "error", err)
    os.Exit(1)
}
```

### D6. `Health()` does not work in OpenRouter mode

**[services/rag/internal/service/rag.go:160-177]**

`Health()` calls `BlueprintURL+"/health"`. OpenRouter does not have a `/health` endpoint. In OpenRouter mode, the health check will always fail, making monitoring tools think the service is down.

**Fix:** Differentiate health behavior:
```go
func (r *RAG) Health(ctx context.Context) error {
    if r.cfg.APIKey != "" {
        return nil // external API -- assume healthy, failures surface on generate
    }
    // ... existing Blueprint health check
}
```

---

## Sugerencias

### S1. Log the mode and model on RAG startup

When OpenRouter mode is active, log the model name (not the key) for operational visibility:
```go
if ragCfg.APIKey != "" {
    slog.Info("rag service starting in external API mode", "port", port, "url", blueprintURL, "model", ragCfg.Model)
} else {
    slog.Info("rag service starting in blueprint mode", "port", port, "url", blueprintURL)
}
```

### S2. Consider caching `service.Auth` instances per tenant

Every request in multi-tenant mode allocates a new `service.Auth` struct (handler/auth.go:75). The struct is lightweight (pool pointer + config), so allocation cost is negligible today. But if `service.Auth` ever gains initialization logic, this becomes a problem. Add a `// TODO: cache service instances per tenant` comment.

### S3. Add `omitempty` to Blueprint-specific RAG fields

In OpenRouter mode, cleared fields still serialize as `"vdb_top_k": 0, "reranker_top_k": 0`. While OpenRouter likely ignores them, `omitempty` tags on `VdbTopK`, `RerankerTopK` would produce cleaner payloads:
```go
VdbTopK      int  `json:"vdb_top_k,omitempty"`
RerankerTopK int  `json:"reranker_top_k,omitempty"`
```

Note: `UseKnowledgeBase` with `omitempty` would omit `false`, which may change Blueprint behavior. Only apply `omitempty` to the integer fields.

### S4. Dotted event types in NATS subjects are inconsistent (pre-existing)

The auth service publishes `auth.login_success` as event type (service/auth.go:193). The NATS publisher builds the subject as `tenant.{slug}.notify.auth.login_success` -- 5 tokens. The consumer filter `tenant.*.notify.>` uses `>` (multi-token wildcard), so this works. But it's inconsistent with 4-token subjects from other services. This is a pre-existing issue documented in agent memory from PR #52.

---

## Lo que esta bien

1. **Dual-mode startup pattern is clean and explicit.** The `if platformDBURL / else if tenantDBURL / else exit` cascade in cmd/main.go is clear, handles both modes, and fails with a helpful error if neither DB URL is provided. Good separation.

2. **Resolver pool caching is correct.** `tenant.Resolver.PostgresPool()` uses mutex with double-check after lock release during network I/O. Pools are cached per slug. `defer resolver.Close()` in main.go prevents leaks.

3. **Backward compatibility fully preserved.** `NewAuth(authSvc)` sets `authSvc` directly. `resolveService()` returns it immediately when non-nil. All 14 existing handler tests and 8 integration tests pass without modification.

4. **API key never logged.** No slog call in the RAG service includes the key value. The key is only used in the `Authorization` header of outgoing requests.

5. **Blueprint-specific fields properly cleared in OpenRouter mode.** `CollectionName`, `UseKnowledgeBase`, `VdbTopK`, `RerankerTopK` are all zeroed (rag.go:76-79). Model injection only happens when the request doesn't already specify one.

6. **SSE streaming proxy is format-agnostic.** Raw byte pass-through (handler/rag.go:79-88) works for both Blueprint and OpenRouter since both use OpenAI-compatible SSE format.

7. **Error responses remain generic to clients.** 502 "tenant not available", 502 "rag server unavailable", 500 "internal error" -- never internal details. Logs include request ID for correlation.

8. **Refresh token security is solid.** Cookie-first with body fallback, HttpOnly + Secure + SameSite=Strict, SHA-256 hashed in DB, rotation (single-use), proper revocation.

9. **Brute force protection works.** 5 failed attempts trigger 15-minute lockout, timing-safe dummy bcrypt for nonexistent users, disabled accounts get the same error as invalid credentials (prevents account enumeration).

10. **Header spoofing protection on protected routes.** `pkg/middleware/auth.go:27-31` deletes all identity headers before JWT verification. This correctly prevents spoofing on routes behind the auth middleware.

11. **Integration tests are thorough.** Login success, wrong password, brute-force lockout, disabled user, nonexistent user, audit log persistence, refresh token storage -- all tested with real PostgreSQL via testcontainers.
