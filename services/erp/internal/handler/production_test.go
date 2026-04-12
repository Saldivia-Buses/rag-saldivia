package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// --- mock ---

type mockProductionService struct {
	centers     []repository.ErpProductionCenter
	center      repository.ErpProductionCenter
	orders      []repository.ListProductionOrdersRow
	order       *service.ProductionOrderDetail
	createdOrder repository.ErpProductionOrder
	units       []repository.ListUnitsRow
	unit        repository.ErpUnit
	inspection  repository.ErpProductionInspection
	err         error
}

func (m *mockProductionService) ListCenters(_ context.Context, _ string) ([]repository.ErpProductionCenter, error) {
	return m.centers, m.err
}

func (m *mockProductionService) CreateCenter(_ context.Context, _, _, _, _, _ string) (repository.ErpProductionCenter, error) {
	if m.err != nil {
		return repository.ErpProductionCenter{}, m.err
	}
	return m.center, nil
}

func (m *mockProductionService) ListOrders(_ context.Context, _, _ string, _, _ int) ([]repository.ListProductionOrdersRow, error) {
	return m.orders, m.err
}

func (m *mockProductionService) GetOrder(_ context.Context, _ pgtype.UUID, _ string) (*service.ProductionOrderDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.order, nil
}

func (m *mockProductionService) CreateOrder(_ context.Context, _ service.CreateProductionOrderRequest) (repository.ErpProductionOrder, error) {
	if m.err != nil {
		return repository.ErpProductionOrder{}, m.err
	}
	return m.createdOrder, nil
}

func (m *mockProductionService) StartOrder(_ context.Context, _ pgtype.UUID, _, _, _ string) error {
	return m.err
}

func (m *mockProductionService) UpdateStep(_ context.Context, _ pgtype.UUID, _, _, _, _, _ string) error {
	return m.err
}

func (m *mockProductionService) CreateInspection(_ context.Context, _ repository.CreateProductionInspectionParams, _, _ string) (repository.ErpProductionInspection, error) {
	if m.err != nil {
		return repository.ErpProductionInspection{}, m.err
	}
	return m.inspection, nil
}

func (m *mockProductionService) ListUnits(_ context.Context, _, _ string, _, _ int) ([]repository.ListUnitsRow, error) {
	return m.units, m.err
}

func (m *mockProductionService) GetUnit(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpUnit, error) {
	if m.err != nil {
		return repository.ErpUnit{}, m.err
	}
	return m.unit, nil
}

func (m *mockProductionService) CreateUnit(_ context.Context, _ repository.CreateUnitParams, _, _ string) (repository.ErpUnit, error) {
	if m.err != nil {
		return repository.ErpUnit{}, m.err
	}
	return m.unit, nil
}

// --- helpers ---

func setupProductionRouter(mock *mockProductionService) *chi.Mux {
	h := NewProduction(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/production", h.Routes(noopMiddleware))
	return r
}

// --- tests ---

func TestProduction_ListOrders_Success(t *testing.T) {
	mock := &mockProductionService{
		orders: []repository.ListProductionOrdersRow{
			{Number: "PO-001"},
			{Number: "PO-002"},
		},
	}
	r := setupProductionRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/production/orders", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProduction_ListOrders_ServiceError_Returns500(t *testing.T) {
	mock := &mockProductionService{err: errors.New("db error")}
	r := setupProductionRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/production/orders", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestProduction_CreateOrder_InvalidBody_Returns400(t *testing.T) {
	r := setupProductionRouter(&mockProductionService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/production/orders", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestProduction_CreateOrder_Success(t *testing.T) {
	mock := &mockProductionService{createdOrder: repository.ErpProductionOrder{Number: "PO-001"}}
	r := setupProductionRouter(mock)

	body := `{"number":"PO-001","date":"2025-01-01","product_id":"00000000-0000-0000-0000-000000000001","quantity":"10"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/production/orders", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProduction_CreateOrder_ServiceError_Returns500(t *testing.T) {
	mock := &mockProductionService{err: errors.New("duplicate number")}
	r := setupProductionRouter(mock)

	body := `{"number":"PO-001","date":"2025-01-01","product_id":"00000000-0000-0000-0000-000000000001","quantity":"10"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/production/orders", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestProduction_StartOrder_InvalidUUID_Returns400(t *testing.T) {
	r := setupProductionRouter(&mockProductionService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/production/orders/bad-uuid/start", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestProduction_StartOrder_Success(t *testing.T) {
	r := setupProductionRouter(&mockProductionService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/production/orders/00000000-0000-0000-0000-000000000001/start", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProduction_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupProductionRouter(&mockProductionService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/production/orders", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
