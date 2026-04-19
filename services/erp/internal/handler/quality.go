package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// QualityService is the interface the Quality handler depends on.
type QualityService interface {
	ListNC(ctx context.Context, tenantID, status, severity string, limit, offset int) ([]repository.ListNonconformitiesRow, error)
	GetNC(ctx context.Context, id pgtype.UUID, tenantID string) (repository.GetNonconformityRow, error)
	CreateNC(ctx context.Context, p repository.CreateNonconformityParams, ip string) (repository.CreateNonconformityRow, error)
	UpdateNCStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error
	ListCA(ctx context.Context, ncID pgtype.UUID, tenantID string) ([]repository.ErpCorrectiveAction, error)
	CreateCA(ctx context.Context, p repository.CreateCorrectiveActionParams, userID, ip string) (repository.ErpCorrectiveAction, error)
	ListAudits(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpAudit, error)
	CreateAudit(ctx context.Context, p repository.CreateAuditParams, userID, ip string) (repository.ErpAudit, error)
	ListAuditFindings(ctx context.Context, auditID pgtype.UUID, tenantID string) ([]repository.ErpAuditFinding, error)
	CreateAuditFinding(ctx context.Context, p repository.CreateAuditFindingParams, userID, ip string) (repository.ErpAuditFinding, error)
	ListDocuments(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ErpControlledDocument, error)
	CreateDocument(ctx context.Context, p repository.CreateControlledDocumentParams, userID, ip string) (repository.ErpControlledDocument, error)
	// NC Origins
	ListNCOrigins(ctx context.Context, tenantID string) ([]repository.ErpNcOrigin, error)
	CreateNCOrigin(ctx context.Context, tenantID, name, userID, ip string) (repository.ErpNcOrigin, error)
	// Action Plans
	ListActionPlans(ctx context.Context, tenantID, ncFilter, statusFilter string, limit, offset int) ([]repository.ListActionPlansRow, error)
	GetActionPlan(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpQualityActionPlan, error)
	CreateActionPlan(ctx context.Context, p repository.CreateActionPlanParams, userID, ip string) (repository.ErpQualityActionPlan, error)
	UpdateActionPlanStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error
	// Action Tasks
	ListActionTasks(ctx context.Context, tenantID string, planID pgtype.UUID) ([]repository.ListActionTasksRow, error)
	CreateActionTask(ctx context.Context, p repository.CreateActionTaskParams, userID, ip string) (repository.ErpQualityActionTask, error)
	CompleteActionTask(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error
	// Indicators
	ListIndicators(ctx context.Context, tenantID, periodFrom, periodTo string) ([]repository.ErpQualityIndicator, error)
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
		r.Get("/nc-origins", h.ListNCOrigins)
		r.Get("/action-plans", h.ListActionPlans)
		r.Get("/action-plans/{id}", h.GetActionPlan)
		r.Get("/action-plans/{id}/tasks", h.ListActionTasks)
		r.Get("/indicators", h.ListIndicators)
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
		r.Post("/nc-origins", h.CreateNCOrigin)
		r.Post("/action-plans", h.CreateActionPlan)
		r.Patch("/action-plans/{id}/status", h.UpdateActionPlanStatus)
		r.Post("/action-plans/{id}/tasks", h.CreateActionTask)
		r.Patch("/action-plans/{id}/tasks/{taskId}/complete", h.CompleteActionTask)
	})
	return r
}

func (h *Quality) ListNC(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	ncs, err := h.svc.ListNC(r.Context(), slug, q.Get("status"), q.Get("severity"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"nonconformities": ncs})
}

