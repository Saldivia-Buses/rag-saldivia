// Package service implements the feedback business logic.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/services/feedback/internal/repository"
)

// FeedbackEvent represents a single feedback event to be stored.
type FeedbackEvent struct {
	Category string          `json:"category"` // response_quality, error_report, nps, usage, etc.
	Module   string          `json:"module"`   // chat, agent, docai, auth, etc.
	UserID   string          `json:"user_id"`
	Score    *int            `json:"score,omitempty"`    // 1-5 for quality, 0-10 for NPS
	Thumbs   string          `json:"thumbs,omitempty"`   // up, down
	Severity string          `json:"severity,omitempty"` // info, warning, error, critical
	Context  json.RawMessage `json:"context"`
	Comment  string          `json:"comment,omitempty"`
}

// Feedback handles feedback operations for a single tenant.
type Feedback struct {
	tenantDB   *pgxpool.Pool
	platformDB *pgxpool.Pool
	repo       *repository.Queries
}

// NewFeedback creates a feedback service.
func NewFeedback(tenantDB, platformDB *pgxpool.Pool) *Feedback {
	return &Feedback{
		tenantDB:   tenantDB,
		platformDB: platformDB,
		repo:       repository.New(tenantDB),
	}
}

// Repo returns the underlying repository queries for direct use by other
// components (e.g. aggregator, handler) that need tenant DB access.
func (f *Feedback) Repo() *repository.Queries {
	return f.repo
}

// RecordEvent persists a feedback event in the tenant DB.
func (f *Feedback) RecordEvent(ctx context.Context, evt FeedbackEvent) error {
	if evt.Category == "" || evt.Module == "" {
		return fmt.Errorf("category and module are required")
	}

	ctxJSON := evt.Context
	if ctxJSON == nil {
		ctxJSON = json.RawMessage("{}")
	}

	err := f.repo.InsertFeedbackEvent(ctx, repository.InsertFeedbackEventParams{
		Category: evt.Category,
		Module:   evt.Module,
		UserID:   pgtext(evt.UserID),
		Score:    pgint(evt.Score),
		Thumbs:   pgtext(evt.Thumbs),
		Severity: pgtext(evt.Severity),
		Context:  ctxJSON,
		Comment:  pgtext(evt.Comment),
	})
	if err != nil {
		return fmt.Errorf("insert feedback event: %w", err)
	}

	slog.Debug("feedback event recorded",
		"category", evt.Category,
		"module", evt.Module,
		"user_id", evt.UserID,
	)
	return nil
}

// CountByCategory returns event counts by category for the given time window.
func (f *Feedback) CountByCategory(ctx context.Context, hours int) (map[string]int, error) {
	rows, err := f.repo.CountByCategory(ctx, int32(hours))
	if err != nil {
		return nil, fmt.Errorf("count by category: %w", err)
	}

	counts := make(map[string]int)
	for _, row := range rows {
		counts[row.Category] = int(row.Count)
	}
	return counts, nil
}

// QualityMetrics returns positive/negative/total counts for AI quality feedback.
func (f *Feedback) QualityMetrics(ctx context.Context, hours int) (positive, negative, total int, avgScore float64, err error) {
	row, err := f.repo.QualityMetrics(ctx, int32(hours))
	if err != nil {
		err = fmt.Errorf("quality metrics: %w", err)
		return
	}
	positive = int(row.Positive)
	negative = int(row.Negative)
	total = int(row.Total)
	avgScore = row.AvgScore
	return
}

// ErrorCounts returns error report counts by severity.
func (f *Feedback) ErrorCounts(ctx context.Context, hours int) (total, critical, open int, err error) {
	row, err := f.repo.ErrorCounts(ctx, int32(hours))
	if err != nil {
		err = fmt.Errorf("error counts: %w", err)
		return
	}
	total = int(row.Total)
	critical = int(row.Critical)
	open = int(row.Open)
	return
}

// PerformancePercentiles returns p50, p95, p99 latency from performance events.
func (f *Feedback) PerformancePercentiles(ctx context.Context, hours int) (p50, p95, p99 float64, err error) {
	row, err := f.repo.PerformancePercentiles(ctx, int32(hours))
	if err != nil {
		err = fmt.Errorf("performance percentiles: %w", err)
		return
	}
	p50 = row.P50
	p95 = row.P95
	p99 = row.P99
	return
}

// pgtext converts a string to pgtype.Text, treating empty strings as NULL.
func pgtext(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// pgint converts a *int to pgtype.Int4.
func pgint(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}
