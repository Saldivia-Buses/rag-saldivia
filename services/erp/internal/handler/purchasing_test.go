package handler

import (
	"context"
	"encoding/json"
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

type mockPurchasingService struct {
	orders      []repository.ListPurchaseOrdersRow
	orderDetail *service.OrderDetail
	receipts    []repository.ListPurchaseReceiptsRow
	inspections []repository.ErpQcInspection
	listInspections []repository.ListInspectionsRow
	inspection  repository.ErpQcInspection
	demerits    []repository.ErpSupplierDemerit
	demeritTotal int32
	err         error
}

func (m *mockPurchasingService) ListOrders(_ context.Context, _, _ string, _, _ int) ([]repository.ListPurchaseOrdersRow, error) {
	return m.orders, m.err
}

func (m *mockPurchasingService) GetOrder(_ context.Context, _ pgtype.UUID, _ string) (*service.OrderDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.orderDetail, nil
}

func (m *mockPurchasingService) CreateOrder(_ context.Context, _ service.CreateOrderRequest) (*service.OrderDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.orderDetail, nil
}

func (m *mockPurchasingService) ApproveOrder(_ context.Context, _ pgtype.UUID, _, _, _ string) error {
	return m.err
}

func (m *mockPurchasingService) Receive(_ context.Context, _ service.ReceiveRequest) error {
	return m.err
}

func (m *mockPurchasingService) ListReceipts(_ context.Context, _ string, _, _ int) ([]repository.ListPurchaseReceiptsRow, error) {
	return m.receipts, m.err
}

func (m *mockPurchasingService) InspectReceipt(_ context.Context, _ string, _ pgtype.UUID, _ []service.InspectionInput, _, _ string) ([]repository.ErpQcInspection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.inspections, nil
}

func (m *mockPurchasingService) ListInspections(_ context.Context, _, _ string, _, _ int) ([]repository.ListInspectionsRow, error) {
	return m.listInspections, m.err
}

func (m *mockPurchasingService) GetInspection(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpQcInspection, error) {
	if m.err != nil {
		return repository.ErpQcInspection{}, m.err
	}
	return m.inspection, nil
}

func (m *mockPurchasingService) ListSupplierDemerits(_ context.Context, _ string, _ pgtype.UUID) ([]repository.ErpSupplierDemerit, error) {
	return m.demerits, m.err
}

func (m *mockPurchasingService) GetSupplierDemeritTotal(_ context.Context, _ string, _ pgtype.UUID) (int32, error) {
	return m.demeritTotal, m.err
}

// --- helpers ---

func setupPurchasingRouter(mock *mockPurchasingService) *chi.Mux {
	h := NewPurchasing(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/purchasing", h.Routes(noopMiddleware))
	return r
}

func decodePurchasingJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests: InspectReceipt ---

func TestPurchasing_InspectReceipt_InvalidID_Returns400(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	body := `{"inspections":[{"result":"pass"}]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/purchasing/receipts/bad-uuid/inspect", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestPurchasing_InspectReceipt_InvalidBody_Returns400(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/purchasing/receipts/00000000-0000-0000-0000-000000000001/inspect", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestPurchasing_InspectReceipt_EmptyInspections_Returns400(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	// Empty inspections array triggers the validation in handler
	body := `{"inspections":[]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/purchasing/receipts/00000000-0000-0000-0000-000000000001/inspect", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty inspections, got %d", rec.Code)
	}
}

func TestPurchasing_InspectReceipt_Success(t *testing.T) {
	mock := &mockPurchasingService{
		inspections: []repository.ErpQcInspection{{Status: "pass"}},
	}
	r := setupPurchasingRouter(mock)

	body := `{"inspections":[{"receipt_line_id":"00000000-0000-0000-0000-000000000001","result":"pass","notes":"OK"}]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/purchasing/receipts/00000000-0000-0000-0000-000000000001/inspect", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodePurchasingJSON(t, rec, &resp)
	inspections, ok := resp["inspections"].([]any)
	if !ok || len(inspections) != 1 {
		t.Errorf("expected 1 inspection in response, got %v", resp["inspections"])
	}
}

func TestPurchasing_InspectReceipt_ServiceError_Returns500(t *testing.T) {
	mock := &mockPurchasingService{err: errors.New("receipt not found")}
	r := setupPurchasingRouter(mock)

	body := `{"inspections":[{"receipt_line_id":"00000000-0000-0000-0000-000000000001","result":"pass"}]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/purchasing/receipts/00000000-0000-0000-0000-000000000001/inspect", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	var resp map[string]string
	decodePurchasingJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q", resp["error"])
	}
}

// --- tests: ListInspections ---

func TestPurchasing_ListInspections_Success(t *testing.T) {
	mock := &mockPurchasingService{
		listInspections: []repository.ListInspectionsRow{{Status: "pass"}, {Status: "fail"}},
	}
	r := setupPurchasingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/inspections", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodePurchasingJSON(t, rec, &resp)
	inspections, ok := resp["inspections"].([]any)
	if !ok || len(inspections) != 2 {
		t.Errorf("expected 2 inspections, got %v", resp["inspections"])
	}
}

func TestPurchasing_ListInspections_ServiceError_Returns500(t *testing.T) {
	mock := &mockPurchasingService{err: errors.New("db error")}
	r := setupPurchasingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/inspections", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// --- tests: ListSupplierDemerits ---

func TestPurchasing_ListSupplierDemerits_InvalidID_Returns400(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/suppliers/bad-uuid/demerits", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPurchasing_ListSupplierDemerits_Success(t *testing.T) {
	mock := &mockPurchasingService{
		demerits:    []repository.ErpSupplierDemerit{{Points: 5}, {Points: 10}},
		demeritTotal: 15,
	}
	r := setupPurchasingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/suppliers/00000000-0000-0000-0000-000000000001/demerits", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodePurchasingJSON(t, rec, &resp)

	demerits, ok := resp["demerits"].([]any)
	if !ok || len(demerits) != 2 {
		t.Errorf("expected 2 demerits, got %v", resp["demerits"])
	}

	total, ok := resp["total_points"].(float64)
	if !ok || total != 15 {
		t.Errorf("expected total_points=15, got %v", resp["total_points"])
	}
}

func TestPurchasing_ListSupplierDemerits_ServiceError_Returns500(t *testing.T) {
	mock := &mockPurchasingService{err: errors.New("db error")}
	r := setupPurchasingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/suppliers/00000000-0000-0000-0000-000000000001/demerits", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// --- tests: orders ---

func TestPurchasing_ListOrders_Success(t *testing.T) {
	mock := &mockPurchasingService{
		orders: []repository.ListPurchaseOrdersRow{{Number: "OC-001"}},
	}
	r := setupPurchasingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/orders", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPurchasing_GetOrder_InvalidID_Returns400(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/orders/bad-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPurchasing_GetOrder_NotFound_Returns404(t *testing.T) {
	mock := &mockPurchasingService{err: errors.New("not found")}
	r := setupPurchasingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/orders/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestPurchasing_CreateOrder_InvalidBody_Returns400(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/purchasing/orders", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPurchasing_CreateOrder_InvalidSupplierID_Returns400(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	body := `{"number":"OC-001","supplier_id":"not-a-uuid","date":"2024-01-01","lines":[{"article_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"100"}]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/purchasing/orders", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid supplier_id, got %d", rec.Code)
	}
}

func TestPurchasing_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupPurchasingRouter(&mockPurchasingService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/purchasing/orders", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
