# Gateway Review -- PR #88 Plan 08 Phase 2

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

**PR:** `feat/plan08-phase-2` -> `2.0.1`
**Commits:** 5 (c49c9ae, 174f977, 521748d, db69097, 79c231a)
**Scope:** 9 security depth fixes: ReadHeaderTimeout, ownership checks, JTI, FailOpen, cachedModelConfig, WriteTimeout, rate limiting

---

## Bloqueantes

### B1. `services/chat/internal/service/chat_integration_test.go:138` -- Test expects `ErrNotOwner` but service returns `ErrSessionNotFound`

`GetSession` filters by `user_id` at the SQL level (`WHERE id = $1 AND user_id = $2`). When a non-owner queries, PostgreSQL returns zero rows, which `pgx` maps to `pgx.ErrNoRows`, which the service maps to `ErrSessionNotFound` (line 98-99 of `chat.go`). The `ErrNotOwner` sentinel is declared but never returned by `GetSession`.

The integration test at line 138 will fail:
```go
_, err = svc.GetSession(ctx, session.ID, "u-2")
if err != ErrNotOwner {
    t.Fatalf("expected ErrNotOwner, got: %v", err)
}
```

**Fix:** Either:
- (a) Change the test to expect `ErrSessionNotFound` (preferred -- simpler, matches SQL-level filtering design), OR
- (b) Change `GetSession` to first query without `user_id`, then compare `row.UserID != userID` and return `ErrNotOwner`. This is slower (two checks or a different query) but preserves the distinction.

If choosing (a), also remove the unused `ErrNotOwner` from the handler's error checks at lines 118, 174, 233 of `chat.go` -- or keep them as defense-in-depth in case a future service method returns it.

### B2. `services/chat/internal/handler/chat_test.go:360` -- Test expects 201 for `system` role but handler now blocks it with 403

`TestAddMessage_ValidRole_Success` iterates over `{"user", "assistant", "system"}` and expects HTTP 201 for all. But the handler (line 213-215) now rejects `system` role with 403:
```go
if req.Role == "system" {
    writeJSON(w, http.StatusForbidden, ...)
    return
}
```

This test will fail on the `system` sub-case.

**Fix:** Update the test to:
1. Remove `"system"` from the valid roles iteration
2. Add a separate test `TestAddMessage_SystemRole_Returns403` that verifies the 403 behavior

---

## Debe corregirse

### D1. `pkg/middleware/auth.go:19` -- Misleading comment on FailOpen

Comment says `FailOpen: true = allow on Redis error (default)` but Go's zero value for `bool` is `false`, so the actual default when using `Auth(publicKey)` -> `AuthWithConfig(publicKey, AuthConfig{})` is `FailOpen: false` (fail-closed). The parenthetical "(default)" is misleading.

**Fix:** Change comment to:
```go
FailOpen  bool  // true = allow on Redis error; false = reject (default is false)
```

### D2. `pkg/config/resolver.go:169` -- API key fetch silently ignores DB errors on cache hit

On cache hit, the code fetches the API key from DB but discards the error:
```go
_ = r.pool.QueryRow(ctx,
    `SELECT COALESCE(api_key, '') FROM llm_models WHERE id = $1`,
    modelID,
).Scan(&apiKey)
mc.APIKey = apiKey
return &mc, nil
```

If the DB query fails (pool exhausted, network error), `apiKey` stays `""` and the caller gets a `ModelConfig` with empty API key. Downstream LLM calls fail with confusing "unauthorized" errors from the LLM provider rather than a clear infrastructure error.

**Fix:** Return an error if the API key fetch fails:
```go
err := r.pool.QueryRow(ctx,
    `SELECT COALESCE(api_key, '') FROM llm_models WHERE id = $1`,
    modelID,
).Scan(&apiKey)
if err != nil {
    return nil, fmt.Errorf("fetch api key for model %s: %w", modelID, err)
}
```

### D3. `pkg/go.mod:61` -- `golang.org/x/time` should be a direct dependency

`pkg/middleware/ratelimit.go` directly imports `golang.org/x/time/rate`, but `go.mod` lists it as `// indirect`. Running `go mod tidy` will reclassify it.

