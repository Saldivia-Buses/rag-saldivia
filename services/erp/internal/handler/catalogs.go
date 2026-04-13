package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// CatalogsService is the interface the Catalogs handler depends on.
type CatalogsService interface {
	List(ctx context.Context, tenantID, catalogType string, activeOnly bool) ([]repository.ErpCatalog, error)
	ListTypes(ctx context.Context, tenantID string) ([]string, error)
	Get(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpCatalog, error)
	Create(ctx context.Context, req service.CreateCatalogRequest) (repository.ErpCatalog, error)
	Update(ctx context.Context, req service.UpdateCatalogRequest) (repository.ErpCatalog, error)
	Delete(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error
}

// Catalogs handles catalog endpoints.
type Catalogs struct {
	svc CatalogsService
}

// NewCatalogs creates a catalog handler.
func NewCatalogs(svc CatalogsService) *Catalogs {
	return &Catalogs{svc: svc}
}

// Routes returns the chi router for catalog endpoints.
func (h *Catalogs) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.catalogs.read"))
		r.Get("/", h.List)
		r.Get("/types", h.ListTypes)
		r.Get("/{id}", h.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.catalogs.write"))
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	return r
}

// List returns catalogs filtered by type query param.
func (h *Catalogs) List(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	catalogType := r.URL.Query().Get("type")
	if catalogType == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("type query param required"))
		return
	}

	activeOnly := r.URL.Query().Get("active") != "false"

	catalogs, err := h.svc.List(r.Context(), slug, catalogType, activeOnly)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"catalogs": catalogs})
}

// ListTypes returns distinct catalog types.
func (h *Catalogs) ListTypes(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	types, err := h.svc.ListTypes(r.Context(), slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"types": types})
}

// Get returns a single catalog entry.
func (h *Catalogs) Get(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}

	catalog, err := h.svc.Get(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("catalog"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}

// Create creates a new catalog entry.
func (h *Catalogs) Create(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	var body struct {
		Type     string           `json:"type"`
		Code     string           `json:"code"`
		Name     string           `json:"name"`
		ParentID *string          `json:"parent_id,omitempty"`
		Sort     int32            `json:"sort_order"`
		Metadata *json.RawMessage `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}

	var parentID pgtype.UUID
	if body.ParentID != nil {
		pid, err := parseUUID(*body.ParentID)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidInput("invalid parent_id"))
			return
		}
		parentID = pid
	}

	var meta []byte
	if body.Metadata != nil {
		meta = []byte(*body.Metadata)
	}

	catalog, err := h.svc.Create(r.Context(), service.CreateCatalogRequest{
		TenantID: slug,
		Type:     body.Type,
		Code:     body.Code,
		Name:     body.Name,
		ParentID: parentID,
		Sort:     body.Sort,
		Metadata: meta,
		UserID:   r.Header.Get("X-User-ID"),
		IP:       r.RemoteAddr,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(catalog)
}

// Update updates an existing catalog entry.
func (h *Catalogs) Update(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}

	var body struct {
		Code     string           `json:"code"`
		Name     string           `json:"name"`
		ParentID *string          `json:"parent_id,omitempty"`
		Sort     int32            `json:"sort_order"`
		Active   *bool            `json:"active,omitempty"`
		Metadata *json.RawMessage `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}

	var parentID pgtype.UUID
	if body.ParentID != nil {
		pid, err := parseUUID(*body.ParentID)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidInput("invalid parent_id"))
			return
		}
		parentID = pid
	}

	active := true
	if body.Active != nil {
		active = *body.Active
	}

	var meta []byte
	if body.Metadata != nil {
		meta = []byte(*body.Metadata)
	}

	catalog, err := h.svc.Update(r.Context(), service.UpdateCatalogRequest{
		ID:       id,
		TenantID: slug,
		Code:     body.Code,
		Name:     body.Name,
		ParentID: parentID,
		Sort:     body.Sort,
		Active:   active,
		Metadata: meta,
		UserID:   r.Header.Get("X-User-ID"),
		IP:       r.RemoteAddr,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}

// Delete soft-deletes a catalog entry.
func (h *Catalogs) Delete(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}

	if err := h.svc.Delete(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("catalog"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
