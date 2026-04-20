package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// TreasuryService is the interface the Treasury handler depends on.
type TreasuryService interface {
	ListBankAccounts(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpBankAccount, error)
	GetBankAccount(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpBankAccount, error)
	CreateBankAccount(ctx context.Context, p repository.CreateBankAccountParams, userID, ip string) (repository.ErpBankAccount, error)
	ListCashRegisters(ctx context.Context, tenantID string) ([]repository.ErpCashRegister, error)
	GetCashRegister(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpCashRegister, error)
	CreateCashRegister(ctx context.Context, tenantID, name string, accountID pgtype.UUID, userID, ip string) (repository.ErpCashRegister, error)
	ListMovements(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Date, typeFilter string, limit, offset int) ([]repository.ListTreasuryMovementsRow, error)
	CreateMovement(ctx context.Context, req service.CreateTreasuryMovementRequest) (repository.CreateTreasuryMovementRow, error)
	ListChecks(ctx context.Context, tenantID, direction, status string) ([]repository.ErpCheck, error)
	CreateCheck(ctx context.Context, p repository.CreateCheckParams, userID, ip string) (repository.ErpCheck, error)
	UpdateCheckStatus(ctx context.Context, id pgtype.UUID, tenantID, newStatus, userID, ip string) error
	GetBalance(ctx context.Context, tenantID string) ([]repository.GetTreasuryBalanceRow, error)
	ListCashCounts(ctx context.Context, tenantID string, cashRegisterID pgtype.UUID, limit, offset int) ([]repository.ErpCashCount, error)
	CreateCashCount(ctx context.Context, p repository.CreateCashCountParams, ip string) (repository.ErpCashCount, error)
	ListReceipts(ctx context.Context, tenantID, typeFilter string, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ListReceiptsRow, error)
	GetReceipt(ctx context.Context, tenantID string, id pgtype.UUID) (*service.ReceiptDetail, error)
	CreateReceipt(ctx context.Context, tenantID string, inp service.ReceiptInput, userID, ip string) (*service.ReceiptDetail, error)
	VoidReceipt(ctx context.Context, tenantID string, receiptID pgtype.UUID, userID, ip string) error
	ListReconciliations(ctx context.Context, tenantID string, bankAccountID pgtype.UUID) ([]repository.ListReconciliationsRow, error)
	GetReconciliation(ctx context.Context, tenantID string, id pgtype.UUID) (*service.ReconciliationDetail, error)
	CreateReconciliation(ctx context.Context, tenantID string, bankAccountID pgtype.UUID, period string, statementBalance, bookBalance string, userID, ip string) (repository.ErpBankReconciliation, error)
	ImportStatementLines(ctx context.Context, tenantID string, reconID pgtype.UUID, lines []service.StatementLineInput) (int, error)
	AutoMatch(ctx context.Context, tenantID string, reconID pgtype.UUID) (*service.AutoMatchResult, error)
	MatchManual(ctx context.Context, tenantID string, reconID, lineID, movementID pgtype.UUID) error
	ConfirmReconciliation(ctx context.Context, tenantID string, reconID pgtype.UUID, userID, ip string) error
	ListBankImports(ctx context.Context, tenantID string, accountFilter, processedFilter int32, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ErpBankImport, error)
	UpdateBankImportProcessed(ctx context.Context, req service.UpdateBankImportRequest) error
	ListCheckHistory(ctx context.Context, tenantID string, entityFilter pgtype.UUID, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ErpCheckHistory, error)
}

type Treasury struct{ svc TreasuryService }

func NewTreasury(svc TreasuryService) *Treasury { return &Treasury{svc: svc} }