**Fix:** Run `go mod tidy` in the `pkg/` module, or move the line into the direct `require` block.

### D4. `pkg/middleware/ratelimit.go:50` -- Token bucket semantics differ from advertised "N requests per window"

The config says `Requests: 5, Window: time.Minute` which reads as "5 requests per minute", but the token bucket implementation allows:
- 5 burst requests immediately
- Then ~0.083 tokens/second refill
- Total possible in first minute: ~9-10 requests (5 burst + 4-5 refilled)

For auth brute-force protection, the difference between 5 and 10 per minute matters.

**Fix (either):**
- (a) Update `RateLimitConfig` doc to say "burst size and refill rate, not a hard per-window cap"
- (b) To enforce a hard cap, use a sliding window counter (Redis-based or `sync.Map` with timestamps) instead of token bucket. This can be a follow-up.

At minimum, document the actual behavior so future maintainers don't assume it's a hard 5/min cap.

### D5. `services/ws/cmd/main.go:106` -- WriteTimeout 30s on WebSocket service is risky

While `coder/websocket` hijacks the connection (so Go's `WriteTimeout` theoretically stops applying), the deadline that `http.Server` sets on the underlying `net.Conn` before hijack persists unless the WebSocket library explicitly clears it. If the library has a path where it doesn't reset the deadline, WebSocket connections could be killed after 30s.

**Fix:** Use `WriteTimeout: 0` for the WS service, as WebSocket connections are inherently long-lived. The WebSocket library manages its own per-message deadlines. A comment explaining why:
```go
WriteTimeout: 0, // WebSocket connections are long-lived; coder/websocket manages its own deadlines
```

---

## Sugerencias

### S1. Rate limiter cleanup goroutine has no stop mechanism

`go lim.cleanup(10 * time.Minute)` at line 55 of `ratelimit.go` spawns a goroutine that runs forever. In production this is fine (one per endpoint, created at startup), but it's a goroutine leak in tests. Consider accepting a `context.Context` and stopping on cancellation.

### S2. `ByIP` uses `r.RemoteAddr` which may include port for direct connections

When running without a reverse proxy, `r.RemoteAddr` is `ip:port`, meaning each TCP connection gets its own rate limit bucket. This makes the rate limiter ineffective in direct-connect dev scenarios. Consider stripping the port:
```go
func ByIP(r *http.Request) string {
    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    if host == "" {
        return r.RemoteAddr
    }
    return host
}
```

### S3. Cache hit path for `ResolveSlot` still hits DB every time

The `cachedModelConfig` approach correctly excludes the API key from Redis, but the trade-off is that every `ResolveSlot` call hits the DB even on cache hit (to fetch the API key). This negates some of the caching benefit. Consider whether the API key changes frequently enough to justify this. If not, a short in-memory cache (e.g., 30s TTL `sync.Map`) for API keys could avoid the DB roundtrip on hot paths.

### S4. Auth service logout endpoint has no rate limiting

`POST /v1/auth/logout` at line 158 of `auth/cmd/main.go` has no rate limit. While logout isn't a brute-force target, an attacker with a stolen token could spam logout to create Redis churn. Low priority.

### S5. Consider adding `X-RateLimit-Limit` and `X-RateLimit-Remaining` headers

The rate limiter returns `Retry-After` but not the standard `X-RateLimit-*` headers. These help clients implement backoff:
```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1617806400
```

---

## Lo que esta bien

1. **ReadHeaderTimeout: 10s on all 10 services** -- Correct mitigation for slowloris attacks. Applied consistently, including services with different ReadTimeout values (ingest has 120s ReadTimeout but 10s ReadHeaderTimeout, which is exactly right).

2. **SQL-level ownership checks** -- `GetSession`, `TouchSession`, `DeleteSession`, `RenameSession` (chat) and `GetJob`, `DeleteJob` (ingest) all use `WHERE id = $1 AND user_id = $2` in the SQL query. This is defense-in-depth at the database layer, not just application logic. All callers correctly pass `userID` from the authenticated request header.

3. **JTI auto-generation in `pkg/jwt`** -- `CreateAccess` and `CreateRefresh` now auto-generate UUIDs if `claims.ID` is empty. This ensures every token has a JTI for blacklist checking, eliminating a class of bypass where forgotten JTI assignment would make tokens non-revocable.

4. **Empty JTI rejection in auth middleware** -- `auth.go:72-74` rejects tokens with empty `claims.ID` when blacklist is configured. Combined with auto-gen in jwt.go, this creates a two-layer guarantee.

5. **FailOpen configurable** -- Auth service correctly sets `FailOpen: false` (line 162 of auth/cmd/main.go), meaning Redis failure = deny access. This is the secure default for the auth service. Other services don't configure a blacklist, so the check is skipped entirely (correct).

6. **`cachedModelConfig` excludes API key from Redis** -- Correct separation of concerns. API keys don't belong in a shared cache. The struct-level exclusion makes it impossible to accidentally cache the key via field addition.

7. **Agent WriteTimeout 5min** -- Appropriate for LLM streaming endpoints. Long enough for generation, short enough to prevent indefinite slowloris on the write side.

8. **Rate limiting placement** -- Auth login (5/min by IP), refresh (10/min by IP), MFA (5/min by IP) are all per-IP which is correct for unauthenticated endpoints. Agent and search (30/min by user) are per-user which is correct for authenticated endpoints where IP would be shared behind a proxy.

9. **writeJSONError uses json.Encoder** -- Eliminates the risk of manual JSON construction bugs. `json.NewEncoder(w).Encode()` is correct and slightly more efficient than `json.Marshal` + `w.Write`.

10. **Comprehensive test coverage** -- Both unit tests (handler mocks) and integration tests (testcontainers) exist for the ownership checks. The mock in `chat_test.go` correctly simulates owner/non-owner paths.

---

## Files reviewed

| File | Status |
|------|--------|
| `pkg/middleware/ratelimit.go` | NEW -- rate limiter implementation |
| `pkg/middleware/auth.go` | MODIFIED -- FailOpen, JTI check, writeJSONError |
| `pkg/jwt/jwt.go` | MODIFIED -- JTI auto-gen |
| `pkg/config/resolver.go` | MODIFIED -- cachedModelConfig |
| `pkg/go.mod` | x/time indirect (should be direct) |
| `services/auth/cmd/main.go` | MODIFIED -- rate limiters + FailOpen |
| `services/agent/cmd/main.go` | MODIFIED -- WriteTimeout 5min, rate limit |
| `services/search/cmd/main.go` | MODIFIED -- rate limit |
| `services/ws/cmd/main.go` | MODIFIED -- ReadHeaderTimeout + WriteTimeout |
| `services/chat/cmd/main.go` | MODIFIED -- ReadHeaderTimeout |
| `services/chat/internal/handler/chat.go` | MODIFIED -- ownership via service |
| `services/chat/internal/service/chat.go` | MODIFIED -- GetSession takes userID |
| `services/chat/internal/repository/chat.sql.go` | MODIFIED (sqlc gen) |
| `services/chat/db/queries/chat.sql` | MODIFIED -- user_id filters |
| `services/chat/internal/handler/chat_test.go` | **FAILS** -- system role sub-test |
| `services/chat/internal/service/chat_integration_test.go` | **FAILS** -- ErrNotOwner |
| `services/ingest/cmd/main.go` | MODIFIED -- ReadHeaderTimeout |
| `services/ingest/internal/handler/ingest.go` | MODIFIED -- ownership via service |
| `services/ingest/internal/service/ingest.go` | MODIFIED -- GetJob takes userID |
| `services/ingest/internal/repository/ingest.sql.go` | MODIFIED (sqlc gen) |
| `services/ingest/db/queries/ingest.sql` | MODIFIED -- user_id filters |
| `services/ingest/internal/handler/ingest_test.go` | OK -- tests pass |
| `services/ingest/internal/service/ingest_integration_test.go` | OK -- tests pass |
| `services/notification/cmd/main.go` | MODIFIED -- ReadHeaderTimeout |
| `services/platform/cmd/main.go` | MODIFIED -- ReadHeaderTimeout |
| `services/traces/cmd/main.go` | MODIFIED -- ReadHeaderTimeout |
| `services/feedback/cmd/main.go` | MODIFIED -- ReadHeaderTimeout |
