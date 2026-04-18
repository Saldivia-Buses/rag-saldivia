package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Camionerou/rag-saldivia/pkg/tenant"
)

// newTestHandler creates a Handler with a nil service and nil audit writer.
// Safe for tests that exercise validation paths (400/401) — these return
// before the service is called.
func newTestHandler() *Handler {
	return &Handler{svc: nil, audit: nil}
}

// tenantContext returns a context with a valid tenant.Info injected.
func tenantContext(r *http.Request) *http.Request {
	ctx := tenant.WithInfo(r.Context(), tenant.Info{ID: "t-abc", Slug: "test-tenant"})
	return r.WithContext(ctx)
}

// TestSearch_MissingTenantContext_Returns401 verifies the C1 invariant:
// requests without tenant context must be rejected immediately.
func TestSearch_MissingTenantContext_Returns401(t *testing.T) {
	h := newTestHandler()

	r := chi.NewRouter()
	r.Post("/search/query", h.SearchDocuments)

	body := `{"query":"what is the revenue?"}`
	req := httptest.NewRequest(http.MethodPost, "/search/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "unauthorized", resp["code"])
}

// TestSearch_EmptyTenantID_Returns401 verifies that a tenant.Info with empty ID
// is treated as missing (defense-in-depth against partially-populated context).
func TestSearch_EmptyTenantID_Returns401(t *testing.T) {
	h := newTestHandler()

	r := chi.NewRouter()
	r.Post("/search/query", h.SearchDocuments)

	body := `{"query":"revenue"}`
	req := httptest.NewRequest(http.MethodPost, "/search/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Inject a tenant.Info with empty ID — this should be rejected.
	ctx := tenant.WithInfo(req.Context(), tenant.Info{ID: "", Slug: "bad"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "unauthorized", resp["code"])
}

// TestSearch_EmptyQuery_Returns400 verifies that an empty query field is rejected
// with 400 before reaching the service layer.
func TestSearch_EmptyQuery_Returns400(t *testing.T) {
	h := newTestHandler()

	r := chi.NewRouter()
	r.Post("/search/query", h.SearchDocuments)

	body := `{"query":""}`
	req := httptest.NewRequest(http.MethodPost, "/search/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = tenantContext(req)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "invalid_input", resp["code"])
}

// TestSearch_MissingQueryField_Returns400 verifies that omitting the query field
// also yields 400 (zero value of string is "").
func TestSearch_MissingQueryField_Returns400(t *testing.T) {
	h := newTestHandler()

	r := chi.NewRouter()
	r.Post("/search/query", h.SearchDocuments)

	body := `{"collection_id":"col-1"}` // no query field
	req := httptest.NewRequest(http.MethodPost, "/search/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = tenantContext(req)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "invalid_input", resp["code"])
}

// TestSearch_InvalidJSON_Returns400 verifies that a malformed request body
// is rejected with 400 before reaching the service layer.
func TestSearch_InvalidJSON_Returns400(t *testing.T) {
	h := newTestHandler()

	r := chi.NewRouter()
	r.Post("/search/query", h.SearchDocuments)

	req := httptest.NewRequest(http.MethodPost, "/search/query", strings.NewReader("not-json{"))
	req.Header.Set("Content-Type", "application/json")
	req = tenantContext(req)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	// Response must be JSON (invariant 7: JSON error responses only)
	ct := rec.Header().Get("Content-Type")
	assert.Contains(t, ct, "application/json",
		"error response Content-Type must be application/json")

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "invalid_input", resp["code"])
}

// TestSearch_EmptyBody_Returns400 verifies that an empty body (EOF on decode)
// is rejected as invalid input.
func TestSearch_EmptyBody_Returns400(t *testing.T) {
	h := newTestHandler()

	r := chi.NewRouter()
	r.Post("/search/query", h.SearchDocuments)

	req := httptest.NewRequest(http.MethodPost, "/search/query", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req = tenantContext(req)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestSearch_ErrorResponses_AreAlwaysJSON verifies that all 4xx error paths
// return JSON with Content-Type application/json (critical invariant 7).
func TestSearch_ErrorResponses_AreAlwaysJSON(t *testing.T) {
	h := newTestHandler()

	cases := []struct {
		name   string
		body   string
		addCtx bool
		wantSt int
	}{
		{"no tenant context", `{"query":"test"}`, false, http.StatusUnauthorized},
		{"empty query with ctx", `{"query":""}`, true, http.StatusBadRequest},
		{"invalid json with ctx", `{bad json`, true, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Post("/search/query", h.SearchDocuments)

			req := httptest.NewRequest(http.MethodPost, "/search/query", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			if tc.addCtx {
				req = tenantContext(req)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tc.wantSt, rec.Code, "status mismatch for case %q", tc.name)

			ct := rec.Header().Get("Content-Type")
			assert.Contains(t, ct, "application/json",
				"error response Content-Type must be application/json, case: %q", tc.name)

			// Body must be parseable as JSON
			var resp map[string]any
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp),
				"error response body must be valid JSON, case: %q", tc.name)

			// Must include an "error" field (never expose raw error details)
			_, hasError := resp["error"]
			assert.True(t, hasError,
				"error response must include 'error' field, case: %q", tc.name)
		})
	}
}

// TDD-ANCHOR: TestSearch_Success_Returns200 and TestSearch_ServiceError_Returns500_Generic
// require a real *service.Search (svc field is a concrete struct, not an interface).
// These cases are covered in service/search_integration_test.go which spins up
// testcontainers postgres + a mock LLM HTTP server.
//
// To make these handler tests possible without integration infra, the production
// handler should accept a SearchService interface instead of *service.Search directly.
// See: services/app/internal/realtime/chat/handler/chat.go for the pattern (ChatService interface).
