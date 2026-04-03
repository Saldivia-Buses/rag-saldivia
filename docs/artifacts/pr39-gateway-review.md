# Gateway Review -- PR #39 feat/notification-platform-tests

**Fecha:** 2026-04-02
**Tipo:** review
**PR:** #39 (feat/notification-platform-tests)
**Intensity:** thorough
**Reviewer:** gateway-reviewer (Opus)

## Archivos revisados

| Archivo | Cambio | Lineas |
|---|---|---|
| `services/notification/internal/handler/notification.go` | Modified -- added `NotificationService` interface, changed field type | 171 |
| `services/notification/internal/handler/notification_test.go` | New -- 11 test cases | 263 |
| `services/platform/internal/handler/platform.go` | Modified -- added `PlatformService` interface, added `context` import | 402 |
| `services/platform/internal/handler/platform_test.go` | New -- 13 test cases | 330 |

### Archivos de contexto (no modificados)

| Archivo | Motivo |
|---|---|
| `services/notification/internal/service/notification.go` | Concrete types (`Notification`, `Preferences`), error sentinels, method signatures |
| `services/notification/cmd/main.go` | Where `handler.NewNotification()` is called with `*service.NotificationService` |
| `services/platform/internal/service/platform.go` | Concrete types (`TenantDetail`, `FeatureFlag`, `ConfigEntry`), error sentinels |
| `services/platform/cmd/main.go` | Where `handler.NewPlatform()` is called with `*service.Platform` |
| `services/platform/db/` | sqlc generated types: `ListTenantsRow`, `CreateTenantParams`, `UpdateTenantParams`, `Module`, `GetEnabledModulesForTenantRow`, `EnableModuleForTenantParams` |
| `pkg/jwt/jwt.go` | `Claims`, `Config`, `DefaultConfig`, `CreateAccess`, `Verify` |

## Resultado

**APROBADO** (0 bloqueantes, 0 debe corregirse, 4 sugerencias)

---

## Hallazgos

### Bloqueantes

None.

### Compile correctness

Both interfaces are correctly defined. Verified each method signature against the concrete types.

**NotificationService interface** (notification.go:19-26) vs `*service.NotificationService`:

| Method | Interface | Concrete | Match |
|---|---|---|---|
| `List` | `(ctx, userID string, unreadOnly bool, limit int) ([]Notification, error)` | Same (service:79) | Yes |
| `UnreadCount` | `(ctx, userID string) (int, error)` | Same (service:112) | Yes |
| `MarkRead` | `(ctx, notifID, userID string) error` | Same (service:125) | Yes |
| `MarkAllRead` | `(ctx, userID string) (int64, error)` | Same (service:150) | Yes |
| `GetPreferences` | `(ctx, userID string) (*Preferences, error)` | Same (service:164) | Yes |
| `UpdatePreferences` | `(ctx, userID string, emailEnabled, inAppEnabled bool, quietStart, quietEnd *string, mutedTypes []string) (*Preferences, error)` | Same (service:199) | Yes |

The interface intentionally excludes `Create` (used only by the NATS consumer, not the HTTP handler). Correct -- principle of least privilege at the type level.

`cmd/main.go:79`: `handler.NewNotification(notifSvc)` where `notifSvc = service.New(pool)` returns `*service.NotificationService`. The concrete type satisfies all 6 interface methods. Compiles.

**PlatformService interface** (platform.go:22-37) vs `*service.Platform`:

| Method | Interface | Concrete | Match |
|---|---|---|---|
| `ListTenants` | `(ctx) ([]db.ListTenantsRow, error)` | Same (service:82) | Yes |
| `GetTenant` | `(ctx, slug string) (TenantDetail, error)` | Same (service:91) | Yes |
| `CreateTenant` | `(ctx, arg db.CreateTenantParams) (TenantDetail, error)` | Same (service:103) | Yes |
| `UpdateTenant` | `(ctx, arg db.UpdateTenantParams) error` | Same (service:119) | Yes |
| `DisableTenant` | `(ctx, id string) error` | Same (service:127) | Yes |
| `EnableTenant` | `(ctx, id string) error` | Same (service:135) | Yes |
| `ListModules` | `(ctx) ([]db.Module, error)` | Same (service:145) | Yes |
| `GetTenantModules` | `(ctx, tenantID string) ([]db.GetEnabledModulesForTenantRow, error)` | Same (service:154) | Yes |
| `EnableModule` | `(ctx, arg db.EnableModuleForTenantParams) error` | Same (service:163) | Yes |
| `DisableModule` | `(ctx, tenantID, moduleID string) error` | Same (service:171) | Yes |
| `ListFeatureFlags` | `(ctx) ([]FeatureFlag, error)` | Same (service:192) | Yes |
| `ToggleFeatureFlag` | `(ctx, id string, enabled bool) error` | Same (service:222) | Yes |
| `GetConfig` | `(ctx, key string) (ConfigEntry, error)` | Same (service:246) | Yes |
| `SetConfig` | `(ctx, key string, value []byte, updatedBy string) error` | Same (service:262) | Yes |

