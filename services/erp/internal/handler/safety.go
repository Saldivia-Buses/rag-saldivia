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

// SafetyService is the interface the Safety handler depends on.
type SafetyService interface {
	ListAccidentTypes(ctx context.Context, tenantID string) ([]repository.ErpAccidentType, error)
	ListBodyParts(ctx context.Context, tenantID string) ([]repository.ErpBodyPart, error)
	ListRiskAgents(ctx context.Context, tenantID, riskType string) ([]repository.ErpRiskAgent, error)
	ListWorkAccidents(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListWorkAccidentsRow, error)
	GetWorkAccident(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpWorkAccident, error)
	CreateWorkAccident(ctx context.Context, p repository.CreateWorkAccidentParams, userID, ip string) (repository.ErpWorkAccident, error)
	UpdateAccidentStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error
	ListEmployeeRiskExposures(ctx context.Context, tenantID, entityID string) ([]repository.ListEmployeeRiskExposuresRow, error)
	CreateRiskExposure(ctx context.Context, p repository.CreateRiskExposureParams, userID, ip string) (repository.ErpEmployeeRiskExposure, error)
	ListMedicalConsultations(ctx context.Context, tenantID, dateFrom, dateTo string, limit, offset int) ([]repository.ErpMedicalConsultation, error)
	CreateMedicalConsultation(ctx context.Context, p repository.CreateMedicalConsultationParams, userID, ip string) (repository.ErpMedicalConsultation, error)
	ListMedicalLeaves(ctx context.Context, tenantID, entityID, leaveType, status string, limit, offset int) ([]repository.ListMedicalLeavesRow, error)
	CreateMedicalLeave(ctx context.Context, p repository.CreateMedicalLeaveParams, userID, ip string) (repository.ErpMedicalLeafe, error)
	ApproveMedicalLeave(ctx context.Context, id pgtype.UUID, tenantID, approvedBy, userID, ip string) error
	GetSafetyKPIs(ctx context.Context, tenantID, dateFrom, dateTo string) (repository.GetSafetyKPIsRow, error)
}

type Safety struct{ svc SafetyService }

func NewSafety(svc SafetyService) *Safety { return &Safety{svc: svc} }

func (h *Safety) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.safety.read"))
		r.Get("/accident-types", h.ListAccidentTypes)
		r.Get("/body-parts", h.ListBodyParts)
		r.Get("/risk-agents", h.ListRiskAgents)
		r.Get("/accidents", h.ListAccidents)
		r.Get("/accidents/{id}", h.GetAccident)
		r.Get("/risk-exposures", h.ListRiskExposures)
		r.Get("/medical-log", h.ListMedicalLog)
		r.Get("/medical-leaves", h.ListMedicalLeaves)
		r.Get("/kpis", h.GetKPIs)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.safety.write"))
		r.Post("/accidents", h.CreateAccident)
		r.Patch("/accidents/{id}/status", h.UpdateAccidentStatus)
		r.Post("/risk-exposures", h.CreateRiskExposure)
		r.Post("/medical-log", h.CreateMedicalConsultation)
		r.Post("/medical-leaves", h.CreateMedicalLeave)
		r.Patch("/medical-leaves/{id}/approve", h.ApproveMedicalLeave)
	})
	return r
}

// ── Catalogs ──────────────────────────────────────────────────────────────────

