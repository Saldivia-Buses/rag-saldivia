package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

type HR struct{ svc *service.HR }

func NewHR(svc *service.HR) *HR { return &HR{svc: svc} }

func (h *HR) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.hr.read"))
		r.Get("/departments", h.ListDepartments)
		r.Get("/employees", h.ListEmployees)
		r.Get("/employees/{id}", h.GetEmployee)
		r.Get("/events", h.ListEvents)
		r.Get("/training", h.ListTraining)
		r.Get("/attendance", h.ListAttendance)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.hr.write"))
		r.Post("/departments", h.CreateDepartment)
		r.Post("/employees", h.UpsertEmployee)
		r.Post("/events", h.CreateEvent)
		r.Post("/training", h.CreateTraining)
		r.Post("/attendance", h.CreateAttendance)
	})
	return r
}

func (h *HR) ListDepartments(w http.ResponseWriter, r *http.Request) {
	depts, err := h.svc.ListDepartments(r.Context(), tenantSlug(r))
	if err != nil {
		slog.Error("list departments failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"departments": depts})
}

func (h *HR) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct{ Code, Name string; ParentID, ManagerID *string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	d, err := h.svc.CreateDepartment(r.Context(), repository.CreateDepartmentParams{
		TenantID: slug, Code: body.Code, Name: body.Name,
		ParentID: optUUID(body.ParentID), ManagerID: optUUID(body.ManagerID),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create department failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func (h *HR) ListEmployees(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	employees, err := h.svc.ListEmployees(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list employees failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"employees": employees})
}

func (h *HR) GetEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	ed, err := h.svc.GetEmployee(r.Context(), id, tenantSlug(r), r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ed)
}

func (h *HR) UpsertEmployee(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		EntityID     string  `json:"entity_id"`
		DepartmentID *string `json:"department_id,omitempty"`
		Position     string  `json:"position"`
		HireDate     *string `json:"hire_date,omitempty"`
		UnionID      *string `json:"union_id,omitempty"`
		HealthPlanID *string `json:"health_plan_id,omitempty"`
		ScheduleType string  `json:"schedule_type"`
		CategoryID   *string `json:"category_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		http.Error(w, `{"error":"invalid entity_id"}`, http.StatusBadRequest)
		return
	}
	var hd string
	if body.HireDate != nil { hd = *body.HireDate }
	if body.ScheduleType == "" { body.ScheduleType = "full_time" }
	ed, err := h.svc.UpsertEmployee(r.Context(), repository.UpsertEmployeeDetailParams{
		TenantID: slug, EntityID: entityID, DepartmentID: optUUID(body.DepartmentID),
		Position: body.Position, HireDate: pgDate(hd),
		UnionID: optUUID(body.UnionID), HealthPlanID: optUUID(body.HealthPlanID),
		ScheduleType: body.ScheduleType, CategoryID: optUUID(body.CategoryID),
		Metadata: []byte(`{}`),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("upsert employee failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ed)
}

func (h *HR) ListEvents(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	events, err := h.svc.ListEvents(r.Context(), slug, optUUID(ptrStr(q.Get("entity_id"))), q.Get("type"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list hr events failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"events": events})
}

func (h *HR) CreateEvent(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		EntityID  string  `json:"entity_id"`
		EventType string  `json:"event_type"`
		DateFrom  string  `json:"date_from"`
		DateTo    *string `json:"date_to,omitempty"`
		Hours     *string `json:"hours,omitempty"`
		ReasonID  *string `json:"reason_id,omitempty"`
		Notes     string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		http.Error(w, `{"error":"invalid entity_id"}`, http.StatusBadRequest)
		return
	}
	var dt string
	if body.DateTo != nil { dt = *body.DateTo }
	var hrs pgNumericVal
	if body.Hours != nil { hrs.set(*body.Hours) }
	ev, err := h.svc.CreateEvent(r.Context(), repository.CreateHREventParams{
		TenantID: slug, EntityID: entityID, EventType: body.EventType,
		DateFrom: pgDate(body.DateFrom), DateTo: pgDate(dt),
		Hours: hrs.n, ReasonID: optUUID(body.ReasonID),
		Notes: body.Notes, UserID: r.Header.Get("X-User-ID"),
	}, r.RemoteAddr)
	if err != nil {
		slog.Error("create hr event failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ev)
}

func (h *HR) ListTraining(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	training, err := h.svc.ListTraining(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list training failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"training": training})
}

func (h *HR) CreateTraining(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Name, Description, Instructor string
		DateFrom, DateTo *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	var df, dt string
	if body.DateFrom != nil { df = *body.DateFrom }
	if body.DateTo != nil { dt = *body.DateTo }
	t, err := h.svc.CreateTraining(r.Context(), repository.CreateTrainingParams{
		TenantID: slug, Name: body.Name, Description: body.Description,
		Instructor: body.Instructor, DateFrom: pgDate(df), DateTo: pgDate(dt),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create training failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *HR) ListAttendance(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	att, err := h.svc.ListAttendance(r.Context(), slug, optUUID(ptrStr(q.Get("entity_id"))),
		pgDate(q.Get("date_from")), pgDate(q.Get("date_to")), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list attendance failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"attendance": att})
}

func (h *HR) CreateAttendance(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		EntityID string  `json:"entity_id"`
		Date     string  `json:"date"`
		ClockIn  *string `json:"clock_in,omitempty"`
		ClockOut *string `json:"clock_out,omitempty"`
		Hours    *string `json:"hours,omitempty"`
		Source   string  `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		http.Error(w, `{"error":"invalid entity_id"}`, http.StatusBadRequest)
		return
	}
	if body.Source == "" { body.Source = "manual" }
	var hrs pgNumericVal
	if body.Hours != nil { hrs.set(*body.Hours) }
	// ClockIn/ClockOut as pgtype.Timestamptz
	var ci, co pgTimestamptzVal
	if body.ClockIn != nil { ci.set(*body.ClockIn) }
	if body.ClockOut != nil { co.set(*body.ClockOut) }
	a, err := h.svc.CreateAttendance(r.Context(), repository.CreateAttendanceParams{
		TenantID: slug, EntityID: entityID, Date: pgDate(body.Date),
		ClockIn: ci.t, ClockOut: co.t, Hours: hrs.n, Source: body.Source,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create attendance failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// helpers for optional pgtype values
type pgNumericVal struct{ n pgtype.Numeric }
func (v *pgNumericVal) set(s string) { _ = v.n.Scan(s) }

type pgTimestamptzVal struct{ t pgtype.Timestamptz }
func (v *pgTimestamptzVal) set(s string) { _ = v.t.Scan(s) }