All 14 methods match. `cmd/main.go:48`: `handler.NewPlatform(platformSvc, jwtSecret)` where `platformSvc = service.New(pool)` returns `*service.Platform`. Compiles.

### JWT token helpers (platform_test.go:114-138)

The `adminToken` and `userToken` helpers use `pkg/jwt` correctly:

- **Secret:** `testSecret = "test-secret-at-least-32-chars-long!!"` (36 bytes) -- passes the `len < 32` guard in `CreateAccess`.
- **Config:** `sdajwt.DefaultConfig(testSecret)` produces 15-minute expiry, issuer `"sda"`, HS256 signing. Tests will not expire mid-run.
- **Admin token claims:** `UserID: "u-admin"`, `TenantID: "platform"`, `Slug: "platform"`, `Role: "admin"`. All required fields populated -- passes `Verify()` validation (jwt.go:105 checks `UserID`, `TenantID`, `Slug` non-empty).
- **User token claims:** `UserID: "u-user"`, `TenantID: "t-1"`, `Slug: "saldivia"`, `Role: "user"`. Same validation passes, but `requirePlatformAdmin` correctly rejects `Role != "admin"`.
- **Signing method:** `gojwt.SigningMethodHS256` hardcoded in `CreateAccess` (jwt.go:65). `Verify` rejects non-HMAC methods (jwt.go:91-93). Immune to `alg: none` attacks.

Token helpers are correct and secure.

### Debe corregirse

None. Unlike PR #38, this PR already includes 500 error path tests for both services:
- `TestList_ServiceError_Returns500_GenericMessage` (notification_test.go:245) -- sets `err: errors.New("database down")`, verifies 500 + generic message
- `TestListTenants_ServiceError_Returns500` (platform_test.go:312) -- sets `err: errors.New("database exploded")`, verifies 500 + generic message

Both confirm that `serverError()` does not leak internal error details. This directly addresses D1 from PR #38's review.

### Sugerencias

**S1. [platform_test.go] Many handler paths have no dedicated test**

The 13 platform tests cover auth middleware (4 tests), tenant CRUD (5 tests), flags (1 test), config (1 test), and error handling (1 test + the 500 test). The following handlers have no direct test:

| Handler | Status |
|---|---|
| `UpdateTenant` | Untested (happy path + validation) |
| `DisableTenant` | Untested |
| `EnableTenant` | Untested |
| `ListModules` | Untested |
| `GetTenantModules` | Untested |
| `EnableModule` | Untested (happy path + missing module_id = 400) |
| `DisableModule` | Untested |
| `SetConfig` | Untested (happy path + invalid body = 400) |
| `ListFeatureFlags` | Untested (happy path) |
| `ToggleFeatureFlag` | Happy path untested (only not-found tested) |
| `GetConfig` | Happy path untested (only not-found tested) |

These are simple passthrough handlers (read param, call service, write response), so the risk is low. But `EnableModule` has validation logic (empty `module_id` check at platform.go:250-252) and `SetConfig` has body parsing -- those would benefit from tests most.

**S2. [notification_test.go] `MarkRead` 500 error path not tested**

The mock's `MarkRead` (notification_test.go:42-51) has dedicated logic that checks the notifications slice and returns `ErrNotificationNotFound` when the ID is not found. However, no test exercises the case where `m.err` is set (a generic service failure like a database timeout). The `TestList_ServiceError` test covers the 500 pattern for `List`, but if someone changes the `MarkRead` handler's error handling, only a `MarkRead`-specific 500 test would catch a regression.

Low priority since the `serverError()` codepath is shared and already proven by `TestList_ServiceError`.

**S3. [notification_test.go:42-51] Mock `MarkRead` does not model the "already read" no-op**

The concrete service (notification.go:125-147) has three behaviors:
1. Notification exists, unread -> mark read, return nil
2. Notification exists, already read -> no-op, return nil
3. Notification not found for this user -> return `ErrNotificationNotFound`

The mock only models cases 1 and 3 (checks exact ID+UserID match in slice). Case 2 (already-read no-op) is not distinguishable from case 1 in the mock. This is harmless since the handler does not differentiate between them (both return 204), but a comment would make this intentional:

```go
// MarkRead: returns nil if found (regardless of read status), ErrNotificationNotFound otherwise.
// Concrete service also returns nil for already-read notifications (no-op).
```

**S4. [platform_test.go:114-138] Tokens generated per test invocation**

Both `adminToken(t)` and `userToken(t)` create a new JWT on every call, and `withAdminAuth` calls `adminToken(t)` each time. In the current 13 tests, this creates ~10 tokens. This is functionally fine and tests are fast, but if the test suite grows, consider caching the token in a `sync.Once` or package-level `TestMain`:

```go
var cachedAdminToken string

func TestMain(m *testing.M) {
    cfg := sdajwt.DefaultConfig(testSecret)
    var err error
    cachedAdminToken, err = sdajwt.CreateAccess(cfg, sdajwt.Claims{...})
    if err != nil {
        panic(err)
    }
    os.Exit(m.Run())
}
```

