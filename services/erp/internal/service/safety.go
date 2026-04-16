package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type Safety struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

func NewSafety(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Safety {
	return &Safety{repo: repo, audit: auditWriter, publisher: publisher}
}

// ── Catalogs ──────────────────────────────────────────────────────────────────

func (s *Safety) ListAccidentTypes(ctx context.Context, tenantID string) ([]repository.ErpAccidentType, error) {
	return s.repo.ListAccidentTypes(ctx, tenantID)
}

func (s *Safety) CreateAccidentType(ctx context.Context, p repository.CreateAccidentTypeParams, userID, ip string) (repository.ErpAccidentType, error) {
	at, err := s.repo.CreateAccidentType(ctx, p)
	if err != nil {
		return repository.ErpAccidentType{}, fmt.Errorf("create accident type: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.safety.accident_type.created", Resource: uuidStr(at.ID), IP: ip})
	return at, nil
}

func (s *Safety) ListBodyParts(ctx context.Context, tenantID string) ([]repository.ErpBodyPart, error) {
	return s.repo.ListBodyParts(ctx, tenantID)
}

func (s *Safety) CreateBodyPart(ctx context.Context, p repository.CreateBodyPartParams, userID, ip string) (repository.ErpBodyPart, error) {
	bp, err := s.repo.CreateBodyPart(ctx, p)
	if err != nil {
		return repository.ErpBodyPart{}, fmt.Errorf("create body part: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.safety.body_part.created", Resource: uuidStr(bp.ID), IP: ip})
	return bp, nil
}

func (s *Safety) ListRiskAgents(ctx context.Context, tenantID, riskType string) ([]repository.ErpRiskAgent, error) {
	return s.repo.ListRiskAgents(ctx, repository.ListRiskAgentsParams{
		TenantID:       tenantID,
		RiskTypeFilter: riskType,
	})
}

func (s *Safety) CreateRiskAgent(ctx context.Context, p repository.CreateRiskAgentParams, userID, ip string) (repository.ErpRiskAgent, error) {
	ra, err := s.repo.CreateRiskAgent(ctx, p)
	if err != nil {
		return repository.ErpRiskAgent{}, fmt.Errorf("create risk agent: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.safety.risk_agent.created", Resource: uuidStr(ra.ID), IP: ip})
	return ra, nil
}

// ── Work Accidents ────────────────────────────────────────────────────────────

func (s *Safety) ListWorkAccidents(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListWorkAccidentsRow, error) {
	return s.repo.ListWorkAccidents(ctx, repository.ListWorkAccidentsParams{
		TenantID:     tenantID,
		StatusFilter: status,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
}

func (s *Safety) GetWorkAccident(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpWorkAccident, error) {
	return s.repo.GetWorkAccident(ctx, repository.GetWorkAccidentParams{ID: id, TenantID: tenantID})
}

func (s *Safety) CreateWorkAccident(ctx context.Context, p repository.CreateWorkAccidentParams, userID, ip string) (repository.ErpWorkAccident, error) {
	wa, err := s.repo.CreateWorkAccident(ctx, p)
	if err != nil {
		return repository.ErpWorkAccident{}, fmt.Errorf("create work accident: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.safety.accident.created", Resource: uuidStr(wa.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_safety", map[string]any{"action": "accident_created", "accident_id": uuidStr(wa.ID)})
	return wa, nil
}

func (s *Safety) UpdateAccidentStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error {
	rows, err := s.repo.UpdateAccidentStatus(ctx, repository.UpdateAccidentStatusParams{
		ID: id, TenantID: tenantID, Status: status,
	})
	if err != nil {
		return fmt.Errorf("update accident status: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("accident not found")
	}
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.safety.accident.status_changed", Resource: uuidStr(id), Details: map[string]any{"status": status}, IP: ip})
	return nil
}

// ── Risk Exposures ────────────────────────────────────────────────────────────

func (s *Safety) ListEmployeeRiskExposures(ctx context.Context, tenantID, entityID string) ([]repository.ListEmployeeRiskExposuresRow, error) {
	return s.repo.ListEmployeeRiskExposures(ctx, repository.ListEmployeeRiskExposuresParams{
		TenantID:     tenantID,
		EntityFilter: entityID,
	})
}

func (s *Safety) CreateRiskExposure(ctx context.Context, p repository.CreateRiskExposureParams, userID, ip string) (repository.ErpEmployeeRiskExposure, error) {
	exp, err := s.repo.CreateRiskExposure(ctx, p)
	if err != nil {
		return repository.ErpEmployeeRiskExposure{}, fmt.Errorf("create risk exposure: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.safety.risk_exposure.created", Resource: uuidStr(exp.ID), IP: ip})
	return exp, nil
}

// ── Medical Consultations ─────────────────────────────────────────────────────

func (s *Safety) ListMedicalConsultations(ctx context.Context, tenantID, dateFrom, dateTo string, limit, offset int) ([]repository.ErpMedicalConsultation, error) {
	return s.repo.ListMedicalConsultations(ctx, repository.ListMedicalConsultationsParams{
		TenantID: tenantID,
		DateFrom: pgDate(dateFrom),
		DateTo:   pgDate(dateTo),
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
}

func (s *Safety) CreateMedicalConsultation(ctx context.Context, p repository.CreateMedicalConsultationParams, userID, ip string) (repository.ErpMedicalConsultation, error) {
	mc, err := s.repo.CreateMedicalConsultation(ctx, p)
	if err != nil {
		return repository.ErpMedicalConsultation{}, fmt.Errorf("create medical consultation: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.safety.medical_consult.created", Resource: uuidStr(mc.ID), IP: ip})
	return mc, nil
}

// ── Medical Leaves ────────────────────────────────────────────────────────────

func (s *Safety) ListMedicalLeaves(ctx context.Context, tenantID, entityID, leaveType, status string, limit, offset int) ([]repository.ListMedicalLeavesRow, error) {
	return s.repo.ListMedicalLeaves(ctx, repository.ListMedicalLeavesParams{
		TenantID:       tenantID,
		EntityFilter:   entityID,
		LeaveTypeFilter: leaveType,
		StatusFilter:   status,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
}

func (s *Safety) CreateMedicalLeave(ctx context.Context, p repository.CreateMedicalLeaveParams, userID, ip string) (repository.ErpMedicalLeafe, error) {
	ml, err := s.repo.CreateMedicalLeave(ctx, p)
	if err != nil {
		return repository.ErpMedicalLeafe{}, fmt.Errorf("create medical leave: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.safety.medical_leave.created", Resource: uuidStr(ml.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_safety", map[string]any{"action": "medical_leave_created", "leave_id": uuidStr(ml.ID)})
	return ml, nil
}

func (s *Safety) ApproveMedicalLeave(ctx context.Context, id pgtype.UUID, tenantID, approvedBy, userID, ip string) error {
	rows, err := s.repo.ApproveMedicalLeave(ctx, repository.ApproveMedicalLeaveParams{
		ID: id, TenantID: tenantID, ApprovedBy: approvedBy,
	})
	if err != nil {
		return fmt.Errorf("approve medical leave: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("medical leave not found")
	}
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.safety.medical_leave.approved", Resource: uuidStr(id), IP: ip})
	return nil
}

// ── KPIs ──────────────────────────────────────────────────────────────────────

func (s *Safety) GetSafetyKPIs(ctx context.Context, tenantID, dateFrom, dateTo string) (repository.GetSafetyKPIsRow, error) {
	return s.repo.GetSafetyKPIs(ctx, repository.GetSafetyKPIsParams{
		TenantID: tenantID,
		DateFrom: pgDate(dateFrom),
		DateTo:   pgDate(dateTo),
	})
}

// pgDate converts a string to pgtype.Date for service-layer use.
func pgDate(s string) pgtype.Date {
	if s == "" {
		return pgtype.Date{}
	}
	var d pgtype.Date
	_ = d.Scan(s)
	return d
}
