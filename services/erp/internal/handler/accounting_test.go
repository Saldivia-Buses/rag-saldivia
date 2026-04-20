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

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// --- mock ---

type mockAccountingService struct {
	accounts      []repository.ErpAccount
	costCenters   []repository.ErpCostCenter
	fiscalYears   []repository.ListFiscalYearsRow
	entries       []repository.ListJournalEntriesRow
	entryDetail   *service.EntryDetail
	balances      []repository.GetAccountBalanceRow
	ledger        []repository.GetLedgerRow
	previewResult *service.PreviewCloseResult
	closeResult   *service.CloseResult
	err           error
}

func (m *mockAccountingService) ListAccounts(_ context.Context, _ string, _ bool) ([]repository.ErpAccount, error) {
	return m.accounts, m.err
}

func (m *mockAccountingService) CreateAccount(_ context.Context, _, _, _ string, _ pgtype.UUID, _ string, _ bool, _ pgtype.UUID, _, _ string) (repository.ErpAccount, error) {
	if m.err != nil {
		return repository.ErpAccount{}, m.err
	}
	if len(m.accounts) > 0 {
		return m.accounts[0], nil
	}
	return repository.ErpAccount{}, nil
}

func (m *mockAccountingService) ListCostCenters(_ context.Context, _ string, _ bool) ([]repository.ErpCostCenter, error) {
	return m.costCenters, m.err
}

func (m *mockAccountingService) GetCostCenter(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpCostCenter, error) {
	if m.err != nil {
		return repository.ErpCostCenter{}, m.err
	}
	if len(m.costCenters) > 0 {
		return m.costCenters[0], nil
	}
	return repository.ErpCostCenter{}, nil
}

func (m *mockAccountingService) CreateCostCenter(_ context.Context, _, _, _ string, _ pgtype.UUID, _, _ string) (repository.ErpCostCenter, error) {
	if m.err != nil {
		return repository.ErpCostCenter{}, m.err
	}
	if len(m.costCenters) > 0 {
		return m.costCenters[0], nil
	}
	return repository.ErpCostCenter{}, nil
}

func (m *mockAccountingService) ListFiscalYears(_ context.Context, _ string) ([]repository.ListFiscalYearsRow, error) {
	return m.fiscalYears, m.err
}

func (m *mockAccountingService) CreateFiscalYear(_ context.Context, _ string, _ int, _, _, _, _ string) (repository.CreateFiscalYearRow, error) {
	if m.err != nil {
		return repository.CreateFiscalYearRow{}, m.err
	}
	return repository.CreateFiscalYearRow{}, nil
}

func (m *mockAccountingService) SetFiscalYearResultAccount(_ context.Context, _ string, _, _ pgtype.UUID, _, _ string) error {
	return m.err
}

func (m *mockAccountingService) PreviewClose(_ context.Context, _ string, _ pgtype.UUID) (*service.PreviewCloseResult, error) {
	return m.previewResult, m.err
}

func (m *mockAccountingService) CloseFiscalYear(_ context.Context, _ string, _ pgtype.UUID, _, _ string) (*service.CloseResult, error) {
	return m.closeResult, m.err
}

func (m *mockAccountingService) ListEntries(_ context.Context, _ string, _, _ pgtype.Date, _ string, _ pgtype.UUID, _, _ int) ([]repository.ListJournalEntriesRow, error) {
	return m.entries, m.err
}

func (m *mockAccountingService) GetEntry(_ context.Context, _ pgtype.UUID, _ string) (*service.EntryDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.entryDetail, nil
}

func (m *mockAccountingService) CreateEntry(_ context.Context, _ service.CreateEntryRequest) (*service.EntryDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.entryDetail, nil
}

func (m *mockAccountingService) PostEntry(_ context.Context, _ pgtype.UUID, _, _, _ string) error {
	return m.err
}

func (m *mockAccountingService) GetBalance(_ context.Context, _ string, _, _ pgtype.Date) ([]repository.GetAccountBalanceRow, error) {
	return m.balances, m.err
}

func (m *mockAccountingService) GetLedger(_ context.Context, _ string, _ pgtype.UUID, _, _ pgtype.Date, _, _ int) ([]repository.GetLedgerRow, error) {
	return m.ledger, m.err
}

// --- helpers ---

// withAdmin injects admin role + tenant slug into the request context,
// bypassing RequirePermission middleware (admin bypasses all permission checks).
func withAdmin(req *http.Request) *http.Request {
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	req.Header.Set("X-User-ID", "u-1")
	ctx := sdamw.WithRole(req.Context(), "admin")
	return req.WithContext(ctx)
}

