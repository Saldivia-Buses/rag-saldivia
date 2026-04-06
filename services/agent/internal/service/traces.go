package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var safeToken = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// TracePublisher publishes execution trace events to NATS for the Traces Service.
type TracePublisher struct {
	nc *nats.Conn
}

// NewTracePublisher creates a trace publisher. If nc is nil, publishing is a no-op.
func NewTracePublisher(nc *nats.Conn) *TracePublisher {
	return &TracePublisher{nc: nc}
}

// TraceStart publishes a trace start event.
func (p *TracePublisher) TraceStart(tenantSlug, sessionID, userID, query string) string {
	traceID := uuid.New().String()
	if p.nc == nil || !validateToken(tenantSlug) {
		return traceID
	}
	evt := map[string]string{
		"trace_id":   traceID,
		"tenant_id":  tenantSlug,
		"session_id": sessionID,
		"user_id":    userID,
		"query":      query,
	}
	p.publish(fmt.Sprintf("tenant.%s.traces.start", tenantSlug), evt)
	return traceID
}

// TraceEnd publishes a trace end event.
func (p *TracePublisher) TraceEnd(tenantSlug, traceID, status string, modelsUsed []string, durationMS, inputTokens, outputTokens, toolCallCount int, costUSD float64) {
	if p.nc == nil || !validateToken(tenantSlug) {
		return
	}
	evt := map[string]any{
		"trace_id":            traceID,
		"tenant_id":           tenantSlug,
		"status":              status,
		"models_used":         modelsUsed,
		"total_duration_ms":   durationMS,
		"total_input_tokens":  inputTokens,
		"total_output_tokens": outputTokens,
		"total_cost_usd":      costUSD,
		"tool_call_count":     toolCallCount,
	}
	p.publish(fmt.Sprintf("tenant.%s.traces.end", tenantSlug), evt)
}

// TraceEvent publishes a single trace event (llm_call, tool_call, error, etc.).
func (p *TracePublisher) TraceEvent(tenantSlug, traceID, eventType string, seq, durationMS int, data any) {
	if p.nc == nil || !validateToken(tenantSlug) {
		return
	}
	evt := map[string]any{
		"trace_id":    traceID,
		"tenant_id":   tenantSlug,
		"seq":         seq,
		"event_type":  eventType,
		"data":        data,
		"duration_ms": durationMS,
	}
	p.publish(fmt.Sprintf("tenant.%s.traces.event", tenantSlug), evt)
}

func (p *TracePublisher) publish(subject string, evt any) {
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Error("marshal trace event", "error", err)
		return
	}
	if err := p.nc.Publish(subject, data); err != nil {
		slog.Error("publish trace event", "error", err, "subject", subject)
	}
}

// PublishFeedback sends a feedback event for the feedback service to consume.
// Categories: response_quality, agent_quality, performance, usage, error_report
func (p *TracePublisher) PublishFeedback(tenantSlug, category string, data any) {
	if p.nc == nil || !validateToken(tenantSlug) || !validateToken(category) {
		return
	}
	evt := map[string]any{
		"category":  category,
		"tenant_id": tenantSlug,
		"data":      data,
	}
	p.publish(fmt.Sprintf("tenant.%s.feedback.%s", tenantSlug, category), evt)
}

// validateToken checks if a string is safe for NATS subject interpolation.
func validateToken(s string) bool {
	return s != "" && safeToken.MatchString(s)
}
