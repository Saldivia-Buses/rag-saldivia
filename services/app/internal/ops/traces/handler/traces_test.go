package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/app/internal/ops/traces/service"
)

// --- mock ---

type mockTracesService struct {
	traces   []service.Trace
	trace    *service.Trace
	events   []service.TraceEvent
	cost     *service.CostSummary
	err      error
}

func (m *mockTracesService) ListTraces(_ context.Context, tenantID string, limit, offset int) ([]service.Trace, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.traces, nil
}

func (m *mockTracesService) GetTraceDetail(_ context.Context, traceID, tenantID string) (*service.Trace, []service.TraceEvent, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.trace, m.events, nil
}

func (m *mockTracesService) GetTenantCost(_ context.Context, tenantID string, from, to time.Time) (*service.CostSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cost, nil
}

// --- helpers ---

func setupTracesRouter(mock TracesService) *chi.Mux {
	h := &Handler{svc: mock}
	r := chi.NewRouter()
	r.Get("/v1/traces", h.ListTraces)
	r.Get("/v1/traces/costs", h.GetTenantCost)
	r.Get("/v1/traces/{traceID}", h.GetTraceDetail)
	return r
}

func withTenantCtx(req *http.Request, id, slug string) *http.Request {
	ctx := tenant.WithInfo(req.Context(), tenant.Info{ID: id, Slug: slug})
	return req.WithContext(ctx)
}

// --- auth guard tests ---

