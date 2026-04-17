package handler

// White-box tests: same package as handler to access unexported fields.
// The handler uses *service.Agent directly (no interface extracted).
// Tests that require calling the real service are marked TDD-ANCHOR
// and covered at integration level. Tests for input validation and
// JSON encoding are fully covered here using nil-safe handler paths.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRouter builds a chi router wiring Query and Confirm without the
// RequirePermission middleware — tests focus on handler logic, not auth.
func newTestRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Post("/query", h.Query)
	r.Post("/confirm", h.Confirm)
	return r
}

// TestQuery_InvalidJSON_Returns400 verifies that a malformed JSON body is
// rejected before the service is ever called.
func TestQuery_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	// svc is nil — handler must return 400 before reaching service call
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"])
}

// TestQuery_EmptyMessage_Returns400 verifies that an empty message string
// is rejected with 400 before the service is called.
func TestQuery_EmptyMessage_Returns400(t *testing.T) {
	t.Parallel()
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	body := `{"message":""}`
	req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["error"])
}

// TestQuery_MissingMessage_Returns400 verifies that a body without the
// message field (zero value) is rejected.
func TestQuery_MissingMessage_Returns400(t *testing.T) {
	t.Parallel()
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(`{"history":[]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestQuery_Success_Returns200
// TDD-ANCHOR: service.Agent requires *llm.Client + *tools.Executor + *TracePublisher
// — no interface extraction exists, so constructing a real Agent is complex.
// Covered at integration level when those components are available.
//
// The shape of the happy path: POST /query with valid message → service.Query()
// → JSON-encoded QueryResult with 200.

// TestQuery_ServiceError_Returns500_Generic
// TDD-ANCHOR: requires service.Agent with injectable LLM mock.
// Key invariant: service errors must map to 500 with generic message,
// not leak internal error details to the client.

// TestConfirm_InvalidJSON_Returns400 verifies that a malformed JSON body
// is rejected before the service is called.
func TestConfirm_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/confirm", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"])
}

// TestConfirm_EmptyTool_Returns400 verifies that an empty tool name is
// rejected with 400 before the service is called.
func TestConfirm_EmptyTool_Returns400(t *testing.T) {
	t.Parallel()
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	body := `{"tool":"","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["error"])
}

// TestConfirm_MissingTool_Returns400 verifies that a body without the
// tool field (zero value) is rejected.
func TestConfirm_MissingTool_Returns400(t *testing.T) {
	t.Parallel()
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/confirm", bytes.NewBufferString(`{"params":{"x":1}}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestConfirm_Success_Returns200
// TDD-ANCHOR: requires service.Agent with injectable mock.
// Key invariant: POST /confirm with valid tool name → ExecuteConfirmed()
// → JSON-encoded tools.Result with 200.

// TestExtractJWT verifies the Authorization header parsing helper.
func TestExtractJWT(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "valid bearer", header: "Bearer my.jwt.token", want: "my.jwt.token"},
		{name: "no header", header: "", want: ""},
		{name: "no bearer prefix", header: "my.jwt.token", want: ""},
		{name: "bearer only no token", header: "Bearer ", want: ""},
		{name: "short header", header: "Bear", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			got := extractJWT(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestQuery_ResponseIsJSON verifies the Content-Type header is set on error
// responses (invariant: all agent API responses must be JSON).
func TestQuery_ResponseIsJSON(t *testing.T) {
	t.Parallel()
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	// Empty message — triggers 400 before any service call
	req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(`{"message":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	ct := rec.Header().Get("Content-Type")
	assert.Contains(t, ct, "application/json", "error response must be JSON")
}

// TestConfirm_ResponseIsJSON verifies error responses from Confirm are JSON.
func TestConfirm_ResponseIsJSON(t *testing.T) {
	t.Parallel()
	h := &Handler{svc: nil}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/confirm", strings.NewReader(`{"tool":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	ct := rec.Header().Get("Content-Type")
	assert.Contains(t, ct, "application/json", "error response must be JSON")
}
