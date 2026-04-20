package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type Admin struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

func NewAdmin(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Admin {
	return &Admin{repo: repo, audit: auditWriter, publisher: publisher}
}

func (s *Admin) ListCommunications(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpCommunication, error) {
	return s.repo.ListCommunications(ctx, repository.ListCommunicationsParams{TenantID: tenantID, Limit: int32(limit), Offset: int32(offset)})
}

func (s *Admin) CreateCommunication(ctx context.Context, tenantID, subject, body, senderID, priority, ip string) (repository.ErpCommunication, error) {
	if subject == "" || body == "" {
		return repository.ErpCommunication{}, fmt.Errorf("subject and body are required")
	}
	validPriority := map[string]bool{"low": true, "normal": true, "high": true, "urgent": true}
	if !validPriority[priority] {
		priority = "normal"
	}
	c, err := s.repo.CreateCommunication(ctx, repository.CreateCommunicationParams{
		TenantID: tenantID, Subject: subject, Body: body, SenderID: senderID, Priority: priority,
	})
	if err != nil { return repository.ErpCommunication{}, fmt.Errorf("create communication: %w", err) }
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: senderID, Action: "erp.communication.created", Resource: uuidStr(c.ID), IP: ip})
	s.publisher.Broadcast(tenantID, "erp_admin", map[string]any{"action": "communication_created", "comm_id": uuidStr(c.ID)})
	return c, nil
}

func (s *Admin) ListCalendarEvents(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Timestamptz) ([]repository.ErpCalendarEvent, error) {
	return s.repo.ListCalendarEvents(ctx, repository.ListCalendarEventsParams{
		TenantID: tenantID, DateFrom: dateFrom, DateTo: dateTo,
	})
}

func (s *Admin) CreateCalendarEvent(ctx context.Context, p repository.CreateCalendarEventParams, ip string) (repository.ErpCalendarEvent, error) {
	if p.Title == "" {
		return repository.ErpCalendarEvent{}, fmt.Errorf("title is required")
	}
	ev, err := s.repo.CreateCalendarEvent(ctx, p)
	if err != nil { return repository.ErpCalendarEvent{}, fmt.Errorf("create event: %w", err) }
	s.audit.Write(ctx, audit.Entry{TenantID: p.TenantID, UserID: p.UserID, Action: "erp.calendar.created", Resource: uuidStr(ev.ID), IP: ip})
	return ev, nil
}

func (s *Admin) ListSurveys(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpSurvey, error) {
	return s.repo.ListSurveys(ctx, repository.ListSurveysParams{TenantID: tenantID, Limit: int32(limit), Offset: int32(offset)})
}

func (s *Admin) CreateSurvey(ctx context.Context, tenantID, title, description, userID, ip string) (repository.ErpSurvey, error) {
	if title == "" {
		return repository.ErpSurvey{}, fmt.Errorf("title is required")
	}
	sv, err := s.repo.CreateSurvey(ctx, repository.CreateSurveyParams{
		TenantID: tenantID, Title: title, Description: description, UserID: userID,
	})
	if err != nil { return repository.ErpSurvey{}, fmt.Errorf("create survey: %w", err) }
	s.audit.Write(ctx, audit.Entry{TenantID: tenantID, UserID: userID, Action: "erp.survey.created", Resource: uuidStr(sv.ID), IP: ip})
	return sv, nil
}

// ─── 2.0.17: read-only catalogs pending dedicated services ──────────────

func (s *Admin) ListProductSections(ctx context.Context, tenantID string) ([]repository.ErpProductSection, error) {
	return s.repo.ListProductSections(ctx, tenantID)
}

func (s *Admin) ListProducts(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpProduct, error) {
	return s.repo.ListProducts(ctx, repository.ListProductsParams{
		TenantID: tenantID, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Admin) ListProductAttributes(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpProductAttribute, error) {
	return s.repo.ListProductAttributes(ctx, repository.ListProductAttributesParams{
		TenantID: tenantID, ActiveOnly: activeOnly,
	})
}

func (s *Admin) ListTools(ctx context.Context, tenantID string, statusFilter int32, articleFilter string, limit, offset int) ([]repository.ErpTool, error) {
	return s.repo.ListTools(ctx, repository.ListToolsParams{
		TenantID:      tenantID,
		Limit:         int32(limit),
		Offset:        int32(offset),
		StatusFilter:  statusFilter,
		ArticleFilter: articleFilter,
	})
}

func (s *Admin) GetTool(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpTool, error) {
	return s.repo.GetTool(ctx, repository.GetToolParams{ID: id, TenantID: tenantID})
}
