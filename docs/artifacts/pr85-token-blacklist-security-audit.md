# Security Audit -- PR #85 Token Blacklist -- 2026-04-05

## Resumen ejecutivo

PR #85 implements a Redis-backed JWT token blacklist for logout/revocation.
The primitives (`pkg/security/blacklist.go`) are correct and well-tested, but
the wiring is incomplete: **no service actually activates the blacklist**, making
the entire feature dead code in production. There are also two design-level
concerns around fail-open behavior and expired-token edge cases that must be
resolved before merge.

---

## CRITICOS (bloquean merge)

### C1. Blacklist is never activated -- entire feature is dead code

**Files:**
- `services/auth/cmd/main.go` -- no call to `SetBlacklist()`, no Redis client created
- `services/auth/internal/handler/auth.go:84` -- `resolveService()` calls `service.NewAuth()` without `SetBlacklist()`
- `services/auth/cmd/main.go:139` -- uses `sdamw.Auth(publicKey)`, not `AuthWithConfig`

**Problem:** Three disconnected pieces prevent the blacklist from functioning:

1. **Auth service main.go** never creates a Redis client, never calls
   `NewTokenBlacklist()`, never calls `SetBlacklist()` on the service. In
   single-tenant mode (line 116-117), `authSvc` is created without blacklist.
   In multi-tenant mode (line 98), the handler's `resolveService()` creates a
   fresh `service.NewAuth()` per request (line 84 of handler) -- also without
   blacklist.

2. **Auth middleware** uses `sdamw.Auth(publicKey)` (line 139 of main.go), which
   calls `AuthWithConfig(publicKey, AuthConfig{})` with a nil `Blacklist`. The
   `IsRevoked` check (middleware line 69) is guarded by `cfg.Blacklist != nil`,
   so it is never executed.

3. **No other service** uses `AuthWithConfig` either. Every service in the repo
   uses `sdamw.Auth(publicKey)`.

**Impact:** A user can log out, the refresh token is revoked in PostgreSQL, but
the access token remains valid for up to 15 minutes. The PR creates the illusion
of token revocation without delivering it.

**Fix:**
- Auth service main.go must create a Redis client (platform Redis or tenant Redis),
  instantiate `NewTokenBlacklist(rdb)`, and either:
  - (a) Call `SetBlacklist(bl)` on the auth service AND pass `AuthConfig{Blacklist: bl}`
    to `AuthWithConfig` for protected routes, OR
  - (b) In multi-tenant mode, the handler needs access to the tenant's Redis client
    via the Resolver to call `SetBlacklist()` on each resolved service instance.
- All services with protected routes must switch from `sdamw.Auth(publicKey)` to
  `sdamw.AuthWithConfig(publicKey, AuthConfig{Blacklist: bl})`. Otherwise a revoked
  token is only rejected by the auth service, not by chat/rag/notification/etc.

### C2. Logout silently fails to blacklist when access token is expired

**File:** `services/auth/internal/handler/auth.go:226-232`

```go
if claims, err := sdajwt.Verify(h.jwtCfg.PublicKey, bearer[7:]); err == nil {
    accessJTI = claims.ID
    ...
}
```

**Problem:** `sdajwt.Verify()` rejects expired tokens (golang-jwt checks `exp`
claim by default). If a user's access token expired 30 seconds before they click
logout, `Verify` returns an error, `accessJTI` stays empty, and the service's
`Logout()` method skips the blacklist call entirely (line 473: `if accessJTI != ""`).

This is actually a lesser concern because an expired token is already rejected.
However, the default fallback expiry `time.Now().Add(15 * time.Minute)` (line 225)
is never used in this path -- it would only apply if `Verify` succeeds but
`ExpiresAt` is nil, which cannot happen with the current `CreateAccess` code.
So the fallback is dead code, but harmless.

More importantly: if `Verify` fails for any reason other than expiry (e.g.,
malformed token, wrong key), the blacklist is also silently skipped. This is
acceptable because a malformed token was never valid, but the distinction is
not logged.

**Severity reduced to ALTO** because the practical exploit window is narrow
(token is already expired), but the pattern is fragile.

**Fix:** Parse the token without expiry validation to extract the JTI, or use
`jwt.ParseUnverified` to get the claims for blacklisting purposes. The token's
signature should still be verified (to prevent arbitrary JTI injection), but
expiry should be ignored during logout.

---

## ALTOS (corregir antes de produccion)

### A1. Fail-open on Redis failure allows revoked tokens through

**File:** `pkg/middleware/auth.go:71-73`

