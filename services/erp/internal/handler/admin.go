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

type Admin struct{ svc *service.Admin }

func NewAdmin(svc *service.Admin) *Admin { return &Admin{svc: svc} }

func (h *Admin) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.admin.read"))
		r.Get("/communications", h.ListCommunications)
		r.Get("/calendar", h.ListCalendarEvents)
		r.Get("/surveys", h.ListSurveys)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.admin.write"))
		r.Post("/communications", h.CreateCommunication)
		r.Post("/calendar", h.CreateCalendarEvent)
		r.Post("/surveys", h.CreateSurvey)
	})
	return r
}

func (h *Admin) ListCommunications(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	comms, err := h.svc.ListCommunications(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil { slog.Error("list communications failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(map[string]any{"communications": comms})
}

func (h *Admin) CreateCommunication(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct{ Subject, Body, Priority string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	c, err := h.svc.CreateCommunication(r.Context(), slug, body.Subject, body.Body, r.Header.Get("X-User-ID"), body.Priority, r.RemoteAddr)
	if err != nil { slog.Error("create communication failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); json.NewEncoder(w).Encode(c)
}

func (h *Admin) ListCalendarEvents(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	q := r.URL.Query()
	var df, dt pgtype.Timestamptz
	if v := q.Get("date_from"); v != "" { _ = df.Scan(v) }
	if v := q.Get("date_to"); v != "" { _ = dt.Scan(v) }
	events, err := h.svc.ListCalendarEvents(r.Context(), slug, df, dt)
	if err != nil { slog.Error("list calendar events failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(map[string]any{"events": events})
}

func (h *Admin) CreateCalendarEvent(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Title, Description string
		StartAt string; EndAt *string
		AllDay bool; EntityID *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	var startAt, endAt pgtype.Timestamptz
	_ = startAt.Scan(body.StartAt)
	if body.EndAt != nil { _ = endAt.Scan(*body.EndAt) }
	ev, err := h.svc.CreateCalendarEvent(r.Context(), repository.CreateCalendarEventParams{
		TenantID: slug, Title: body.Title, Description: body.Description,
		StartAt: startAt, EndAt: endAt, AllDay: body.AllDay,
		EntityID: optUUID(body.EntityID), UserID: r.Header.Get("X-User-ID"),
	}, r.RemoteAddr)
	if err != nil { slog.Error("create calendar event failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); json.NewEncoder(w).Encode(ev)
}

func (h *Admin) ListSurveys(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	surveys, err := h.svc.ListSurveys(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil { slog.Error("list surveys failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(map[string]any{"surveys": surveys})
}

func (h *Admin) CreateSurvey(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct{ Title, Description string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	sv, err := h.svc.CreateSurvey(r.Context(), slug, body.Title, body.Description, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil { slog.Error("create survey failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); json.NewEncoder(w).Encode(sv)
}
