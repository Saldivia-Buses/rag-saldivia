// Package service implements the Traces Service business logic.
// Receives trace events via NATS, persists to Platform DB, exposes cost/query APIs.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Traces manages execution trace persistence and querying.
type Traces struct {
	pool *pgxpool.Pool
}

// New creates a Traces service.
func New(pool *pgxpool.Pool) *Traces {
	return &Traces{pool: pool}
}

// TraceStartEvent is published when a query begins.
type TraceStartEvent struct {
	TraceID   string `json:"trace_id"`
	TenantID  string `json:"tenant_id"`
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Query     string `json:"query"`
}

// TraceEndEvent is published when a query completes.
type TraceEndEvent struct {
	TraceID          string   `json:"trace_id"`
	TenantID         string   `json:"tenant_id"`
	Status           string   `json:"status"`
	ModelsUsed       []string `json:"models_used"`
	TotalDurationMS  int      `json:"total_duration_ms"`
	TotalInputTokens int      `json:"total_input_tokens"`
	TotalOutputTokens int     `json:"total_output_tokens"`
	TotalCostUSD     float64  `json:"total_cost_usd"`
	ToolCallCount    int      `json:"tool_call_count"`
	Error            string   `json:"error,omitempty"`
}

// TraceEvent is a single event within a trace (llm_call, tool_call, error, etc.).
type TraceEvent struct {
	TraceID    string          `json:"trace_id"`
	TenantID   string          `json:"tenant_id"`
	Seq        int             `json:"seq"`
	EventType  string          `json:"event_type"`
	Data       json.RawMessage `json:"data"`
	DurationMS int             `json:"duration_ms,omitempty"`
}

// RecordTraceStart creates a new execution_trace record.
func (t *Traces) RecordTraceStart(ctx context.Context, evt TraceStartEvent) error {
	_, err := t.pool.Exec(ctx,
		`INSERT INTO execution_traces (id, tenant_id, session_id, user_id, query, status)
		 VALUES ($1, $2, $3, $4, $5, 'running')
		 ON CONFLICT (id) DO NOTHING`,
		evt.TraceID, evt.TenantID, evt.SessionID, evt.UserID, evt.Query,
	)
	return err
}

// RecordTraceEnd updates an execution_trace with final stats.
func (t *Traces) RecordTraceEnd(ctx context.Context, evt TraceEndEvent) error {
	_, err := t.pool.Exec(ctx,
		`UPDATE execution_traces SET
			status = $1, models_used = $2, total_duration_ms = $3,
			total_input_tokens = $4, total_output_tokens = $5,
			total_cost_usd = $6, tool_call_count = $7, error = $8
		 WHERE id = $9`,
		evt.Status, evt.ModelsUsed, evt.TotalDurationMS,
		evt.TotalInputTokens, evt.TotalOutputTokens,
		evt.TotalCostUSD, evt.ToolCallCount, nilIfEmpty(evt.Error),
		evt.TraceID,
	)
	return err
}

// RecordEvent inserts a trace_event.
func (t *Traces) RecordEvent(ctx context.Context, evt TraceEvent) error {
	_, err := t.pool.Exec(ctx,
		`INSERT INTO trace_events (trace_id, seq, event_type, data, duration_ms)
		 VALUES ($1, $2, $3, $4, $5)`,
		evt.TraceID, evt.Seq, evt.EventType, evt.Data, evt.DurationMS,
	)
	return err
}

