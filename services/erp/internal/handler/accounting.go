package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// Accounting handles accounting endpoints.
type Accounting struct{ svc *service.Accounting }

func NewAccounting(svc *service.Accounting) *Accounting { return &Accounting{svc: svc} }

func (h *Accounting) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.accounting.read"))
		r.Get("/accounts", h.ListAccounts)
		r.Get("/cost-centers", h.ListCostCenters)
		r.Get("/fiscal-years", h.ListFiscalYears)
		r.Get("/entries", h.ListEntries)
		r.Get("/entries/{id}", h.GetEntry)
		r.Get("/balance", h.GetBalance)
		r.Get("/ledger", h.GetLedger)
	})

	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.accounting.write"))
		r.Post("/accounts", h.CreateAccount)
		r.Post("/cost-centers", h.CreateCostCenter)
		r.Post("/fiscal-years", h.CreateFiscalYear)
		r.Post("/entries", h.CreateEntry)
	})

	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.accounting.reverse"))
		r.Post("/entries/{id}/post", h.PostEntry)
	})

	return r
}

func (h *Accounting) ListAccounts(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	activeOnly := r.URL.Query().Get("active") != "false"
	accounts, err := h.svc.ListAccounts(r.Context(), slug, activeOnly)
	if err != nil {
		slog.Error("list accounts failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"accounts": accounts})
}

func (h *Accounting) CreateAccount(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		Code         string  `json:"code"`
		Name         string  `json:"name"`
		ParentID     *string `json:"parent_id,omitempty"`
		AccountType  string  `json:"account_type"`
		IsDetail     bool    `json:"is_detail"`
		CostCenterID *string `json:"cost_center_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	acct, err := h.svc.CreateAccount(r.Context(), slug, body.Code, body.Name,
		optUUID(body.ParentID), body.AccountType, body.IsDetail, optUUID(body.CostCenterID),
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create account failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acct)
}

func (h *Accounting) ListCostCenters(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	activeOnly := r.URL.Query().Get("active") != "false"
	centers, err := h.svc.ListCostCenters(r.Context(), slug, activeOnly)
	if err != nil {
		slog.Error("list cost centers failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"cost_centers": centers})
}

func (h *Accounting) CreateCostCenter(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	cc, err := h.svc.CreateCostCenter(r.Context(), slug, body.Code, body.Name,
		optUUID(body.ParentID), r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create cost center failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cc)
}

func (h *Accounting) ListFiscalYears(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	years, err := h.svc.ListFiscalYears(r.Context(), slug)
	if err != nil {
		slog.Error("list fiscal years failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"fiscal_years": years})
}

func (h *Accounting) CreateFiscalYear(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		Year      int    `json:"year"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	fy, err := h.svc.CreateFiscalYear(r.Context(), slug, body.Year, body.StartDate, body.EndDate)
	if err != nil {
		slog.Error("create fiscal year failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fy)
}

func (h *Accounting) ListEntries(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	status := r.URL.Query().Get("status")
	dateFrom := pgDate(r.URL.Query().Get("date_from"))
	dateTo := pgDate(r.URL.Query().Get("date_to"))

	entries, err := h.svc.ListEntries(r.Context(), slug, dateFrom, dateTo, status, p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list entries failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"entries": entries})
}

func (h *Accounting) GetEntry(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	detail, err := h.svc.GetEntry(r.Context(), id, slug)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (h *Accounting) CreateEntry(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	var body struct {
		Number       string `json:"number"`
		Date         string `json:"date"`
		FiscalYearID *string `json:"fiscal_year_id,omitempty"`
		Concept      string `json:"concept"`
		EntryType    string `json:"entry_type"`
		Lines        []struct {
			AccountID    string `json:"account_id"`
			CostCenterID *string `json:"cost_center_id,omitempty"`
			Debit        string `json:"debit"`
			Credit       string `json:"credit"`
			Description  string `json:"description"`
		} `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	var lines []service.CreateLineRequest
	for _, l := range body.Lines {
		acctID, err := parseUUID(l.AccountID)
		if err != nil {
			http.Error(w, `{"error":"invalid account_id in line"}`, http.StatusBadRequest)
			return
		}
		lines = append(lines, service.CreateLineRequest{
			AccountID:    acctID,
			CostCenterID: optUUID(l.CostCenterID),
			Debit:        l.Debit,
			Credit:       l.Credit,
			Description:  l.Description,
		})
	}

	detail, err := h.svc.CreateEntry(r.Context(), service.CreateEntryRequest{
		TenantID:     slug,
		Number:       body.Number,
		Date:         pgDate(body.Date),
		FiscalYearID: optUUID(body.FiscalYearID),
		Concept:      body.Concept,
		EntryType:    body.EntryType,
		UserID:       r.Header.Get("X-User-ID"),
		IP:           r.RemoteAddr,
		Lines:        lines,
	})
	if err != nil {
		slog.Error("create entry failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(detail)
}

func (h *Accounting) PostEntry(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.PostEntry(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		slog.Error("post entry failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Accounting) GetBalance(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	dateFrom := pgDate(r.URL.Query().Get("date_from"))
	dateTo := pgDate(r.URL.Query().Get("date_to"))

	balances, err := h.svc.GetBalance(r.Context(), slug, dateFrom, dateTo)
	if err != nil {
		slog.Error("get balance failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"balances": balances})
}

func (h *Accounting) GetLedger(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	accountID := optUUID(ptrStr(r.URL.Query().Get("account_id")))
	dateFrom := pgDate(r.URL.Query().Get("date_from"))
	dateTo := pgDate(r.URL.Query().Get("date_to"))

	ledger, err := h.svc.GetLedger(r.Context(), slug, accountID, dateFrom, dateTo, p.Limit(), p.Offset())
	if err != nil {
		slog.Error("get ledger failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ledger": ledger})
}

// pgDate parses a date string into pgtype.Date. Returns invalid if empty.
func pgDate(s string) pgtype.Date {
	if s == "" {
		return pgtype.Date{}
	}
	var d pgtype.Date
	_ = d.Scan(s)
	return d
}
