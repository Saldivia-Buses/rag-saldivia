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

type mockTreasuryService struct {
	bankAccounts    []repository.ErpBankAccount
	cashRegisters   []repository.ErpCashRegister
	movements       []repository.ListTreasuryMovementsRow
	checks          []repository.ErpCheck
	balances        []repository.GetTreasuryBalanceRow
	cashCounts      []repository.ErpCashCount
	receipts        []repository.ListReceiptsRow
	receiptDetail   *service.ReceiptDetail
	reconciliations []repository.ListReconciliationsRow
	reconDetail     *service.ReconciliationDetail
	autoMatchResult *service.AutoMatchResult
	movement        repository.CreateTreasuryMovementRow
	recon           repository.ErpBankReconciliation
	check           repository.ErpCheck
	cashCount       repository.ErpCashCount
	bankAccount     repository.ErpBankAccount
	cashRegister    repository.ErpCashRegister
	err             error
}

func (m *mockTreasuryService) ListBankAccounts(_ context.Context, _ string, _ bool) ([]repository.ErpBankAccount, error) {
	return m.bankAccounts, m.err
}
func (m *mockTreasuryService) CreateBankAccount(_ context.Context, _ repository.CreateBankAccountParams, _, _ string) (repository.ErpBankAccount, error) {
	if m.err != nil {
		return repository.ErpBankAccount{}, m.err
	}
	return m.bankAccount, nil
}
func (m *mockTreasuryService) ListCashRegisters(_ context.Context, _ string) ([]repository.ErpCashRegister, error) {
	return m.cashRegisters, m.err
}
func (m *mockTreasuryService) CreateCashRegister(_ context.Context, _, _ string, _ pgtype.UUID, _, _ string) (repository.ErpCashRegister, error) {
	if m.err != nil {
		return repository.ErpCashRegister{}, m.err
	}
	return m.cashRegister, nil
}
func (m *mockTreasuryService) ListMovements(_ context.Context, _ string, _, _ pgtype.Date, _ string, _, _ int) ([]repository.ListTreasuryMovementsRow, error) {
	return m.movements, m.err
}
func (m *mockTreasuryService) CreateMovement(_ context.Context, _ service.CreateTreasuryMovementRequest) (repository.CreateTreasuryMovementRow, error) {
	if m.err != nil {
		return repository.CreateTreasuryMovementRow{}, m.err
	}
	return m.movement, nil
}
func (m *mockTreasuryService) ListChecks(_ context.Context, _, _, _ string) ([]repository.ErpCheck, error) {
	return m.checks, m.err
}
func (m *mockTreasuryService) CreateCheck(_ context.Context, _ repository.CreateCheckParams, _, _ string) (repository.ErpCheck, error) {
	if m.err != nil {
		return repository.ErpCheck{}, m.err
	}
	return m.check, nil
}
func (m *mockTreasuryService) UpdateCheckStatus(_ context.Context, _ pgtype.UUID, _, _, _, _ string) error {
	return m.err
}
func (m *mockTreasuryService) GetBalance(_ context.Context, _ string) ([]repository.GetTreasuryBalanceRow, error) {
	return m.balances, m.err
}
func (m *mockTreasuryService) ListCashCounts(_ context.Context, _ string, _, _ int) ([]repository.ErpCashCount, error) {
	return m.cashCounts, m.err
}
func (m *mockTreasuryService) CreateCashCount(_ context.Context, _ repository.CreateCashCountParams, _ string) (repository.ErpCashCount, error) {
	if m.err != nil {
		return repository.ErpCashCount{}, m.err
	}
	return m.cashCount, nil
}
func (m *mockTreasuryService) ListReceipts(_ context.Context, _, _ string, _, _ pgtype.Date, _, _ int) ([]repository.ListReceiptsRow, error) {
	return m.receipts, m.err
}
func (m *mockTreasuryService) GetReceipt(_ context.Context, _ string, _ pgtype.UUID) (*service.ReceiptDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.receiptDetail, nil
}
func (m *mockTreasuryService) CreateReceipt(_ context.Context, _ string, _ service.ReceiptInput, _, _ string) (*service.ReceiptDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.receiptDetail, nil
}
func (m *mockTreasuryService) VoidReceipt(_ context.Context, _ string, _ pgtype.UUID, _, _ string) error {
	return m.err
}
func (m *mockTreasuryService) ListReconciliations(_ context.Context, _ string) ([]repository.ListReconciliationsRow, error) {
	return m.reconciliations, m.err
}
func (m *mockTreasuryService) GetReconciliation(_ context.Context, _ string, _ pgtype.UUID) (*service.ReconciliationDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reconDetail, nil
}
func (m *mockTreasuryService) CreateReconciliation(_ context.Context, _ string, _ pgtype.UUID, _, _, _, _, _ string) (repository.ErpBankReconciliation, error) {
	if m.err != nil {
		return repository.ErpBankReconciliation{}, m.err
	}
	return m.recon, nil
}
func (m *mockTreasuryService) ImportStatementLines(_ context.Context, _ string, _ pgtype.UUID, _ []service.StatementLineInput) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 0, nil
}
func (m *mockTreasuryService) AutoMatch(_ context.Context, _ string, _ pgtype.UUID) (*service.AutoMatchResult, error) {
	return m.autoMatchResult, m.err
}
func (m *mockTreasuryService) MatchManual(_ context.Context, _ string, _, _, _ pgtype.UUID) error {
	return m.err
}
func (m *mockTreasuryService) ConfirmReconciliation(_ context.Context, _ string, _ pgtype.UUID, _, _ string) error {
	return m.err
}