func (h *Treasury) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.treasury.read"))
		r.Get("/bank-accounts", h.ListBankAccounts)
		r.Get("/bank-accounts/{id}", h.GetBankAccount)
		r.Get("/cash-registers", h.ListCashRegisters)
		r.Get("/cash-registers/{id}", h.GetCashRegister)
		r.Get("/movements", h.ListMovements)
		r.Get("/checks", h.ListChecks)
		r.Get("/balance", h.GetBalance)
		r.Get("/cash-counts", h.ListCashCounts)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.treasury.write"))
		r.Post("/bank-accounts", h.CreateBankAccount)
		r.Post("/cash-registers", h.CreateCashRegister)
		r.Post("/movements", h.CreateMovement)
		r.Post("/checks", h.CreateCheck)
		r.Patch("/checks/{id}/status", h.UpdateCheckStatus)
		r.Post("/cash-counts", h.CreateCashCount)
	})

	// Receipts
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.treasury.read"))
		r.Get("/receipts", h.ListReceiptsH)
		r.Get("/receipts/{id}", h.GetReceiptH)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.treasury.receipt"))
		r.Post("/receipts", h.CreateReceiptH)
		r.Post("/receipts/{id}/void", h.VoidReceiptH)
	})

	// Reconciliation
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.treasury.read"))
		r.Get("/reconciliations", h.ListReconciliations)
		r.Get("/reconciliations/{id}", h.GetReconciliation)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.treasury.reconcile"))
		r.Post("/reconciliations", h.CreateReconciliation)
		r.Post("/reconciliations/{id}/import", h.ImportStatementLines)
		r.Post("/reconciliations/{id}/auto-match", h.AutoMatch)
		r.Patch("/reconciliations/{id}/lines/{lineId}/match", h.MatchManual)
		r.Post("/reconciliations/{id}/confirm", h.ConfirmReconciliation)
	})

	// Bank imports (BCS_IMPORTACION parity)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.treasury.read"))
		r.Get("/imports", h.ListBankImportsH)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.treasury.write"))
		r.Patch("/imports/{id}", h.UpdateBankImportProcessedH)
	})

	// Check history (CARCHEHI parity)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.treasury.read"))
		r.Get("/check-history", h.ListCheckHistoryH)
	})

	return r
}

func (h *Treasury) ListBankAccounts(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	accounts, err := h.svc.ListBankAccounts(r.Context(), slug, r.URL.Query().Get("active") != "false")
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"bank_accounts": accounts})
}

func (h *Treasury) GetBankAccount(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	ba, err := h.svc.GetBankAccount(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("bank account"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ba)
}

func (h *Treasury) CreateBankAccount(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		BankName      string  `json:"bank_name"`
		Branch        string  `json:"branch"`
		AccountNumber string  `json:"account_number"`
		CBU           *string `json:"cbu,omitempty"`
		Alias         *string `json:"alias,omitempty"`
		CurrencyID    *string `json:"currency_id,omitempty"`
		AccountID     *string `json:"account_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	ba, err := h.svc.CreateBankAccount(r.Context(), repository.CreateBankAccountParams{
		TenantID: slug, BankName: body.BankName, Branch: body.Branch,
		AccountNumber: body.AccountNumber,
		Cbu:           pgTextOpt(body.CBU),
		Alias:         pgTextOpt(body.Alias),
		CurrencyID:    optUUID(body.CurrencyID),
		AccountID:     optUUID(body.AccountID),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ba)
}

func (h *Treasury) ListCashRegisters(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	regs, err := h.svc.ListCashRegisters(r.Context(), slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"cash_registers": regs})
}

func (h *Treasury) GetCashRegister(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	cr, err := h.svc.GetCashRegister(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("cash register"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cr)
}

func (h *Treasury) CreateCashRegister(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Name      string  `json:"name"`
		AccountID *string `json:"account_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	cr, err := h.svc.CreateCashRegister(r.Context(), slug, body.Name, optUUID(body.AccountID),
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cr)
}

func (h *Treasury) ListMovements(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	dateFrom := pgDate(r.URL.Query().Get("date_from"))
	dateTo := pgDate(r.URL.Query().Get("date_to"))
	typeFilter := r.URL.Query().Get("type")
	movements, err := h.svc.ListMovements(r.Context(), slug, dateFrom, dateTo, typeFilter, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"movements": movements})
}

func (h *Treasury) CreateMovement(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Date           string  `json:"date"`
		Number         string  `json:"number"`
		MovementType   string  `json:"movement_type"`
		Amount         string  `json:"amount"`
		CurrencyID     *string `json:"currency_id,omitempty"`
		BankAccountID  *string `json:"bank_account_id,omitempty"`
		CashRegisterID *string `json:"cash_register_id,omitempty"`
		EntityID       *string `json:"entity_id,omitempty"`
		ConceptID      *string `json:"concept_id,omitempty"`
		PaymentMethod  *string `json:"payment_method,omitempty"`
		Notes          string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	var pm pgtype.Text
	if body.PaymentMethod != nil {
		pm = pgtype.Text{String: *body.PaymentMethod, Valid: true}
	}
	mov, err := h.svc.CreateMovement(r.Context(), service.CreateTreasuryMovementRequest{
		CreateTreasuryMovementParams: repository.CreateTreasuryMovementParams{
			TenantID:       slug,
			Date:           pgDate(body.Date),
			Number:         body.Number,
			MovementType:   body.MovementType,
			Amount:         pgNumericH(body.Amount),
			CurrencyID:     optUUID(body.CurrencyID),
			BankAccountID:  optUUID(body.BankAccountID),
			CashRegisterID: optUUID(body.CashRegisterID),
			EntityID:       optUUID(body.EntityID),
			ConceptID:      optUUID(body.ConceptID),
			PaymentMethod:  pm,
			UserID:         r.Header.Get("X-User-ID"),
			Notes:          body.Notes,
		},
		UserIDVal: r.Header.Get("X-User-ID"),
		IP:        r.RemoteAddr,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(mov)
}

func (h *Treasury) ListChecks(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	direction := r.URL.Query().Get("direction")
	status := r.URL.Query().Get("status")
	checks, err := h.svc.ListChecks(r.Context(), slug, direction, status)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"checks": checks})
}

