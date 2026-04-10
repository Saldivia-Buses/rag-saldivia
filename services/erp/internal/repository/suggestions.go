// Package repository provides database access for ERP modules.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Suggestion represents a row in erp_suggestions.
type Suggestion struct {
	ID            uuid.UUID `json:"id"`
	TenantID      string    `json:"tenant_id"`
	UserID        string    `json:"user_id"`
	Origin        string    `json:"origin"`
	Body          string    `json:"body"`
	IsRead        bool      `json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ResponseCount int       `json:"response_count,omitempty"`
}

// SuggestionResponse represents a row in erp_suggestion_responses.
type SuggestionResponse struct {
	ID           uuid.UUID `json:"id"`
	TenantID     string    `json:"tenant_id"`
	SuggestionID uuid.UUID `json:"suggestion_id"`
	UserID       string    `json:"user_id"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
}

// Queries provides all suggestion-related database operations.
type Queries struct {
	db *pgxpool.Pool
}

// New creates a new Queries instance.
func New(db *pgxpool.Pool) *Queries {
	return &Queries{db: db}
}

// ListSuggestions returns paginated suggestions with response count.
func (q *Queries) ListSuggestions(ctx context.Context, tenantID string, limit, offset int) ([]Suggestion, error) {
	rows, err := q.db.Query(ctx,
		`SELECT s.id, s.tenant_id, s.user_id, s.origin, s.body, s.is_read, s.created_at, s.updated_at,
		        COUNT(r.id)::INT AS response_count
		 FROM erp_suggestions s
		 LEFT JOIN erp_suggestion_responses r ON r.suggestion_id = s.id AND r.tenant_id = s.tenant_id
		 WHERE s.tenant_id = $1
		 GROUP BY s.id
		 ORDER BY s.created_at DESC
		 LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []Suggestion
	for rows.Next() {
		var s Suggestion
		if err := rows.Scan(&s.ID, &s.TenantID, &s.UserID, &s.Origin, &s.Body,
			&s.IsRead, &s.CreatedAt, &s.UpdatedAt, &s.ResponseCount); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, s)
	}
	if suggestions == nil {
		suggestions = []Suggestion{}
	}
	return suggestions, nil
}

// GetSuggestion returns a single suggestion by ID.
func (q *Queries) GetSuggestion(ctx context.Context, id uuid.UUID, tenantID string) (*Suggestion, error) {
	var s Suggestion
	err := q.db.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, origin, body, is_read, created_at, updated_at
		 FROM erp_suggestions
		 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID).Scan(&s.ID, &s.TenantID, &s.UserID, &s.Origin, &s.Body,
		&s.IsRead, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CreateSuggestion inserts a new suggestion.
func (q *Queries) CreateSuggestion(ctx context.Context, tenantID, userID, origin, body string) (*Suggestion, error) {
	var s Suggestion
	err := q.db.QueryRow(ctx,
		`INSERT INTO erp_suggestions (tenant_id, user_id, origin, body)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, tenant_id, user_id, origin, body, is_read, created_at, updated_at`,
		tenantID, userID, origin, body).Scan(&s.ID, &s.TenantID, &s.UserID, &s.Origin, &s.Body,
		&s.IsRead, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// MarkSuggestionRead marks a suggestion as read.
func (q *Queries) MarkSuggestionRead(ctx context.Context, id uuid.UUID, tenantID string) error {
	_, err := q.db.Exec(ctx,
		`UPDATE erp_suggestions SET is_read = true, updated_at = now()
		 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID)
	return err
}

// ListResponses returns all responses for a suggestion.
func (q *Queries) ListResponses(ctx context.Context, suggestionID uuid.UUID, tenantID string) ([]SuggestionResponse, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, tenant_id, suggestion_id, user_id, body, created_at
		 FROM erp_suggestion_responses
		 WHERE suggestion_id = $1 AND tenant_id = $2
		 ORDER BY created_at ASC`,
		suggestionID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []SuggestionResponse
	for rows.Next() {
		var r SuggestionResponse
		if err := rows.Scan(&r.ID, &r.TenantID, &r.SuggestionID, &r.UserID, &r.Body, &r.CreatedAt); err != nil {
			return nil, err
		}
		responses = append(responses, r)
	}
	if responses == nil {
		responses = []SuggestionResponse{}
	}
	return responses, nil
}

// CreateResponse inserts a new response to a suggestion.
func (q *Queries) CreateResponse(ctx context.Context, tenantID string, suggestionID uuid.UUID, userID, body string) (*SuggestionResponse, error) {
	var r SuggestionResponse
	err := q.db.QueryRow(ctx,
		`INSERT INTO erp_suggestion_responses (tenant_id, suggestion_id, user_id, body)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, tenant_id, suggestion_id, user_id, body, created_at`,
		tenantID, suggestionID, userID, body).Scan(&r.ID, &r.TenantID, &r.SuggestionID, &r.UserID, &r.Body, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CountUnread returns the number of unread suggestions.
func (q *Queries) CountUnread(ctx context.Context, tenantID string) (int, error) {
	var count int
	err := q.db.QueryRow(ctx,
		`SELECT COUNT(*)::INT FROM erp_suggestions WHERE tenant_id = $1 AND is_read = false`,
		tenantID).Scan(&count)
	return count, err
}