// --- helpers ---

func setupTreasuryRouter(mock *mockTreasuryService) *chi.Mux {
	h := NewTreasury(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/treasury", h.Routes(noopMiddleware))
	return r
}

func decodeTreasuryJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests: reconciliations ---

func TestTreasury_ListReconciliations_Success(t *testing.T) {
	mock := &mockTreasuryService{
		reconciliations: []repository.ListReconciliationsRow{{Period: "2024-01"}, {Period: "2024-02"}},
	}
	r := setupTreasuryRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/treasury/reconciliations", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeTreasuryJSON(t, rec, &resp)
	recons, ok := resp["reconciliations"].([]any)
	if !ok || len(recons) != 2 {
		t.Errorf("expected 2 reconciliations, got %v", resp["reconciliations"])
	}
}

func TestTreasury_ListReconciliations_ServiceError_Returns500(t *testing.T) {
	mock := &mockTreasuryService{err: errors.New("db error")}
	r := setupTreasuryRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/treasury/reconciliations", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	var resp map[string]string
	decodeTreasuryJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q", resp["error"])
	}
}

func TestTreasury_CreateReconciliation_InvalidBankAccountID_Returns400(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	body := `{"bank_account_id":"not-a-uuid","period":"2024-01","statement_balance":"10000","book_balance":"9800"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid bank_account_id, got %d", rec.Code)
	}
}

func TestTreasury_CreateReconciliation_InvalidBody_Returns400(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestTreasury_CreateReconciliation_Success(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	body := `{"bank_account_id":"00000000-0000-0000-0000-000000000001","period":"2024-01","statement_balance":"10000","book_balance":"9800"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTreasury_AutoMatch_InvalidID_Returns400(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations/bad-uuid/auto-match", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTreasury_AutoMatch_Success(t *testing.T) {
	mock := &mockTreasuryService{
		autoMatchResult: &service.AutoMatchResult{Matched: 5, Unmatched: 2},
	}
	r := setupTreasuryRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations/00000000-0000-0000-0000-000000000001/auto-match", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTreasury_AutoMatch_ServiceError_Returns500(t *testing.T) {
	mock := &mockTreasuryService{err: errors.New("db error")}
	r := setupTreasuryRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations/00000000-0000-0000-0000-000000000001/auto-match", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestTreasury_ConfirmReconciliation_InvalidID_Returns400(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations/bad-uuid/confirm", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTreasury_ConfirmReconciliation_Success(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations/00000000-0000-0000-0000-000000000001/confirm", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTreasury_ConfirmReconciliation_NotConfirmedError_Returns400(t *testing.T) {
	mock := &mockTreasuryService{
		err: errors.New("not confirmed: unmatched lines remain"),
	}
	r := setupTreasuryRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/reconciliations/00000000-0000-0000-0000-000000000001/confirm", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for business error, got %d", rec.Code)
	}
}

// --- tests: receipts ---

func TestTreasury_CreateReceiptH_InvalidBody_Returns400(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/receipts", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestTreasury_CreateReceiptH_Success(t *testing.T) {
	mock := &mockTreasuryService{
		receiptDetail: &service.ReceiptDetail{},
	}
	r := setupTreasuryRouter(mock)

	body := `{"type":"customer","entity_id":"00000000-0000-0000-0000-000000000001","date":"2024-01-15","amount":"5000","movements":[]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/receipts", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTreasury_CreateReceiptH_BalanceError_Returns400(t *testing.T) {
	// Payments don't balance triggers writeSafeErr
	mock := &mockTreasuryService{
		err: errors.New("payments don't balance: expected 5000, got 4500"),
	}
	r := setupTreasuryRouter(mock)

	body := `{"type":"customer","entity_id":"00000000-0000-0000-0000-000000000001","date":"2024-01-15","amount":"5000","movements":[]}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/receipts", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for balance error, got %d", rec.Code)
	}
}

func TestTreasury_VoidReceiptH_InvalidID_Returns400(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/receipts/bad-uuid/void", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTreasury_VoidReceiptH_Success(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/receipts/00000000-0000-0000-0000-000000000001/void", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTreasury_VoidReceiptH_AlreadyVoidedError_Returns400(t *testing.T) {
	mock := &mockTreasuryService{
		err: errors.New("receipt already voided"),
	}
	r := setupTreasuryRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/receipts/00000000-0000-0000-0000-000000000001/void", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for business error, got %d", rec.Code)
	}
}

// --- tests: movements ---

func TestTreasury_ListMovements_Success(t *testing.T) {
	mock := &mockTreasuryService{
		movements: []repository.ListTreasuryMovementsRow{{Number: "MOV-001"}},
	}
	r := setupTreasuryRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/treasury/movements", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTreasury_CreateMovement_InvalidBody_Returns400(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/treasury/movements", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTreasury_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupTreasuryRouter(&mockTreasuryService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/treasury/reconciliations", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
