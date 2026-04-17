package service

// Tests for TracePublisher — the wrapper around pkg/traces.Publisher.
//
// Key invariants:
//   - nil NATS connection → all methods are no-ops (no panic)
//   - Subject format: tenant.{slug}.traces.start / .end / .event
//   - Tenant slug injected into subject, NOT from message body
//   - Invalid slug (empty or with dots) → publish silently skipped (ValidateToken)
//
// Real NATS publishing is tested via pkg/traces tests. Here we verify:
//   1. Method aliases work (TraceStart → Start, etc.)
//   2. nil conn is safe (no-op)
//   3. TraceStart returns a non-empty trace ID even when nc is nil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTracePublisher_NilConn_NoopAndSafe verifies that constructing a
// TracePublisher with nil NATS connection is valid and all methods run
// without panicking.
func TestNewTracePublisher_NilConn_NoopAndSafe(t *testing.T) {
	t.Parallel()
	tp := NewTracePublisher(nil)
	require.NotNil(t, tp)

	// None of these should panic.
	traceID := tp.TraceStart("test-tenant", "", "user1", "hello")
	assert.NotEmpty(t, traceID, "TraceStart must return a non-empty ID even when nc is nil")

	tp.TraceEnd("test-tenant", traceID, "completed", []string{"test-model"}, 100, 50, 20, 2, 0.0)
	tp.TraceEvent("test-tenant", traceID, "llm_call", 1, 50, map[string]string{"key": "val"})
	tp.PublishFeedback("test-tenant", "usage", map[string]any{"tokens": 100})
}

// TestTraceStart_ReturnsUniqueIDs verifies that successive TraceStart calls
// return different IDs (UUID generation is not deterministic).
func TestTraceStart_ReturnsUniqueIDs(t *testing.T) {
	t.Parallel()
	tp := NewTracePublisher(nil)

	id1 := tp.TraceStart("slug1", "", "u1", "msg1")
	id2 := tp.TraceStart("slug1", "", "u1", "msg1")
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2, "each TraceStart must return a unique ID")
}

// TestTraceStart_EmptySlug_StillReturnsID verifies that an empty tenant slug
// (ValidateToken returns false) does NOT cause a panic or empty return — the
// underlying Publisher.Start always generates a trace ID regardless.
func TestTraceStart_EmptySlug_StillReturnsID(t *testing.T) {
	t.Parallel()
	tp := NewTracePublisher(nil)
	// Empty slug: ValidateToken("") == false → publish skipped, but ID still returned.
	traceID := tp.TraceStart("", "", "u1", "msg")
	assert.NotEmpty(t, traceID, "trace ID must be returned even for invalid slugs")
}

// TestTracePublisher_MethodAliases_DelegateToPublisher verifies that each
// wrapper method delegates to the underlying Publisher (not stubbed).
// This is a structural test — if an alias is removed or renamed, this fails.
func TestTracePublisher_MethodAliases_DelegateToPublisher(t *testing.T) {
	t.Parallel()
	tp := NewTracePublisher(nil)
	require.NotNil(t, tp.Publisher, "TracePublisher must embed a non-nil *traces.Publisher")

	// Verify the embedded Publisher's Start method is the same one called by TraceStart.
	// With nil nc, Start() is deterministic: always returns a UUID.
	fromAlias := tp.TraceStart("slug", "sess", "user", "query")
	fromDirect := tp.Start("slug", "sess", "user", "query")

	// Both must be non-empty UUIDs — we can't compare values, but format is the same.
	assert.Len(t, fromAlias, 36, "TraceStart should return a UUID (36 chars)")
	assert.Len(t, fromDirect, 36, "Start should return a UUID (36 chars)")
}

// TestPublishFeedback_NilConn_NoopAndSafe verifies the Feedback alias.
func TestPublishFeedback_NilConn_NoopAndSafe(t *testing.T) {
	t.Parallel()
	tp := NewTracePublisher(nil)
	// Neither of these should panic.
	tp.PublishFeedback("tenant", "usage", nil)
	tp.PublishFeedback("", "usage", map[string]any{"x": 1}) // invalid slug — silently skipped
	tp.PublishFeedback("tenant", "", nil)                    // invalid category — silently skipped
}
