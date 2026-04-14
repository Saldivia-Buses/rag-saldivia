package service

import (
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/traces"
)

// TracePublisher wraps pkg/traces.Publisher with backward-compatible method names.
type TracePublisher struct {
	*traces.Publisher
}

// NewTracePublisher creates a trace publisher. If nc is nil, publishing is a no-op.
func NewTracePublisher(nc *nats.Conn) *TracePublisher {
	return &TracePublisher{Publisher: traces.NewPublisher(nc)}
}

// TraceStart is the backward-compatible name for Publisher.Start.
func (t *TracePublisher) TraceStart(tenantSlug, sessionID, userID, query string) string {
	return t.Start(tenantSlug, sessionID, userID, query)
}

// TraceEnd is the backward-compatible name for Publisher.End.
func (t *TracePublisher) TraceEnd(tenantSlug, traceID, status string, modelsUsed []string, durationMS, inputTokens, outputTokens, toolCallCount int, costUSD float64) {
	t.End(tenantSlug, traceID, status, modelsUsed, durationMS, inputTokens, outputTokens, toolCallCount, costUSD)
}

// TraceEvent is the backward-compatible name for Publisher.Event.
func (t *TracePublisher) TraceEvent(tenantSlug, traceID, eventType string, seq, durationMS int, data any) {
	t.Event(tenantSlug, traceID, eventType, seq, durationMS, data)
}

// PublishFeedback is the backward-compatible name for Publisher.Feedback.
func (t *TracePublisher) PublishFeedback(tenantSlug, category string, data any) {
	t.Feedback(tenantSlug, category, data)
}
