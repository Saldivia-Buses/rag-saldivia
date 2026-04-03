# Gateway Review -- PR #38 feat/handler-tests

**Fecha:** 2026-04-02
**Tipo:** review
**PR:** #38 (feat/handler-tests)
**Intensity:** thorough
**Reviewer:** gateway-reviewer (Opus)

## Archivos revisados

| Archivo | Cambio | Lineas |
|---|---|---|
| `services/chat/internal/handler/chat.go` | Modified -- added `ChatService` interface, changed field type | 235 |
| `services/chat/internal/handler/chat_test.go` | New -- 16 test cases | 443 |
| `services/ingest/internal/handler/ingest.go` | Modified -- added `IngestService` interface, changed field type | 186 |
| `services/ingest/internal/handler/ingest_test.go` | New -- 14 test cases | 320 |

### Archivos de contexto (no modificados)

| Archivo | Motivo |
|---|---|
| `services/chat/internal/service/chat.go` | Concrete types, error sentinels, method signatures |
| `services/chat/cmd/main.go` | Where `handler.NewChat()` is called with `*service.Chat` |
| `services/ingest/internal/service/ingest.go` | Concrete types, error sentinels, method signatures |
| `services/ingest/cmd/main.go` | Where `handler.NewIngest()` is called with `*service.Ingest` |
| `pkg/middleware/auth.go` | Shared auth middleware -- sets identity headers |

## Resultado

**CAMBIOS REQUERIDOS** (0 bloqueantes, 2 debe corregirse, 5 sugerencias)

---

## Hallazgos

### Bloqueantes

None.

### Compile correctness

Both interfaces are correctly defined. Verified each method signature against the concrete types:

**ChatService interface** (chat.go:18-26) vs `*service.Chat`:

| Method | Interface | Concrete | Match |
|---|---|---|---|
| `CreateSession` | `(ctx, userID, title string, collection *string) (*Session, error)` | Same | Yes |
| `GetSession` | `(ctx, sessionID, userID string) (*Session, error)` | Same | Yes |
| `ListSessions` | `(ctx, userID string) ([]Session, error)` | Same | Yes |
| `DeleteSession` | `(ctx, sessionID, userID string) error` | Same | Yes |
| `RenameSession` | `(ctx, sessionID, userID, title string) error` | Same | Yes |
| `AddMessage` | `(ctx, sessionID, userID, role, content string, sources, metadata []byte) (*Message, error)` | Same | Yes |
| `GetMessages` | `(ctx, sessionID string) ([]Message, error)` | Same | Yes |

**IngestService interface** (ingest.go:31-36) vs `*service.Ingest`:

| Method | Interface | Concrete | Match |
|---|---|---|---|
| `Submit` | `(ctx, tenantSlug, userID, collection, fileName string, fileSize int64, file multipart.File) (*Job, error)` | Same | Yes |
| `ListJobs` | `(ctx, userID string, limit int) ([]Job, error)` | Same | Yes |
| `GetJob` | `(ctx, jobID, userID string) (*Job, error)` | Same | Yes |
| `DeleteJob` | `(ctx, jobID, userID string) error` | Same | Yes |

Both `cmd/main.go` files pass `*service.Chat` and `*service.Ingest` to `handler.NewChat()` and `handler.NewIngest()` respectively. Since the concrete types implement all interface methods, compilation is guaranteed. The `IngestService` interface intentionally excludes `UpdateJobStatus` (worker-only method), which is correct -- the handler does not need it.

### Debe corregirse

**D1. [chat_test.go] Missing 500 error path tests -- mock has `err` field but no test uses it**

Both `mockChatService` and `mockIngestService` have an `err` field designed to inject service-layer failures, but neither test file exercises it. This means the `serverError()` code path (chat.go:230-234) and the equivalent error logging paths in ingest handlers are completely untested.

The `serverError()` function is security-relevant: it must NOT leak internal error details to the client (it correctly returns a generic `"internal error"` message). Without a test, a future regression could expose stack traces.

**Fix:** Add at least one test per service that sets `err` on the mock and verifies the handler returns 500 with a generic error message. Example for chat:

