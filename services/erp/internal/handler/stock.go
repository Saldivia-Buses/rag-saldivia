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

// StockService is the interface the Stock handler depends on.
type StockService interface {
	ListArticles(ctx context.Context, tenantID, search, articleType string, activeOnly bool, limit, offset int) ([]repository.ErpArticle, error)
	GetArticle(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpArticle, error)
	CreateArticle(ctx context.Context, req service.CreateArticleRequest) (repository.ErpArticle, error)
	ListWarehouses(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpWarehouse, error)
	CreateWarehouse(ctx context.Context, tenantID, code, name, location, userID, ip string) (repository.ErpWarehouse, error)
	GetStockLevels(ctx context.Context, tenantID string, articleID, warehouseID pgtype.UUID) ([]repository.GetStockLevelsRow, error)
	ListMovements(ctx context.Context, tenantID string, articleID pgtype.UUID, limit, offset int) ([]repository.ListStockMovementsRow, error)
	CreateMovement(ctx context.Context, req service.CreateMovementRequest) (repository.ErpStockMovement, error)
	ListBOM(ctx context.Context, tenantID string, parentID pgtype.UUID) ([]repository.ListBOMRow, error)
}

// Stock handles stock & warehouse endpoints.
type Stock struct{ svc StockService }

func NewStock(svc StockService) *Stock { return &Stock{svc: svc} }

func (h *Stock) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.stock.read"))
		r.Get("/articles", h.ListArticles)
		r.Get("/articles/{id}", h.GetArticle)
		r.Get("/warehouses", h.ListWarehouses)
		r.Get("/levels", h.GetStockLevels)
		r.Get("/movements", h.ListMovements)
		r.Get("/articles/{id}/bom", h.ListBOM)
	})

	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.stock.write"))
		r.Post("/articles", h.CreateArticle)
		r.Post("/warehouses", h.CreateWarehouse)
		r.Post("/movements", h.CreateMovement)
	})

	return r
}

func (h *Stock) ListArticles(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	search := r.URL.Query().Get("search")
	artType := r.URL.Query().Get("type")
	activeOnly := r.URL.Query().Get("active") != "false"

	articles, err := h.svc.ListArticles(r.Context(), slug, search, artType, activeOnly, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"articles": articles})
}

func (h *Stock) GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	article, err := h.svc.GetArticle(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("article"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(article)
}

func (h *Stock) CreateArticle(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	var body struct {
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		FamilyID    *string `json:"family_id,omitempty"`
		CategoryID  *string `json:"category_id,omitempty"`
		UnitID      *string `json:"unit_id,omitempty"`
		ArticleType string  `json:"article_type"`
		MinStock    string  `json:"min_stock"`
		MaxStock    string  `json:"max_stock"`
		ReorderPt   string  `json:"reorder_point"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}

	article, err := h.svc.CreateArticle(r.Context(), service.CreateArticleRequest{
		TenantID:    slug,
		Code:        body.Code,
		Name:        body.Name,
		FamilyID:    optUUID(body.FamilyID),
		CategoryID:  optUUID(body.CategoryID),
		UnitID:      optUUID(body.UnitID),
		ArticleType: body.ArticleType,
		MinStock:    body.MinStock,
		MaxStock:    body.MaxStock,
		ReorderPt:   body.ReorderPt,
		UserID:      r.Header.Get("X-User-ID"),
		IP:          r.RemoteAddr,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(article)
}

func (h *Stock) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	activeOnly := r.URL.Query().Get("active") != "false"
	warehouses, err := h.svc.ListWarehouses(r.Context(), slug, activeOnly)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"warehouses": warehouses})
}

func (h *Stock) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		Code     string `json:"code"`
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	wh, err := h.svc.CreateWarehouse(r.Context(), slug, body.Code, body.Name, body.Location, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(wh)
}

func (h *Stock) GetStockLevels(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	articleID := optUUID(ptrStr(r.URL.Query().Get("article_id")))
	warehouseID := optUUID(ptrStr(r.URL.Query().Get("warehouse_id")))

	levels, err := h.svc.GetStockLevels(r.Context(), slug, articleID, warehouseID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"levels": levels})
}

func (h *Stock) ListMovements(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	articleID := optUUID(ptrStr(r.URL.Query().Get("article_id")))

	movements, err := h.svc.ListMovements(r.Context(), slug, articleID, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"movements": movements})
}

func (h *Stock) CreateMovement(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		ArticleID   string  `json:"article_id"`
		WarehouseID string  `json:"warehouse_id"`
		Type        string  `json:"movement_type"`
		Quantity    string  `json:"quantity"`
		UnitCost    string  `json:"unit_cost"`
		Notes       string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}

	artID, err := parseUUID(body.ArticleID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID("article_id"))
		return
	}
	whID, err := parseUUID(body.WarehouseID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID("warehouse_id"))
		return
	}

	mov, err := h.svc.CreateMovement(r.Context(), service.CreateMovementRequest{
		TenantID:     slug,
		ArticleID:    artID,
		WarehouseID:  whID,
		MovementType: body.Type,
		Quantity:      body.Quantity,
		UnitCost:      body.UnitCost,
		UserID:       r.Header.Get("X-User-ID"),
		Notes:        body.Notes,
		IP:           r.RemoteAddr,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(mov)
}

func (h *Stock) ListBOM(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	bom, err := h.svc.ListBOM(r.Context(), slug, id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"bom": bom})
}

// optUUID parses an optional UUID string pointer into pgtype.UUID.
func optUUID(s *string) pgtype.UUID {
	if s == nil || *s == "" {
		return pgtype.UUID{}
	}
	id, err := parseUUID(*s)
	if err != nil {
		return pgtype.UUID{}
	}
	return id
}

func ptrStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
