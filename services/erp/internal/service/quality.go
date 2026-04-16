package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type Quality struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

func NewQuality(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Quality {
	return &Quality{repo: repo, audit: auditWriter, publisher: publisher}
}

func (s *Quality) ListNC(ctx context.Context, tenantID, status, severity string, limit, offset int) ([]repository.ListNonconformitiesRow, error) {
	return s.repo.ListNonconformities(ctx, repository.ListNonconformitiesParams{
		TenantID: tenantID, StatusFilter: status, SeverityFilter: severity,
		Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Quality) GetNC(ctx context.Context, id pgtype.UUID, tenantID string) (repository.GetNonconformityRow, error) {
	return s.repo.GetNonconformity(ctx, repository.GetNonconformityParams{ID: id, TenantID: tenantID})
}

func (s *Quality) CreateNC(ctx context.Context, p repository.CreateNonconformityParams, ip string) (repository.CreateNonconformityRow, error) {
	nc, err := s.repo.CreateNonconformity(ctx, p)
	if err != nil {
		return repository.CreateNonconformityRow{}, fmt.Errorf("create NC: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: p.UserID, Action: "erp.nc.created", Resource: uuidStr(nc.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_quality", map[string]any{"action": "nc_created", "nc_id": uuidStr(nc.ID)})
	return nc, nil
}

func (s *Quality) UpdateNCStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error {
	rows, err := s.repo.UpdateNonconformityStatus(ctx, repository.UpdateNonconformityStatusParams{
		ID: id, TenantID: tenantID, Status: status,
	})
	if err != nil {
		return fmt.Errorf("update NC status: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("NC not found")
	}
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.nc.status_changed", Resource: uuidStr(id), Details: map[string]any{"status": status}, IP: ip})
	return nil
}

func (s *Quality) ListCA(ctx context.Context, ncID pgtype.UUID, tenantID string) ([]repository.ErpCorrectiveAction, error) {
	return s.repo.ListCorrectiveActions(ctx, repository.ListCorrectiveActionsParams{NcID: ncID, TenantID: tenantID})
}

func (s *Quality) CreateCA(ctx context.Context, p repository.CreateCorrectiveActionParams, userID, ip string) (repository.ErpCorrectiveAction, error) {
	ca, err := s.repo.CreateCorrectiveAction(ctx, p)
	if err != nil {
		return repository.ErpCorrectiveAction{}, fmt.Errorf("create CA: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.ca.created", Resource: uuidStr(ca.ID), IP: ip})
	return ca, nil
}

func (s *Quality) ListAudits(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpAudit, error) {
	return s.repo.ListAudits(ctx, repository.ListAuditsParams{TenantID: tenantID, Limit: int32(limit), Offset: int32(offset)})
}

func (s *Quality) CreateAudit(ctx context.Context, p repository.CreateAuditParams, userID, ip string) (repository.ErpAudit, error) {
	a, err := s.repo.CreateAudit(ctx, p)
	if err != nil {
		return repository.ErpAudit{}, fmt.Errorf("create audit: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.quality_audit.created", Resource: uuidStr(a.ID), IP: ip})
	return a, nil
}

func (s *Quality) ListAuditFindings(ctx context.Context, auditID pgtype.UUID, tenantID string) ([]repository.ErpAuditFinding, error) {
	return s.repo.ListAuditFindings(ctx, repository.ListAuditFindingsParams{AuditID: auditID, TenantID: tenantID})
}

func (s *Quality) CreateAuditFinding(ctx context.Context, p repository.CreateAuditFindingParams, userID, ip string) (repository.ErpAuditFinding, error) {
	f, err := s.repo.CreateAuditFinding(ctx, p)
	if err != nil {
		return repository.ErpAuditFinding{}, fmt.Errorf("create finding: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.audit_finding.created", Resource: uuidStr(f.ID), IP: ip})
	return f, nil
}

func (s *Quality) ListDocuments(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ErpControlledDocument, error) {
	return s.repo.ListControlledDocuments(ctx, repository.ListControlledDocumentsParams{
		TenantID: tenantID, StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Quality) CreateDocument(ctx context.Context, p repository.CreateControlledDocumentParams, userID, ip string) (repository.ErpControlledDocument, error) {
	d, err := s.repo.CreateControlledDocument(ctx, p)
	if err != nil {
		return repository.ErpControlledDocument{}, fmt.Errorf("create document: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.document.created", Resource: uuidStr(d.ID), IP: ip})
	return d, nil
}

func (s *Quality) ListNCOrigins(ctx context.Context, tenantID string) ([]repository.ErpNcOrigin, error) {
	return s.repo.ListNCOrigins(ctx, tenantID)
}

func (s *Quality) CreateNCOrigin(ctx context.Context, tenantID, name, userID, ip string) (repository.ErpNcOrigin, error) {
	o, err := s.repo.CreateNCOrigin(ctx, repository.CreateNCOriginParams{TenantID: tenantID, Name: name})
	if err != nil {
		return repository.ErpNcOrigin{}, fmt.Errorf("create nc origin: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.nc_origin.created", Resource: uuidStr(o.ID), IP: ip})
	return o, nil
}

func (s *Quality) ListActionPlans(ctx context.Context, tenantID, ncFilter, statusFilter string, limit, offset int) ([]repository.ListActionPlansRow, error) {
	return s.repo.ListActionPlans(ctx, repository.ListActionPlansParams{
		TenantID: tenantID, NcFilter: ncFilter, StatusFilter: statusFilter,
		Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Quality) GetActionPlan(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpQualityActionPlan, error) {
	p, err := s.repo.GetActionPlan(ctx, repository.GetActionPlanParams{ID: id, TenantID: tenantID})
	if err != nil {
		return repository.ErpQualityActionPlan{}, fmt.Errorf("get action plan: %w", err)
	}
	return p, nil
}

func (s *Quality) CreateActionPlan(ctx context.Context, p repository.CreateActionPlanParams, userID, ip string) (repository.ErpQualityActionPlan, error) {
	plan, err := s.repo.CreateActionPlan(ctx, p)
	if err != nil {
		return repository.ErpQualityActionPlan{}, fmt.Errorf("create action plan: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.action_plan.created", Resource: uuidStr(plan.ID), IP: ip})
	s.publisher.Broadcast(p.TenantID, "erp_quality", map[string]any{"action": "action_plan_created", "plan_id": uuidStr(plan.ID)})
	return plan, nil
}

func (s *Quality) UpdateActionPlanStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error {
	rows, err := s.repo.UpdateActionPlanStatus(ctx, repository.UpdateActionPlanStatusParams{ID: id, TenantID: tenantID, Status: status})
	if err != nil {
		return fmt.Errorf("update action plan status: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("action plan not found")
	}
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.action_plan.status_changed", Resource: uuidStr(id), Details: map[string]any{"status": status}, IP: ip})
	return nil
}

func (s *Quality) ListActionTasks(ctx context.Context, tenantID string, planID pgtype.UUID) ([]repository.ListActionTasksRow, error) {
	return s.repo.ListActionTasks(ctx, repository.ListActionTasksParams{TenantID: tenantID, PlanID: planID})
}

func (s *Quality) CreateActionTask(ctx context.Context, p repository.CreateActionTaskParams, userID, ip string) (repository.ErpQualityActionTask, error) {
	t, err := s.repo.CreateActionTask(ctx, p)
	if err != nil {
		return repository.ErpQualityActionTask{}, fmt.Errorf("create action task: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: userID, Action: "erp.action_task.created", Resource: uuidStr(t.ID), IP: ip})
	return t, nil
}

func (s *Quality) CompleteActionTask(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	rows, err := s.repo.CompleteActionTask(ctx, repository.CompleteActionTaskParams{ID: id, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("complete action task: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("action task not found")
	}
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.action_task.completed", Resource: uuidStr(id), IP: ip})
	return nil
}
