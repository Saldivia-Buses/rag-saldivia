package spine_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"

	"github.com/Camionerou/rag-saldivia/pkg/spine"
)

func TestSubjectSlug(t *testing.T) {
	tests := []struct {
		subject string
		want    string
	}{
		{"tenant.saldivia.notify.chat.new_message", "saldivia"},
		{"tenant.foo.bar", "foo"},
		{"platform.lifecycle.tenant_created", ""},
		{"tenant.x", ""},
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.subject, func(t *testing.T) {
			if got := spine.SubjectSlug(tc.subject); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateTenantMatch(t *testing.T) {
	t.Run("matches", func(t *testing.T) {
		err := spine.ValidateTenantMatch("tenant.saldivia.notify.x",
			spine.Header{TenantID: "saldivia"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
	t.Run("mismatch", func(t *testing.T) {
		err := spine.ValidateTenantMatch("tenant.saldivia.notify.x",
			spine.Header{TenantID: "other"})
		if err == nil || !errors.Is(err, spine.ErrTenantMismatch) {
			t.Errorf("expected ErrTenantMismatch, got %v", err)
		}
	})
	t.Run("non-tenant subject is permissive", func(t *testing.T) {
		err := spine.ValidateTenantMatch("platform.lifecycle.tenant_created",
			spine.Header{TenantID: "saldivia"})
		if err != nil {
			t.Errorf("expected nil for platform-wide subject, got %v", err)
		}
	})
}

func TestBackoff(t *testing.T) {
	base := 1 * time.Second
	max := 16 * time.Second
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 16 * time.Second}, // capped
		{100, 16 * time.Second},
	}
	for _, tc := range tests {
		got := spine.Backoff(tc.attempt, base, max)
		if got != tc.want {
			t.Errorf("attempt=%d: got %v, want %v", tc.attempt, got, tc.want)
		}
	}
}

func TestExtractTraceContext_FromEnvelope(t *testing.T) {
	traceIDHex := "0af7651916cd43dd8448eb211c80319c"
	spanIDHex := "b7ad6b7169203331"
	header := spine.Header{TraceID: traceIDHex, SpanID: spanIDHex}

	ctx := spine.ExtractTraceContext(context.Background(), header, nats.Header{})

	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		t.Fatal("expected valid span context from envelope TraceID + SpanID")
	}
	if sc.TraceID().String() != traceIDHex {
		t.Errorf("got TraceID %s, want %s", sc.TraceID(), traceIDHex)
	}
	if sc.SpanID().String() != spanIDHex {
		t.Errorf("got SpanID %s, want %s", sc.SpanID(), spanIDHex)
	}
	if !sc.IsRemote() {
		t.Error("expected Remote span context")
	}
}

func TestExtractTraceContext_NoTrace(t *testing.T) {
	ctx := spine.ExtractTraceContext(context.Background(), spine.Header{}, nats.Header{})
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		t.Error("expected invalid span context when no trace info")
	}
}

func TestExtractTraceContext_BadHexFallsBackToHeaders(t *testing.T) {
	header := spine.Header{TraceID: "not-hex", SpanID: "also-not"}
	ctx := spine.ExtractTraceContext(context.Background(), header, nats.Header{})
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		t.Error("expected invalid span context when TraceID is malformed and no headers")
	}
}