```go
func TestListSessions_ServiceError_Returns500(t *testing.T) {
    mock := &mockChatService{err: errors.New("db connection lost")}
    r := setupRouter(mock)

    req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil), "u-1")
    rec := httptest.NewRecorder()
    r.ServeHTTP(rec, req)

    if rec.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", rec.Code)
    }

    var resp map[string]string
    json.NewDecoder(rec.Body).Decode(&resp)
    if resp["error"] != "internal error" {
        t.Errorf("expected generic error message, got %q", resp["error"])
    }
}
```

Same pattern for ingest. This also verifies the security property that internal error details are not leaked.

**D2. [chat.go:71,83,106,123,140,162,188] Chat handler never validates `X-User-ID` is non-empty**

This is a pre-existing issue not introduced by this PR, but the test suite should document the expected behavior. Every chat handler method reads `X-User-ID` with `r.Header.Get("X-User-ID")` but never checks for empty string. Compare with ingest, which has `requireIdentity()` (ingest.go:59-63) returning 401 when headers are missing.

In production, the auth middleware (pkg/middleware/auth.go:53) always sets `X-User-ID` from JWT claims, so an empty value should be impossible. But if a service misconfiguration skips the middleware (as is currently the case -- see Sugerencias S5), an empty `X-User-ID` would cause:
- `ListSessions` to return all sessions with `user_id = ''` (likely empty set, not harmful)
- `CreateSession` to create a session with empty `user_id` (data integrity issue)
- `GetSession`/`DeleteSession` to fail with not-found (benign)

**Fix for this PR:** Add a test that documents the current behavior (request with empty `X-User-ID`). Optionally, add a `requireIdentity()` function to the chat handler matching ingest's pattern, or add a test comment documenting that auth middleware is the guard.

### Sugerencias

**S1. [chat_test.go:67-79] Mock `DeleteSession` does not distinguish non-owner from not-found**

The mock's `DeleteSession` returns `ErrSessionNotFound` for both "session exists but different user" and "session does not exist". This happens to match the concrete service behavior (which does `DELETE WHERE id=$1 AND user_id=$2` and checks rows affected), so the mock is functionally correct. However, if the concrete service ever adds explicit `ErrNotOwner` handling for delete (as `GetSession` already does), the mock would diverge silently.

Consider adding a comment in the mock explaining this is intentional:

```go
// DeleteSession matches the concrete service: non-owner returns ErrSessionNotFound
// (DELETE WHERE user_id=$2 yields 0 rows), not ErrNotOwner.
```

**S2. [ingest_test.go] Missing test for upload without file field**

The handler has explicit handling for `r.FormFile("file")` failure (ingest.go:80-84, returns 400 "file is required"), but no test exercises this path. A multipart request with the `collection` field but no `file` field would cover this.

```go
func TestUpload_MissingFile_Returns400(t *testing.T) {
    r := setupIngestRouter(&mockIngestService{})

    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)
    writer.WriteField("collection", "test")
    writer.Close()

    req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("X-User-ID", "u-1")
    req.Header.Set("X-Tenant-Slug", "saldivia")
    rec := httptest.NewRecorder()
    r.ServeHTTP(rec, req)

    if rec.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for missing file, got %d", rec.Code)
    }
}
```

**S3. [chat_test.go:163] `TestCreateSession_Success` does not assert `collection` field**

The test sends `"collection":"docs"` in the body and verifies `Title` and `UserID` but not `Collection`. Since the collection pass-through is a meaningful handler behavior (the `*string` handling), it is worth asserting:

```go
if session.Collection == nil || *session.Collection != "docs" {
    t.Errorf("expected collection 'docs', got %v", session.Collection)
}
```

**S4. [ingest_test.go:37-51] Mock `ListJobs` ignores `limit` parameter**

The mock's `ListJobs` filters by `userID` but ignores the `limit` argument. Since the handler's limit-parsing logic (ingest.go:119-124) passes the value to the service, a test could verify the handler sends the correct default (50) and respects query params like `?limit=10`. This is a low-priority gap since limit enforcement is in the service layer, but the handler's parsing of the query param is untested.