func (h *Treasury) CreateCheck(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Direction  string  `json:"direction"`
		Number     string  `json:"number"`
		BankName   string  `json:"bank_name"`
		Amount     string  `json:"amount"`
		IssueDate  string  `json:"issue_date"`
		DueDate    string  `json:"due_date"`
		EntityID   *string `json:"entity_id,omitempty"`
		MovementID *string `json:"movement_id,omitempty"`
		Notes      string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	chk, err := h.svc.CreateCheck(r.Context(), repository.CreateCheckParams{
		TenantID:   slug,
		Direction:  body.Direction,
		Number:     body.Number,
		BankName:   body.BankName,
		Amount:     pgNumericH(body.Amount),
		IssueDate:  pgDate(body.IssueDate),
		DueDate:    pgDate(body.DueDate),
		EntityID:   optUUID(body.EntityID),
		MovementID: optUUID(body.MovementID),
		Notes:      body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(chk)
}

func (h *Treasury) UpdateCheckStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	if err := h.svc.UpdateCheckStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "invalid transition") {
			erperrors.WriteError(w, r, erperrors.InvalidInput(msg))
		} else if strings.Contains(msg, "not found") {
			erperrors.WriteError(w, r, erperrors.NotFound("check"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Treasury) GetBalance(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	balances, err := h.svc.GetBalance(r.Context(), slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"balances": balances})
}

func (h *Treasury) ListCashCounts(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	var cashRegisterID pgtype.UUID
	if s := r.URL.Query().Get("cash_register_id"); s != "" {
		id, err := parseUUID(s)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidID("cash_register_id"))
			return
		}
		cashRegisterID = id
	}
	counts, err := h.svc.ListCashCounts(r.Context(), slug, cashRegisterID, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"cash_counts": counts})
}

func (h *Treasury) CreateCashCount(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		CashRegisterID string `json:"cash_register_id"`
		Date           string `json:"date"`
		Expected       string `json:"expected"`
		Counted        string `json:"counted"`
		Notes          string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	crID, err := parseUUID(body.CashRegisterID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid cash_register_id"))
		return
	}
	// Calculate difference server-side (counted - expected)
	expected := pgNumericH(body.Expected)
	counted := pgNumericH(body.Counted)
	// Simple approach: parse strings to compute difference
	var eF, cF float64
	if body.Expected != "" {
		eF = parseFloat(body.Expected)
	}
	if body.Counted != "" {
		cF = parseFloat(body.Counted)
	}
	var diff pgtype.Numeric
	_ = diff.Scan(cF - eF)
	cc, err := h.svc.CreateCashCount(r.Context(), repository.CreateCashCountParams{
		TenantID:       slug,
		CashRegisterID: crID,
		Date:           pgDate(body.Date),
		Expected:       expected,
		Counted:        counted,
		Difference:     diff,
		UserID:         r.Header.Get("X-User-ID"),
		Notes:          body.Notes,
	}, r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cc)
}

// ============================================================
// Receipt handlers (Plan 18 Fase 4)
// ============================================================

func (h *Treasury) ListReceiptsH(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	receipts, err := h.svc.ListReceipts(r.Context(), slug, q.Get("type"),
		pgDate(q.Get("date_from")), pgDate(q.Get("date_to")), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"receipts": receipts})
}

