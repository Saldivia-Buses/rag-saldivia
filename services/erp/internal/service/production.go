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

type Production struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     *audit.Writer
	publisher *traces.Publisher
}

func NewProduction(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Production {
	return &Production{repo: repo, pool: pool, audit: auditWriter, publisher: publisher}
}

func (s *Production) ListCenters(ctx context.Context, tenantID string) ([]repository.ErpProductionCenter, error) {
	return s.repo.ListProductionCenters(ctx, tenantID)
}

func (s *Production) CreateCenter(ctx context.Context, tenantID, code, name, userID, ip string) (repository.ErpProductionCenter, error) {
	c, err := s.repo.CreateProductionCenter(ctx, repository.CreateProductionCenterParams{
		TenantID: tenantID, Code: code, Name: name,
	})
	if err != nil {
		return repository.ErpProductionCenter{}, fmt.Errorf("create center: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.production_center.created", Resource: uuidStr(c.ID), IP: ip,
	})
	return c, nil
}

func (s *Production) ListOrders(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListProductionOrdersRow, error) {
	return s.repo.ListProductionOrders(ctx, repository.ListProductionOrdersParams{
		TenantID: tenantID, StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

type ProductionOrderDetail struct {
	Order       repository.ErpProductionOrder            `json:"order"`
	Materials   []repository.ListProductionMaterialsRow  `json:"materials"`
	Steps       []repository.ErpProductionStep           `json:"steps"`
	Inspections []repository.ErpProductionInspection     `json:"inspections"`
}

func (s *Production) GetOrder(ctx context.Context, id pgtype.UUID, tenantID string) (*ProductionOrderDetail, error) {
	order, err := s.repo.GetProductionOrder(ctx, repository.GetProductionOrderParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	materials, err := s.repo.ListProductionMaterials(ctx, repository.ListProductionMaterialsParams{
		OrderID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list materials: %w", err)
	}
	steps, err := s.repo.ListProductionSteps(ctx, repository.ListProductionStepsParams{
		OrderID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list steps: %w", err)
	}
	return &ProductionOrderDetail{Order: order, Materials: materials, Steps: steps}, nil
}

type CreateProductionOrderRequest struct {
	TenantID  string
	Number    string
	Date      pgtype.Date
	ProductID pgtype.UUID
	CenterID  pgtype.UUID
	Quantity  string
	Priority  int32
	OrderID   pgtype.UUID
	StartDate pgtype.Date
	EndDate   pgtype.Date
	UserID    string
	Notes     string
	IP        string
}

func (s *Production) CreateOrder(ctx context.Context, req CreateProductionOrderRequest) (repository.ErpProductionOrder, error) {
	if req.Number == "" {
		return repository.ErpProductionOrder{}, fmt.Errorf("number is required")
	}
	o, err := s.repo.CreateProductionOrder(ctx, repository.CreateProductionOrderParams{
		TenantID: req.TenantID, Number: req.Number, Date: req.Date,
		ProductID: req.ProductID, CenterID: req.CenterID,
		Quantity: pgNumeric(req.Quantity), Priority: req.Priority,
		OrderID: req.OrderID, StartDate: req.StartDate, EndDate: req.EndDate,
		UserID: req.UserID, Notes: req.Notes,
	})
	if err != nil {
		return repository.ErpProductionOrder{}, fmt.Errorf("create order: %w", err)
	}
	idStr := uuidStr(o.ID)
	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.production_order.created", Resource: idStr, IP: req.IP,
	})
	s.publisher.Broadcast(req.TenantID, "erp_production", map[string]any{
		"action": "order_created", "order_id": idStr,
	})
	slog.Info("production order created", "id", idStr, "number", req.Number)
	return o, nil
}

func (s *Production) StartOrder(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	rows, err := s.repo.StartProductionOrder(ctx, repository.StartProductionOrderParams{
		ID: id, TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("start order: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("order not found")
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.production_order.started", Resource: uuidStr(id), IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_production", map[string]any{
		"action": "order_started", "order_id": uuidStr(id),
	})
	return nil
}

func (s *Production) UpdateStep(ctx context.Context, id pgtype.UUID, tenantID, status, notes, userID, ip string) error {
	rows, err := s.repo.UpdateProductionStep(ctx, repository.UpdateProductionStepParams{
		ID: id, TenantID: tenantID, Status: status, Column4: notes,
	})
	if err != nil {
		return fmt.Errorf("update step: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("step not found")
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.production_step.updated", Resource: uuidStr(id),
		Details: map[string]any{"status": status}, IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_production", map[string]any{
		"action": "step_updated", "step_id": uuidStr(id),
	})
	return nil
}

func (s *Production) CreateInspection(ctx context.Context, p repository.CreateProductionInspectionParams, userID, ip string) (repository.ErpProductionInspection, error) {
	insp, err := s.repo.CreateProductionInspection(ctx, p)
	if err != nil {
		return repository.ErpProductionInspection{}, fmt.Errorf("create inspection: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.production_inspection.created", Resource: uuidStr(insp.ID), IP: ip,
	})
	return insp, nil
}

func (s *Production) ListUnits(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListUnitsRow, error) {
	return s.repo.ListUnits(ctx, repository.ListUnitsParams{
		TenantID: tenantID, StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Production) GetUnit(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpUnit, error) {
	return s.repo.GetUnit(ctx, repository.GetUnitParams{ID: id, TenantID: tenantID})
}

func (s *Production) CreateUnit(ctx context.Context, p repository.CreateUnitParams, userID, ip string) (repository.ErpUnit, error) {
	u, err := s.repo.CreateUnit(ctx, p)
	if err != nil {
		return repository.ErpUnit{}, fmt.Errorf("create unit: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.unit.created", Resource: uuidStr(u.ID), IP: ip,
	})
	s.publisher.Broadcast(p.TenantID, "erp_production", map[string]any{
		"action": "unit_created", "unit_id": uuidStr(u.ID),
	})
	return u, nil
}