**S5. [services/chat/cmd/main.go] Chat service does not use shared auth middleware**

Pre-existing, not introduced by this PR. The ingest service applies `sdamw.Auth(jwtSecret)` (ingest cmd/main.go:99), but the chat service does not import or use `pkg/middleware`. This means the chat service currently relies entirely on an upstream gateway (Traefik) for JWT verification. If Traefik is misconfigured or bypassed, chat endpoints are unprotected.

This is outside the scope of PR #38 but should be tracked as a separate issue.

### Lo que esta bien

1. **Interface placement follows Go idiom.** Both `ChatService` and `IngestService` are defined in the consumer package (`handler/`), not the provider package (`service/`). This is textbook Go interface design -- the consumer defines what it needs, the provider implicitly satisfies it.

2. **IngestService correctly excludes `UpdateJobStatus`.** The interface only includes methods the handler needs (Submit, ListJobs, GetJob, DeleteJob), leaving out worker-only methods. This enforces the principle of least privilege at the type level.

3. **Ownership tests are thorough.** Both test suites verify that non-owners get 404 (not 403) for session and job access. This is the correct security pattern -- returning 404 instead of 403 prevents information leakage about resource existence.

4. **Chat test covers `ErrNotOwner` vs `ErrSessionNotFound` distinction.** The mock's `GetSession` correctly returns `ErrNotOwner` when the session exists but belongs to a different user (chat_test.go:44-45), and the handler maps both to 404 (chat.go:111). This subtle distinction is well-tested.

5. **Ingest tests cover both identity headers independently.** `TestUpload_MissingUserID_Returns401` tests no headers at all, while `TestUpload_MissingTenantSlug_Returns401` tests `X-User-ID` present but `X-Tenant-Slug` missing. This validates the `&&` in `requireIdentity()`.

6. **File extension validation is well-tested.** Both unsupported extensions (exe, sh, png, mp4) and supported ones (pdf, docx, txt, csv, xlsx) are covered with table-driven subtests.

7. **Role validation is exhaustive.** `TestAddMessage_ValidRole_Success` iterates all three valid roles (user, assistant, system) with subtests, and `TestAddMessage_InvalidRole_Returns400` tests an invalid role ("admin").

8. **Request body size limits.** Handlers use `http.MaxBytesReader` for body-accepting endpoints (CreateSession, RenameSession, AddMessage, Upload). The tests exercise the JSON parsing paths, which would fail gracefully on oversized bodies.

9. **Mock design is clean and minimal.** Both mocks use a single `err` field for error injection and slice fields for data, avoiding complex mock frameworks. The `err` check comes first in every mock method, providing clean override capability.

10. **`writeJSON` and `serverError` are properly factored.** The error handler logs the request ID and internal error via slog, then returns a generic "internal error" to the client -- no stack trace leakage.

---

## Coverage summary

| Handler | Happy path | Error path | Ownership | Validation | 500 error |
|---|---|---|---|---|---|
| Chat: ListSessions | Tested | -- | Filtered by userID | -- | **Not tested** |
| Chat: CreateSession | Tested | Invalid JSON | -- | Empty title default | **Not tested** |
| Chat: GetSession | Tested | Not found | Non-owner = 404 | -- | **Not tested** |
| Chat: DeleteSession | Tested | Not found | Non-owner = 404 | -- | **Not tested** |
| Chat: RenameSession | Tested | Not found | -- | Empty title = 400 | **Not tested** |
| Chat: AddMessage | Tested | -- | Non-owner = 404 | Role + content | **Not tested** |
| Chat: GetMessages | Tested | -- | Non-owner = 404 | -- | **Not tested** |
| Ingest: Upload | Tested | Missing identity | -- | Extension, collection | **Not tested** |
| Ingest: ListJobs | Tested | Missing identity | Filtered by userID | -- | **Not tested** |
| Ingest: GetJob | Tested | Not found | Non-owner = 404 | -- | **Not tested** |
| Ingest: DeleteJob | Tested | Not found | -- | -- | **Not tested** |

The "Not tested" column is D1. Everything else has solid coverage.