Very low priority -- only matters if you end up with 100+ platform tests.

### Lo que esta bien

1. **Interface placement follows Go idiom.** Both `NotificationService` and `PlatformService` are defined in the consumer package (`handler/`), not the provider package (`service/`). The consumer defines what it needs, the provider implicitly satisfies. Consistent with the pattern established in PR #38.

2. **NotificationService correctly excludes `Create`.** The `Create` method is only called by the NATS consumer (`service.NewConsumer`), not the HTTP handler. The interface enforces this boundary at compile time.

3. **PlatformService interface is comprehensive (14 methods) but still minimal.** Every method in the interface is called by at least one handler. No dead methods in the interface.

4. **Platform JWT middleware tests are thorough.** Four auth tests cover: (1) no token -> 401, (2) invalid token -> 401, (3) valid user-role token -> 403, (4) valid admin token -> 200. This is the minimum necessary set and all four are present.

5. **500 error tests included from day one.** Both test files include a test that injects a generic `errors.New(...)` into the mock and verifies the handler returns 500 with `"internal error"` -- no stack trace or internal detail leakage. This was the main finding in PR #38's review (D1), and it has been proactively addressed here.

6. **Tenant creation validation is well-tested.** Three error scenarios: missing required fields (400), invalid slug via `ErrInvalidSlug` (400), duplicate slug via `ErrSlugTaken` (409). Plus the happy path (201). All four HTTP status codes are asserted.

7. **Real JWT generation in tests, not mocked.** Platform tests use the actual `pkg/jwt` package to generate tokens, verifying the full middleware chain (header parsing -> JWT verification -> role check -> header propagation). This catches integration issues that a mocked JWT parser would miss.

8. **`requirePlatformAdmin` propagates identity correctly.** After JWT verification, the middleware sets `r.Header.Set("X-User-ID", claims.UserID)` (platform.go:386), making the admin's identity available to downstream handlers like `EnableModule` and `SetConfig`. The `withAdminAuth` test helper mirrors the real flow.

9. **Error messages are safe.** All client-facing errors use generic descriptions: `"missing authorization"`, `"invalid token"`, `"platform admin access required"`, `"tenant not found"`, `"internal error"`. No JWT details, no SQL errors, no stack traces.

10. **Both services use `http.MaxBytesReader` on body-accepting endpoints.** Notification: `UpdatePreferences` (notification.go:133). Platform: `CreateTenant`, `UpdateTenant`, `EnableModule`, `ToggleFeatureFlag`, `SetConfig`. All capped at 1MB. Consistent and safe.

11. **Notification `requireUserID` middleware is clean.** A single middleware applied to all routes (notification.go:41) that returns 401 if `X-User-ID` is empty. Tested by `TestList_MissingUserID_Returns401`. This is the correct pattern for services behind the gateway auth middleware.

---

## Coverage summary

### Notification service

| Handler | Happy path | Error path | Ownership | Validation | 500 error |
|---|---|---|---|---|---|
| List | Tested | -- | Filtered by userID | -- | Tested |
| UnreadCount | Tested | -- | Filtered by userID | -- | -- |
| MarkRead | Tested | Not found = 404 | Owner match in mock | -- | -- |
| MarkAllRead | Tested | -- | Filtered by userID | -- | -- |
| GetPreferences | Tested | -- | Filtered by userID | -- | -- |
| UpdatePreferences | Tested | Invalid JSON = 400 | Filtered by userID | -- | -- |
| requireUserID | -- | Missing = 401 | -- | -- | -- |

### Platform service

| Handler | Happy path | Error path | Auth | Validation | 500 error |
|---|---|---|---|---|---|
| requirePlatformAdmin | Admin = 200 | No token = 401, invalid = 401 | User role = 403 | -- | -- |
| ListTenants | Tested (via auth test) | -- | -- | -- | Tested |
| GetTenant | -- | Not found = 404 | -- | -- | -- |
| CreateTenant | Tested | Slug invalid = 400, taken = 409 | -- | Missing fields = 400 | -- |
| UpdateTenant | **Not tested** | -- | -- | -- | -- |
| DisableTenant | **Not tested** | -- | -- | -- | -- |
| EnableTenant | **Not tested** | -- | -- | -- | -- |
| ListModules | **Not tested** | -- | -- | -- | -- |
| GetTenantModules | **Not tested** | -- | -- | -- | -- |
| EnableModule | **Not tested** | -- | -- | -- | -- |
| DisableModule | **Not tested** | -- | -- | -- | -- |
| ListFeatureFlags | **Not tested** | -- | -- | -- | -- |
| ToggleFeatureFlag | -- | Not found = 404 | -- | -- | -- |
| GetConfig | -- | Not found = 404 | -- | -- | -- |
| SetConfig | **Not tested** | -- | -- | -- | -- |

Notification coverage is solid (all 6 handlers + auth middleware). Platform has good coverage on the critical paths (auth, tenant creation, error handling) but 9 out of 14 handlers lack a happy-path test. These are low-risk passthrough handlers, but completing coverage would be a worthwhile follow-up.