func (h *Safety) ListAccidentTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.svc.ListAccidentTypes(r.Context(), tenantSlug(r))
	if err != nil {
		slog.Error("list accident types failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"accident_types": types})
}

func (h *Safety) ListBodyParts(w http.ResponseWriter, r *http.Request) {
	parts, err := h.svc.ListBodyParts(r.Context(), tenantSlug(r))
	if err != nil {
		slog.Error("list body parts failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"body_parts": parts})
}

func (h *Safety) ListRiskAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.svc.ListRiskAgents(r.Context(), tenantSlug(r), r.URL.Query().Get("risk_type"))
	if err != nil {
		slog.Error("list risk agents failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"risk_agents": agents})
}

// ── Work Accidents ────────────────────────────────────────────────────────────

func (h *Safety) ListAccidents(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	accidents, err := h.svc.ListWorkAccidents(r.Context(), tenantSlug(r), r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list accidents failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"accidents": accidents})
}

func (h *Safety) GetAccident(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	wa, err := h.svc.GetWorkAccident(r.Context(), id, tenantSlug(r))
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wa)
}

func (h *Safety) CreateAccident(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		IncidentDate   string  `json:"incident_date"`
		RecoveryDate   *string `json:"recovery_date"`
		ReportedBy     string  `json:"reported_by"`
		Observations   string  `json:"observations"`
		LostDays       int32   `json:"lost_days"`
		EntityID       *string `json:"entity_id"`
		AccidentTypeID *string `json:"accident_type_id"`
		BodyPartID     *string `json:"body_part_id"`
		SectionID      *string `json:"section_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if body.IncidentDate == "" {
		http.Error(w, `{"error":"incident_date is required"}`, http.StatusBadRequest)
		return
	}
	var rd string
	if body.RecoveryDate != nil {
		rd = *body.RecoveryDate
	}
	wa, err := h.svc.CreateWorkAccident(r.Context(), repository.CreateWorkAccidentParams{
		TenantID:       slug,
		EntityID:       optUUID(body.EntityID),
		AccidentTypeID: optUUID(body.AccidentTypeID),
		BodyPartID:     optUUID(body.BodyPartID),
		SectionID:      optUUID(body.SectionID),
		IncidentDate:   pgDate(body.IncidentDate),
		RecoveryDate:   pgDate(rd),
		LostDays:       body.LostDays,
		Observations:   body.Observations,
		ReportedBy:     body.ReportedBy,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create accident failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wa)
}

func (h *Safety) UpdateAccidentStatus(w http.ResponseWriter, r *http.Request) {
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
	validStatus := map[string]bool{"open": true, "investigating": true, "closed": true}
	if !validStatus[body.Status] {
		http.Error(w, `{"error":"invalid status (open, investigating, closed)"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateAccidentStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "accident not found" {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		} else {
			slog.Error("update accident status failed", "error", err)
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Risk Exposures ────────────────────────────────────────────────────────────

func (h *Safety) ListRiskExposures(w http.ResponseWriter, r *http.Request) {
	exposures, err := h.svc.ListEmployeeRiskExposures(r.Context(), tenantSlug(r), r.URL.Query().Get("entity_id"))
	if err != nil {
		slog.Error("list risk exposures failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"risk_exposures": exposures})
}

func (h *Safety) CreateRiskExposure(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		EntityID     string  `json:"entity_id"`
		RiskAgentID  string  `json:"risk_agent_id"`
		SectionID    *string `json:"section_id"`
		ExposedFrom  string  `json:"exposed_from"`
		ExposedUntil *string `json:"exposed_until"`
		Notes        string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if body.EntityID == "" || body.RiskAgentID == "" || body.ExposedFrom == "" {
		http.Error(w, `{"error":"entity_id, risk_agent_id and exposed_from are required"}`, http.StatusBadRequest)
		return
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		http.Error(w, `{"error":"invalid entity_id"}`, http.StatusBadRequest)
		return
	}
	riskAgentID, err := parseUUID(body.RiskAgentID)
	if err != nil {
		http.Error(w, `{"error":"invalid risk_agent_id"}`, http.StatusBadRequest)
		return
	}
	var eu string
	if body.ExposedUntil != nil {
		eu = *body.ExposedUntil
	}
	exp, err := h.svc.CreateRiskExposure(r.Context(), repository.CreateRiskExposureParams{
		TenantID:     slug,
		EntityID:     entityID,
		RiskAgentID:  riskAgentID,
		SectionID:    optUUID(body.SectionID),
		ExposedFrom:  pgDate(body.ExposedFrom),
		ExposedUntil: pgDate(eu),
		Notes:        body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create risk exposure failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(exp)
}

// ── Medical Consultations ─────────────────────────────────────────────────────

func (h *Safety) ListMedicalLog(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	consultations, err := h.svc.ListMedicalConsultations(r.Context(), tenantSlug(r), q.Get("date_from"), q.Get("date_to"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list medical log failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"consultations": consultations})
}

func (h *Safety) CreateMedicalConsultation(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		PatientName  string  `json:"patient_name"`
		ConsultDate  string  `json:"consult_date"`
		ConsultTime  string  `json:"consult_time"`
		Symptoms     string  `json:"symptoms"`
		Prescription string  `json:"prescription"`
		MedicUser    string  `json:"medic_user"`
		EntityID     *string `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if body.ConsultDate == "" {
		http.Error(w, `{"error":"consult_date is required"}`, http.StatusBadRequest)
		return
	}
	mc, err := h.svc.CreateMedicalConsultation(r.Context(), repository.CreateMedicalConsultationParams{
		TenantID:     slug,
		EntityID:     optUUID(body.EntityID),
		PatientName:  body.PatientName,
		ConsultDate:  pgDate(body.ConsultDate),
		ConsultTime:  pgTime(body.ConsultTime),
		Symptoms:     body.Symptoms,
		Prescription: body.Prescription,
		MedicUser:    body.MedicUser,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create medical consultation failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mc)
}

// ── Medical Leaves ────────────────────────────────────────────────────────────

func (h *Safety) ListMedicalLeaves(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	leaves, err := h.svc.ListMedicalLeaves(r.Context(), tenantSlug(r), q.Get("entity_id"), q.Get("leave_type"), q.Get("status"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list medical leaves failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"medical_leaves": leaves})
}

func (h *Safety) CreateMedicalLeave(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		EntityID     string  `json:"entity_id"`
		LeaveType    string  `json:"leave_type"`
		DateFrom     string  `json:"date_from"`
		DateTo       string  `json:"date_to"`
		WorkingDays  int32   `json:"working_days"`
		Observations string  `json:"observations"`
		BodyPartID   *string `json:"body_part_id"`
		AccidentID   *string `json:"accident_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if body.EntityID == "" || body.DateFrom == "" || body.DateTo == "" {
		http.Error(w, `{"error":"entity_id, date_from and date_to are required"}`, http.StatusBadRequest)
		return
	}
	validLeaveType := map[string]bool{"illness": true, "accident": true, "vacation": true, "leave": true, "other": true}
	if body.LeaveType != "" && !validLeaveType[body.LeaveType] {
		http.Error(w, `{"error":"invalid leave_type (illness, accident, vacation, leave, other)"}`, http.StatusBadRequest)
		return
	}
	if body.LeaveType == "" {
		body.LeaveType = "illness"
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		http.Error(w, `{"error":"invalid entity_id"}`, http.StatusBadRequest)
		return
	}
	ml, err := h.svc.CreateMedicalLeave(r.Context(), repository.CreateMedicalLeaveParams{
		TenantID:     slug,
		EntityID:     entityID,
		BodyPartID:   optUUID(body.BodyPartID),
		AccidentID:   optUUID(body.AccidentID),
		LeaveType:    body.LeaveType,
		DateFrom:     pgDate(body.DateFrom),
		DateTo:       pgDate(body.DateTo),
		WorkingDays:  body.WorkingDays,
		Observations: body.Observations,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create medical leave failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ml)
}

func (h *Safety) ApproveMedicalLeave(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		ApprovedBy string `json:"approved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if body.ApprovedBy == "" {
		body.ApprovedBy = r.Header.Get("X-User-ID")
	}
	if err := h.svc.ApproveMedicalLeave(r.Context(), id, slug, body.ApprovedBy, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "medical leave not found" {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		} else {
			slog.Error("approve medical leave failed", "error", err)
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── KPIs ──────────────────────────────────────────────────────────────────────

func (h *Safety) GetKPIs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	kpis, err := h.svc.GetSafetyKPIs(r.Context(), tenantSlug(r), q.Get("date_from"), q.Get("date_to"))
	if err != nil {
		slog.Error("get safety KPIs failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kpis)
}

// pgTime converts a time string "HH:MM" or "HH:MM:SS" to pgtype.Time.
func pgTime(s string) pgtype.Time {
	if s == "" {
		return pgtype.Time{}
	}
	var t pgtype.Time
	_ = t.Scan(s)
	return t
}