// Trace is a full execution trace with events.
type Trace struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	SessionID         string    `json:"session_id"`
	UserID            string    `json:"user_id"`
	Query             string    `json:"query"`
	Status            string    `json:"status"`
	ModelsUsed        []string  `json:"models_used"`
	TotalDurationMS   *int      `json:"total_duration_ms"`
	TotalInputTokens  int       `json:"total_input_tokens"`
	TotalOutputTokens int       `json:"total_output_tokens"`
	TotalCostUSD      float64   `json:"total_cost_usd"`
	ToolCallCount     int       `json:"tool_call_count"`
	Error             *string   `json:"error,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// ListTraces returns traces filtered by tenant, with pagination.
func (t *Traces) ListTraces(ctx context.Context, tenantID string, limit, offset int) ([]Trace, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := t.pool.Query(ctx,
		`SELECT id, tenant_id, session_id, user_id, query, status, models_used,
			total_duration_ms, total_input_tokens, total_output_tokens,
			total_cost_usd, tool_call_count, error, created_at
		 FROM execution_traces
		 WHERE tenant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traces []Trace
	for rows.Next() {
		var tr Trace
		if err := rows.Scan(
			&tr.ID, &tr.TenantID, &tr.SessionID, &tr.UserID, &tr.Query,
			&tr.Status, &tr.ModelsUsed, &tr.TotalDurationMS,
			&tr.TotalInputTokens, &tr.TotalOutputTokens,
			&tr.TotalCostUSD, &tr.ToolCallCount, &tr.Error, &tr.CreatedAt,
		); err != nil {
			return nil, err
		}
		traces = append(traces, tr)
	}
	return traces, rows.Err()
}

// GetTraceDetail returns a trace with all its events. Enforces tenant isolation.
func (t *Traces) GetTraceDetail(ctx context.Context, traceID, tenantID string) (*Trace, []TraceEvent, error) {
	var tr Trace
	err := t.pool.QueryRow(ctx,
		`SELECT id, tenant_id, session_id, user_id, query, status, models_used,
			total_duration_ms, total_input_tokens, total_output_tokens,
			total_cost_usd, tool_call_count, error, created_at
		 FROM execution_traces WHERE id = $1 AND tenant_id = $2`, traceID, tenantID,
	).Scan(
		&tr.ID, &tr.TenantID, &tr.SessionID, &tr.UserID, &tr.Query,
		&tr.Status, &tr.ModelsUsed, &tr.TotalDurationMS,
		&tr.TotalInputTokens, &tr.TotalOutputTokens,
		&tr.TotalCostUSD, &tr.ToolCallCount, &tr.Error, &tr.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("trace not found: %w", err)
	}

	rows, err := t.pool.Query(ctx,
		`SELECT trace_id, seq, event_type, data, duration_ms
		 FROM trace_events WHERE trace_id = $1 ORDER BY seq`, traceID,
	)
	if err != nil {
		return &tr, nil, err
	}
	defer rows.Close()

	var events []TraceEvent
	for rows.Next() {
		var evt TraceEvent
		if err := rows.Scan(&evt.TraceID, &evt.Seq, &evt.EventType, &evt.Data, &evt.DurationMS); err != nil {
			return &tr, nil, err
		}
		events = append(events, evt)
	}
	return &tr, events, rows.Err()
}

// CostSummary is the cost breakdown for a tenant.
type CostSummary struct {
	TenantID          string  `json:"tenant_id"`
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalQueries      int     `json:"total_queries"`
	TotalInputTokens  int     `json:"total_input_tokens"`
	TotalOutputTokens int     `json:"total_output_tokens"`
	AvgCostPerQuery   float64 `json:"avg_cost_per_query"`
}

// GetTenantCost returns cost summary for a tenant in a date range.
func (t *Traces) GetTenantCost(ctx context.Context, tenantID string, from, to time.Time) (*CostSummary, error) {
	var cs CostSummary
	cs.TenantID = tenantID

	err := t.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_cost_usd), 0),
			COUNT(*),
			COALESCE(SUM(total_input_tokens), 0),
			COALESCE(SUM(total_output_tokens), 0)
		 FROM execution_traces
		 WHERE tenant_id = $1 AND created_at >= $2 AND created_at < $3
		   AND status = 'completed'`,
		tenantID, from, to,
	).Scan(&cs.TotalCostUSD, &cs.TotalQueries, &cs.TotalInputTokens, &cs.TotalOutputTokens)
	if err != nil {
		return nil, err
	}

	if cs.TotalQueries > 0 {
		cs.AvgCostPerQuery = cs.TotalCostUSD / float64(cs.TotalQueries)
	}
	return &cs, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
