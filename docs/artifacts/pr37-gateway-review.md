# Gateway Review -- PR #37 feat/shared-pkg-tests

**Fecha:** 2026-04-02
**Tipo:** review
**PR:** #37 (feat/shared-pkg-tests)
**Intensity:** thorough
**Reviewer:** gateway-reviewer (Opus)

## Archivos revisados

| Archivo | Cambio | Lineas |
|---|---|---|
| `pkg/middleware/auth_test.go` | New -- 9 test cases | 220 |
| `pkg/nats/publisher_test.go` | New -- 6 test groups | 193 |

### Archivos de contexto (no modificados)

| Archivo | Motivo |
|---|---|
| `pkg/middleware/auth.go` | Production code under test |
| `pkg/nats/publisher.go` | Production code under test |
| `pkg/jwt/jwt.go` | JWT creation/verification used by auth tests |
| `pkg/jwt/jwt_test.go` | Existing test patterns to compare against |
| `pkg/tenant/context.go` | Tenant context used by middleware |

## Resultado

**CAMBIOS REQUERIDOS** (1 bloqueante, 2 debe corregirse, 6 sugerencias)

---

## Hallazgos

### Bloqueantes

**B1. [pkg/middleware/auth_test.go:35] Compile error: `tenant.FromContext` returns two values, only one assigned**

The `echoHandler` helper calls:

```go
info := tenant.FromContext(r.Context())
```

But `tenant.FromContext` (context.go:34) returns `(Info, error)`. This is a compile error -- the test file will not build. Every test case that uses `echoHandler` is affected (5 of the 9 tests).

**Fix:** Handle both return values:

```go
info, err := tenant.FromContext(r.Context())
if err != nil {
    http.Error(w, "no tenant in context", http.StatusInternalServerError)
    return
}
```

Alternatively, use `tenant.SlugFromContext(r.Context())` which returns a single string (but panics if no tenant -- acceptable in a test handler that should always have one).

**Severity:** Bloqueante. The test file does not compile. `go test ./pkg/middleware/...` will fail with a build error before any test runs.

---

### Debe corregirse

**D1. [pkg/middleware/auth_test.go] Missing test: expired token should return 401**

The JWT package tests include `TestVerify_ExpiredToken` (jwt_test.go:90), but the middleware tests do not have an equivalent. The middleware is the enforcement point -- it is important to verify that an expired token actually produces a 401 at the HTTP level, not just that `jwt.Verify` returns an error.

This is a real risk because the middleware could, in theory, catch the error but mishandle it (e.g., fall through to `next.ServeHTTP`).

**Fix:** Add a test case:

```go
func TestAuth_ExpiredToken_Returns401(t *testing.T) {
    t.Helper()
    cfg := sdajwt.Config{
        Secret:       testSecret,
        AccessExpiry: -1 * time.Hour, // already expired
        Issuer:       "sda",
    }
    token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
        UserID:   "u-123",
        TenantID: "t-456",
        Slug:     "saldivia",
        Role:     "admin",
    })
    if err != nil {
        t.Fatalf("create expired token: %v", err)
    }

    handler := Auth(testSecret)(echoHandler())
    req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401 for expired token, got %d", rec.Code)
    }
}
```

This requires adding `"time"` to the imports.

---

**D2. [pkg/nats/publisher_test.go] Missing tests: `Broadcast` channel parameter is not validated, and no test covers this gap**

The production code (`publisher.go:82`) concatenates the `channel` parameter directly into the NATS subject without validation:

```go
subject := "tenant." + tenantSlug + "." + channel
```

While `tenantSlug` is validated with `isValidSubjectToken`, `channel` is not. This means a channel value like `"notify.>"` creates subject `"tenant.saldivia.notify.>"` (a NATS wildcard) and `"*"` creates a single-level wildcard. This is a subject injection bug in the production code, and the test suite should expose it.

**Fix:** Two parts:

1. **Test file** -- add a test that demonstrates the gap:

```go
func TestPublisher_Broadcast_InvalidChannel_Returns_Error(t *testing.T) {
    nc := startTestNATS(t)
    pub := New(nc)

    badChannels := []struct {
        name    string
        channel string
    }{
        {"empty", ""},
        {"dots", "notify.something"},
        {"wildcard star", "notify*"},
        {"wildcard gt", "notify>"},
        {"spaces", "notify something"},
    }

    for _, tt := range badChannels {
        t.Run(tt.name, func(t *testing.T) {
            err := pub.Broadcast("saldivia", tt.channel, "data")
            if err == nil {
                t.Errorf("expected error for channel %q", tt.channel)
            }
        })
    }
}
```

2. **Production code** (separate PR or same one) -- add validation to `Broadcast`:

```go
if !isValidSubjectToken(channel) {
    return fmt.Errorf("invalid channel for NATS subject: %q", channel)
}
```

