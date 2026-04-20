package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type Maintenance struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

func NewMaintenance(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Maintenance {
	return &Maintenance{repo: repo, audit: auditWriter, publisher: publisher}
}

func (s *Maintenance) ListAssets(ctx context.Context, tenantID, assetType string, activeOnly bool) ([]repository.ErpMaintenanceAsset, error) {
	return s.repo.ListMaintenanceAssets(ctx, repository.ListMaintenanceAssetsParams{
		TenantID: tenantID, ActiveOnly: activeOnly, TypeFilter: assetType,
	})
}

func (s *Maintenance) GetAsset(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpMaintenanceAsset, error) {
	return s.repo.GetMaintenanceAsset(ctx, repository.GetMaintenanceAssetParams{ID: id, TenantID: tenantID})
}

var validAssetTypes = map[string]bool{"vehicle": true, "machine": true, "tool": true, "facility": true}

func (s *Maintenance) CreateAsset(ctx context.Context, p repository.CreateMaintenanceAssetParams, userID, ip string) (repository.ErpMaintenanceAsset, error) {
	if p.Code == "" || p.Name == "" {
		return repository.ErpMaintenanceAsset{}, fmt.Errorf("code and name are required")
	}
	if !validAssetTypes[p.AssetType] {
		return repository.ErpMaintenanceAsset{}, fmt.Errorf("invalid asset_type (vehicle, machine, tool, facility)")
	}
	a, err := s.repo.CreateMaintenanceAsset(ctx, p)
	if err != nil { return repository.ErpMaintenanceAsset{}, fmt.Errorf("create asset: %w", err) }
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.maint_asset.created", Resource: uuidStr(a.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_maintenance", map[string]any{"action": "asset_created", "asset_id": uuidStr(a.ID)})
	return a, nil
}

func (s *Maintenance) ListPlans(ctx context.Context, tenantID string, assetID pgtype.UUID) ([]repository.ErpMaintenancePlan, error) {
	return s.repo.ListMaintenancePlans(ctx, repository.ListMaintenancePlansParams{TenantID: tenantID, AssetID: assetID})
}

func (s *Maintenance) CreatePlan(ctx context.Context, p repository.CreateMaintenancePlanParams, userID, ip string) (repository.ErpMaintenancePlan, error) {
	pl, err := s.repo.CreateMaintenancePlan(ctx, p)
	if err != nil { return repository.ErpMaintenancePlan{}, fmt.Errorf("create plan: %w", err) }
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.maint_plan.created", Resource: uuidStr(pl.ID), IP: ip})
	return pl, nil
}

func (s *Maintenance) ListWorkOrders(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListWorkOrdersRow, error) {
	return s.repo.ListWorkOrders(ctx, repository.ListWorkOrdersParams{
		TenantID: tenantID, StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

type WorkOrderDetail struct {
	Order repository.ErpWorkOrder            `json:"order"`
	Parts []repository.ListWorkOrderPartsRow `json:"parts"`
}

func (s *Maintenance) GetWorkOrder(ctx context.Context, id pgtype.UUID, tenantID string) (*WorkOrderDetail, error) {
	wo, err := s.repo.GetWorkOrder(ctx, repository.GetWorkOrderParams{ID: id, TenantID: tenantID})
	if err != nil { return nil, fmt.Errorf("get work order: %w", err) }
	parts, err := s.repo.ListWorkOrderParts(ctx, repository.ListWorkOrderPartsParams{WorkOrderID: id, TenantID: tenantID})
	if err != nil { return nil, fmt.Errorf("list parts: %w", err) }
	return &WorkOrderDetail{Order: wo, Parts: parts}, nil
}

func (s *Maintenance) CreateWorkOrder(ctx context.Context, p repository.CreateWorkOrderParams, ip string) (repository.ErpWorkOrder, error) {
	if p.Number == "" || p.Description == "" {
		return repository.ErpWorkOrder{}, fmt.Errorf("number and description are required")
	}
	validTypes := map[string]bool{"preventive": true, "corrective": true, "inspection": true}
	if !validTypes[p.WorkType] {
		return repository.ErpWorkOrder{}, fmt.Errorf("invalid work_type")
	}
	validPriorities := map[string]bool{"low": true, "normal": true, "high": true, "urgent": true}
	if !validPriorities[p.Priority] {
		return repository.ErpWorkOrder{}, fmt.Errorf("invalid priority")
	}
	wo, err := s.repo.CreateWorkOrder(ctx, p)
	if err != nil { return repository.ErpWorkOrder{}, fmt.Errorf("create work order: %w", err) }
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: p.UserID, Action: "erp.work_order.created", Resource: uuidStr(wo.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_maintenance", map[string]any{"action": "wo_created", "wo_id": uuidStr(wo.ID)})
	return wo, nil
}

func (s *Maintenance) UpdateWorkOrderStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error {
	validStatus := map[string]bool{"open": true, "in_progress": true, "completed": true, "cancelled": true}
	if !validStatus[status] {
		return fmt.Errorf("invalid status")
	}
	rows, err := s.repo.UpdateWorkOrderStatus(ctx, repository.UpdateWorkOrderStatusParams{ID: id, TenantID: tenantID, Status: status})
	if err != nil { return fmt.Errorf("update status: %w", err) }
	if rows == 0 { return fmt.Errorf("work order not found") }
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.work_order.status_changed", Resource: uuidStr(id), Details: map[string]any{"status": status}, IP: ip})
	s.publisher.Broadcast(tenantID, "erp_maintenance", map[string]any{"action": "wo_status_changed", "wo_id": uuidStr(id)})
	return nil
}

func (s *Maintenance) ListFuelLogs(ctx context.Context, tenantID string, assetID pgtype.UUID, limit, offset int) ([]repository.ListFuelLogsRow, error) {
	return s.repo.ListFuelLogs(ctx, repository.ListFuelLogsParams{
		TenantID: tenantID, AssetFilter: assetID, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Maintenance) CreateFuelLog(ctx context.Context, p repository.CreateFuelLogParams, ip string) (repository.ErpFuelLog, error) {
	fl, err := s.repo.CreateFuelLog(ctx, p)
	if err != nil { return repository.ErpFuelLog{}, fmt.Errorf("create fuel log: %w", err) }
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: p.UserID, Action: "erp.fuel_log.created", Resource: uuidStr(fl.ID), IP: ip})
	return fl, nil
}
