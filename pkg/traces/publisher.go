// Package traces provides a shared trace event publisher for NATS.
// Used by any service that publishes execution traces.
package traces

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var safeToken = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Publisher publishes execution trace events to NATS for the Traces Service.
type Publisher struct {
	nc *nats.Conn
}

// NewPublisher creates a trace publisher. If nc is nil, publishing is a no-op.
func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

// Start publishes a trace start event and returns a new trace ID.
func (p *Publisher) Start(tenantSlug, sessionID, userID, query string) string {
	traceID := uuid.New().String()
	if p.nc == nil || !ValidateToken(tenantSlug) {
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

// End publishes a trace end event.
func (p *Publisher) End(tenantSlug, traceID, status string, modelsUsed []string, durationMS, inputTokens, outputTokens, toolCallCount int, costUSD float64) {
	if p.nc == nil || !ValidateToken(tenantSlug) {
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

// Event publishes a single trace event (llm_call, tool_call, technique, error, etc.).
func (p *Publisher) Event(tenantSlug, traceID, eventType string, seq, durationMS int, data any) {
	if p.nc == nil || !ValidateToken(tenantSlug) {
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

// Feedback sends a feedback event for the feedback service to consume.
func (p *Publisher) Feedback(tenantSlug, category string, data any) {
	if p.nc == nil || !ValidateToken(tenantSlug) || !ValidateToken(category) {
		return
	}
	evt := map[string]any{
		"category":  category,
		"tenant_id": tenantSlug,
		"data":      data,
	}
	p.publish(fmt.Sprintf("tenant.%s.feedback.%s", tenantSlug, category), evt)
}

// Notify publishes to the notification subject for the notification service.
func (p *Publisher) Notify(tenantSlug, eventType string, data any) {
	if p.nc == nil || !ValidateToken(tenantSlug) || !ValidateToken(eventType) {
		return
	}
	evt := map[string]any{
		"type":      eventType,
		"tenant_id": tenantSlug,
		"data":      data,
	}
	p.publish(fmt.Sprintf("tenant.%s.notify.%s", tenantSlug, eventType), evt)
}

// Broadcast publishes a state-change event for the WS Hub.
func (p *Publisher) Broadcast(tenantSlug, channel string, data any) {
	if p.nc == nil || !ValidateToken(tenantSlug) || !ValidateToken(channel) {
		return
	}
	evt := map[string]any{
		"channel":   channel,
		"tenant_id": tenantSlug,
		"data":      data,
	}
	p.publish(fmt.Sprintf("tenant.%s.%s", tenantSlug, channel), evt)
}

func (p *Publisher) publish(subject string, evt any) {
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Error("marshal trace event", "error", err)
		return
	}
	if err := p.nc.Publish(subject, data); err != nil {
		slog.Error("publish trace event", "error", err, "subject", subject)
	}
}

// ValidateToken checks if a string is safe for NATS subject interpolation.
func ValidateToken(s string) bool {
	return s != "" && safeToken.MatchString(s)
}
