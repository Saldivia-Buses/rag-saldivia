package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type HR struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

func NewHR(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *HR {
	return &HR{repo: repo, audit: auditWriter, publisher: publisher}
}

func (s *HR) ListDepartments(ctx context.Context, tenantID string) ([]repository.ErpDepartment, error) {
	return s.repo.ListDepartments(ctx, tenantID)
}

func (s *HR) CreateDepartment(ctx context.Context, p repository.CreateDepartmentParams, userID, ip string) (repository.ErpDepartment, error) {
	d, err := s.repo.CreateDepartment(ctx, p)
	if err != nil {
		return repository.ErpDepartment{}, fmt.Errorf("create department: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.department.created", Resource: uuidStr(d.ID), IP: ip})
	return d, nil
}

func (s *HR) ListEmployees(ctx context.Context, tenantID string, limit, offset int) ([]repository.ListEmployeeDetailsRow, error) {
	return s.repo.ListEmployeeDetails(ctx, repository.ListEmployeeDetailsParams{
		TenantID: tenantID, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *HR) GetEmployee(ctx context.Context, entityID pgtype.UUID, tenantID string) (repository.ErpEmployeeDetail, error) {
	return s.repo.GetEmployeeDetail(ctx, repository.GetEmployeeDetailParams{EntityID: entityID, TenantID: tenantID})
}

func (s *HR) UpsertEmployee(ctx context.Context, p repository.UpsertEmployeeDetailParams, userID, ip string) (repository.ErpEmployeeDetail, error) {
	ed, err := s.repo.UpsertEmployeeDetail(ctx, p)
	if err != nil {
		return repository.ErpEmployeeDetail{}, fmt.Errorf("upsert employee: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.employee.upserted", Resource: uuidStr(ed.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_hr", map[string]any{"action": "employee_updated", "entity_id": uuidStr(ed.EntityID)})
	return ed, nil
}

func (s *HR) ListEvents(ctx context.Context, tenantID string, entityID pgtype.UUID, typeFilter string, limit, offset int) ([]repository.ErpHrEvent, error) {
	return s.repo.ListHREvents(ctx, repository.ListHREventsParams{
		TenantID: tenantID, EntityFilter: entityID, TypeFilter: typeFilter,
		Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *HR) CreateEvent(ctx context.Context, p repository.CreateHREventParams, ip string) (repository.ErpHrEvent, error) {
	ev, err := s.repo.CreateHREvent(ctx, p)
	if err != nil {
		return repository.ErpHrEvent{}, fmt.Errorf("create event: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: p.UserID, Action: "erp.hr_event.created", Resource: uuidStr(ev.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_hr", map[string]any{"action": "event_created", "event_id": uuidStr(ev.ID)})
	return ev, nil
}

func (s *HR) ListTraining(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpTraining, error) {
	return s.repo.ListTraining(ctx, repository.ListTrainingParams{TenantID: tenantID, Limit: int32(limit), Offset: int32(offset)})
}

func (s *HR) CreateTraining(ctx context.Context, p repository.CreateTrainingParams, userID, ip string) (repository.ErpTraining, error) {
	t, err := s.repo.CreateTraining(ctx, p)
	if err != nil {
		return repository.ErpTraining{}, fmt.Errorf("create training: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.training.created", Resource: uuidStr(t.ID), IP: ip})
	return t, nil
}

func (s *HR) ListAttendance(ctx context.Context, tenantID string, entityID pgtype.UUID, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ErpAttendance, error) {
	return s.repo.ListAttendance(ctx, repository.ListAttendanceParams{
		TenantID: tenantID, EntityFilter: entityID, DateFrom: dateFrom, DateTo: dateTo,
		Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *HR) CreateAttendance(ctx context.Context, p repository.CreateAttendanceParams, userID, ip string) (repository.ErpAttendance, error) {
	a, err := s.repo.CreateAttendance(ctx, p)
	if err != nil {
		return repository.ErpAttendance{}, fmt.Errorf("create attendance: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.attendance.created", Resource: uuidStr(a.ID), IP: ip})
	return a, nil
}
