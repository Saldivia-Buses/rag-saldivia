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

type mockInvoicingService struct {
	invoices     []repository.ListInvoicesRow
	invoiceDetail *service.InvoiceDetail
	taxBook      []repository.GetTaxBookRow
	withholdings []repository.ListWithholdingsRow
	voidPreview  *service.VoidPreviewResult
	voidResult   *service.VoidResult
	err          error
}

func (m *mockInvoicingService) ListInvoices(_ context.Context, _ string, _, _, _ string, _, _ pgtype.Date, _, _ int) ([]repository.ListInvoicesRow, error) {
	return m.invoices, m.err
}

func (m *mockInvoicingService) GetInvoice(_ context.Context, _ pgtype.UUID, _ string) (*service.InvoiceDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.invoiceDetail, nil
}

func (m *mockInvoicingService) CreateInvoice(_ context.Context, _ service.CreateInvoiceRequest) (*service.InvoiceDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.invoiceDetail, nil
}

func (m *mockInvoicingService) PostInvoice(_ context.Context, _ pgtype.UUID, _, _, _ string) error {
	return m.err
}

func (m *mockInvoicingService) GetTaxBook(_ context.Context, _, _, _ string) ([]repository.GetTaxBookRow, error) {
	return m.taxBook, m.err
}

func (m *mockInvoicingService) ListWithholdings(_ context.Context, _, _ string, _, _ int) ([]repository.ListWithholdingsRow, error) {
	return m.withholdings, m.err
}

func (m *mockInvoicingService) CreateWithholding(_ context.Context, _ repository.CreateWithholdingParams, _, _ string) (repository.ErpWithholding, error) {
	if m.err != nil {
		return repository.ErpWithholding{}, m.err
	}
	return repository.ErpWithholding{}, nil
}

func (m *mockInvoicingService) VoidPreview(_ context.Context, _ pgtype.UUID, _ string) (*service.VoidPreviewResult, error) {
	return m.voidPreview, m.err
}

func (m *mockInvoicingService) VoidInvoice(_ context.Context, _ pgtype.UUID, _, _, _, _ string) (*service.VoidResult, error) {
	return m.voidResult, m.err
}

func (m *mockInvoicingService) ListInvoiceNotes(_ context.Context, _ string, _ pgtype.UUID, _, _ pgtype.Date, _, _ int) ([]repository.ErpInvoiceNote, error) {
	return nil, m.err
}

// --- helpers ---

func setupInvoicingRouter(mock *mockInvoicingService) *chi.Mux {
	h := NewInvoicing(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/invoicing", h.Routes(noopMiddleware))
	return r
}

func decodeInvoicingJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests ---

func TestInvoicing_ListInvoices_Success(t *testing.T) {
	mock := &mockInvoicingService{
		invoices: []repository.ListInvoicesRow{{Number: "FAC-001"}, {Number: "FAC-002"}},
	}
	r := setupInvoicingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeInvoicingJSON(t, rec, &resp)
	invoices, ok := resp["invoices"].([]any)
	if !ok || len(invoices) != 2 {
		t.Errorf("expected 2 invoices in response, got %v", resp["invoices"])
	}
}

func TestInvoicing_ListInvoices_ServiceError_Returns500(t *testing.T) {
	mock := &mockInvoicingService{err: errors.New("db error")}
	r := setupInvoicingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	var resp map[string]string
	decodeInvoicingJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error message, got %q", resp["error"])
	}
}

func TestInvoicing_GetInvoice_InvalidID_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestInvoicing_GetInvoice_NotFound_Returns404(t *testing.T) {
	mock := &mockInvoicingService{err: errors.New("not found")}
	r := setupInvoicingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestInvoicing_CreateInvoice_InvalidBody_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestInvoicing_CreateInvoice_InvalidEntityID_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	body := `{"number":"FAC-001","invoice_type":"invoice_a","direction":"issued","entity_id":"not-a-uuid","lines":[{"description":"Svc","quantity":"1","unit_price":"100","tax_rate":"21"}]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid entity_id, got %d", rec.Code)
	}
}

func TestInvoicing_CreateInvoice_Success(t *testing.T) {
	mock := &mockInvoicingService{
		invoiceDetail: &service.InvoiceDetail{},
	}
	r := setupInvoicingRouter(mock)

	body := `{"number":"FAC-001","invoice_type":"invoice_a","direction":"issued","entity_id":"00000000-0000-0000-0000-000000000001","lines":[{"description":"Svc","quantity":"1","unit_price":"100","tax_rate":"21"}]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvoicing_CreateInvoice_ServiceError_Returns500(t *testing.T) {
	mock := &mockInvoicingService{err: errors.New("duplicate invoice number")}
	r := setupInvoicingRouter(mock)

	body := `{"number":"FAC-001","invoice_type":"invoice_a","direction":"issued","entity_id":"00000000-0000-0000-0000-000000000001","lines":[{"description":"Svc","quantity":"1","unit_price":"100","tax_rate":"21"}]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestInvoicing_PostInvoice_InvalidID_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices/bad-uuid/post", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestInvoicing_PostInvoice_Success(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001/post", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvoicing_PostInvoice_ServiceError_Returns500(t *testing.T) {
	mock := &mockInvoicingService{err: errors.New("already posted")}
	r := setupInvoicingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001/post", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestInvoicing_GetTaxBook_MissingPeriod_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	// No period query param
	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/tax-book", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing period, got %d", rec.Code)
	}
}

func TestInvoicing_GetTaxBook_WithPeriod_Success(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{taxBook: []repository.GetTaxBookRow{}})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/tax-book?period=2024-01", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvoicing_VoidPreview_InvalidID_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices/bad-uuid/void-preview", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestInvoicing_VoidPreview_Success(t *testing.T) {
	mock := &mockInvoicingService{
		voidPreview: &service.VoidPreviewResult{TaxEntryCount: 2},
	}
	r := setupInvoicingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001/void-preview", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvoicing_VoidPreview_NotPosted_Returns400(t *testing.T) {
	mock := &mockInvoicingService{
		err: errors.New("solo se pueden anular facturas posted/paid"),
	}
	r := setupInvoicingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001/void-preview", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unpostable void preview, got %d", rec.Code)
	}
}

func TestInvoicing_VoidInvoice_InvalidID_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	body := `{"reason":"Error de imputacion"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices/bad-uuid/void", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestInvoicing_VoidInvoice_InvalidBody_Returns400(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001/void", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestInvoicing_VoidInvoice_Success(t *testing.T) {
	mock := &mockInvoicingService{
		voidResult: &service.VoidResult{TaxEntriesReversed: 2},
	}
	r := setupInvoicingRouter(mock)

	body := `{"reason":"Error de imputacion"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001/void", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInvoicing_VoidInvoice_CAEBlocked_Returns400(t *testing.T) {
	// Invoice with CAE cannot be voided without Plan 19
	mock := &mockInvoicingService{
		err: errors.New("factura con CAE requiere Nota de Crédito AFIP (Plan 19)"),
	}
	r := setupInvoicingRouter(mock)

	body := `{"reason":"Error"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/invoicing/invoices/00000000-0000-0000-0000-000000000001/void", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for CAE-blocked void, got %d", rec.Code)
	}
}

func TestInvoicing_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupInvoicingRouter(&mockInvoicingService{})

	// No role in context → RequirePermission blocks
	req := httptest.NewRequest(http.MethodGet, "/v1/erp/invoicing/invoices", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
