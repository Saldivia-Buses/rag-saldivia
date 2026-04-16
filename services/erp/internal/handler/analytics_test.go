package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// --- mock ---

// mockAnalyticsService returns a nil *repository.Queries.
// Tests that actually call Repo() will panic — only test permission layer here.
type mockAnalyticsService struct{}

func (m *mockAnalyticsService) Repo() *repository.Queries { return nil }

// --- helpers ---

func setupAnalyticsRouter(mock AnalyticsService) *chi.Mux {
	h := NewAnalytics(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/analytics", h.Routes(noopMiddleware))
	return r
}

// --- tests ---

func TestAnalytics_RequirePermission_AccountingWithoutRole_Returns403(t *testing.T) {
	r := setupAnalyticsRouter(&mockAnalyticsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/analytics/accounting/balance-evolution", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}

func TestAnalytics_RequirePermission_TreasuryWithoutRole_Returns403(t *testing.T) {
	r := setupAnalyticsRouter(&mockAnalyticsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/analytics/treasury/cash-flow", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for treasury without auth, got %d", rec.Code)
	}
}

func TestAnalytics_RequirePermission_StockWithoutRole_Returns403(t *testing.T) {
	r := setupAnalyticsRouter(&mockAnalyticsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/analytics/stock/valuation", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for stock without auth, got %d", rec.Code)
	}
}

func TestAnalytics_RequirePermission_HRWithoutRole_Returns403(t *testing.T) {
	r := setupAnalyticsRouter(&mockAnalyticsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/analytics/hr/headcount", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for hr without auth, got %d", rec.Code)
	}
}

func TestAnalytics_RequirePermission_DashboardWithoutRole_Returns403(t *testing.T) {
	r := setupAnalyticsRouter(&mockAnalyticsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/analytics/dashboard/kpis", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for dashboard without auth, got %d", rec.Code)
	}
}

func TestAnalytics_RequirePermission_ProductionWithoutRole_Returns403(t *testing.T) {
	r := setupAnalyticsRouter(&mockAnalyticsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/analytics/production/orders-by-status", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for production without auth, got %d", rec.Code)
	}
}