func (h *Quality) GetNC(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	nc, err := h.svc.GetNC(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("nonconformity"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(nc)
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body")); return
	}
	if body.Number == "" || body.Description == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("number and description are required")); return
	}
	validSev := map[string]bool{"minor": true, "major": true, "critical": true}
	if !validSev[body.Severity] {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid severity (minor, major, critical)")); return
	}
	nc, err := h.svc.CreateNC(r.Context(), repository.CreateNonconformityParams{
		TenantID: slug, Number: body.Number, Date: pgDate(body.Date),
		TypeID: optUUID(body.TypeID), OriginID: optUUID(body.OriginID),
		Description: body.Description, Severity: body.Severity,
		AssignedTo: optUUID(body.AssignedTo), UserID: r.Header.Get("X-User-ID"),
	}, r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(nc)
}

func (h *Quality) UpdateNCStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id"))); return }
	var body struct{ Status string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body")); return
	}
	validNCStatus := map[string]bool{"open": true, "investigating": true, "corrective_action": true, "closed": true}
	if !validNCStatus[body.Status] {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid status (open, investigating, corrective_action, closed)")); return
	}
	if err := h.svc.UpdateNCStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "NC not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("nonconformity"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Quality) ListCA(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id"))); return }
	cas, err := h.svc.ListCA(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"corrective_actions": cas})
}

func (h *Quality) CreateCA(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	ncID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id"))); return }
	var body struct {
		ActionType, Description string
		ResponsibleID *string
		DueDate *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body")); return
	}
	validActionType := map[string]bool{"corrective": true, "preventive": true}
	if !validActionType[body.ActionType] || body.Description == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("valid action_type and description required")); return
	}
	var dd string
	if body.DueDate != nil { dd = *body.DueDate }
	ca, err := h.svc.CreateCA(r.Context(), repository.CreateCorrectiveActionParams{
		TenantID: slug, NcID: ncID, ActionType: body.ActionType,
		Description: body.Description, ResponsibleID: optUUID(body.ResponsibleID),
		DueDate: pgDate(dd),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ca)
}

func (h *Quality) ListAudits(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	audits, err := h.svc.ListAudits(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"audits": audits})
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body")); return
	}
	a, err := h.svc.CreateAudit(r.Context(), repository.CreateAuditParams{
		TenantID: slug, Number: body.Number, Date: pgDate(body.Date),
		AuditType: body.AuditType, Scope: body.Scope,
		LeadAuditorID: optUUID(body.LeadAuditorID), Notes: body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(a)
}

func (h *Quality) ListFindings(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id"))); return }
	findings, err := h.svc.ListAuditFindings(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"findings": findings})
}

func (h *Quality) CreateFinding(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	auditID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id"))); return }
	var body struct {
		FindingType, Description string
		NcID *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body")); return
	}
	f, err := h.svc.CreateAuditFinding(r.Context(), repository.CreateAuditFindingParams{
		TenantID: slug, AuditID: auditID, FindingType: body.FindingType,
		Description: body.Description, NcID: optUUID(body.NcID),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(f)
}

func (h *Quality) ListDocuments(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	docs, err := h.svc.ListDocuments(r.Context(), tenantSlug(r), r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"documents": docs})
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body")); return
	}
	if body.Revision == 0 { body.Revision = 1 }
	d, err := h.svc.CreateDocument(r.Context(), repository.CreateControlledDocumentParams{
		TenantID: slug, Code: body.Code, Title: body.Title,
		Revision: body.Revision, DocTypeID: optUUID(body.DocTypeID), FileKey: body.FileKey,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err)); return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(d)
}

// ─── NC Origins ────────────────────────────────────────────────────────────

func (h *Quality) ListNCOrigins(w http.ResponseWriter, r *http.Request) {
	origins, err := h.svc.ListNCOrigins(r.Context(), tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"nc_origins": origins})
}

func (h *Quality) CreateNCOrigin(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct{ Name string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Name == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("name is required"))
		return
	}
	o, err := h.svc.CreateNCOrigin(r.Context(), slug, body.Name, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(o)
}

// ─── Action Plans ──────────────────────────────────────────────────────────

func (h *Quality) ListActionPlans(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	plans, err := h.svc.ListActionPlans(r.Context(), tenantSlug(r), q.Get("nc_id"), q.Get("status"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"action_plans": plans})
}

