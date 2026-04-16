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

type mockMaintenanceService struct {
	assets    []repository.ErpMaintenanceAsset
	asset     repository.ErpMaintenanceAsset
	plans     []repository.ErpMaintenancePlan
	plan      repository.ErpMaintenancePlan
	workOrders []repository.ListWorkOrdersRow
	workOrder  *service.WorkOrderDetail
	createdWO  repository.ErpWorkOrder
	fuelLogs  []repository.ListFuelLogsRow
	fuelLog   repository.ErpFuelLog
	err       error
}

func (m *mockMaintenanceService) ListAssets(_ context.Context, _, _ string, _ bool) ([]repository.ErpMaintenanceAsset, error) {
	return m.assets, m.err
}

func (m *mockMaintenanceService) CreateAsset(_ context.Context, _ repository.CreateMaintenanceAssetParams, _, _ string) (repository.ErpMaintenanceAsset, error) {
	if m.err != nil {
		return repository.ErpMaintenanceAsset{}, m.err
	}
	return m.asset, nil
}

func (m *mockMaintenanceService) ListPlans(_ context.Context, _ string, _ pgtype.UUID) ([]repository.ErpMaintenancePlan, error) {
	return m.plans, m.err
}

func (m *mockMaintenanceService) CreatePlan(_ context.Context, _ repository.CreateMaintenancePlanParams, _, _ string) (repository.ErpMaintenancePlan, error) {
	if m.err != nil {
		return repository.ErpMaintenancePlan{}, m.err
	}
	return m.plan, nil
}

func (m *mockMaintenanceService) ListWorkOrders(_ context.Context, _, _ string, _, _ int) ([]repository.ListWorkOrdersRow, error) {
	return m.workOrders, m.err
}

func (m *mockMaintenanceService) GetWorkOrder(_ context.Context, _ pgtype.UUID, _ string) (*service.WorkOrderDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.workOrder, nil
}

func (m *mockMaintenanceService) CreateWorkOrder(_ context.Context, _ repository.CreateWorkOrderParams, _ string) (repository.ErpWorkOrder, error) {
	if m.err != nil {
		return repository.ErpWorkOrder{}, m.err
	}
	return m.createdWO, nil
}

func (m *mockMaintenanceService) UpdateWorkOrderStatus(_ context.Context, _ pgtype.UUID, _, _, _, _ string) error {
	return m.err
}

func (m *mockMaintenanceService) ListFuelLogs(_ context.Context, _ string, _ pgtype.UUID, _, _ int) ([]repository.ListFuelLogsRow, error) {
	return m.fuelLogs, m.err
}

func (m *mockMaintenanceService) CreateFuelLog(_ context.Context, _ repository.CreateFuelLogParams, _ string) (repository.ErpFuelLog, error) {
	if m.err != nil {
		return repository.ErpFuelLog{}, m.err
	}
	return m.fuelLog, nil
}

// --- helpers ---

func setupMaintenanceRouter(mock *mockMaintenanceService) *chi.Mux {
	h := NewMaintenance(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/maintenance", h.Routes(noopMiddleware))
	return r
}

// --- tests ---

func TestMaintenance_ListAssets_Success(t *testing.T) {
	mock := &mockMaintenanceService{
		assets: []repository.ErpMaintenanceAsset{
			{Code: "BUS-001", Name: "Bus 001"},
		},
	}
	r := setupMaintenanceRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/maintenance/assets", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMaintenance_ListAssets_ServiceError_Returns500(t *testing.T) {
	mock := &mockMaintenanceService{err: errors.New("db error")}
	r := setupMaintenanceRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/maintenance/assets", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestMaintenance_CreateWorkOrder_InvalidBody_Returns400(t *testing.T) {
	r := setupMaintenanceRouter(&mockMaintenanceService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/maintenance/work-orders", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMaintenance_CreateWorkOrder_Success(t *testing.T) {
	mock := &mockMaintenanceService{createdWO: repository.ErpWorkOrder{Number: "WO-001"}}
	r := setupMaintenanceRouter(mock)

	body := `{"Number":"WO-001","AssetID":"00000000-0000-0000-0000-000000000001","WorkType":"preventive","Description":"Oil change","Date":"2025-01-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/maintenance/work-orders", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMaintenance_CreateWorkOrder_ServiceError_Returns500(t *testing.T) {
	mock := &mockMaintenanceService{err: errors.New("number and description are required")}
	r := setupMaintenanceRouter(mock)

	body := `{"Number":"WO-001","AssetID":"00000000-0000-0000-0000-000000000001","WorkType":"preventive","Description":"Oil change","Date":"2025-01-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/maintenance/work-orders", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestMaintenance_UpdateWorkOrderStatus_InvalidUUID_Returns400(t *testing.T) {
	r := setupMaintenanceRouter(&mockMaintenanceService{})

	body := `{"status":"completed"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/maintenance/work-orders/bad-uuid/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMaintenance_UpdateWorkOrderStatus_Success(t *testing.T) {
	r := setupMaintenanceRouter(&mockMaintenanceService{})

	body := `{"status":"completed"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/maintenance/work-orders/00000000-0000-0000-0000-000000000001/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMaintenance_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupMaintenanceRouter(&mockMaintenanceService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/maintenance/assets", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
