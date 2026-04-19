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

// AdminService is the interface the Admin handler depends on.
type AdminService interface {
	ListCommunications(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpCommunication, error)
	CreateCommunication(ctx context.Context, tenantID, subject, body, senderID, priority, ip string) (repository.ErpCommunication, error)
	ListCalendarEvents(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Timestamptz) ([]repository.ErpCalendarEvent, error)
	CreateCalendarEvent(ctx context.Context, p repository.CreateCalendarEventParams, ip string) (repository.ErpCalendarEvent, error)
	ListSurveys(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpSurvey, error)
	CreateSurvey(ctx context.Context, tenantID, title, description, userID, ip string) (repository.ErpSurvey, error)
	ListProductSections(ctx context.Context, tenantID string) ([]repository.ErpProductSection, error)
	ListProducts(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpProduct, error)
	ListProductAttributes(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpProductAttribute, error)
	ListTools(ctx context.Context, tenantID string, statusFilter int32, articleFilter string, limit, offset int) ([]repository.ErpTool, error)
	GetTool(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpTool, error)
}

type Admin struct{ svc AdminService }

func NewAdmin(svc AdminService) *Admin { return &Admin{svc: svc} }

func (h *Admin) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.admin.read"))
		r.Get("/communications", h.ListCommunications)
		r.Get("/calendar", h.ListCalendarEvents)
		r.Get("/surveys", h.ListSurveys)
		r.Get("/product-sections", h.ListProductSectionsH)
		r.Get("/products", h.ListProductsH)
		r.Get("/product-attributes", h.ListProductAttributesH)
		r.Get("/tools", h.ListToolsH)
		r.Get("/tools/{id}", h.GetToolH)
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
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"communications": comms})
}

func (h *Admin) CreateCommunication(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct{ Subject, Body, Priority string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	c, err := h.svc.CreateCommunication(r.Context(), slug, body.Subject, body.Body, r.Header.Get("X-User-ID"), body.Priority, r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

func (h *Admin) ListCalendarEvents(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	q := r.URL.Query()
	var df, dt pgtype.Timestamptz
	if v := q.Get("date_from"); v != "" { _ = df.Scan(v) }
	if v := q.Get("date_to"); v != "" { _ = dt.Scan(v) }
	events, err := h.svc.ListCalendarEvents(r.Context(), slug, df, dt)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"events": events})
}

func (h *Admin) CreateCalendarEvent(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Title, Description string
		StartAt string
		EndAt *string
		AllDay bool
		EntityID *string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	var startAt, endAt pgtype.Timestamptz
	if err := startAt.Scan(body.StartAt); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid start_at"))
		return
	}
	if body.EndAt != nil { _ = endAt.Scan(*body.EndAt) }
	ev, err := h.svc.CreateCalendarEvent(r.Context(), repository.CreateCalendarEventParams{
		TenantID: slug, Title: body.Title, Description: body.Description,
		StartAt: startAt, EndAt: endAt, AllDay: body.AllDay,
		EntityID: optUUID(body.EntityID), UserID: r.Header.Get("X-User-ID"),
	}, r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ev)
}

func (h *Admin) ListSurveys(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	surveys, err := h.svc.ListSurveys(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"surveys": surveys})
}

func (h *Admin) CreateSurvey(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct{ Title, Description string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	sv, err := h.svc.CreateSurvey(r.Context(), slug, body.Title, body.Description, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sv)
}

// ─── 2.0.17 catalog read handlers ───────────────────────────────────────

func (h *Admin) ListProductSectionsH(w http.ResponseWriter, r *http.Request) {
	sections, err := h.svc.ListProductSections(r.Context(), tenantSlug(r))
	if err != nil { erperrors.WriteError(w, r, erperrors.Internal(err)); return }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"sections": sections})
}

func (h *Admin) ListProductsH(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	products, err := h.svc.ListProducts(r.Context(), tenantSlug(r), p.Limit(), p.Offset())
	if err != nil { erperrors.WriteError(w, r, erperrors.Internal(err)); return }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"products": products})
}

func (h *Admin) ListProductAttributesH(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") != "false"
	attrs, err := h.svc.ListProductAttributes(r.Context(), tenantSlug(r), activeOnly)
	if err != nil { erperrors.WriteError(w, r, erperrors.Internal(err)); return }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"attributes": attrs})
}

func (h *Admin) ListToolsH(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	status := int32(-1)
	if s := q.Get("status"); s != "" {
		var v int32
		_, _ = fmt.Sscanf(s, "%d", &v)
		status = v
	}
	tools, err := h.svc.ListTools(r.Context(), tenantSlug(r), status, q.Get("article_code"), p.Limit(), p.Offset())
	if err != nil { erperrors.WriteError(w, r, erperrors.Internal(err)); return }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"tools": tools})
}

func (h *Admin) GetToolH(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	tool, err := h.svc.GetTool(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("tool"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tool)
}
