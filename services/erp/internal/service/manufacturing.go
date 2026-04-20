package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Manufacturing handles manufacturing pipeline business logic.
type Manufacturing struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     *audit.Writer
	publisher *traces.Publisher
}

// NewManufacturing creates a manufacturing service.
func NewManufacturing(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Manufacturing {
	return &Manufacturing{repo: repo, pool: pool, audit: auditWriter, publisher: publisher}
}

// ─── Chassis Brands ───────────────────────────────────────────────────────────

// ListChassisBrands returns all active chassis brands for a tenant.
func (s *Manufacturing) ListChassisBrands(ctx context.Context, tenantID string) ([]repository.ErpChassisBrand, error) {
	return s.repo.ListChassisBrands(ctx, tenantID)
}

// CreateChassisBrand creates a new chassis brand with audit and NATS broadcast.
func (s *Manufacturing) CreateChassisBrand(ctx context.Context, tenantID, code, name, userID, ip string) (repository.ErpChassisBrand, error) {
	brand, err := s.repo.CreateChassisBrand(ctx, repository.CreateChassisBrandParams{
		TenantID: tenantID,
		Code:     code,
		Name:     name,
	})
	if err != nil {
		return repository.ErpChassisBrand{}, fmt.Errorf("create chassis brand: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.manufacturing.chassis_brand_created", Resource: uuidStr(brand.ID), IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_manufacturing", map[string]any{
		"action": "chassis_brand_created", "brand_id": uuidStr(brand.ID),
	})
	return brand, nil
}

// ─── Chassis Models ────────────────────────────────────────────────────────────

// ListChassisModels returns active chassis models, optionally filtered by brand UUID string.
func (s *Manufacturing) ListChassisModels(ctx context.Context, tenantID, brandFilter string) ([]repository.ListChassisModelsRow, error) {
	return s.repo.ListChassisModels(ctx, repository.ListChassisModelsParams{
		TenantID:    tenantID,
		BrandFilter: brandFilter,
	})
}

// GetChassisModel returns a chassis model by ID, scoped to tenant.
func (s *Manufacturing) GetChassisModel(ctx context.Context, id pgtype.UUID, tenantID string) (repository.GetChassisModelRow, error) {
	return s.repo.GetChassisModel(ctx, repository.GetChassisModelParams{ID: id, TenantID: tenantID})
}

// ─── Carroceria Models ─────────────────────────────────────────────────────────

// ListCarroceriaModels returns all active carroceria models for a tenant.
func (s *Manufacturing) ListCarroceriaModels(ctx context.Context, tenantID string) ([]repository.ErpCarroceriaModel, error) {
	return s.repo.ListCarroceriaModels(ctx, tenantID)
}

// GetCarroceriaModel returns a carroceria model by ID, scoped to tenant.
func (s *Manufacturing) GetCarroceriaModel(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpCarroceriaModel, error) {
	return s.repo.GetCarroceriaModel(ctx, repository.GetCarroceriaModelParams{ID: id, TenantID: tenantID})
}

// CreateCarroceriaModel creates a new carroceria model with audit and NATS broadcast.
func (s *Manufacturing) CreateCarroceriaModel(ctx context.Context, p repository.CreateCarroceriaModelParams, userID, ip string) (repository.ErpCarroceriaModel, error) {
	model, err := s.repo.CreateCarroceriaModel(ctx, p)
	if err != nil {
		return repository.ErpCarroceriaModel{}, fmt.Errorf("create carroceria model: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.manufacturing.carroceria_model_created", Resource: uuidStr(model.ID), IP: ip,
	})
	s.publisher.Broadcast(p.TenantID, "erp_manufacturing", map[string]any{
		"action": "carroceria_model_created", "model_id": uuidStr(model.ID),
	})
	return model, nil
}

// ─── Carroceria BOM ────────────────────────────────────────────────────────────

// GetCarroceriaBOM returns the bill of materials for a carroceria model.
func (s *Manufacturing) GetCarroceriaBOM(ctx context.Context, tenantID string, modelID pgtype.UUID) ([]repository.GetCarroceriaBOMRow, error) {
	return s.repo.GetCarroceriaBOM(ctx, repository.GetCarroceriaBOMParams{
		CarroceriaModelID: modelID,
		TenantID:          tenantID,
	})
}

// AddBOMItem adds a new item to a carroceria BOM with audit.
func (s *Manufacturing) AddBOMItem(ctx context.Context, p repository.AddBOMItemParams, userID, ip string) (repository.ErpCarroceriaBom, error) {
	item, err := s.repo.AddBOMItem(ctx, p)
	if err != nil {
		return repository.ErpCarroceriaBom{}, fmt.Errorf("add bom item: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.manufacturing.bom_item_added", Resource: uuidStr(item.ID), IP: ip,
	})
	return item, nil
}

// DeleteBOMItem removes a BOM item by ID, scoped to tenant, with audit.
func (s *Manufacturing) DeleteBOMItem(ctx context.Context, tenantID string, bomID pgtype.UUID, userID, ip string) error {
	rows, err := s.repo.DeleteBOMItem(ctx, repository.DeleteBOMItemParams{
		ID:       bomID,
		TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("delete bom item: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bom item not found")
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.manufacturing.bom_item_deleted", Resource: uuidStr(bomID), IP: ip,
	})
	return nil
}

// ─── Manufacturing Units ───────────────────────────────────────────────────────

// ListUnits returns paginated manufacturing units, optionally filtered by status.
func (s *Manufacturing) ListUnits(ctx context.Context, tenantID, statusFilter string, limit, offset int) ([]repository.ListManufacturingUnitsRow, error) {
	return s.repo.ListManufacturingUnits(ctx, repository.ListManufacturingUnitsParams{
		TenantID:     tenantID,
		StatusFilter: statusFilter,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
}

// GetUnit returns a single manufacturing unit by ID with joined names.
func (s *Manufacturing) GetUnit(ctx context.Context, id pgtype.UUID, tenantID string) (repository.GetManufacturingUnitRow, error) {
	return s.repo.GetManufacturingUnit(ctx, repository.GetManufacturingUnitParams{ID: id, TenantID: tenantID})
}

// CreateUnit registers a new manufacturing unit with audit and NATS broadcast.
func (s *Manufacturing) CreateUnit(ctx context.Context, p repository.CreateManufacturingUnitParams, userID, ip string) (repository.ErpManufacturingUnit, error) {
	unit, err := s.repo.CreateManufacturingUnit(ctx, p)
	if err != nil {
		return repository.ErpManufacturingUnit{}, fmt.Errorf("create manufacturing unit: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action:   "erp.manufacturing.unit_created",
		Resource: uuidStr(unit.ID),
		Details:  map[string]any{"work_order": unit.WorkOrderNumber, "chassis_serial": unit.ChassisSerial},
		IP:       ip,
	})
	s.publisher.Broadcast(p.TenantID, "erp_manufacturing", map[string]any{
		"action":  "unit_created",
		"unit_id": uuidStr(unit.ID),
	})
	return unit, nil
}

// UpdateManufacturingUnitStatus transitions a unit's status with audit and NATS broadcast.
func (s *Manufacturing) UpdateManufacturingUnitStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error {
	rows, err := s.repo.UpdateManufacturingUnitStatus(ctx, repository.UpdateManufacturingUnitStatusParams{
		ID:       id,
		TenantID: tenantID,
		Status:   status,
	})
	if err != nil {
		return fmt.Errorf("update manufacturing unit status: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("manufacturing unit not found")
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action:   "erp.manufacturing.unit_status_changed",
		Resource: uuidStr(id),
		Details:  map[string]any{"status": status},
		IP:       ip,
	})
	s.publisher.Broadcast(tenantID, "erp_manufacturing", map[string]any{
		"action":  "unit_status_changed",
		"unit_id": uuidStr(id),
		"status":  status,
	})
	return nil
}

// ─── Production Controls ──────────────────────────────────────────────────────

// ListProductionControls returns production controls for a unit.
// statusFilter and pagination params are accepted for API consistency but the
// underlying query filters only by tenantID + unitID; callers should rely on
// GetUnitPendingControls for status-based filtering.
func (s *Manufacturing) ListProductionControls(ctx context.Context, tenantID string, unitID pgtype.UUID, _ string, _, _ int) ([]repository.ListProductionControlsRow, error) {
	return s.repo.ListProductionControls(ctx, repository.ListProductionControlsParams{
		TenantID: tenantID,
		UnitID:   unitID,
	})
}

// CreateProductionControl registers a new production control for a unit with audit.
func (s *Manufacturing) CreateProductionControl(ctx context.Context, p repository.CreateProductionControlParams, userID, ip string) (repository.ErpProductionControl, error) {
	ctrl, err := s.repo.CreateProductionControl(ctx, p)
	if err != nil {
		return repository.ErpProductionControl{}, fmt.Errorf("create production control: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action:   "erp.manufacturing.control_created",
		Resource: uuidStr(ctrl.ID),
		Details:  map[string]any{"station": ctrl.Station},
		IP:       ip,
	})
	return ctrl, nil
}

// GetUnitControlExecutions returns all execution time entries for a unit's controls.
func (s *Manufacturing) GetUnitControlExecutions(ctx context.Context, tenantID string, unitID pgtype.UUID) ([]repository.GetUnitControlExecutionsRow, error) {
	return s.repo.GetUnitControlExecutions(ctx, repository.GetUnitControlExecutionsParams{
		UnitID:   unitID,
		TenantID: tenantID,
	})
}

// GetUnitPendingControls returns controls in pending/in_progress/blocked/rework states for a unit.
func (s *Manufacturing) GetUnitPendingControls(ctx context.Context, tenantID string, unitID pgtype.UUID) ([]repository.ErpProductionControl, error) {
	return s.repo.GetUnitPendingControls(ctx, repository.GetUnitPendingControlsParams{
		TenantID: tenantID,
		UnitID:   unitID,
	})
}

// ExecuteControl records a time entry for a control execution with audit and NATS broadcast.
func (s *Manufacturing) ExecuteControl(ctx context.Context, p repository.ExecuteControlParams, userID, ip string) (repository.ErpProductionControlExecution, error) {
	exec, err := s.repo.ExecuteControl(ctx, p)
	if err != nil {
		return repository.ErpProductionControlExecution{}, fmt.Errorf("execute control: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action:   "erp.manufacturing.control_executed",
		Resource: uuidStr(exec.ID),
		Details:  map[string]any{"control_id": uuidStr(exec.ControlID), "exec_type": exec.ExecType},
		IP:       ip,
	})
	s.publisher.Broadcast(p.TenantID, "erp_manufacturing_control", map[string]any{
		"action":     "control_executed",
		"exec_id":    uuidStr(exec.ID),
		"control_id": uuidStr(exec.ControlID),
	})
	return exec, nil
}

// ─── LCM (Material Consumption per Unit) ─────────────────────────────────────

// ListLCM returns LCM records, optionally filtered by unit and/or status.
func (s *Manufacturing) ListLCM(ctx context.Context, tenantID string, unitID pgtype.UUID, statusFilter string) ([]repository.ListLCMRow, error) {
	unitFilter := ""
	if unitID.Valid {
		unitFilter = uuidStr(unitID)
	}
	return s.repo.ListLCM(ctx, repository.ListLCMParams{
		TenantID:     tenantID,
		UnitFilter:   unitFilter,
		StatusFilter: statusFilter,
		Limit:        1000,
		Offset:       0,
	})
}

// CreateLCM creates a new material consumption list for a unit with audit and NATS broadcast.
func (s *Manufacturing) CreateLCM(ctx context.Context, p repository.CreateLCMParams, userID, ip string) (repository.ErpManufacturingLcm, error) {
	lcm, err := s.repo.CreateLCM(ctx, p)
	if err != nil {
		return repository.ErpManufacturingLcm{}, fmt.Errorf("create lcm: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.manufacturing.lcm_created", Resource: uuidStr(lcm.ID), IP: ip,
	})
	s.publisher.Broadcast(p.TenantID, "erp_certification", map[string]any{
		"action": "lcm_created", "lcm_id": uuidStr(lcm.ID),
	})
	return lcm, nil
}

// ─── CNRT Work Orders ─────────────────────────────────────────────────────────

// ListCNRTWork returns all CNRT inspection work orders for a unit.
func (s *Manufacturing) ListCNRTWork(ctx context.Context, tenantID string, unitID pgtype.UUID) ([]repository.ErpCnrtWorkOrder, error) {
	return s.repo.ListCNRTWork(ctx, repository.ListCNRTWorkParams{
		TenantID: tenantID,
		UnitID:   unitID,
	})
}

// CreateCNRTWork registers a new CNRT inspection work order with audit and NATS broadcast.
func (s *Manufacturing) CreateCNRTWork(ctx context.Context, p repository.CreateCNRTWorkParams, userID, ip string) (repository.ErpCnrtWorkOrder, error) {
	order, err := s.repo.CreateCNRTWork(ctx, p)
	if err != nil {
		return repository.ErpCnrtWorkOrder{}, fmt.Errorf("create cnrt work order: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action:   "erp.manufacturing.cnrt_created",
		Resource: uuidStr(order.ID),
		Details:  map[string]any{"cnrt_number": order.CnrtNumber},
		IP:       ip,
	})
	s.publisher.Broadcast(p.TenantID, "erp_certification", map[string]any{
		"action":   "cnrt_created",
		"order_id": uuidStr(order.ID),
	})
	return order, nil
}

// ─── Manufacturing Certificates ───────────────────────────────────────────────

// GetCertificate returns the most recent certificate for a unit.
func (s *Manufacturing) GetCertificate(ctx context.Context, tenantID string, unitID pgtype.UUID) (repository.ErpManufacturingCertificate, error) {
	return s.repo.GetCertificate(ctx, repository.GetCertificateParams{
		TenantID: tenantID,
		UnitID:   unitID,
	})
}

// CreateCertificate creates a new manufacturing certificate with audit and NATS broadcast.
func (s *Manufacturing) CreateCertificate(ctx context.Context, p repository.CreateCertificateParams, userID, ip string) (repository.ErpManufacturingCertificate, error) {
	cert, err := s.repo.CreateCertificate(ctx, p)
	if err != nil {
		return repository.ErpManufacturingCertificate{}, fmt.Errorf("create certificate: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action:   "erp.manufacturing.certificate_created",
		Resource: uuidStr(cert.ID),
		Details:  map[string]any{"cert_number": cert.CertificateNumber, "cert_type": cert.CertType},
		IP:       ip,
	})
	s.publisher.Broadcast(p.TenantID, "erp_certification", map[string]any{
		"action":  "certificate_created",
		"cert_id": uuidStr(cert.ID),
	})
	return cert, nil
}

// IssueCertificate transitions a certificate to issued state with audit and NATS broadcast.
func (s *Manufacturing) IssueCertificate(ctx context.Context, tenantID string, certID pgtype.UUID, userID, ip string) error {
	rows, err := s.repo.IssueCertificate(ctx, repository.IssueCertificateParams{
		ID:       certID,
		TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("issue certificate: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("certificate not found")
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.manufacturing.certificate_issued", Resource: uuidStr(certID), IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_certification", map[string]any{
		"action":  "certificate_issued",
		"cert_id": uuidStr(certID),
	})
	return nil
}