func TestListTraces_MissingTenantContext_Returns401(t *testing.T) {
	// Handler with nil service — we only test the tenant check
	h := &Handler{svc: nil}

	r := chi.NewRouter()
	r.Get("/v1/traces", h.ListTraces)

	req := httptest.NewRequest(http.MethodGet, "/v1/traces", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestGetTraceDetail_MissingTenantContext_Returns401(t *testing.T) {
	h := &Handler{svc: nil}

	r := chi.NewRouter()
	r.Get("/v1/traces/{traceID}", h.GetTraceDetail)

	req := httptest.NewRequest(http.MethodGet, "/v1/traces/some-id", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestGetTenantCost_MissingTenantContext_Returns401(t *testing.T) {
	h := &Handler{svc: nil}

	r := chi.NewRouter()
	r.Get("/v1/traces/costs", h.GetTenantCost)

	req := httptest.NewRequest(http.MethodGet, "/v1/traces/costs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// --- list traces ---

func TestListTraces_Success_ReturnsJSON(t *testing.T) {
	dur := 100
	mock := &mockTracesService{
		traces: []service.Trace{
			{ID: "tr-1", TenantID: "t-1", Query: "what is X", Status: "completed", TotalDurationMS: &dur},
		},
	}
	r := setupTracesRouter(mock)

	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces", nil), "t-1", "test")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var traces []service.Trace
	if err := json.NewDecoder(rec.Body).Decode(&traces); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(traces) != 1 || traces[0].ID != "tr-1" {
		t.Errorf("unexpected traces: %+v", traces)
	}
}

func TestListTraces_ServiceError_Returns500(t *testing.T) {
	mock := &mockTracesService{err: errors.New("db down")}
	r := setupTracesRouter(mock)

	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces", nil), "t-1", "test")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestListTraces_EmptyResult_ReturnsEmptyArray(t *testing.T) {
	mock := &mockTracesService{traces: nil}
	r := setupTracesRouter(mock)

	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces", nil), "t-1", "test")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var traces []service.Trace
	_ = json.NewDecoder(rec.Body).Decode(&traces)
	if len(traces) != 0 {
		t.Errorf("expected empty array, got %d elements", len(traces))
	}
}

// --- get trace detail ---

func TestGetTraceDetail_Success_ReturnsTraceAndEvents(t *testing.T) {
	tr := &service.Trace{ID: "tr-1", TenantID: "t-1", Query: "test query", Status: "completed"}
	evts := []service.TraceEvent{{TraceID: "tr-1", Seq: 1, EventType: "llm_call"}}
	mock := &mockTracesService{trace: tr, events: evts}
	r := setupTracesRouter(mock)

	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces/tr-1", nil), "t-1", "test")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := resp["trace"]; !ok {
		t.Error("expected 'trace' key in response")
	}
	if _, ok := resp["events"]; !ok {
		t.Error("expected 'events' key in response")
	}
}

func TestGetTraceDetail_NotFound_Returns404(t *testing.T) {
	mock := &mockTracesService{err: errors.New("trace not found: no rows")}
	r := setupTracesRouter(mock)

	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces/nonexistent", nil), "t-1", "test")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// --- get tenant cost ---

func TestGetTenantCost_Success_ReturnsSummary(t *testing.T) {
	mock := &mockTracesService{
		cost: &service.CostSummary{
			TenantID:     "t-1",
			TotalCostUSD: 1.23,
			TotalQueries: 5,
		},
	}
	r := setupTracesRouter(mock)

	req := withTenantCtx(
		httptest.NewRequest(http.MethodGet, "/v1/traces/costs?from=2026-01-01&to=2026-02-01", nil),
		"t-1", "test",
	)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var cs service.CostSummary
	if err := json.NewDecoder(rec.Body).Decode(&cs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if cs.TotalCostUSD != 1.23 {
		t.Errorf("expected TotalCostUSD 1.23, got %f", cs.TotalCostUSD)
	}
}

func TestGetTenantCost_ServiceError_Returns500(t *testing.T) {
	mock := &mockTracesService{err: errors.New("query failed")}
	r := setupTracesRouter(mock)

	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces/costs", nil), "t-1", "test")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetTenantCost_InvalidDateParams_UsesDefaults(t *testing.T) {
	// Invalid date params should fall back to defaults (last 30 days) — no error
	mock := &mockTracesService{
		cost: &service.CostSummary{TenantID: "t-1"},
	}
	r := setupTracesRouter(mock)

	req := withTenantCtx(
		httptest.NewRequest(http.MethodGet, "/v1/traces/costs?from=not-a-date&to=also-not", nil),
		"t-1", "test",
	)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should still return 200 — handler falls back to time.Now()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with fallback dates, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- tenant isolation ---

// tenantFilteredMock is a mock that enforces tenant isolation: it only returns
// traces whose TenantID matches the tenantID argument. This tests that the
// handler actually passes the correct tenant ID to the service, not a fixed or
// zero value.
type tenantFilteredMock struct {
	allTraces  []service.Trace
	allDetails map[string]*service.Trace // traceID → trace
	err        error
}

func (m *tenantFilteredMock) ListTraces(_ context.Context, tenantID string, limit, offset int) ([]service.Trace, error) {
	if m.err != nil {
		return nil, m.err
	}
	var out []service.Trace
	for _, tr := range m.allTraces {
		if tr.TenantID == tenantID {
			out = append(out, tr)
		}
	}
	return out, nil
}

func (m *tenantFilteredMock) GetTraceDetail(_ context.Context, traceID, tenantID string) (*service.Trace, []service.TraceEvent, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	if tr, ok := m.allDetails[traceID]; ok && tr.TenantID == tenantID {
		return tr, nil, nil
	}
	return nil, nil, errors.New("trace not found: no rows")
}

func (m *tenantFilteredMock) GetTenantCost(_ context.Context, tenantID string, from, to time.Time) (*service.CostSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &service.CostSummary{TenantID: tenantID}, nil
}

// TestListTraces_OnlyReturnsCurrentTenantTraces verifies that ListTraces passes
// the tenant ID from context to the service and that cross-tenant traces are
// not visible to another tenant.
func TestListTraces_OnlyReturnsCurrentTenantTraces(t *testing.T) {
	mock := &tenantFilteredMock{
		allTraces: []service.Trace{
			{ID: "tr-t1-1", TenantID: "t-1", Query: "query from tenant 1"},
			{ID: "tr-t2-1", TenantID: "t-2", Query: "query from tenant 2"},
			{ID: "tr-t1-2", TenantID: "t-1", Query: "another query from tenant 1"},
		},
	}

	r := setupTracesRouter(mock)

	// Tenant 1 should only see its own 2 traces
	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces", nil), "t-1", "tenant1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var traces []service.Trace
	if err := json.NewDecoder(rec.Body).Decode(&traces); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(traces) != 2 {
		t.Fatalf("tenant t-1 should see 2 traces, got %d", len(traces))
	}
	for _, tr := range traces {
		if tr.TenantID != "t-1" {
			t.Errorf("cross-tenant trace leak: trace %q with tenant_id %q visible to t-1", tr.ID, tr.TenantID)
		}
	}

	// Tenant 2 should only see its own 1 trace
	req = withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces", nil), "t-2", "tenant2")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var traces2 []service.Trace
	if err := json.NewDecoder(rec.Body).Decode(&traces2); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(traces2) != 1 {
		t.Fatalf("tenant t-2 should see 1 trace, got %d", len(traces2))
	}
}

// TestGetTrace_OtherTenantTrace_Returns404 verifies that requesting a trace ID
// that exists but belongs to a different tenant returns 404, not the actual trace.
// The handler must pass tenant ID to the service so cross-tenant lookups fail.
func TestGetTrace_OtherTenantTrace_Returns404(t *testing.T) {
	mock := &tenantFilteredMock{
		allDetails: map[string]*service.Trace{
			// trace belongs to t-1
			"tr-owned-by-t1": {ID: "tr-owned-by-t1", TenantID: "t-1", Query: "secret query"},
		},
	}

	r := setupTracesRouter(mock)

	// Tenant t-2 requests a trace that belongs to t-1
	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces/tr-owned-by-t1", nil), "t-2", "tenant2")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-tenant trace access, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestGetTrace_SameTenant_Returns200 is the positive case: the owner tenant can
// access the trace.
func TestGetTrace_SameTenant_Returns200(t *testing.T) {
	mock := &tenantFilteredMock{
		allDetails: map[string]*service.Trace{
			"tr-owned-by-t1": {ID: "tr-owned-by-t1", TenantID: "t-1", Query: "my query"},
		},
	}

	r := setupTracesRouter(mock)

	req := withTenantCtx(httptest.NewRequest(http.MethodGet, "/v1/traces/tr-owned-by-t1", nil), "t-1", "tenant1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for owner tenant, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- JSON error format (invariant #7) ---

// TestTracesHandler_ErrorResponses_AreJSON verifies that all HTTP error
// responses from the traces handler are valid JSON. The handler uses
// http.Error with a JSON string literal — this test catches any path
// where a plain-text error response is returned instead.
func TestTracesHandler_ErrorResponses_AreJSON(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func() (*chi.Mux, *http.Request)
	}{
		{
			name: "list traces — missing tenant context",
			setupFn: func() (*chi.Mux, *http.Request) {
				h := &Handler{svc: &mockTracesService{}}
				r := chi.NewRouter()
				r.Get("/v1/traces", h.ListTraces)
				return r, httptest.NewRequest(http.MethodGet, "/v1/traces", nil)
			},
		},
		{
			name: "get trace detail — missing tenant context",
			setupFn: func() (*chi.Mux, *http.Request) {
				h := &Handler{svc: &mockTracesService{}}
				r := chi.NewRouter()
				r.Get("/v1/traces/{traceID}", h.GetTraceDetail)
				return r, httptest.NewRequest(http.MethodGet, "/v1/traces/some-id", nil)
			},
		},
		{
			name: "get tenant cost — missing tenant context",
			setupFn: func() (*chi.Mux, *http.Request) {
				h := &Handler{svc: &mockTracesService{}}
				r := chi.NewRouter()
				r.Get("/v1/traces/costs", h.GetTenantCost)
				return r, httptest.NewRequest(http.MethodGet, "/v1/traces/costs", nil)
			},
		},
		{
			name: "list traces — service error",
			setupFn: func() (*chi.Mux, *http.Request) {
				h := &Handler{svc: &mockTracesService{err: errors.New("db down")}}
				r := chi.NewRouter()
				r.Get("/v1/traces", h.ListTraces)
				req := httptest.NewRequest(http.MethodGet, "/v1/traces", nil)
				ctx := tenant.WithInfo(req.Context(), tenant.Info{ID: "t-1", Slug: "test"})
				return r, req.WithContext(ctx)
			},
		},
		{
			name: "get trace detail — not found",
			setupFn: func() (*chi.Mux, *http.Request) {
				h := &Handler{svc: &mockTracesService{err: errors.New("trace not found: no rows")}}
				r := chi.NewRouter()
				r.Get("/v1/traces/{traceID}", h.GetTraceDetail)
				req := httptest.NewRequest(http.MethodGet, "/v1/traces/nope", nil)
				ctx := tenant.WithInfo(req.Context(), tenant.Info{ID: "t-1", Slug: "test"})
				return r, req.WithContext(ctx)
			},
		},
		{
			name: "get tenant cost — service error",
			setupFn: func() (*chi.Mux, *http.Request) {
				h := &Handler{svc: &mockTracesService{err: errors.New("query failed")}}
				r := chi.NewRouter()
				r.Get("/v1/traces/costs", h.GetTenantCost)
				req := httptest.NewRequest(http.MethodGet, "/v1/traces/costs", nil)
				ctx := tenant.WithInfo(req.Context(), tenant.Info{ID: "t-1", Slug: "test"})
				return r, req.WithContext(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, req := tt.setupFn()
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code < 400 {
				t.Fatalf("expected an error status, got %d", rec.Code)
			}

			// Must parse as JSON
			var resp map[string]any
			body := rec.Body.String()
			if err := json.Unmarshal([]byte(strings.TrimSpace(body)), &resp); err != nil {
				t.Errorf("error response is not valid JSON: status=%d body=%q", rec.Code, body)
				return
			}
			if _, ok := resp["error"]; !ok {
				t.Errorf("JSON error response missing 'error' key: %v", resp)
			}
		})
	}
}

// TestListTraces_WithTenantContext_RequiresService verifies tenant context passes the auth check.
func TestListTraces_WithTenantContext_RequiresService(t *testing.T) {
	// Verify that tenant context passes the auth check
	// (will panic/fail at service call since svc is nil — that's expected)
	h := &Handler{svc: nil}

	r := chi.NewRouter()
	r.Get("/v1/traces", h.ListTraces)

	req := httptest.NewRequest(http.MethodGet, "/v1/traces", nil)
	ctx := tenant.WithInfo(req.Context(), tenant.Info{ID: "t-1", Slug: "test"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Expect 500 (nil pointer on svc.ListTraces) not 401
	defer func() {
		if r := recover(); r != nil {
			// nil service panics — tenant check passed, which is what we're testing
			return
		}
	}()
	r.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Error("should have passed tenant check with valid context")
	}
}
