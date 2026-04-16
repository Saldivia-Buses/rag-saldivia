package traces

import (
	"testing"
)

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		// Valid tokens
		{name: "simple slug", token: "tenant-a", want: true},
		{name: "alphanumeric", token: "abc123", want: true},
		{name: "with underscore", token: "my_tenant", want: true},
		{name: "with hyphen", token: "my-tenant", want: true},
		{name: "single char", token: "a", want: true},
		{name: "uppercase", token: "TenantA", want: true},
		{name: "mixed case with numbers", token: "Tenant123_test-slug", want: true},

		// Invalid tokens
		{name: "empty string", token: "", want: false},
		{name: "contains dot", token: "tenant.a", want: false},
		{name: "contains space", token: "tenant a", want: false},
		{name: "contains star", token: "tenant*", want: false},
		{name: "contains greater than", token: "tenant>a", want: false},
		{name: "contains slash", token: "tenant/a", want: false},
		{name: "nats wildcard star", token: "*", want: false},
		{name: "nats wildcard gt", token: ">", want: false},
		{name: "contains newline", token: "tenant\na", want: false},
		{name: "contains tab", token: "tenant\ta", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateToken(tt.token); got != tt.want {
				t.Errorf("ValidateToken(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestNewPublisher_NilConn(t *testing.T) {
	p := NewPublisher(nil)
	if p == nil {
		t.Fatal("NewPublisher(nil) should return non-nil Publisher")
	}
}

func TestStart_NilConn_ReturnsTraceID(t *testing.T) {
	p := NewPublisher(nil)

	traceID := p.Start("tenant-a", "session-1", "user-1", "test query")
	if traceID == "" {
		t.Error("Start with nil conn should still return a non-empty trace ID")
	}

	// UUID format: 8-4-4-4-12
	if len(traceID) != 36 {
		t.Errorf("trace ID length = %d, want 36 (UUID format)", len(traceID))
	}
}

func TestStart_NilConn_UniqueIDs(t *testing.T) {
	p := NewPublisher(nil)

	id1 := p.Start("tenant-a", "s1", "u1", "q1")
	id2 := p.Start("tenant-a", "s1", "u1", "q1")

	if id1 == id2 {
		t.Error("Start should return unique trace IDs on each call")
	}
}

func TestStart_InvalidTenant_StillReturnsTraceID(t *testing.T) {
	p := NewPublisher(nil)

	// Invalid tenant slug — should not panic, should still return a trace ID
	traceID := p.Start("", "session-1", "user-1", "query")
	if traceID == "" {
		t.Error("Start with invalid tenant should still return a trace ID")
	}
}

func TestEnd_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Should not panic
	p.End("tenant-a", "trace-123", "success", []string{"model-a"}, 100, 500, 200, 3, 0.05)
}

func TestEvent_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Should not panic
	p.Event("tenant-a", "trace-123", "llm_call", 1, 50, map[string]string{"model": "test"})
}

func TestFeedback_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Should not panic
	p.Feedback("tenant-a", "thumbs_up", map[string]string{"comment": "good"})
}

func TestNotify_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Should not panic
	p.Notify("tenant-a", "alert", map[string]string{"msg": "test"})
}

func TestBroadcast_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Should not panic
	p.Broadcast("tenant-a", "chat_update", map[string]string{"session": "s1"})
}

func TestFeedback_InvalidCategory_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Invalid category with dots — should not panic
	p.Feedback("tenant-a", "bad.category", nil)
}

func TestNotify_InvalidEventType_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Invalid event type — should not panic
	p.Notify("tenant-a", "bad event type", nil)
}

func TestBroadcast_InvalidChannel_NilConn_NoPanic(t *testing.T) {
	p := NewPublisher(nil)
	// Invalid channel — should not panic
	p.Broadcast("tenant-a", "", nil)
}
