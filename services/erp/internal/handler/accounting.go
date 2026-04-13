package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// AccountingService is the interface the Accounting handler depends on.
type AccountingService interface {
	ListAccounts(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpAccount, error)
	CreateAccount(ctx context.Context, tenantID, code, name string, parentID pgtype.UUID, accountType string, isDetail bool, costCenterID pgtype.UUID, userID, ip string) (repository.ErpAccount, error)
	ListCostCenters(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpCostCenter, error)
	CreateCostCenter(ctx context.Context, tenantID, code, name string, parentID pgtype.UUID, userID, ip string) (repository.ErpCostCenter, error)
	ListFiscalYears(ctx context.Context, tenantID string) ([]repository.ListFiscalYearsRow, error)
	CreateFiscalYear(ctx context.Context, tenantID string, year int, startDate, endDate, userID, ip string) (repository.CreateFiscalYearRow, error)
	SetFiscalYearResultAccount(ctx context.Context, tenantID string, yearID, accountID pgtype.UUID, userID, ip string) error
	PreviewClose(ctx context.Context, tenantID string, yearID pgtype.UUID) (*service.PreviewCloseResult, error)
	CloseFiscalYear(ctx context.Context, tenantID string, yearID pgtype.UUID, userID, ip string) (*service.CloseResult, error)
	ListEntries(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Date, status string, limit, offset int) ([]repository.ListJournalEntriesRow, error)
	GetEntry(ctx context.Context, id pgtype.UUID, tenantID string) (*service.EntryDetail, error)
	CreateEntry(ctx context.Context, req service.CreateEntryRequest) (*service.EntryDetail, error)
	PostEntry(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error
	GetBalance(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Date) ([]repository.GetAccountBalanceRow, error)
	GetLedger(ctx context.Context, tenantID string, accountID pgtype.UUID, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.GetLedgerRow, error)
}

// Accounting handles accounting endpoints.
type Accounting struct{ svc AccountingService }

func NewAccounting(svc AccountingService) *Accounting { return &Accounting{svc: svc} }

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
		r.Use(sdamw.RequirePermission("erp.accounting.post"))
		r.Post("/entries/{id}/post", h.PostEntry)
	})

	// Fiscal year close
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.accounting.read"))
		r.Get("/fiscal-years/{id}/preview-close", h.PreviewClose)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.accounting.close"))
		r.Patch("/fiscal-years/{id}/result-account", h.SetResultAccount)
		r.Post("/fiscal-years/{id}/close", h.CloseFiscalYear)
	})

	return r
}

func (h *Accounting) ListAccounts(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	activeOnly := r.URL.Query().Get("active") != "false"
	accounts, err := h.svc.ListAccounts(r.Context(), slug, activeOnly)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}

	acct, err := h.svc.CreateAccount(r.Context(), slug, body.Code, body.Name,
		optUUID(body.ParentID), body.AccountType, body.IsDetail, optUUID(body.CostCenterID),
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	cc, err := h.svc.CreateCostCenter(r.Context(), slug, body.Code, body.Name,
		optUUID(body.ParentID), r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	fy, err := h.svc.CreateFiscalYear(r.Context(), slug, body.Year, body.StartDate, body.EndDate, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"entries": entries})
}

func (h *Accounting) GetEntry(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	detail, err := h.svc.GetEntry(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("entry"))
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}

	var lines []service.CreateLineRequest
	for _, l := range body.Lines {
		acctID, err := parseUUID(l.AccountID)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidInput("invalid account_id in line"))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	if err := h.svc.PostEntry(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ledger": ledger})
}

// SetResultAccount sets the result account on a fiscal year (required before close).
func (h *Accounting) SetResultAccount(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	yearID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid fiscal year id"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		ResultAccountID string `json:"result_account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ResultAccountID == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("result_account_id required"))
		return
	}
	accountID, err := parseUUID(body.ResultAccountID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid account id"))
		return
	}
	if err := h.svc.SetFiscalYearResultAccount(r.Context(), slug, yearID, accountID,
		r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PreviewClose returns what the fiscal year close would do.
func (h *Accounting) PreviewClose(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	yearID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid fiscal year id"))
		return
	}
	preview, err := h.svc.PreviewClose(r.Context(), slug, yearID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// CloseFiscalYear closes a fiscal year and creates the next one.
func (h *Accounting) CloseFiscalYear(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	yearID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid fiscal year id"))
		return
	}
	result, err := h.svc.CloseFiscalYear(r.Context(), slug, yearID,
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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
