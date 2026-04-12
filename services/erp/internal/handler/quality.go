package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// QualityService is the interface the Quality handler depends on.
type QualityService interface {
	ListNC(ctx context.Context, tenantID, status, severity string, limit, offset int) ([]repository.ListNonconformitiesRow, error)
	GetNC(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpNonconformity, error)
	CreateNC(ctx context.Context, p repository.CreateNonconformityParams, ip string) (repository.ErpNonconformity, error)
	UpdateNCStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error
	ListCA(ctx context.Context, ncID pgtype.UUID, tenantID string) ([]repository.ErpCorrectiveAction, error)
	CreateCA(ctx context.Context, p repository.CreateCorrectiveActionParams, userID, ip string) (repository.ErpCorrectiveAction, error)
	ListAudits(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpAudit, error)
	CreateAudit(ctx context.Context, p repository.CreateAuditParams, userID, ip string) (repository.ErpAudit, error)
	ListAuditFindings(ctx context.Context, auditID pgtype.UUID, tenantID string) ([]repository.ErpAuditFinding, error)
	CreateAuditFinding(ctx context.Context, p repository.CreateAuditFindingParams, userID, ip string) (repository.ErpAuditFinding, error)
	ListDocuments(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ErpControlledDocument, error)
	CreateDocument(ctx context.Context, p repository.CreateControlledDocumentParams, userID, ip string) (repository.ErpControlledDocument, error)
}

type Quality struct{ svc QualityService }

func NewQuality(svc QualityService) *Quality { return &Quality{svc: svc} }

func (h *Quality) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.quality.read"))
		r.Get("/nc", h.ListNC)
		r.Get("/nc/{id}", h.GetNC)
		r.Get("/nc/{id}/actions", h.ListCA)
		r.Get("/audits", h.ListAudits)
		r.Get("/audits/{id}/findings", h.ListFindings)
		r.Get("/documents", h.ListDocuments)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.quality.write"))
		r.Post("/nc", h.CreateNC)
		r.Patch("/nc/{id}/status", h.UpdateNCStatus)
		r.Post("/nc/{id}/actions", h.CreateCA)
		r.Post("/audits", h.CreateAudit)
		r.Post("/audits/{id}/findings", h.CreateFinding)
		r.Post("/documents", h.CreateDocument)
	})
	return r
}

func (h *Quality) ListNC(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	ncs, err := h.svc.ListNC(r.Context(), slug, q.Get("status"), q.Get("severity"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list NC failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"nonconformities": ncs})
}

func (h *Quality) GetNC(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	nc, err := h.svc.GetNC(r.Context(), id, tenantSlug(r))
	if err != nil { http.Error(w, `{"error":"not found"}`, http.StatusNotFound); return }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nc)
}

func (h *Quality) CreateNC(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Number, Description, Severity string
		TypeID, OriginID, AssignedTo *string
		Date string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return
	}
	if body.Number == "" || body.Description == "" {
		http.Error(w, `{"error":"number and description are required"}`, http.StatusBadRequest); return
	}
	validSev := map[string]bool{"minor": true, "major": true, "critical": true}
	if !validSev[body.Severity] {
		http.Error(w, `{"error":"invalid severity (minor, major, critical)"}`, http.StatusBadRequest); return
	}
	nc, err := h.svc.CreateNC(r.Context(), repository.CreateNonconformityParams{
		TenantID: slug, Number: body.Number, Date: pgDate(body.Date),
		TypeID: optUUID(body.TypeID), OriginID: optUUID(body.OriginID),
		Description: body.Description, Severity: body.Severity,
		AssignedTo: optUUID(body.AssignedTo), UserID: r.Header.Get("X-User-ID"),
	}, r.RemoteAddr)
	if err != nil {
		slog.Error("create NC failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nc)
}

func (h *Quality) UpdateNCStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	var body struct{ Status string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return
	}
	validNCStatus := map[string]bool{"open": true, "investigating": true, "corrective_action": true, "closed": true}
	if !validNCStatus[body.Status] {
		http.Error(w, `{"error":"invalid status (open, investigating, corrective_action, closed)"}`, http.StatusBadRequest); return
	}
	if err := h.svc.UpdateNCStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "NC not found" {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		} else {
			slog.Error("update NC status failed", "error", err)
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Quality) ListCA(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	cas, err := h.svc.ListCA(r.Context(), id, tenantSlug(r))
	if err != nil {
		slog.Error("list CA failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"corrective_actions": cas})
}

func (h *Quality) CreateCA(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	ncID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	var body struct {
		ActionType, Description string
		ResponsibleID *string
		DueDate *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return
	}
	validActionType := map[string]bool{"corrective": true, "preventive": true}
	if !validActionType[body.ActionType] || body.Description == "" {
		http.Error(w, `{"error":"valid action_type and description required"}`, http.StatusBadRequest); return
	}
	var dd string
	if body.DueDate != nil { dd = *body.DueDate }
	ca, err := h.svc.CreateCA(r.Context(), repository.CreateCorrectiveActionParams{
		TenantID: slug, NcID: ncID, ActionType: body.ActionType,
		Description: body.Description, ResponsibleID: optUUID(body.ResponsibleID),
		DueDate: pgDate(dd),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create CA failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ca)
}

func (h *Quality) ListAudits(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	audits, err := h.svc.ListAudits(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list audits failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"audits": audits})
}

func (h *Quality) CreateAudit(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Number, AuditType, Scope, Notes string
		Date string
		LeadAuditorID *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return
	}
	a, err := h.svc.CreateAudit(r.Context(), repository.CreateAuditParams{
		TenantID: slug, Number: body.Number, Date: pgDate(body.Date),
		AuditType: body.AuditType, Scope: body.Scope,
		LeadAuditorID: optUUID(body.LeadAuditorID), Notes: body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create audit failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func (h *Quality) ListFindings(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	findings, err := h.svc.ListAuditFindings(r.Context(), id, tenantSlug(r))
	if err != nil {
		slog.Error("list findings failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"findings": findings})
}

func (h *Quality) CreateFinding(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	auditID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	var body struct {
		FindingType, Description string
		NcID *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return
	}
	f, err := h.svc.CreateAuditFinding(r.Context(), repository.CreateAuditFindingParams{
		TenantID: slug, AuditID: auditID, FindingType: body.FindingType,
		Description: body.Description, NcID: optUUID(body.NcID),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create finding failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

func (h *Quality) ListDocuments(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	docs, err := h.svc.ListDocuments(r.Context(), tenantSlug(r), r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list documents failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"documents": docs})
}

func (h *Quality) CreateDocument(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Code, Title, FileKey string
		Revision int32
		DocTypeID *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return
	}
	if body.Revision == 0 { body.Revision = 1 }
	d, err := h.svc.CreateDocument(r.Context(), repository.CreateControlledDocumentParams{
		TenantID: slug, Code: body.Code, Title: body.Title,
		Revision: body.Revision, DocTypeID: optUUID(body.DocTypeID), FileKey: body.FileKey,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create document failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}