func setupAccountingRouter(mock *mockAccountingService) *chi.Mux {
	h := NewAccounting(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/accounting", h.Routes(noopMiddleware))
	return r
}

func decodeAccountingJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests ---

func TestAccounting_ListAccounts_Success(t *testing.T) {
	mock := &mockAccountingService{
		accounts: []repository.ErpAccount{
			{Code: "1.1.1", Name: "Caja"},
			{Code: "1.1.2", Name: "Banco"},
		},
	}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/accounts", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeAccountingJSON(t, rec, &resp)
	accounts, ok := resp["accounts"].([]any)
	if !ok || len(accounts) != 2 {
		t.Errorf("expected 2 accounts in response, got %v", resp["accounts"])
	}
}

func TestAccounting_ListAccounts_ServiceError_Returns500(t *testing.T) {
	mock := &mockAccountingService{err: errors.New("db connection lost")}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/accounts", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	decodeAccountingJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error message, got %q", resp["error"])
	}
}

func TestAccounting_CreateAccount_Success(t *testing.T) {
	mock := &mockAccountingService{
		accounts: []repository.ErpAccount{{Code: "1.1.1", Name: "Caja"}},
	}
	r := setupAccountingRouter(mock)

	body := `{"code":"1.1.1","name":"Caja","account_type":"asset","is_detail":true}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/accounts", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccounting_CreateAccount_InvalidBody_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/accounts", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAccounting_CreateAccount_ServiceError_Returns500(t *testing.T) {
	mock := &mockAccountingService{err: errors.New("duplicate code")}
	r := setupAccountingRouter(mock)

	body := `{"code":"1.1.1","name":"Caja","account_type":"asset","is_detail":true}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/accounts", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestAccounting_ListFiscalYears_Success(t *testing.T) {
	mock := &mockAccountingService{
		fiscalYears: []repository.ListFiscalYearsRow{{Year: 2024}},
	}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/fiscal-years", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccounting_CreateFiscalYear_InvalidBody_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/fiscal-years", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAccounting_CreateFiscalYear_Success(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	body := `{"year":2025,"start_date":"2025-01-01","end_date":"2025-12-31"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/fiscal-years", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccounting_SetResultAccount_InvalidFiscalYearID_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	body := `{"result_account_id":"00000000-0000-0000-0000-000000000001"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/accounting/fiscal-years/not-a-uuid/result-account", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestAccounting_SetResultAccount_MissingAccountID_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	body := `{}` // result_account_id missing
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/accounting/fiscal-years/00000000-0000-0000-0000-000000000001/result-account", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing result_account_id, got %d", rec.Code)
	}
}

func TestAccounting_SetResultAccount_Success(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	body := `{"result_account_id":"00000000-0000-0000-0000-000000000002"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/accounting/fiscal-years/00000000-0000-0000-0000-000000000001/result-account", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccounting_PreviewClose_InvalidID_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/fiscal-years/bad-uuid/preview-close", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestAccounting_PreviewClose_Success(t *testing.T) {
	mock := &mockAccountingService{
		previewResult: &service.PreviewCloseResult{CanClose: true},
	}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/fiscal-years/00000000-0000-0000-0000-000000000001/preview-close", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeAccountingJSON(t, rec, &resp)
	canClose, ok := resp["can_close"].(bool)
	if !ok || !canClose {
		t.Errorf("expected can_close=true in response, got %v", resp)
	}
}

func TestAccounting_PreviewClose_BusinessError_Returns400(t *testing.T) {
	mock := &mockAccountingService{
		err: errors.New("fiscal year is not open"),
	}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/fiscal-years/00000000-0000-0000-0000-000000000001/preview-close", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for business error, got %d", rec.Code)
	}
}

func TestAccounting_CloseFiscalYear_InvalidID_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/fiscal-years/bad-uuid/close", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAccounting_CloseFiscalYear_Success(t *testing.T) {
	mock := &mockAccountingService{
		closeResult: &service.CloseResult{},
	}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/fiscal-years/00000000-0000-0000-0000-000000000001/close", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccounting_CloseFiscalYear_DraftEntriesError_Returns400(t *testing.T) {
	mock := &mockAccountingService{
		err: errors.New("draft entries must be posted or deleted before closing"),
	}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/fiscal-years/00000000-0000-0000-0000-000000000001/close", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for draft entries error, got %d", rec.Code)
	}
}

func TestAccounting_PostEntry_InvalidID_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/entries/bad-uuid/post", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAccounting_PostEntry_Success(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/entries/00000000-0000-0000-0000-000000000001/post", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccounting_PostEntry_ServiceError_Returns500(t *testing.T) {
	mock := &mockAccountingService{err: errors.New("db error")}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/accounting/entries/00000000-0000-0000-0000-000000000001/post", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestAccounting_GetEntry_InvalidID_Returns400(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/entries/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAccounting_GetEntry_NotFound_Returns404(t *testing.T) {
	mock := &mockAccountingService{err: errors.New("not found")}
	r := setupAccountingRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/entries/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestAccounting_RequirePermission_WithoutAdminRole_Returns403(t *testing.T) {
	r := setupAccountingRouter(&mockAccountingService{})

	// Request without role → RequirePermission should block with 403
	req := httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/accounts", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	// No role injected in context → no permissions → 403
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
