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

type mockSalesService struct {
	quotations []repository.ListQuotationsRow
	quotation  *service.QuotationDetail
	orders     []repository.ListOrdersRow
	order      repository.ErpOrder
	priceLists []repository.ErpPriceList
	err        error
}

func (m *mockSalesService) ListQuotations(_ context.Context, _, _ string, _, _ int) ([]repository.ListQuotationsRow, error) {
	return m.quotations, m.err
}

func (m *mockSalesService) GetQuotation(_ context.Context, _ pgtype.UUID, _ string) (*service.QuotationDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quotation, nil
}

func (m *mockSalesService) CreateQuotation(_ context.Context, _ service.CreateQuotationRequest) (*service.QuotationDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quotation, nil
}

func (m *mockSalesService) ListOrders(_ context.Context, _, _, _ string, _, _ int) ([]repository.ListOrdersRow, error) {
	return m.orders, m.err
}

func (m *mockSalesService) CreateOrder(_ context.Context, _, _ string, _ pgtype.Date, _ string, _, _ pgtype.UUID, _ pgtype.Numeric, _, _, _ string) (repository.ErpOrder, error) {
	if m.err != nil {
		return repository.ErpOrder{}, m.err
	}
	return m.order, nil
}

func (m *mockSalesService) UpdateOrderStatus(_ context.Context, _ pgtype.UUID, _, _, _, _ string) error {
	return m.err
}

func (m *mockSalesService) ListPriceLists(_ context.Context, _ string) ([]repository.ErpPriceList, error) {
	return m.priceLists, m.err
}

func (m *mockSalesService) GetPriceList(_ context.Context, _ pgtype.UUID, _ string) (*service.PriceListDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, errors.New("not implemented in mock")
}

// --- helpers ---

func setupSalesRouter(mock SalesService) *chi.Mux {
	h := NewSales(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/sales", h.Routes(noopMiddleware))
	return r
}

func decodeSalesJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests ---

func TestSales_ListOrders_Success(t *testing.T) {
	mock := &mockSalesService{
		orders: []repository.ListOrdersRow{
			{Number: "ORD-001"},
			{Number: "ORD-002"},
		},
	}
	r := setupSalesRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/sales/orders", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeSalesJSON(t, rec, &resp)
	orders, ok := resp["orders"].([]any)
	if !ok || len(orders) != 2 {
		t.Errorf("expected 2 orders in response, got %v", resp["orders"])
	}
}

func TestSales_ListOrders_ServiceError_Returns500(t *testing.T) {
	mock := &mockSalesService{err: errors.New("db error")}
	r := setupSalesRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/sales/orders", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	decodeSalesJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q", resp["error"])
	}
}

func TestSales_CreateOrder_InvalidBody_Returns400(t *testing.T) {
	r := setupSalesRouter(&mockSalesService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/sales/orders", strings.NewReader("bad json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSales_CreateOrder_Success(t *testing.T) {
	mock := &mockSalesService{order: repository.ErpOrder{Number: "ORD-001"}}
	r := setupSalesRouter(mock)

	body := `{"number":"ORD-001","date":"2025-01-15","order_type":"sale","total":"1500.00"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/sales/orders", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSales_CreateOrder_ServiceError_Returns500(t *testing.T) {
	mock := &mockSalesService{err: errors.New("db error")}
	r := setupSalesRouter(mock)

	body := `{"number":"ORD-001","date":"2025-01-15","order_type":"sale","total":"1500.00"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/sales/orders", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestSales_UpdateOrderStatus_InvalidUUID_Returns400(t *testing.T) {
	r := setupSalesRouter(&mockSalesService{})

	body := `{"status":"confirmed"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/sales/orders/not-a-uuid/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestSales_UpdateOrderStatus_Success(t *testing.T) {
	r := setupSalesRouter(&mockSalesService{})

	body := `{"status":"confirmed"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/sales/orders/00000000-0000-0000-0000-000000000001/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSales_UpdateOrderStatus_NotFound_Returns404(t *testing.T) {
	mock := &mockSalesService{err: errors.New("order not found")}
	r := setupSalesRouter(mock)

	body := `{"status":"confirmed"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/sales/orders/00000000-0000-0000-0000-000000000001/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// UpdateOrderStatus returns 404 on service error (see handler)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on not found, got %d", rec.Code)
	}
}

func TestSales_GetQuotation_InvalidUUID_Returns400(t *testing.T) {
	r := setupSalesRouter(&mockSalesService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/sales/quotations/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestSales_GetQuotation_NotFound_Returns404(t *testing.T) {
	mock := &mockSalesService{err: errors.New("not found")}
	r := setupSalesRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/sales/quotations/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSales_GetQuotation_Success(t *testing.T) {
	mock := &mockSalesService{
		quotation: &service.QuotationDetail{
			Quotation: repository.ErpQuotation{Number: "Q-001"},
		},
	}
	r := setupSalesRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/sales/quotations/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSales_ListPriceLists_Success(t *testing.T) {
	mock := &mockSalesService{
		priceLists: []repository.ErpPriceList{{Name: "Lista A"}},
	}
	r := setupSalesRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/sales/price-lists", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSales_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupSalesRouter(&mockSalesService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/sales/orders", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
