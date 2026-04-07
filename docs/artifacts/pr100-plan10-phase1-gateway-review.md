# Gateway Review — PR #100 Plan 10 Phase 1: Blacklist Wiring

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### 1. auth multi-tenant mode: blacklist never wired
**`services/auth/cmd/main.go:107`**

In the `platformDBURL != ""` branch, `NewMultiTenantAuth(resolver, jwtCfg, publisher)` is called without passing `blacklist`. In the single-tenant branch (line 126), `authSvc.SetBlacklist(blacklist)` is called correctly. Multi-tenant auth resolves per-request services via `resolveService()`, which calls `service.NewAuth(...)` — that service is never given the blacklist.

This means any logout in multi-tenant mode (production path) does not actually block revoked tokens.

Fix: add `blacklist` to `NewMultiTenantAuth` and thread it into each resolved `AuthService`:

```go
// handler/auth.go
func NewMultiTenantAuth(resolver *tenant.Resolver, jwtCfg sdajwt.Config, publisher EventPublisher, blacklist *security.TokenBlacklist) *Auth {
    return &Auth{resolver: resolver, jwtCfg: jwtCfg, publisher: publisher, blacklist: blacklist}
}

// in resolveService(), after creating svc:
svc := service.NewAuth(pool, h.jwtCfg, tenantID, slug, h.publisher)
svc.SetBlacklist(h.blacklist)
```

And in `cmd/main.go`:
```go
authHandler = handler.NewMultiTenantAuth(resolver, jwtCfg, publisher, blacklist)
```

---

### 2. auth Redis client not closed in multi-tenant path
**`services/auth/cmd/main.go:71-77`**

When `platformDBURL != ""`, the code creates `rdb` and potentially reaches the `rdb.Ping` check, but never calls `defer rdb.Close()` in that branch — the `defer` is only registered if `rdb != nil` after Ping inside the `else` block. Actually the structure is:

```go
rdb := redis.NewClient(...)
if err := rdb.Ping(ctx).Err(); err != nil {
    slog.Warn(...)
    rdb = nil     // ← no Close() called on the failed client
} else {
    defer rdb.Close()
}
```

A failed Ping leaves the `redis.NewClient` connection pool open (it leaks the underlying TCP connection attempt). The fix is to call `rdb.Close()` on the failure path, or use `InitBlacklist` (which this PR introduces) consistently in auth too:

```go
// Replace the auth Redis block entirely with:
blacklist = security.InitBlacklist(ctx, redisURL)
```

This also removes the duplicated inline init logic vs `pkg/security/init.go`.

---

## Debe corregirse

### 3. `platform.requirePlatformAdmin` only sets `X-User-ID` after verify — missing other identity headers
**`services/platform/internal/handler/platform.go:408`**

After JWT verification, the middleware re-injects only `X-User-ID`. Handlers `EnableModule` (line 248) and `SetConfig` (line 353) read `X-User-ID` for audit logging — that works. But the pattern is inconsistent with `pkg/middleware/auth.go` which also injects `X-User-Email`, `X-User-Role`, `X-Tenant-ID`, `X-Tenant-Slug`. If any future platform handler reads those headers, they'll be empty strings from spoofed requests that were stripped. Add the full set for consistency:

```go
r.Header.Set("X-User-ID", claims.UserID)
r.Header.Set("X-User-Email", claims.Email)
r.Header.Set("X-User-Role", claims.Role)
r.Header.Set("X-Tenant-ID", claims.TenantID)
r.Header.Set("X-Tenant-Slug", claims.Slug)
```

### 4. `FailOpen: true` in `feedback` platform admin route
**`services/feedback/cmd/main.go:148`**

The `/v1/platform/feedback` route group uses `AuthWithConfig{FailOpen: true}`. Platform-facing routes should fail closed (`FailOpen: false`) — if the blacklist is unavailable, a revoked admin token should still be rejected if the JWT itself is valid but on the revoke list. Using `FailOpen: true` means a Redis outage leaves revoked tokens fully operative on platform routes.

The tenant-scoped `/v1/feedback` route (line 141) is fine with `FailOpen: true`.

---

## Sugerencias

- `pkg/security/init.go` does `redis.NewClient` with only `Addr`. If `REDIS_URL` is a full `redis://` URL (which is standard), `redis.Options{Addr: redisURL}` will fail silently (Ping will error, blacklist disabled). Consider using `redis.ParseURL(redisURL)` and falling back to treating the value as a bare host:port. Alternatively document that `REDIS_URL` must be `host:port` format (not the URL scheme).

- The auth service has its own inline Redis init (blocker #2 above), which duplicates `InitBlacklist`. After fixing, auth should use `security.InitBlacklist` for consistency. This also means auth gets the same graceful-degradation log message as every other service.

---

## Lo que está bien

- `InitBlacklist` returns nil on both empty URL and Ping failure — callers never need to nil-check twice. Clean design.
- All 7 new services call `InitBlacklist` unconditionally at the top of `main()`, before any other resource setup. Correct position (not inside an if/else branch).
- `requirePlatformAdmin` strips all 5 identity headers before parsing the token (lines 375-379). Order is correct: strip → verify → re-inject.
- Blacklist check in `requirePlatformAdmin` correctly guards on `claims.ID != ""` before calling `IsRevoked` — avoids spurious Redis calls on tokens without JTI.
- `FailOpen: true` is correct for the 7 tenant services — a Redis outage should not take down chat/search/agent/etc.
- auth single-tenant path (`SetBlacklist`) and gRPC interceptor config (search) both wired correctly.
- Platform `NewPlatform` 4th param (`blacklist`) added cleanly; test updated with `nil` blacklist — correct for unit tests.
- Coverage: ws correctly excluded (does JWT inline in upgrade handler), auth already had it. All 10 services accounted for.