func (h *Treasury) GetReceiptH(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	detail, err := h.svc.GetReceipt(r.Context(), slug, id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("receipt"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *Treasury) CreateReceiptH(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body service.ReceiptInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	detail, err := h.svc.CreateReceipt(r.Context(), slug, body,
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		if strings.Contains(err.Error(), "don't balance") {
			writeSafeErr(w, err, http.StatusBadRequest)
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *Treasury) VoidReceiptH(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	if err := h.svc.VoidReceipt(r.Context(), slug, id,
		r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		writeSafeErr(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// Reconciliation handlers (Plan 18 Fase 1)
// ============================================================

func (h *Treasury) ListReconciliations(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	var bankAccountID pgtype.UUID
	if s := r.URL.Query().Get("bank_account_id"); s != "" {
		id, err := parseUUID(s)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidID("bank_account_id"))
			return
		}
		bankAccountID = id
	}
	recons, err := h.svc.ListReconciliations(r.Context(), slug, bankAccountID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"reconciliations": recons})
}

func (h *Treasury) CreateReconciliation(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	var body struct {
		BankAccountID    string `json:"bank_account_id"`
		Period           string `json:"period"`
		StatementBalance string `json:"statement_balance"`
		BookBalance      string `json:"book_balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	baID, err := parseUUID(body.BankAccountID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid bank_account_id"))
		return
	}
	recon, err := h.svc.CreateReconciliation(r.Context(), slug, baID, body.Period,
		body.StatementBalance, body.BookBalance, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(recon)
}

func (h *Treasury) GetReconciliation(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	detail, err := h.svc.GetReconciliation(r.Context(), slug, id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *Treasury) ImportStatementLines(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	reconID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB for statement lines
	var body struct {
		Lines []service.StatementLineInput `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	count, err := h.svc.ImportStatementLines(r.Context(), slug, reconID, body.Lines)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"imported": count})
}

func (h *Treasury) AutoMatch(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	reconID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	result, err := h.svc.AutoMatch(r.Context(), slug, reconID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Treasury) MatchManual(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	reconID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid reconciliation id"))
		return
	}
	lineID, err := parseUUID(chi.URLParam(r, "lineId"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid line id"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		MovementID string `json:"movement_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MovementID == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("movement_id required"))
		return
	}
	movID, err := parseUUID(body.MovementID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid movement_id"))
		return
	}
	if err := h.svc.MatchManual(r.Context(), slug, reconID, lineID, movID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already matched") {
			writeSafeErr(w, err, http.StatusBadRequest)
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Treasury) ConfirmReconciliation(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	reconID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	if err := h.svc.ConfirmReconciliation(r.Context(), slug, reconID,
		r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		writeSafeErr(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// pgNumericH parses a numeric string in the handler layer.
func pgNumericH(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if s == "" {
		s = "0"
	}
	_ = n.Scan(s)
	return n
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ListBankImportsH lists bank-import staging rows.
// Query params: account (int32, 0=all), processed (int32, -1=all / 0=pending /
// 1=done / 2=cancelled), date_from, date_to, page, page_size.
func (h *Treasury) ListBankImportsH(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()

	account, _ := strconv.Atoi(q.Get("account"))
	processed := -1
	if s := q.Get("processed"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			processed = v
		}
	}

	imports, err := h.svc.ListBankImports(r.Context(), slug,
		int32(account), int32(processed),
		pgDate(q.Get("date_from")), pgDate(q.Get("date_to")),
		p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"imports": imports})
}

// ListCheckHistoryH lists historical cheque rows (CARCHEHI).
// Query params: entity_id, date_from, date_to, page, page_size.
func (h *Treasury) ListCheckHistoryH(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()

	entityFilter := optUUID(ptrStr(q.Get("entity_id")))

	history, err := h.svc.ListCheckHistory(r.Context(), slug,
		entityFilter,
		pgDate(q.Get("date_from")), pgDate(q.Get("date_to")),
		p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"history": history})
}

// UpdateBankImportProcessedH toggles the processed flag on a bank-import row.
// Body: {processed: 0|1|2, treasury_movement_id: "<uuid>" (optional)}.
func (h *Treasury) UpdateBankImportProcessedH(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 2<<10)
	var body struct {
		Processed          int32  `json:"processed"`
		TreasuryMovementID string `json:"treasury_movement_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Processed < 0 || body.Processed > 2 {
		erperrors.WriteError(w, r, erperrors.InvalidInput(
			"processed must be 0 (pending), 1 (processed) or 2 (cancelled)"))
		return
	}

	if err := h.svc.UpdateBankImportProcessed(r.Context(), service.UpdateBankImportRequest{
		ID:                 id,
		TenantID:           slug,
		Processed:          body.Processed,
		TreasuryMovementID: optUUID(&body.TreasuryMovementID),
		UserID:             r.Header.Get("X-User-ID"),
		IP:                 r.RemoteAddr,
	}); err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func pgTextOpt(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}
