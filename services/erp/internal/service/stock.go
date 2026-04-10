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

// Stock handles stock & warehouse business logic.
type Stock struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     *audit.Writer
	publisher *traces.Publisher
}

// NewStock creates a stock service.
func NewStock(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Stock {
	return &Stock{repo: repo, pool: pool, audit: auditWriter, publisher: publisher}
}

var validMovementTypes = map[string]bool{"in": true, "out": true, "transfer": true, "adjustment": true}

// ListArticles returns paginated articles.
func (s *Stock) ListArticles(ctx context.Context, tenantID, search, articleType string, activeOnly bool, limit, offset int) ([]repository.ErpArticle, error) {
	return s.repo.ListArticles(ctx, repository.ListArticlesParams{
		TenantID:          tenantID,
		ActiveOnly:        activeOnly,
		Search:            likeEscaper.Replace(search),
		ArticleTypeFilter: articleType,
		Limit:             int32(limit),
		Offset:            int32(offset),
	})
}

// GetArticle returns an article by ID.
func (s *Stock) GetArticle(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpArticle, error) {
	return s.repo.GetArticle(ctx, repository.GetArticleParams{ID: id, TenantID: tenantID})
}

// CreateArticleRequest holds data for creating an article.
type CreateArticleRequest struct {
	TenantID    string
	Code        string
	Name        string
	FamilyID    pgtype.UUID
	CategoryID  pgtype.UUID
	UnitID      pgtype.UUID
	ArticleType string
	MinStock    string
	MaxStock    string
	ReorderPt   string
	Metadata    []byte
	UserID      string
	IP          string
}

// CreateArticle creates a new article.
func (s *Stock) CreateArticle(ctx context.Context, req CreateArticleRequest) (repository.ErpArticle, error) {
	if req.Code == "" || req.Name == "" {
		return repository.ErpArticle{}, fmt.Errorf("code and name are required")
	}
	if req.Metadata == nil {
		req.Metadata = []byte(`{}`)
	}

	article, err := s.repo.CreateArticle(ctx, repository.CreateArticleParams{
		TenantID:    req.TenantID,
		Code:        req.Code,
		Name:        req.Name,
		FamilyID:    req.FamilyID,
		CategoryID:  req.CategoryID,
		UnitID:      req.UnitID,
		ArticleType: req.ArticleType,
		MinStock:    pgNumeric(req.MinStock),
		MaxStock:    pgNumeric(req.MaxStock),
		ReorderPoint: pgNumeric(req.ReorderPt),
		Metadata:    req.Metadata,
	})
	if err != nil {
		return repository.ErpArticle{}, fmt.Errorf("create article: %w", err)
	}

	idStr := uuidStr(article.ID)
	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.article.created", Resource: idStr,
		Details: map[string]any{"code": req.Code}, IP: req.IP,
	})
	s.publisher.Broadcast(req.TenantID, "erp_stock", map[string]any{
		"action": "article_created", "article_id": idStr,
	})

	slog.Info("article created", "id", idStr, "code", req.Code)
	return article, nil
}

// ListWarehouses returns all warehouses.
func (s *Stock) ListWarehouses(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpWarehouse, error) {
	return s.repo.ListWarehouses(ctx, repository.ListWarehousesParams{
		TenantID: tenantID, ActiveOnly: activeOnly,
	})
}

// CreateWarehouse creates a new warehouse.
func (s *Stock) CreateWarehouse(ctx context.Context, tenantID, code, name, location, userID, ip string) (repository.ErpWarehouse, error) {
	wh, err := s.repo.CreateWarehouse(ctx, repository.CreateWarehouseParams{
		TenantID: tenantID, Code: code, Name: name, Location: location,
	})
	if err != nil {
		return repository.ErpWarehouse{}, fmt.Errorf("create warehouse: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.warehouse.created", Resource: uuidStr(wh.ID), IP: ip,
	})
	return wh, nil
}

