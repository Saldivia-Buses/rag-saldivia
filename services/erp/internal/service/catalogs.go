package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

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
func (s *Catalogs) List(ctx context.Context, tenantID, catalogType string, activeOnly bool) ([]repository.Catalog, error) {
	return s.repo.ListCatalogs(ctx, tenantID, catalogType, activeOnly)
}

// ListTypes returns distinct catalog types.
func (s *Catalogs) ListTypes(ctx context.Context, tenantID string) ([]string, error) {
	return s.repo.ListCatalogTypes(ctx, tenantID)
}

// Get returns a single catalog entry.
func (s *Catalogs) Get(ctx context.Context, id uuid.UUID, tenantID string) (*repository.Catalog, error) {
	return s.repo.GetCatalog(ctx, id, tenantID)
}

// CreateCatalogRequest holds data for creating a catalog entry.
type CreateCatalogRequest struct {
	TenantID string
	Type     string
	Code     string
	Name     string
	ParentID *uuid.UUID
	Sort     int
	Metadata json.RawMessage
	UserID   string
	IP       string
}

// Create creates a new catalog entry.
func (s *Catalogs) Create(ctx context.Context, req CreateCatalogRequest) (*repository.Catalog, error) {
	if req.Type == "" || req.Code == "" || req.Name == "" {
		return nil, fmt.Errorf("type, code, and name are required")
	}

	catalog, err := s.repo.CreateCatalog(ctx, req.TenantID, req.Type, req.Code, req.Name, req.ParentID, req.Sort, req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("create catalog: %w", err)
	}

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.catalog.created",
		Resource: catalog.ID.String(),
		Details:  map[string]any{"type": req.Type, "code": req.Code},
		IP:       req.IP,
	})

	s.publisher.Broadcast(req.TenantID, "erp_catalogs", map[string]any{
		"action":     "created",
		"catalog_id": catalog.ID.String(),
		"type":       req.Type,
	})

	slog.Info("catalog created", "id", catalog.ID, "type", req.Type, "code", req.Code)
	return catalog, nil
}

// UpdateCatalogRequest holds data for updating a catalog entry.
type UpdateCatalogRequest struct {
	ID       uuid.UUID
	TenantID string
	Code     string
	Name     string
	ParentID *uuid.UUID
	Sort     int
	Active   bool
	Metadata json.RawMessage
	UserID   string
	IP       string
}

// Update updates an existing catalog entry.
func (s *Catalogs) Update(ctx context.Context, req UpdateCatalogRequest) (*repository.Catalog, error) {
	if req.Code == "" || req.Name == "" {
		return nil, fmt.Errorf("code and name are required")
	}

	catalog, err := s.repo.UpdateCatalog(ctx, req.ID, req.TenantID, req.Code, req.Name, req.ParentID, req.Sort, req.Active, req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("update catalog: %w", err)
	}

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.catalog.updated",
		Resource: catalog.ID.String(),
		Details:  map[string]any{"code": req.Code},
		IP:       req.IP,
	})

	s.publisher.Broadcast(req.TenantID, "erp_catalogs", map[string]any{
		"action":     "updated",
		"catalog_id": catalog.ID.String(),
	})

	return catalog, nil
}

// Delete soft-deletes a catalog entry.
func (s *Catalogs) Delete(ctx context.Context, id uuid.UUID, tenantID, userID, ip string) error {
	if err := s.repo.DeleteCatalog(ctx, id, tenantID); err != nil {
		return err
	}

	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID,
		UserID:   userID,
		Action:   "erp.catalog.deleted",
		Resource: id.String(),
		IP:       ip,
	})

	s.publisher.Broadcast(tenantID, "erp_catalogs", map[string]any{
		"action":     "deleted",
		"catalog_id": id.String(),
	})

	return nil
}
