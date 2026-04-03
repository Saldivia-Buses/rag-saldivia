# Gateway Review -- PR #40 RAG Handler Tests

**Fecha:** 2026-04-02
**Tipo:** review
**Intensity:** standard
**Branch:** feat/rag-handler-tests
**Reviewer:** gateway-reviewer (Opus)

## Resultado

**CAMBIOS REQUERIDOS** (1 bloqueante, 1 debe corregirse)

---

## Hallazgos

### Bloqueantes

1. **[services/rag/internal/handler/rag.go:110] Health endpoint leaks Blueprint error details to client**

   The Health handler returns `err.Error()` directly in the JSON response:

   ```go
   writeJSON(w, http.StatusServiceUnavailable, map[string]string{
       "status": "unhealthy", "service": "rag", "error": err.Error(),
   })
   ```

   The concrete `service.RAG.Health()` wraps errors with context like
   `"blueprint unreachable: dial tcp 127.0.0.1:8081: connect: connection refused"`.
   This exposes internal network topology (blueprint IP/port, OS-level error strings)
   to any client that can hit `/health`.

   The bible is explicit: "Errores internos no exponen stack traces al cliente."
   While this is not a stack trace, leaking `dial tcp <ip>:<port>` is the same class
   of information disclosure.

   **Fix:** Return a generic message; log the real error server-side:

   ```go
   func (h *RAG) Health(w http.ResponseWriter, r *http.Request) {
       if err := h.ragSvc.Health(r.Context()); err != nil {
           slog.Error("health check failed", "error", err)
           writeJSON(w, http.StatusServiceUnavailable, map[string]string{
               "status": "unhealthy", "service": "rag",
           })
           return
       }
       writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "rag"})
   }
   ```

   This is a blocker because it is a security issue in production handler code
   (not just the tests), and since the PR modifies `rag.go` to add the interface,
   it is in scope for this review.

   **Test impact:** `TestHealth_BlueprintDown_Returns503` currently does not assert
   on the error field, so the test needs no change. Good.

### Debe corregirse

2. **[services/rag/internal/handler/rag_test.go:20-29] Mock `err` field controls all methods uniformly -- masks independent failure paths**

   The `mockRAGService` has a single `err` field shared by `GenerateStream` and
   `ListCollections`. The `healthErr` field is separate (good), but if a future test
   needs GenerateStream to succeed while ListCollections fails (or vice versa), the
   mock cannot express that.

   More importantly, the current design means `TestListCollections_BlueprintDown_Returns502`
   would also cause `GenerateStream` to fail if called, which is not visible in the
   test but is a coupling that can produce false positives or confusing failures if
   tests grow.

   **Fix:** Split into per-method error fields:

   ```go
   type mockRAGService struct {
       collections    []string
       streamBody     string
       healthErr      error
       generateErr    error
       collectionsErr error
   }
   ```

   Then use `m.generateErr` in `GenerateStream` and `m.collectionsErr` in
   `ListCollections`. This is low-effort and makes each test's intent explicit.

### Sugerencias

3. **[rag_test.go] Missing test: oversized body (>1MB) returns 400/413**

   The handler sets `http.MaxBytesReader(w, r.Body, 1<<20)` on line 44 of `rag.go`.
   No test verifies this limit. A test sending a body >1MB would confirm the
   protection works and document the limit for future readers:

   ```go
   func TestGenerate_OversizedBody_Rejected(t *testing.T) {
       r := setupRAGRouter(&mockRAGService{})
       bigBody := strings.NewReader(strings.Repeat("x", (1<<20)+1))
       req := httptest.NewRequest(http.MethodPost, "/v1/rag/generate", bigBody)
       req.Header.Set("Content-Type", "application/json")
       req.Header.Set("X-Tenant-Slug", "saldivia")
       rec := httptest.NewRecorder()
       r.ServeHTTP(rec, req)
       if rec.Code != http.StatusBadRequest {
           t.Fatalf("expected 400 for oversized body, got %d", rec.Code)
       }
   }
   ```

4. **[rag_test.go] Missing test: empty X-Tenant-Slug header**

   Both `Generate` and `ListCollections` read `X-Tenant-Slug` from the header and
   pass it straight to the service layer. If it is empty, the service will namespace
   collections with a bare `-` prefix. There is no validation at the handler level.
   This is either a gap in the handler (should return 400 if empty) or the middleware
   is expected to guarantee it. Either way, a test documenting current behavior
   would clarify the contract.

5. **[rag_test.go] Consider table-driven tests for Generate error paths**

   The bible specifies "Tests: Table-driven, archivo `_test.go` junto al codigo".
   The three Generate error tests (empty messages, invalid JSON, blueprint error)
   are structurally identical and could be a single table-driven test. This is a
   style nit, not a correctness issue -- the current tests are readable.

6. **[rag.go:84-87] Generate streaming loop does not check w.Write error**

   Not introduced by this PR, but visible in the handler under test: the streaming
   loop ignores the return value of `w.Write(buf[:n])`. If the client disconnects
   mid-stream, writes will fail silently and the loop continues until the upstream
   body ends. Consider breaking on write error to free the Blueprint connection
   sooner. Low priority -- the context cancellation from the client disconnect
   will eventually propagate, but the write error would be faster.

### Lo que esta bien

- **Interface extraction is clean.** `RAGService` lives in the handler package (consumer-side),
  follows the Go convention of "accept interfaces, return structs", and uses the `-er`
  suffix convention from the bible. The concrete `*service.RAG` satisfies it without
  adapter code.

- **Compile correctness verified.** All three method signatures on `*service.RAG` match
  `handler.RAGService` exactly. `main.go:40` passes `*service.RAG` to `handler.NewRAG()`
  which accepts `RAGService` -- implicit interface satisfaction works.

- **Error messages to clients are generic** (in Generate and ListCollections).
  `"rag server unavailable"` does not leak Blueprint internals. Only the Health
  endpoint has this issue.

- **9 tests, good coverage of the handler layer.** Happy path + error path for all
  three endpoints. SSE streaming test correctly asserts on Content-Type and streamed
  content.

- **MaxBytesReader protection** on Generate is good practice -- prevents DoS via
  oversized payloads.

- **Mock is simple and appropriate** for unit tests. No over-engineering.

---

## Interface Conformance Matrix

| Interface Method | Concrete `*service.RAG` Signature | Match |
|---|---|---|
| `GenerateStream(ctx, tenantSlug, req) (io.ReadCloser, string, error)` | `func (r *RAG) GenerateStream(ctx context.Context, tenantSlug string, req GenerateRequest) (io.ReadCloser, string, error)` | OK |
| `ListCollections(ctx, tenantSlug) ([]string, error)` | `func (r *RAG) ListCollections(ctx context.Context, tenantSlug string) ([]string, error)` | OK |
| `Health(ctx) error` | `func (r *RAG) Health(ctx context.Context) error` | OK |

## Coverage Summary

| Handler Method | Happy Path | Error: Bad Input | Error: Service Failure | Edge Cases |
|---|---|---|---|---|
| Generate | TestGenerate_Success_StreamsSSE | TestGenerate_EmptyMessages_Returns400, TestGenerate_InvalidJSON_Returns400 | TestGenerate_BlueprintError_Returns502 | Missing: oversized body, empty tenant |
| ListCollections | TestListCollections_Success | -- | TestListCollections_BlueprintDown_Returns502 | Missing: empty tenant |
| Health | TestHealth_BlueprintHealthy_Returns200 | -- | TestHealth_BlueprintDown_Returns503 | -- |
