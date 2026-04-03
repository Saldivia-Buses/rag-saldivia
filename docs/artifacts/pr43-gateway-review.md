# Gateway Review -- PR #43 Expand Test Coverage

**Fecha:** 2026-04-02
**Tipo:** review
**Intensity:** quick
**Branch:** feat/expand-test-coverage
**Reviewer:** gateway-reviewer (Opus)

## Resultado

**CAMBIOS REQUERIDOS** (1 correctness bug, no blockers)

## Files Reviewed

| File | Verdict |
|---|---|
| `services/auth/internal/handler/auth_test.go` | OK |
| `services/ws/internal/handler/ws_test.go` | OK |
| `services/platform/internal/service/platform_integration_test.go` | 1 issue |

---

## Hallazgos

### Bloqueantes

(none)

### Debe corregirse

#### 1. `TestDisableTenant_and_EnableTenant_Integration` passes for the wrong reason

**File:** `services/platform/internal/service/platform_integration_test.go:358-365`

After `DisableTenant`, the test calls `svc.GetTenant(ctx, "toggleme")` and checks `got.Enabled == false`. But `GetTenant` internally calls `GetTenantBySlug` which has a SQL filter `WHERE slug = $1 AND enabled = true` (see `services/platform/db/tenants.sql.go:78`). This means:

- After disabling, `GetTenantBySlug` returns `pgx.ErrNoRows`
- `GetTenant` maps that to `ErrTenantNotFound`
- The error is silently discarded (`got, _ := svc.GetTenant(...)`)
- `got` is a zero-value `TenantDetail{Enabled: false}`
- The assertion `if got.Enabled` passes -- but because of the zero value, not because it read the actual disabled tenant

The test is green but is not actually testing what it claims. If `DisableTenant` silently did nothing, the test would still pass.

**Fix:** Either:

**(a)** Assert the error behavior directly -- after disabling, expect `GetTenant` to return `ErrTenantNotFound` (which is the actual current behavior since the query filters by `enabled = true`):

```go
// Disable
if err := svc.DisableTenant(ctx, tenant.ID); err != nil {
    t.Fatalf("disable: %v", err)
}
_, err := svc.GetTenant(ctx, "toggleme")
if err != ErrTenantNotFound {
    t.Fatalf("expected disabled tenant to be hidden from GetTenant, got err: %v", err)
}
```

**(b)** Query the DB directly to verify the `enabled` column changed:

```go
var enabled bool
pool.QueryRow(ctx, `SELECT enabled FROM tenants WHERE id = $1`, tenant.ID).Scan(&enabled)
if enabled {
    t.Error("expected enabled=false after DisableTenant")
}
```

Option (a) is preferred because it also documents the expected API behavior (disabled tenants are invisible via `GetTenant`). Option (b) is a good addition on top of (a) if you want to verify the DB state too.

### Sugerencias

#### 1. Auth test: assert `Decode` errors are not silently swallowed

In `TestLogin_Success` (line 55) and `TestLogin_InvalidCredentials_Returns401` (line 124), `json.NewDecoder(rec.Body).Decode(...)` errors are silently ignored. Low risk since test failures would surface as wrong field values anyway, but for rigor:

```go
if err := json.NewDecoder(rec.Body).Decode(&tokens); err != nil {
    t.Fatalf("decode response: %v", err)
}
```

#### 2. WS tests: consider testing query-param token rejection

`ws.go:33` explicitly only accepts `Authorization: Bearer` header (comment says "not query param, to avoid log leakage"). A test that sends `?token=xxx` and confirms it is rejected (401) would lock in this security decision. Not required for this PR but good to track.

#### 3. Platform integration tests: unchecked errors on setup calls

Several tests call `svc.CreateTenant(...)` during setup and discard the error (e.g., lines 165, 207-213, 232, 264, 295-301). If a setup call fails silently, subsequent assertions may produce confusing failures. Consider:

```go
tenant, err := svc.CreateTenant(ctx, ...)
if err != nil {
    t.Fatalf("setup: create tenant: %v", err)
}
```

This is a style nit -- the existing tests that are the "subject under test" do check errors properly.

---

## Lo que esta bien

- **Auth `TestLogin_PropagatesIPAndUserAgent`:** Solid test. Verifies all three fields (IP, UserAgent, Email) are forwarded to the service layer. The mock enhancement (`lastReq` field) is clean and minimal.

- **WS Upgrade tests:** Correctly test the two main rejection paths (no token, invalid token) without needing a real WebSocket upgrade. The direct struct init `&WS{jwtSecret: ...}` with a nil hub is a smart shortcut -- the handler rejects before touching the hub.

- **Platform integration tests cover the full lifecycle:** CreateTenant, UpdateTenant, DisableTenant, EnableTenant, ToggleFeatureFlag, ListModules, SetConfig, GetConfig. Good coverage of both happy path and error cases (duplicate slug, invalid slug, not found).

- **Integration tests use testcontainers correctly:** Real PostgreSQL 16-alpine, proper cleanup with `defer`, and the migration seed data matches the sqlc schema. The 30-second startup timeout with 2-occurrence wait is the right pattern for PostgreSQL readiness.

- **Test naming follows Go conventions:** `TestX_Y_Integration` suffix makes it clear which tests need Docker/`-tags=integration`. Table-driven tests where appropriate (`TestParseOrigins`, `TestExtractBearerToken`, `TestCreateTenant_InvalidSlug_Integration`).

- **No security concerns in test code:** No real secrets, no hardcoded connection strings leaking to logs. Test JWT secrets are clearly fake values.