```go
if err != nil {
    slog.Error("blacklist check failed", "error", err)
    // fail-open: if Redis is down, don't block all requests
}
```

**Problem:** When Redis is unreachable, all requests pass through including
revoked tokens. An attacker who can take down Redis (or trigger network
partition between app and Redis) can use revoked tokens indefinitely.

**Context:** Fail-open is a defensible choice for availability-first systems.
But the bible says "La seguridad no es un tradeoff. Es una restriccion." This
warrants fail-closed.

**Tradeoff to discuss:** Fail-closed means Redis downtime = service outage for
all authenticated requests. Mitigation: circuit breaker with short timeout (50ms)
so a Redis blip doesn't cause cascading failures, but sustained outage does
block requests.

**Fix:** Change to fail-closed, OR implement a circuit breaker with configurable
behavior (`BLACKLIST_FAIL_MODE=closed|open`), defaulting to closed.

### A2. Blacklist key prefix has no tenant namespace -- cross-tenant collision risk

**File:** `pkg/security/blacklist.go:26`

```go
prefix: "sda:token:blacklist:",
```

**Problem:** The key is `sda:token:blacklist:{jti}`. JTIs are UUIDs (extremely
low collision risk), but the design assumes all tenants share one Redis or that
each tenant has their own Redis instance. In the current architecture, Redis is
per-tenant, so this is safe today. However:

- If the platform Redis is used for the blacklist (likely, since the auth service
  doesn't currently connect to tenant Redis), all tenants share one key space.
  Tenant A could theoretically forge a JTI to un-blacklist Tenant B's token
  (impossible because JTIs are random UUIDs, but the defense-in-depth principle
  suggests adding the tenant slug to the prefix).
- The `RevokeAll` method takes bare JTIs with no tenant scoping. A bug in the
  caller could revoke tokens across tenants.

**Fix:** Change prefix to `sda:token:blacklist:{tenantSlug}:{jti}` or accept
the risk with a documented ADR.

### A3. RevokeAll not wired to password change flow

**File:** `pkg/security/blacklist.go:50` (method exists), plan07 Phase 5 (line 277)

**Problem:** Plan 07 explicitly requires: "Password change: llamar
`blacklist.RevokeAll()` para invalidar todos los tokens del usuario." The method
exists but is not called anywhere. There is no password change endpoint yet, but
this should be tracked as a blocker for that feature.

**Fix:** Wire `RevokeAll` into the password change flow when it is implemented.
Add a TODO or tracking issue.

### A4. Multi-tenant resolveService creates ephemeral auth services without blacklist

**File:** `services/auth/internal/handler/auth.go:84`

```go
return service.NewAuth(pool, h.jwtCfg, tenantID, slug, h.publisher), nil
```

**Problem:** In multi-tenant mode, every request creates a new `service.Auth`
instance. `SetBlacklist()` is never called on these ephemeral instances. Even
if `SetBlacklist` were called in `main.go` for the single-tenant case, the
multi-tenant path would still skip it.

**Fix:** Either:
- Pass the `*TokenBlacklist` into `NewMultiTenantAuth()` so `resolveService()`
  can call `SetBlacklist()` on each instance, OR
- Add `blacklist` as a parameter to `service.NewAuth()` directly, OR
- Have the handler's `resolveService()` resolve the tenant's Redis client via
  `resolver.RedisClient()` and create a per-tenant blacklist.

---

## MEDIOS (backlog prioritario)

### M1. writeJSONError in middleware uses string concatenation -- JSON injection vector

**File:** `pkg/middleware/auth.go:125`

```go
w.Write([]byte(`{"error":"` + msg + `"}`))
```

**Problem:** If `msg` contains a double quote or backslash, the JSON output
is malformed. Currently all callers pass string literals, so this is not
exploitable today. But if a future caller passes user-derived input (e.g., an
error message), it becomes a JSON injection vector.

**Fix:** Use `json.Marshal` or `json.NewEncoder`:
```go
json.NewEncoder(w).Encode(map[string]string{"error": msg})
```

### M2. Test cleanup calls FlushDB on shared Redis

**File:** `pkg/security/blacklist_test.go:24`

```go
t.Cleanup(func() { rdb.FlushDB(ctx); rdb.Close() })
```

**Problem:** `FlushDB` wipes all keys in the Redis database, not just test keys.
If tests run against a shared Redis instance (CI, dev), this destroys other
data. Not a security vulnerability per se, but demonstrates the exact attack
vector from checklist item 3: if application code (or test code) can call
`FlushDB`, so can an attacker with Redis access.

**Fix:** Use `DEL` with a pattern scan for the test prefix, or use a dedicated
Redis database number for tests (`SELECT 15`).

### M3. Logout handler silently swallows all errors

**File:** `services/auth/internal/handler/auth.go:234`

```go
_ = svc.Logout(r.Context(), refreshToken, accessJTI, accessExpiry)
```

**Problem:** If `Logout()` fails (DB error, Redis error), the user sees
`{"status":"logged_out"}` but their tokens are not actually revoked. The
cookie is cleared so the browser forgets the refresh token, but the access
token and the server-side refresh token remain active.

**Fix:** Log the error at minimum. Consider returning a 500 if the refresh
token revocation fails (the critical path). The blacklist failure is already
logged inside `Logout()` as non-fatal.

### M4. No rate limiting on logout endpoint

**File:** `services/auth/cmd/main.go:135`

```go
r.Post("/v1/auth/logout", authHandler.Logout)
```

**Problem:** The logout endpoint is unauthenticated (no middleware). An attacker
could send millions of logout requests with random refresh tokens, causing:
- DB load from `GetRefreshTokenOwner` and `RevokeRefreshTokenByHash` queries
- Redis load from `blacklist.Revoke` calls

**Fix:** Add rate limiting on `/v1/auth/logout` or require the auth middleware.

---

## BAJOS (nice to have)

### B1. TTL rounding in Revoke could miss the last second

**File:** `pkg/security/blacklist.go:33-34`

```go
ttl := time.Until(expiresAt)
if ttl <= 0 { return nil }
```

**Problem:** If the token expires in 500ms, the Redis key TTL is set to ~500ms.
Due to clock skew between the app server and Redis, the key could expire slightly
before the token does. The window is sub-second and practically unexploitable.

**Fix:** Add a small buffer: `ttl := time.Until(expiresAt) + 30*time.Second`.

### B2. Logout should also accept the access token JTI from a request body field

**Problem:** CLI and MCP clients send the refresh token in the body but the
access token JTI is only extracted from the `Authorization` header. If a client
does not send the header (e.g., sends JTI as a body field), blacklisting is
silently skipped.

**Fix:** Accept `access_jti` in the request body as a fallback.

---

## Tenant isolation audit

**Verdict: ACCEPTABLE with caveats.**

- The blacklist prefix `sda:token:blacklist:{jti}` does not include tenant
  scoping. Since JTIs are random UUIDs, cross-tenant collision is statistically
  impossible, but it violates defense-in-depth (see A2).
- The `resolveService` in multi-tenant mode correctly resolves the tenant's
  PostgreSQL pool, so refresh token operations (revoke, lookup) are isolated.
- The Redis instance for the blacklist has not been decided yet (platform Redis
  vs. tenant Redis). If platform Redis is used, all tenants share one blacklist
  namespace -- acceptable only if tenant slug is added to the key prefix.

---

## Missing security pieces (from spec/plan that should exist)

1. **Blacklist wiring** -- the entire point of this PR is not connected (C1)
2. **Password change revocation** -- `RevokeAll` exists but is not wired (A3)
3. **Integration test** -- Plan 07 Phase 5 requires "login -> token -> logout ->
   use token -> 401" test. Not present in this PR.
4. **`AuthWithConfig` adoption** -- no service uses it, so even if auth service
   checks the blacklist, chat/rag/notification/etc. would still accept revoked tokens.

---

## CVEs

| Dependency | Version | Known CVEs |
|---|---|---|
| golang-jwt/jwt/v5 | v5.3.1 | None known |
| redis/go-redis/v9 | v9.18.0 | None known |
| jackc/pgx/v5 | v5.9.1 | None known |
| golang.org/x/crypto | v0.49.0 | None known |

All dependencies are recent versions with no known CVEs as of 2026-04-05.

---

## Veredicto: NO APTO para merge

**Razon principal:** The PR introduces blacklist primitives and modifies the
middleware and service interfaces to support them, but does not connect any of
the pieces. A user logging out today gets no benefit from this code -- the
access token remains valid for its full 15-minute lifetime. This creates a false
sense of security.

**To become APTO:**

1. **[C1]** Wire the blacklist end-to-end: Redis client in auth main.go,
   `SetBlacklist()` called, `AuthWithConfig` used in routing.
2. **[A1]** Decide and document fail-open vs fail-closed behavior.
3. **[A4]** Multi-tenant `resolveService()` must inject blacklist.
4. **[M3]** Do not swallow `Logout()` errors silently.
5. Add the integration test specified in Plan 07 Phase 5.
