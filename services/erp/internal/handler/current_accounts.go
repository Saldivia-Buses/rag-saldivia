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
	ListComplaints(ctx context.Context, tenantID string, statusFilter int16, entityID pgtype.UUID, limit, offset int) ([]repository.ErpPaymentComplaint, error)
	CreateComplaint(ctx context.Context, req service.CreateComplaintRequest) (repository.ErpPaymentComplaint, error)
	UpdateComplaintStatus(ctx context.Context, req service.UpdateComplaintStatusRequest) error
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
		r.Get("/complaints", h.ListComplaints)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.accounts.write"))
		r.Post("/allocate", h.Allocate)
		r.Post("/complaints", h.CreateComplaint)
		r.Patch("/complaints/{id}/status", h.UpdateComplaintStatus)
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

// ─── Payment complaints ───

// ListComplaints GET /complaints?status=0|1|-1&entity_id=<uuid>&limit=&offset=
func (h *CurrentAccounts) ListComplaints(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	status := int16(-1)
	if raw := q.Get("status"); raw != "" {
		switch raw {
		case "0":
			status = 0
		case "1":
			status = 1
		}
	}
	entityID := optUUID(ptrStr(q.Get("entity_id")))
	complaints, err := h.svc.ListComplaints(r.Context(), slug, status, entityID, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"complaints": complaints})
}

// CreateComplaint POST /complaints {entity_id, entity_legacy_code, observation, date?}
func (h *CurrentAccounts) CreateComplaint(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var body struct {
		EntityID         string `json:"entity_id"`
		EntityLegacyCode int32  `json:"entity_legacy_code"`
		Observation      string `json:"observation"`
		Date             string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	entityID := optUUID(&body.EntityID)
	complaint, err := h.svc.CreateComplaint(r.Context(), service.CreateComplaintRequest{
		TenantID:         slug,
		ComplaintDate:    pgDate(body.Date),
		EntityID:         entityID,
		EntityLegacyCode: body.EntityLegacyCode,
		Observation:      body.Observation,
		Login:            r.Header.Get("X-User-Login"),
		UserID:           r.Header.Get("X-User-ID"),
		IP:               r.RemoteAddr,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"complaint": complaint})
}

// UpdateComplaintStatus PATCH /complaints/{id}/status {status: 0|1}
func (h *CurrentAccounts) UpdateComplaintStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	idStr := chi.URLParam(r, "id")
	id, err := parseUUID(idStr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid id"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
	var body struct {
		Status int16 `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Status != 0 && body.Status != 1 {
		erperrors.WriteError(w, r, erperrors.InvalidInput("status must be 0 or 1"))
		return
	}
	if err := h.svc.UpdateComplaintStatus(r.Context(), service.UpdateComplaintStatusRequest{
		TenantID: slug, ID: id, Status: body.Status,
		UserID: r.Header.Get("X-User-ID"), IP: r.RemoteAddr,
	}); err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