The test will fail initially (because the production code doesn't validate yet). That's fine -- write the test first, then fix the production code. Or ship both together. Either way, the test must exist to prevent regression.

---

### Sugerencias

**S1. [pkg/middleware/auth_test.go] Consider table-driven tests for consistency with jwt_test.go and bible conventions**

The bible convention says "Table-driven, archivo `_test.go` junto al codigo". The JWT test file (`jwt_test.go`) uses individual test functions (not table-driven), so the auth tests are consistent with the existing pattern in this repo. However, the four negative-case tests (`NoAuthHeader`, `InvalidToken`, `WrongSecret`, `BearerPrefix`) share identical structure and could be collapsed into a single table-driven test:

```go
func TestAuth_InvalidRequests_Return401(t *testing.T) {
    tests := []struct {
        name string
        auth string // Authorization header value
    }{
        {"no header", ""},
        {"invalid token", "Bearer invalid.jwt.token"},
        {"wrong secret token", "Bearer " + tokenSignedWithWrongSecret(t)},
        {"missing Bearer prefix", validToken(t)},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := Auth(testSecret)(echoHandler())
            req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
            if tt.auth != "" {
                req.Header.Set("Authorization", tt.auth)
            }
            rec := httptest.NewRecorder()
            handler.ServeHTTP(rec, req)

            if rec.Code != http.StatusUnauthorized {
                t.Errorf("expected 401, got %d", rec.Code)
            }
        })
    }
}
```

Not blocking -- the current form is readable and correct (once B1 is fixed). Just noting the opportunity.

---

**S2. [pkg/middleware/auth_test.go:92-97] TestAuth_NoAuthHeader should also verify error body does not leak internals**

The test checks status code 401 and Content-Type, but does not assert on the response body. Since the middleware returns `{"error":"missing authorization"}`, it would be good to verify:
- The error message is generic (no JWT details leaked)
- The body is valid JSON

```go
var errResp map[string]string
if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
    t.Errorf("response not valid JSON: %v", err)
}
if strings.Contains(errResp["error"], "token") || strings.Contains(errResp["error"], "jwt") {
    t.Error("error response leaks token details")
}
```

Same applies to `TestAuth_InvalidToken_Returns401` -- the middleware returns `"invalid token"` which is fine, but a test assertion locks it down.

---

**S3. [pkg/nats/publisher_test.go:13-25] `startTestNATS` uses `t.Skip` for unavailable NATS -- consider documenting this pattern**

The `t.Skipf` approach is pragmatic -- tests pass in CI without NATS and actually test against a real server when one is available. However, this means the NATS publisher tests provide zero coverage in environments without a running NATS server. This is the opposite of the JWT tests which are fully self-contained.

Options to consider:
1. **Current approach (skip)** -- fine for integration tests, but document the requirement in the test file or a comment in the Makefile so CI can opt into running NATS.
2. **Embedded NATS server** -- `github.com/nats-io/nats-server/v2/server` can be started in-process in tests. This is the pattern used by the NATS team itself:

```go
func startTestNATS(t *testing.T) *nats.Conn {
    t.Helper()
    opts := &server.Options{Port: -1} // random port
    ns, err := server.NewServer(opts)
    if err != nil {
        t.Fatalf("start test NATS: %v", err)
    }
    ns.Start()
    t.Cleanup(ns.Shutdown)
    if !ns.ReadyForConnections(2 * time.Second) {
        t.Fatal("NATS not ready")
    }
    nc, err := nats.Connect(ns.ClientURL())
    if err != nil {
        t.Fatalf("connect: %v", err)
    }
    t.Cleanup(func() { nc.Close() })
    return nc
}
```

This makes the tests deterministic and self-contained. It does add a dependency (`nats-server/v2`) to `go.mod` but only in test scope.

---

**S4. [pkg/nats/publisher_test.go] Missing test: `Notify` with malicious event type containing NATS wildcards**

`Notify` extracts `parsed.Type` from the event JSON and puts it directly in the NATS subject (line 57 of publisher.go). While dots in the type (e.g., `"chat.new_message"`) are fine for NATS subject hierarchy, wildcards like `"*"` or `">"` could be dangerous.

The production code does not validate the event type for NATS special characters. A test should document whether this is acceptable or a bug:

```go
func TestPublisher_Notify_MaliciousEventType(t *testing.T) {
    nc := startTestNATS(t)
    pub := New(nc)

    evt := Event{Type: ">", UserID: "u-1", Title: "t", Body: "b"}
    err := pub.Notify("saldivia", evt)
    if err == nil {
        t.Error("expected error for wildcard event type")
    }
}
```

If the decision is that event types are trusted (internal-only), add a comment to the test explaining why this is not tested. Either way, the decision should be explicit.

---

**S5. [pkg/nats/publisher_test.go:117-118] Subscription error ignored in `TestPublisher_Notify_ValidSlug_Accepted`**

```go
sub, _ := nc.SubscribeSync("tenant." + slug + ".notify.test.event")
```

The subscription error is discarded. If `SubscribeSync` fails (unlikely but possible), `sub` is nil and `sub.NextMsg` panics. Use `t.Fatalf` on error for consistency with `TestPublisher_Notify_ValidEvent` (line 33) which does check the error.

---

**S6. [pkg/middleware/auth_test.go:126-158] Health bypass tests could cover path traversal edge cases**

The tests check `/health` and `/health/` which is good. Consider also testing:
- `/health?foo=bar` (query params should not break bypass)
- `/v1/health` (should NOT bypass -- different prefix)
- `/healthz` (common alternative, should NOT bypass)

These edge cases ensure `strings.TrimRight(r.URL.Path, "/") == "/health"` does exactly what is expected and nothing more.

---

## Hallazgos en codigo de produccion (fuera de scope del PR pero relevantes)

While reviewing the test files, I identified a production code bug that the tests should eventually cover:

**[pkg/nats/publisher.go:68-86] `Broadcast` does not validate the `channel` parameter**

The `channel` parameter is concatenated into the NATS subject without any validation. While `tenantSlug` is validated with `isValidSubjectToken`, `channel` is not. This allows subject injection if a caller passes a channel like `"notify.>"`. See D2 above for the full analysis.

This should be tracked as a follow-up fix to `publisher.go`, not just a test gap.

---

## Lo que esta bien

1. **Header spoofing tests are excellent.** `TestAuth_SpoofedHeadersStripped` (line 160) verifies that injected `X-User-ID`, `X-User-Role`, and `X-Tenant-Slug` headers are overwritten by JWT claims. `TestAuth_SpoofedHeaders_NoToken_NotLeaked` (line 191) verifies that without a valid token, the spoofed headers never reach the handler. This pair covers both attack vectors (valid-token-with-spoofed-headers and no-token-with-spoofed-headers). This is security-critical and well done.

2. **NATS slug injection tests are thorough.** `TestPublisher_Notify_InvalidSlug_Returns_Error` (line 79) covers 6 attack vectors: empty, dots, wildcard-star, wildcard-gt, spaces, and tabs. The table-driven format is clean and easy to extend. The complementary `TestPublisher_Notify_ValidSlug_Accepted` (line 107) tests positive cases, confirming that valid slugs work. This is a good balance of positive and negative testing.

3. **`isValidSubjectToken` is tested independently.** `TestIsValidSubjectToken` (line 168) has 11 cases including edge cases like newlines and carriage returns. Testing the validation function directly in addition to testing it through `Notify` and `Broadcast` is the right approach -- it catches regressions at the lowest level.

4. **Test helpers use `t.Helper()` correctly.** Both `validToken` (auth_test.go:16) and `startTestNATS` (publisher_test.go:13) call `t.Helper()`, so failure stack traces point to the calling test, not the helper.

5. **`t.Fatalf` vs `t.Errorf` usage is correct.** Fatal errors are used for setup failures (token creation, NATS connection) that make the rest of the test meaningless. `t.Errorf` is used for assertion failures that should report all failures, not just the first. This matches the pattern in `jwt_test.go`.

6. **`t.Cleanup` is used for NATS connection teardown.** The `startTestNATS` helper (line 23) registers `nc.Close()` via `t.Cleanup`, ensuring connections are closed even if a test panics. This is better than manual `defer` in each test.

7. **Bearer prefix test catches a real attack vector.** `TestAuth_BearerPrefix_Required` (line 208) verifies that sending a raw JWT without the `"Bearer "` prefix is rejected. Some middleware implementations accept both forms -- this test locks down the strict behavior.

8. **echoHandler pattern is well designed.** Writing identity headers back as JSON (line 33) makes assertions straightforward. The handler also reads from context via `tenant.FromContext`, verifying that the middleware sets context correctly -- not just headers. This tests the full integration path. (Pending the B1 compile fix.)

9. **Health bypass trailing slash test.** `TestAuth_HealthBypass_TrailingSlash` (line 146) catches the edge case where chi or a reverse proxy appends a trailing slash. Most middleware tests miss this.

10. **NATS tests verify the full publish-subscribe cycle.** Instead of just calling `Notify` and checking for no error, the tests subscribe first and then verify the received message payload. This catches serialization bugs, subject format bugs, and NATS routing bugs in one shot.

---

## Resumen de acciones

| ID | Severidad | Archivo | Fix |
|---|---|---|---|
| B1 | Bloqueante | `pkg/middleware/auth_test.go:35` | Handle second return value from `tenant.FromContext` (compile error) |
| D1 | Debe corregirse | `pkg/middleware/auth_test.go` | Add `TestAuth_ExpiredToken_Returns401` |
| D2 | Debe corregirse | `pkg/nats/publisher_test.go` | Add `TestPublisher_Broadcast_InvalidChannel_Returns_Error` (exposes production bug) |
| S1 | Sugerencia | `pkg/middleware/auth_test.go` | Consider table-driven tests for 401 cases |
| S2 | Sugerencia | `pkg/middleware/auth_test.go:92` | Assert error body content, not just status code |
| S3 | Sugerencia | `pkg/nats/publisher_test.go:13` | Consider embedded NATS server for deterministic tests |
| S4 | Sugerencia | `pkg/nats/publisher_test.go` | Add test for wildcard event type in `Notify` |
| S5 | Sugerencia | `pkg/nats/publisher_test.go:117` | Don't ignore `SubscribeSync` error |
| S6 | Sugerencia | `pkg/middleware/auth_test.go:126` | Add health bypass edge cases (`/v1/health`, `/healthz`, query params) |
