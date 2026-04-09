package metrics

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/metric"
)

func TestCounters_Initialized(t *testing.T) {
	counters := map[string]metric.Int64Counter{
		"QueriesTotal":          QueriesTotal,
		"LLMTokensTotal":       LLMTokensTotal,
		"DocumentsIngestedTotal": DocumentsIngestedTotal,
		"ToolCallsTotal":       ToolCallsTotal,
		"AuthLoginsTotal":      AuthLoginsTotal,
		"NATSMessagesTotal":    NATSMessagesTotal,
	}

	for name, counter := range counters {
		if counter == nil {
			t.Errorf("%s is nil after init()", name)
		}
	}
}

func TestGauges_Initialized(t *testing.T) {
	if WSConnectionsActive == nil {
		t.Error("WSConnectionsActive is nil after init()")
	}
}

func TestHistograms_Initialized(t *testing.T) {
	if LLMRequestDuration == nil {
		t.Error("LLMRequestDuration is nil after init()")
	}
}

func TestCounters_AddDoesNotPanic(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		counter metric.Int64Counter
	}{
		{name: "QueriesTotal", counter: QueriesTotal},
		{name: "LLMTokensTotal", counter: LLMTokensTotal},
		{name: "DocumentsIngestedTotal", counter: DocumentsIngestedTotal},
		{name: "ToolCallsTotal", counter: ToolCallsTotal},
		{name: "AuthLoginsTotal", counter: AuthLoginsTotal},
		{name: "NATSMessagesTotal", counter: NATSMessagesTotal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.counter.Add(ctx, 1)
		})
	}
}

func TestWSConnectionsActive_AddDoesNotPanic(t *testing.T) {
	ctx := context.Background()
	// Increment and decrement — neither should panic
	WSConnectionsActive.Add(ctx, 1)
	WSConnectionsActive.Add(ctx, -1)
}

func TestLLMRequestDuration_RecordDoesNotPanic(t *testing.T) {
	ctx := context.Background()
	// Various durations — none should panic
	LLMRequestDuration.Record(ctx, 0.0)
	LLMRequestDuration.Record(ctx, 1.5)
	LLMRequestDuration.Record(ctx, 100.0)
}
