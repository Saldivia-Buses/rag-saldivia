package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type Workshop struct {
	repo      *repository.Queries
	auditLog  *audit.Writer
	publisher *traces.Publisher
}

func NewWorkshop(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Workshop {
	return &Workshop{repo: repo, auditLog: auditWriter, publisher: publisher}
}

// ─── Customer Vehicles ─────────────────────────────────────────────────────

func (s *Workshop) ListCustomerVehicles(ctx context.Context, tenantID, ownerFilter string, limit, offset int) ([]repository.ListCustomerVehiclesRow, error) {
	return s.repo.ListCustomerVehicles(ctx, repository.ListCustomerVehiclesParams{
		TenantID:    tenantID,
		OwnerFilter: ownerFilter,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
}

func (s *Workshop) GetCustomerVehicle(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpCustomerVehicle, error) {
	return s.repo.GetCustomerVehicle(ctx, repository.GetCustomerVehicleParams{ID: id, TenantID: tenantID})
}

func (s *Workshop) CreateCustomerVehicle(ctx context.Context, p repository.CreateCustomerVehicleParams, userID, ip string) (repository.ErpCustomerVehicle, error) {
	v, err := s.repo.CreateCustomerVehicle(ctx, p)
	if err != nil {
		return repository.ErpCustomerVehicle{}, fmt.Errorf("create customer vehicle: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.vehicle.created", Resource: uuidStr(v.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_workshop", map[string]any{"action": "vehicle_created", "vehicle_id": uuidStr(v.ID)})
	return v, nil
}

func (s *Workshop) UpdateCustomerVehicle(ctx context.Context, p repository.UpdateCustomerVehicleParams, userID, ip string) error {
	rows, err := s.repo.UpdateCustomerVehicle(ctx, p)
	if err != nil {
		return fmt.Errorf("update customer vehicle: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("vehicle not found")
	}
	s.auditLog.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.vehicle.updated", Resource: uuidStr(p.ID), IP: ip})
	return nil
}

// ─── Incident Types ────────────────────────────────────────────────────────

func (s *Workshop) ListIncidentTypes(ctx context.Context, tenantID string) ([]repository.ErpVehicleIncidentType, error) {
	return s.repo.ListVehicleIncidentTypes(ctx, tenantID)
}

func (s *Workshop) CreateIncidentType(ctx context.Context, tenantID, name, userID, ip string) (repository.ErpVehicleIncidentType, error) {
	t, err := s.repo.CreateVehicleIncidentType(ctx, repository.CreateVehicleIncidentTypeParams{TenantID: tenantID, Name: name})
	if err != nil {
		return repository.ErpVehicleIncidentType{}, fmt.Errorf("create incident type: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.incident_type.created", Resource: uuidStr(t.ID), IP: ip})
	return t, nil
}

// ─── Vehicle Incidents ─────────────────────────────────────────────────────

func (s *Workshop) ListVehicleIncidents(ctx context.Context, tenantID, vehicleFilter, statusFilter string, limit, offset int) ([]repository.ListVehicleIncidentsRow, error) {
	return s.repo.ListVehicleIncidents(ctx, repository.ListVehicleIncidentsParams{
		TenantID:      tenantID,
		VehicleFilter: vehicleFilter,
		StatusFilter:  statusFilter,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
}

func (s *Workshop) CreateVehicleIncident(ctx context.Context, p repository.CreateVehicleIncidentParams, userID, ip string) (repository.ErpVehicleIncident, error) {
	inc, err := s.repo.CreateVehicleIncident(ctx, p)
	if err != nil {
		return repository.ErpVehicleIncident{}, fmt.Errorf("create vehicle incident: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.incident.created", Resource: uuidStr(inc.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_workshop", map[string]any{"action": "incident_created", "incident_id": uuidStr(inc.ID)})
	return inc, nil
}

func (s *Workshop) ResolveVehicleIncident(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	rows, err := s.repo.ResolveVehicleIncident(ctx, repository.ResolveVehicleIncidentParams{ID: id, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("resolve vehicle incident: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("incident not found")
	}
	s.auditLog.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.incident.resolved", Resource: uuidStr(id), IP: ip})
	return nil
}

// ─── KPIs ──────────────────────────────────────────────────────────────────

func (s *Workshop) GetKPIs(ctx context.Context, tenantID string) (repository.GetWorkshopKPIsRow, error) {
	return s.repo.GetWorkshopKPIs(ctx, tenantID)
}
