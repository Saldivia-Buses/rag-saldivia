package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/pkg/tenant"
)

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
