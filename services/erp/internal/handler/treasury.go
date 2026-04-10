package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

type Treasury struct{ svc *service.Treasury }

func NewTreasury(svc *service.Treasury) *Treasury { return &Treasury{svc: svc} }

func (h *Treasury) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.treasury.read"))
		r.Get("/bank-accounts", h.ListBankAccounts)
		r.Get("/cash-registers", h.ListCashRegisters)
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
	return r
}

func (h *Treasury) ListBankAccounts(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	accounts, err := h.svc.ListBankAccounts(r.Context(), slug, r.URL.Query().Get("active") != "false")
	if err != nil {
		slog.Error("list bank accounts failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"bank_accounts": accounts})
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
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	ba, err := h.svc.CreateBankAccount(r.Context(), repository.CreateBankAccountParams{
		TenantID: slug, BankName: body.BankName, Branch: body.Branch,
		AccountNumber: body.AccountNumber,
		Cbu:         pgTextOpt(body.CBU),
		Alias:       pgTextOpt(body.Alias),
		CurrencyID:  optUUID(body.CurrencyID),
		AccountID:   optUUID(body.AccountID),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create bank account failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ba)
}

func (h *Treasury) ListCashRegisters(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	regs, err := h.svc.ListCashRegisters(r.Context(), slug)
	if err != nil {
		slog.Error("list cash registers failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"cash_registers": regs})
}

func (h *Treasury) CreateCashRegister(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Name      string  `json:"name"`
		AccountID *string `json:"account_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	cr, err := h.svc.CreateCashRegister(r.Context(), slug, body.Name, optUUID(body.AccountID),
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create cash register failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cr)
}

func (h *Treasury) ListMovements(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	dateFrom := pgDate(r.URL.Query().Get("date_from"))
	dateTo := pgDate(r.URL.Query().Get("date_to"))
	typeFilter := r.URL.Query().Get("type")
	movements, err := h.svc.ListMovements(r.Context(), slug, dateFrom, dateTo, typeFilter, p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list treasury movements failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"movements": movements})
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
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
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
		slog.Error("create treasury movement failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mov)
}

func (h *Treasury) ListChecks(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	direction := r.URL.Query().Get("direction")
	status := r.URL.Query().Get("status")
	checks, err := h.svc.ListChecks(r.Context(), slug, direction, status)
	if err != nil {
		slog.Error("list checks failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"checks": checks})
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
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
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
		slog.Error("create check failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chk)
}

func (h *Treasury) UpdateCheckStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateCheckStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Treasury) GetBalance(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	balances, err := h.svc.GetBalance(r.Context(), slug)
	if err != nil {
		slog.Error("get treasury balance failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"balances": balances})
}

func (h *Treasury) ListCashCounts(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	counts, err := h.svc.ListCashCounts(r.Context(), slug, p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list cash counts failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"cash_counts": counts})
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
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	crID, err := parseUUID(body.CashRegisterID)
	if err != nil {
		http.Error(w, `{"error":"invalid cash_register_id"}`, http.StatusBadRequest)
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
		slog.Error("create cash count failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cc)
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
	var f float64
	for i, c := range s {
		if c == '.' || c == '-' || (c >= '0' && c <= '9') {
			continue
		}
		s = s[:i]
		break
	}
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func pgTextOpt(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}
