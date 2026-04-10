package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Catalogs handles catalog business logic.
type Catalogs struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

// NewCatalogs creates a catalogs service.
func NewCatalogs(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Catalogs {
	return &Catalogs{
		repo:      repo,
		audit:     auditWriter,
		publisher: publisher,
	}
}

// List returns catalogs filtered by type.
func (s *Catalogs) List(ctx context.Context, tenantID, catalogType string, activeOnly bool) ([]repository.ErpCatalog, error) {
	return s.repo.ListCatalogs(ctx, repository.ListCatalogsParams{
		TenantID:   tenantID,
		Type:       catalogType,
		ActiveOnly: activeOnly,
	})
}

// ListTypes returns distinct catalog types.
func (s *Catalogs) ListTypes(ctx context.Context, tenantID string) ([]string, error) {
	return s.repo.ListCatalogTypes(ctx, tenantID)
}

// Get returns a single catalog entry.
func (s *Catalogs) Get(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpCatalog, error) {
	return s.repo.GetCatalog(ctx, repository.GetCatalogParams{
		ID:       id,
		TenantID: tenantID,
	})
}

// CreateCatalogRequest holds data for creating a catalog entry.
type CreateCatalogRequest struct {
	TenantID string
	Type     string
	Code     string
	Name     string
	ParentID pgtype.UUID
	Sort     int32
	Metadata []byte
	UserID   string
	IP       string
}

// Create creates a new catalog entry.
func (s *Catalogs) Create(ctx context.Context, req CreateCatalogRequest) (repository.ErpCatalog, error) {
	if req.Type == "" || req.Code == "" || req.Name == "" {
		return repository.ErpCatalog{}, fmt.Errorf("type, code, and name are required")
	}

	if req.Metadata == nil {
		req.Metadata = []byte(`{}`)
	}

	catalog, err := s.repo.CreateCatalog(ctx, repository.CreateCatalogParams{
		TenantID:  req.TenantID,
		Type:      req.Type,
		Code:      req.Code,
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.Sort,
		Metadata:  req.Metadata,
	})
	if err != nil {
		return repository.ErpCatalog{}, fmt.Errorf("create catalog: %w", err)
	}

	idStr := uuidStr(catalog.ID)

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.catalog.created",
		Resource: idStr,
		Details:  map[string]any{"type": req.Type, "code": req.Code},
		IP:       req.IP,
	})

	s.publisher.Broadcast(req.TenantID, "erp_catalogs", map[string]any{
		"action":     "created",
		"catalog_id": idStr,
		"type":       req.Type,
	})

	slog.Info("catalog created", "id", idStr, "type", req.Type, "code", req.Code)
	return catalog, nil
}

// UpdateCatalogRequest holds data for updating a catalog entry.
type UpdateCatalogRequest struct {
	ID       pgtype.UUID
	TenantID string
	Code     string
	Name     string
	ParentID pgtype.UUID
	Sort     int32
	Active   bool
	Metadata []byte
	UserID   string
	IP       string
}

// Update updates an existing catalog entry.
func (s *Catalogs) Update(ctx context.Context, req UpdateCatalogRequest) (repository.ErpCatalog, error) {
	if req.Code == "" || req.Name == "" {
		return repository.ErpCatalog{}, fmt.Errorf("code and name are required")
	}

	if req.Metadata == nil {
		req.Metadata = []byte(`{}`)
	}

	catalog, err := s.repo.UpdateCatalog(ctx, repository.UpdateCatalogParams{
		ID:        req.ID,
		TenantID:  req.TenantID,
		Code:      req.Code,
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.Sort,
		Active:    req.Active,
		Metadata:  req.Metadata,
	})
	if err != nil {
		return repository.ErpCatalog{}, fmt.Errorf("update catalog: %w", err)
	}

	idStr := uuidStr(catalog.ID)

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.catalog.updated",
		Resource: idStr,
		Details:  map[string]any{"code": req.Code},
		IP:       req.IP,
	})

	s.publisher.Broadcast(req.TenantID, "erp_catalogs", map[string]any{
		"action":     "updated",
		"catalog_id": idStr,
	})

	return catalog, nil
}

// Delete soft-deletes a catalog entry.
func (s *Catalogs) Delete(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	rows, err := s.repo.DeleteCatalog(ctx, repository.DeleteCatalogParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("delete catalog: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("catalog not found")
	}

	idStr := uuidStr(id)

	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID,
		UserID:   userID,
		Action:   "erp.catalog.deleted",
		Resource: idStr,
		IP:       ip,
	})

	s.publisher.Broadcast(tenantID, "erp_catalogs", map[string]any{
		"action":     "deleted",
		"catalog_id": idStr,
	})

	return nil
}
