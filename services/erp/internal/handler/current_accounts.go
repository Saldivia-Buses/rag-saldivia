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

// CurrentAccountsService is the interface the CurrentAccounts handler depends on.
type CurrentAccountsService interface {
	ListMovements(ctx context.Context, tenantID string, entityID pgtype.UUID, direction string, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ListAccountMovementsRow, error)
	GetBalances(ctx context.Context, tenantID, direction string) ([]repository.GetEntityBalancesRow, error)
	GetOverdue(ctx context.Context, tenantID string) ([]repository.GetOverdueInvoicesRow, error)
	Allocate(ctx context.Context, req service.AllocateRequest) error
}

type CurrentAccounts struct{ svc CurrentAccountsService }

func NewCurrentAccounts(svc CurrentAccountsService) *CurrentAccounts {
	return &CurrentAccounts{svc: svc}
}

func (h *CurrentAccounts) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.accounts.read"))
		r.Get("/statement", h.ListMovements)
		r.Get("/balances", h.GetBalances)
		r.Get("/overdue", h.GetOverdue)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.accounts.write"))
		r.Post("/allocate", h.Allocate)
	})
	return r
}

func (h *CurrentAccounts) ListMovements(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	entityID := optUUID(ptrStr(q.Get("entity_id")))
	movements, err := h.svc.ListMovements(r.Context(), slug, entityID, q.Get("direction"),
		pgDate(q.Get("date_from")), pgDate(q.Get("date_to")), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"movements": movements})
}

func (h *CurrentAccounts) GetBalances(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	direction := r.URL.Query().Get("direction")
	balances, err := h.svc.GetBalances(r.Context(), slug, direction)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"balances": balances})
}

func (h *CurrentAccounts) GetOverdue(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	overdue, err := h.svc.GetOverdue(r.Context(), slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"overdue": overdue})
}

func (h *CurrentAccounts) Allocate(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		PaymentID string `json:"payment_id"`
		InvoiceID string `json:"invoice_id"`
		Amount    string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	payID, err := parseUUID(body.PaymentID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid payment_id"))
		return
	}
	invID, err := parseUUID(body.InvoiceID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid invoice_id"))
		return
	}
	if err := h.svc.Allocate(r.Context(), service.AllocateRequest{
		TenantID: slug, PaymentID: payID, InvoiceID: invID, Amount: body.Amount,
		UserID: r.Header.Get("X-User-ID"), IP: r.RemoteAddr,
	}); err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
