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

// Suggestions handles suggestion business logic.
type Suggestions struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

// NewSuggestions creates a suggestions service.
func NewSuggestions(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Suggestions {
	return &Suggestions{
		repo:      repo,
		audit:     auditWriter,
		publisher: publisher,
	}
}

// List returns paginated suggestions with response count.
func (s *Suggestions) List(ctx context.Context, tenantID string, limit, offset int) ([]repository.ListSuggestionsRow, error) {
	return s.repo.ListSuggestions(ctx, repository.ListSuggestionsParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
}

// Get returns a single suggestion with its responses.
func (s *Suggestions) Get(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpSuggestion, []repository.ErpSuggestionResponse, error) {
	suggestion, err := s.repo.GetSuggestion(ctx, repository.GetSuggestionParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return repository.ErpSuggestion{}, nil, fmt.Errorf("get suggestion: %w", err)
	}

	responses, err := s.repo.ListResponses(ctx, repository.ListResponsesParams{
		SuggestionID: id,
		TenantID:     tenantID,
	})
	if err != nil {
		return repository.ErpSuggestion{}, nil, fmt.Errorf("list responses: %w", err)
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
func (s *Suggestions) Create(ctx context.Context, req CreateRequest) (repository.ErpSuggestion, error) {
	if req.Body == "" {
		return repository.ErpSuggestion{}, fmt.Errorf("suggestion body is required")
	}

	suggestion, err := s.repo.CreateSuggestion(ctx, repository.CreateSuggestionParams{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Origin:   req.Origin,
		Body:     req.Body,
	})
	if err != nil {
		return repository.ErpSuggestion{}, fmt.Errorf("create suggestion: %w", err)
	}

	idStr := uuidStr(suggestion.ID)

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.suggestion.created",
		Resource: idStr,
		Details:  map[string]any{"origin": req.Origin},
		IP:       req.IP,
	})

	s.publisher.Notify(req.TenantID, "new_suggestion", map[string]any{
		"suggestion_id": idStr,
		"user_id":       req.UserID,
		"origin":        req.Origin,
		"preview":       truncate(req.Body, 100),
	})

	s.publisher.Broadcast(req.TenantID, "erp_suggestions", map[string]any{
		"action":        "created",
		"suggestion_id": idStr,
	})

	slog.Info("suggestion created", "id", idStr, "user", req.UserID)
	return suggestion, nil
}

// RespondRequest holds data for responding to a suggestion.
type RespondRequest struct {
	TenantID     string
	SuggestionID pgtype.UUID
	UserID       string
	Body         string
	IP           string
}

// Respond adds a response to a suggestion and marks it as read.
func (s *Suggestions) Respond(ctx context.Context, req RespondRequest) (repository.ErpSuggestionResponse, error) {
	if req.Body == "" {
		return repository.ErpSuggestionResponse{}, fmt.Errorf("response body is required")
	}

	// Verify suggestion exists
	_, err := s.repo.GetSuggestion(ctx, repository.GetSuggestionParams{
		ID:       req.SuggestionID,
		TenantID: req.TenantID,
	})
	if err != nil {
		return repository.ErpSuggestionResponse{}, fmt.Errorf("suggestion not found: %w", err)
	}

	response, err := s.repo.CreateResponse(ctx, repository.CreateResponseParams{
		TenantID:     req.TenantID,
		SuggestionID: req.SuggestionID,
		UserID:       req.UserID,
		Body:         req.Body,
	})
	if err != nil {
		return repository.ErpSuggestionResponse{}, fmt.Errorf("create response: %w", err)
	}

	// Auto-mark as read when responded
	_ = s.repo.MarkSuggestionRead(ctx, repository.MarkSuggestionReadParams{
		ID:       req.SuggestionID,
		TenantID: req.TenantID,
	})

	idStr := uuidStr(req.SuggestionID)
	respIDStr := uuidStr(response.ID)

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.suggestion.responded",
		Resource: idStr,
		Details:  map[string]any{"response_id": respIDStr},
		IP:       req.IP,
	})

	s.publisher.Broadcast(req.TenantID, "erp_suggestions", map[string]any{
		"action":        "responded",
		"suggestion_id": idStr,
		"response_id":   respIDStr,
	})

	return response, nil
}

// MarkRead marks a suggestion as read.
func (s *Suggestions) MarkRead(ctx context.Context, id pgtype.UUID, tenantID string) error {
	return s.repo.MarkSuggestionRead(ctx, repository.MarkSuggestionReadParams{
		ID:       id,
		TenantID: tenantID,
	})
}

// CountUnread returns the number of unread suggestions.
func (s *Suggestions) CountUnread(ctx context.Context, tenantID string) (int32, error) {
	return s.repo.CountUnread(ctx, tenantID)
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

// uuidStr converts a pgtype.UUID to string. Returns "" if not valid.
func uuidStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