func (h *Quality) GetActionPlan(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	plan, err := h.svc.GetActionPlan(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("action plan"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(plan)
}

func (h *Quality) CreateActionPlan(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Description        string
		NonconformityID    *string
		ResponsibleID      *string
		SectionID          *string
		PlannedStart       *string
		TargetDate         *string
		TimeSavingsHours   *float64
		CostSavings        *float64
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Description == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("description is required"))
		return
	}
	var plannedStart, targetDate string
	if body.PlannedStart != nil { plannedStart = *body.PlannedStart }
	if body.TargetDate != nil { targetDate = *body.TargetDate }
	var timeSavings, costSavings pgtype.Numeric
	if body.TimeSavingsHours != nil {
		timeSavings = numericFromFloat(*body.TimeSavingsHours)
	}
	if body.CostSavings != nil {
		costSavings = numericFromFloat(*body.CostSavings)
	}
	plan, err := h.svc.CreateActionPlan(r.Context(), repository.CreateActionPlanParams{
		TenantID:        slug,
		NonconformityID: optUUID(body.NonconformityID),
		ResponsibleID:   optUUID(body.ResponsibleID),
		SectionID:       optUUID(body.SectionID),
		Description:     body.Description,
		PlannedStart:    pgDate(plannedStart),
		TargetDate:      pgDate(targetDate),
		TimeSavingsHours: timeSavings,
		CostSavings:     costSavings,
		CreatedBy:       r.Header.Get("X-User-ID"),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(plan)
}

func (h *Quality) UpdateActionPlanStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct{ Status string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	validStatus := map[string]bool{"draft": true, "active": true, "closed": true, "cancelled": true}
	if !validStatus[body.Status] {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid status (draft, active, closed, cancelled)"))
		return
	}
	if err := h.svc.UpdateActionPlanStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "action plan not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("action plan"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Action Tasks ──────────────────────────────────────────────────────────

func (h *Quality) ListActionTasks(w http.ResponseWriter, r *http.Request) {
	planID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	tasks, err := h.svc.ListActionTasks(r.Context(), tenantSlug(r), planID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"tasks": tasks})
}

func (h *Quality) CreateActionTask(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	planID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		Description  string
		LeaderID     *string
		PlannedStart *string
		TargetDate   *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Description == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("description is required"))
		return
	}
	var plannedStart, targetDate string
	if body.PlannedStart != nil { plannedStart = *body.PlannedStart }
	if body.TargetDate != nil { targetDate = *body.TargetDate }
	t, err := h.svc.CreateActionTask(r.Context(), repository.CreateActionTaskParams{
		TenantID:     slug,
		PlanID:       planID,
		Description:  body.Description,
		LeaderID:     optUUID(body.LeaderID),
		PlannedStart: pgDate(plannedStart),
		TargetDate:   pgDate(targetDate),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

func (h *Quality) CompleteActionTask(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	taskID, err := parseUUID(chi.URLParam(r, "taskId"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "taskId")))
		return
	}
	if err := h.svc.CompleteActionTask(r.Context(), taskID, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "action task not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("action task"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// numericFromFloat converts float64 to pgtype.Numeric for action plan savings fields.
func numericFromFloat(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.2f", f))
	return n
}

// ListIndicators returns quality KPI rows in a period range.
// Query params: period_from, period_to (both TEXT — e.g. '2025-01').
// When omitted, defaults cover the full migrated range.
func (h *Quality) ListIndicators(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	periodFrom := q.Get("period_from")
	if periodFrom == "" {
		periodFrom = "0000-00"
	}
	periodTo := q.Get("period_to")
	if periodTo == "" {
		periodTo = "9999-99"
	}

	indicators, err := h.svc.ListIndicators(r.Context(), tenantSlug(r), periodFrom, periodTo)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"indicators": indicators})
}