// GetStockLevels returns stock levels filtered by article/warehouse.
func (s *Stock) GetStockLevels(ctx context.Context, tenantID string, articleID, warehouseID pgtype.UUID) ([]repository.GetStockLevelsRow, error) {
	return s.repo.GetStockLevels(ctx, repository.GetStockLevelsParams{
		TenantID: tenantID, ArticleFilter: articleID, WarehouseFilter: warehouseID,
	})
}

// ListMovements returns paginated stock movements.
func (s *Stock) ListMovements(ctx context.Context, tenantID string, articleID pgtype.UUID, limit, offset int) ([]repository.ListStockMovementsRow, error) {
	return s.repo.ListStockMovements(ctx, repository.ListStockMovementsParams{
		TenantID: tenantID, ArticleFilter: articleID,
		Limit: int32(limit), Offset: int32(offset),
	})
}

// CreateMovementRequest holds data for a stock movement.
type CreateMovementRequest struct {
	TenantID     string
	ArticleID    pgtype.UUID
	WarehouseID  pgtype.UUID
	MovementType string
	Quantity     string
	UnitCost     string
	RefType      *string
	RefID        pgtype.UUID
	ConceptID    pgtype.UUID
	UserID       string
	Notes        string
	IP           string
}

// CreateMovement registers a stock movement and updates stock levels in a single transaction.
func (s *Stock) CreateMovement(ctx context.Context, req CreateMovementRequest) (repository.ErpStockMovement, error) {
	if !validMovementTypes[req.MovementType] {
		return repository.ErpStockMovement{}, fmt.Errorf("invalid movement type: %s", req.MovementType)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.ErpStockMovement{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	var refType pgtype.Text
	if req.RefType != nil {
		refType = pgtype.Text{String: *req.RefType, Valid: true}
	}

	mov, err := qtx.CreateStockMovement(ctx, repository.CreateStockMovementParams{
		TenantID:      req.TenantID,
		ArticleID:     req.ArticleID,
		WarehouseID:   req.WarehouseID,
		MovementType:  req.MovementType,
		Quantity:      pgNumeric(req.Quantity),
		UnitCost:      pgNumeric(req.UnitCost),
		ReferenceType: refType,
		ReferenceID:   req.RefID,
		ConceptID:     req.ConceptID,
		UserID:        req.UserID,
		Notes:         req.Notes,
	})
	if err != nil {
		return repository.ErpStockMovement{}, fmt.Errorf("create movement: %w", err)
	}

	// Update stock level in same transaction
	delta := pgNumeric(req.Quantity)
	if req.MovementType == "out" {
		delta = pgNumericNeg(req.Quantity)
	}
	if err := qtx.UpsertStockLevel(ctx, repository.UpsertStockLevelParams{
		TenantID: req.TenantID, ArticleID: req.ArticleID,
		WarehouseID: req.WarehouseID, Quantity: delta,
	}); err != nil {
		return repository.ErpStockMovement{}, fmt.Errorf("upsert stock level: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return repository.ErpStockMovement{}, fmt.Errorf("commit movement: %w", err)
	}

	idStr := uuidStr(mov.ID)
	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.stock.movement", Resource: idStr,
		Details: map[string]any{"type": req.MovementType, "qty": req.Quantity}, IP: req.IP,
	})
	s.publisher.Broadcast(req.TenantID, "erp_stock", map[string]any{
		"action": "movement_created", "movement_id": idStr,
	})

	return mov, nil
}

// ListBOM returns BOM entries for an article.
func (s *Stock) ListBOM(ctx context.Context, tenantID string, parentID pgtype.UUID) ([]repository.ListBOMRow, error) {
	return s.repo.ListBOM(ctx, repository.ListBOMParams{TenantID: tenantID, ParentID: parentID})
}

// pgNumeric converts a string to pgtype.Numeric.
func pgNumeric(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if s == "" {
		s = "0"
	}
	_ = n.Scan(s)
	return n
}

func pgNumericNeg(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if s == "" {
		s = "0"
	}
	_ = n.Scan("-" + s)
	return n
}
