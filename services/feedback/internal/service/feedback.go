// Package service implements the feedback business logic.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
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
}

// NewFeedback creates a feedback service.
func NewFeedback(tenantDB, platformDB *pgxpool.Pool) *Feedback {
	return &Feedback{tenantDB: tenantDB, platformDB: platformDB}
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

	_, err := f.tenantDB.Exec(ctx,
		`INSERT INTO feedback_events (category, module, user_id, score, thumbs, severity, context, comment)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		evt.Category, evt.Module, nullIfEmpty(evt.UserID),
		evt.Score, nullIfEmpty(evt.Thumbs), nullIfEmpty(evt.Severity),
		ctxJSON, nullIfEmpty(evt.Comment),
	)
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
	rows, err := f.tenantDB.Query(ctx,
		`SELECT category, COUNT(*) FROM feedback_events
		 WHERE created_at > now() - make_interval(hours => $1)
		 GROUP BY category`,
		hours,
	)
	if err != nil {
		return nil, fmt.Errorf("count by category: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var cat string
		var count int
		if err := rows.Scan(&cat, &count); err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		counts[cat] = count
	}
	return counts, nil
}

// QualityMetrics returns positive/negative/total counts for AI quality feedback.
func (f *Feedback) QualityMetrics(ctx context.Context, hours int) (positive, negative, total int, avgScore float64, err error) {
	err = f.tenantDB.QueryRow(ctx,
		`SELECT
			COUNT(*) FILTER (WHERE thumbs = 'up') AS positive,
			COUNT(*) FILTER (WHERE thumbs = 'down') AS negative,
			COUNT(*) AS total,
			COALESCE(AVG(score), 0) AS avg_score
		 FROM feedback_events
		 WHERE category IN ('response_quality', 'agent_quality', 'extraction', 'detection')
		   AND created_at > now() - make_interval(hours => $1)`,
		hours,
	).Scan(&positive, &negative, &total, &avgScore)
	if err != nil {
		err = fmt.Errorf("quality metrics: %w", err)
	}
	return
}

// ErrorCounts returns error report counts by severity.
func (f *Feedback) ErrorCounts(ctx context.Context, hours int) (total, critical, open int, err error) {
	err = f.tenantDB.QueryRow(ctx,
		`SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE severity = 'critical') AS critical,
			COUNT(*) FILTER (WHERE status = 'open') AS open
		 FROM feedback_events
		 WHERE category = 'error_report'
		   AND created_at > now() - make_interval(hours => $1)`,
		hours,
	).Scan(&total, &critical, &open)
	if err != nil {
		err = fmt.Errorf("error counts: %w", err)
	}
	return
}

// PerformancePercentiles returns p50, p95, p99 latency from performance events.
func (f *Feedback) PerformancePercentiles(ctx context.Context, hours int) (p50, p95, p99 float64, err error) {
	err = f.tenantDB.QueryRow(ctx,
		`SELECT
			COALESCE(percentile_cont(0.50) WITHIN GROUP (ORDER BY (context->>'latency_ms')::numeric), 0),
			COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY (context->>'latency_ms')::numeric), 0),
			COALESCE(percentile_cont(0.99) WITHIN GROUP (ORDER BY (context->>'latency_ms')::numeric), 0)
		 FROM feedback_events
		 WHERE category = 'performance'
		   AND created_at > now() - make_interval(hours => $1)`,
		hours,
	).Scan(&p50, &p95, &p99)
	if err != nil {
		err = fmt.Errorf("performance percentiles: %w", err)
	}
	return
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
