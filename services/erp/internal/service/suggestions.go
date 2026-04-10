// Package service provides business logic for ERP modules.
package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Suggestions handles suggestion business logic.
type Suggestions struct {
	repo       *repository.Queries
	audit      *audit.Writer
	publisher  *traces.Publisher
	tenantSlug string
}

// NewSuggestions creates a suggestions service.
func NewSuggestions(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher, tenantSlug string) *Suggestions {
	return &Suggestions{
		repo:       repo,
		audit:      auditWriter,
		publisher:  publisher,
		tenantSlug: tenantSlug,
	}
}

// List returns paginated suggestions with response count.
func (s *Suggestions) List(ctx context.Context, tenantID string, limit, offset int) ([]repository.Suggestion, error) {
	return s.repo.ListSuggestions(ctx, tenantID, limit, offset)
}

// Get returns a single suggestion with its responses.
func (s *Suggestions) Get(ctx context.Context, id uuid.UUID, tenantID string) (*repository.Suggestion, []repository.SuggestionResponse, error) {
	suggestion, err := s.repo.GetSuggestion(ctx, id, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("get suggestion: %w", err)
	}

	responses, err := s.repo.ListResponses(ctx, id, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("list responses: %w", err)
	}

	return suggestion, responses, nil
}

// CreateRequest holds data for creating a suggestion.
type CreateRequest struct {
	TenantID string
	UserID   string
	Origin   string
	Body     string
	IP       string
}

// Create creates a new suggestion and publishes a NATS notification.
func (s *Suggestions) Create(ctx context.Context, req CreateRequest) (*repository.Suggestion, error) {
	if req.Body == "" {
		return nil, fmt.Errorf("suggestion body is required")
	}

	suggestion, err := s.repo.CreateSuggestion(ctx, req.TenantID, req.UserID, req.Origin, req.Body)
	if err != nil {
		return nil, fmt.Errorf("create suggestion: %w", err)
	}

	// Audit
	s.audit.Write(ctx, audit.Entry{
		TenantID: s.tenantSlug,
		UserID:   req.UserID,
		Action:   "erp.suggestion.created",
		Resource: suggestion.ID.String(),
		Details:  map[string]any{"origin": req.Origin},
		IP:       req.IP,
	})

	// Notify via NATS → notification service can alert admins
	s.publisher.Notify(s.tenantSlug, "new_suggestion", map[string]any{
		"suggestion_id": suggestion.ID.String(),
		"user_id":       req.UserID,
		"origin":        req.Origin,
		"preview":       truncate(req.Body, 100),
	})

	// Broadcast via NATS → WS Hub pushes to connected admins in real-time
	s.publisher.Broadcast(s.tenantSlug, "erp_suggestions", map[string]any{
		"action":        "created",
		"suggestion_id": suggestion.ID.String(),
	})

	slog.Info("suggestion created", "id", suggestion.ID, "user", req.UserID)
	return suggestion, nil
}

// RespondRequest holds data for responding to a suggestion.
type RespondRequest struct {
	TenantID     string
	SuggestionID uuid.UUID
	UserID       string
	Body         string
	IP           string
}

// Respond adds a response to a suggestion and marks it as read.
func (s *Suggestions) Respond(ctx context.Context, req RespondRequest) (*repository.SuggestionResponse, error) {
	if req.Body == "" {
		return nil, fmt.Errorf("response body is required")
	}

	// Verify suggestion exists
	_, err := s.repo.GetSuggestion(ctx, req.SuggestionID, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("suggestion not found: %w", err)
	}

	response, err := s.repo.CreateResponse(ctx, req.TenantID, req.SuggestionID, req.UserID, req.Body)
	if err != nil {
		return nil, fmt.Errorf("create response: %w", err)
	}

	// Auto-mark as read when responded
	if err := s.repo.MarkSuggestionRead(ctx, req.SuggestionID, req.TenantID); err != nil {
		slog.Error("mark read failed", "error", err, "suggestion_id", req.SuggestionID)
	}

	// Audit
	s.audit.Write(ctx, audit.Entry{
		TenantID: s.tenantSlug,
		UserID:   req.UserID,
		Action:   "erp.suggestion.responded",
		Resource: req.SuggestionID.String(),
		Details:  map[string]any{"response_id": response.ID.String()},
		IP:       req.IP,
	})

	// Broadcast for real-time update
	s.publisher.Broadcast(s.tenantSlug, "erp_suggestions", map[string]any{
		"action":        "responded",
		"suggestion_id": req.SuggestionID.String(),
		"response_id":   response.ID.String(),
	})

	return response, nil
}

// MarkRead marks a suggestion as read.
func (s *Suggestions) MarkRead(ctx context.Context, id uuid.UUID, tenantID string) error {
	return s.repo.MarkSuggestionRead(ctx, id, tenantID)
}

// CountUnread returns the number of unread suggestions.
func (s *Suggestions) CountUnread(ctx context.Context, tenantID string) (int, error) {
	return s.repo.CountUnread(ctx, tenantID)
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
