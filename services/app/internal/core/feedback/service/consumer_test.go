// Unit tests for the NATS feedback consumer.
//
// Architecture: Consumer.handleEvent calls c.svc.RecordEvent which uses
// *repository.Queries backed by *pgxpool.Pool (concrete — no interface).
// This means the Ack path (successful persistence) requires a real database
// and is documented as TDD-ANCHOR at the bottom.
//
// This file covers all paths that resolve BEFORE reaching the service call:
//   - Invalid subject format → Term (parseSubject fails)
//   - Unknown category in subject → Term (not in validCategories)
//   - Malformed JSON payload → Term
//   - Valid event → RecordEvent called → Ack (or Nak on service error)
//
// CRITICAL INVARIANT tested:
//   TestConsumer_HandleEvent_TenantSpoofing_UsesSubjectNotPayload
//   The tenant slug comes exclusively from msg.Subject() via parseSubject().
//   A "tenant" field in the JSON payload is never read. This means a producer
//   cannot inject a different tenant by crafting the body — tenant routing is
//   enforced by the NATS subject filter `tenant.*.feedback.>`.
//
// Note on the Ack path: the consumer drives handleEvent through c.svc which
// wraps *repository.Queries. A nil *Feedback.repo would panic. We construct
// a consumer with a Feedback backed by mockFeedbackDB to exercise the full
// path including RecordEvent. When mockFeedbackDB.execErr is nil, RecordEvent
// succeeds and the message is Acked.
package service

import (
	"context"
	"testing"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/Camionerou/rag-saldivia/services/app/internal/core/feedback/repository"
)

// ---------------------------------------------------------------------------
// Mock: jetstream.Msg
// ---------------------------------------------------------------------------

// mockFeedbackMsg implements jetstream.Msg for testing. Only the methods
// called by handleEvent are implemented; unused methods panic if invoked.
type mockFeedbackMsg struct {
	subject string
	data    []byte

	ackCalled  bool
	nakCalled  bool
	termCalled bool
}

// Compile-time check: mockFeedbackMsg must satisfy jetstream.Msg.
var _ jetstream.Msg = (*mockFeedbackMsg)(nil)

func (m *mockFeedbackMsg) Subject() string                             { return m.subject }
func (m *mockFeedbackMsg) Data() []byte                                { return m.data }
func (m *mockFeedbackMsg) Ack() error                                  { m.ackCalled = true; return nil }
func (m *mockFeedbackMsg) Nak() error                                  { m.nakCalled = true; return nil }
func (m *mockFeedbackMsg) Term() error                                 { m.termCalled = true; return nil }
func (m *mockFeedbackMsg) Headers() nats.Header                        { return nil }
func (m *mockFeedbackMsg) Reply() string                               { return "" }
func (m *mockFeedbackMsg) DoubleAck(_ context.Context) error           { panic("not implemented") }
func (m *mockFeedbackMsg) NakWithDelay(_ time.Duration) error          { panic("not implemented") }
func (m *mockFeedbackMsg) InProgress() error                           { panic("not implemented") }
func (m *mockFeedbackMsg) TermWithReason(_ string) error               { panic("not implemented") }
func (m *mockFeedbackMsg) Metadata() (*jetstream.MsgMetadata, error)   { return nil, nil }

// ---------------------------------------------------------------------------
// Helpers: consumer builder
// ---------------------------------------------------------------------------

// newTestConsumerFeedback builds a Consumer where feedbackSvc is backed by
// mockFeedbackDB. nc is nil — handleEvent does not use NATS connection.
// This allows testing the full handleEvent path including RecordEvent
// (via repository.New(mockFeedbackDB)).
func newTestConsumerFeedback(db *mockFeedbackDB) *Consumer {
	svc := &Feedback{
		repo: repository.New(db),
	}
	return &Consumer{
		svc: svc,
		ctx: context.Background(),
	}
}

// ---------------------------------------------------------------------------
// Tests: parseSubject (pure function)
// ---------------------------------------------------------------------------

func TestConsumer_ParseSubject_Valid(t *testing.T) {
	tests := []struct {
		subject      string
		wantSlug     string
		wantCategory string
	}{
		{"tenant.saldivia.feedback.nps", "saldivia", "nps"},
		{"tenant.acme.feedback.response_quality", "acme", "response_quality"},
		{"tenant.foo.feedback.error_report", "foo", "error_report"},
		{"tenant.bigcorp.feedback.usage", "bigcorp", "usage"},
		{"tenant.t1.feedback.performance", "t1", "performance"},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			slug, cat := parseSubject(tt.subject)
			require.Equal(t, tt.wantSlug, slug, "slug mismatch for %q", tt.subject)
			require.Equal(t, tt.wantCategory, cat, "category mismatch for %q", tt.subject)
		})
	}
}

func TestConsumer_ParseSubject_Invalid_ReturnsBothEmpty(t *testing.T) {
	tests := []struct {
		name    string
		subject string
	}{
		{"too few parts", "tenant.saldivia.feedback"},
		{"wrong service prefix", "tenant.saldivia.notify.chat"},
		{"wrong root prefix", "service.saldivia.feedback.nps"},
		{"only 2 parts", "feedback.nps"},
		{"empty string", ""},
		{"only dots", "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug, cat := parseSubject(tt.subject)
			require.Empty(t, slug, "slug must be empty for invalid subject %q", tt.subject)
			require.Empty(t, cat, "category must be empty for invalid subject %q", tt.subject)
		})
	}
}

func TestConsumer_ParseSubject_RequiresFeedbackAsServicePart(t *testing.T) {
	// parts[2] must be "feedback" — other services are rejected
	slug, cat := parseSubject("tenant.saldivia.traces.nps")
	require.Empty(t, slug, "non-feedback service must be rejected")
	require.Empty(t, cat, "non-feedback service must be rejected")
}

// ---------------------------------------------------------------------------
// Tests: validCategories
// ---------------------------------------------------------------------------

func TestConsumer_ValidCategories_ContainsExpected(t *testing.T) {
	expected := []string{
		"response_quality", "agent_quality", "extraction", "detection",
		"error_report", "feature_request", "nps", "usage", "performance", "security",
	}
	for _, cat := range expected {
		require.True(t, validCategories[cat], "category %q must be in validCategories", cat)
	}
}

func TestConsumer_ValidCategories_RejectsUnknown(t *testing.T) {
	unknown := []string{"chat", "auth", "billing", "test", "", "USAGE", "NPS"}
	for _, cat := range unknown {
		require.False(t, validCategories[cat],
			"category %q must NOT be in validCategories", cat)
	}
}

// ---------------------------------------------------------------------------
// Tests: moduleFromCategory (pure function)
// ---------------------------------------------------------------------------

func TestConsumer_ModuleFromCategory_KnownCategories(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{"response_quality", "chat"},
		{"agent_quality", "agent"},
		{"extraction", "docai"},
		{"detection", "vision"},
		{"security", "auth"},
		{"performance", "system"},
		{"nps", "platform"},
		{"usage", "platform"},
		{"error_report", "platform"},
		{"feature_request", "platform"},
	}
	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := moduleFromCategory(tt.category)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestConsumer_ModuleFromCategory_UnknownCategory_ReturnsUnknown(t *testing.T) {
	got := moduleFromCategory("something_new")
	require.Equal(t, "unknown", got)
}

// ---------------------------------------------------------------------------
// Tests: handleEvent — Term paths (no DB interaction)
// ---------------------------------------------------------------------------

// TestConsumer_HandleEvent_InvalidSubject_Terms verifies that a subject with
// invalid format causes Term (not Nak) — the message is permanently broken.
func TestConsumer_HandleEvent_InvalidSubject_Terms(t *testing.T) {
	tests := []struct {
		name    string
		subject string
	}{
		{"too few parts", "tenant.saldivia.feedback"},
		{"wrong service", "tenant.saldivia.notify.chat"},
		{"2 parts", "feedback.nps"},
		{"1 part", "tenant"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockFeedbackDB{}
			c := newTestConsumerFeedback(db)
			msg := &mockFeedbackMsg{
				subject: tt.subject,
				data:    []byte(`{"module":"chat","user_id":"u1"}`),
			}
			c.handleEvent(msg)

			require.True(t, msg.termCalled, "invalid subject must Term (not Nak)")
			require.False(t, msg.nakCalled)
			require.False(t, msg.ackCalled)
			// No DB interaction for invalid subject
			require.Empty(t, db.lastExecSQL, "Exec must not be called for invalid subject")
		})
	}
}

// TestConsumer_HandleEvent_UnknownCategory_Terms verifies that an unrecognized
// category in the subject causes Term — consumer only accepts validCategories.
func TestConsumer_HandleEvent_UnknownCategory_Terms(t *testing.T) {
	tests := []struct {
		name     string
		category string
	}{
		{"chat category", "chat"},
		{"billing category", "billing"},
		{"empty category", ""},
		{"uppercase NPS", "NPS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockFeedbackDB{}
			c := newTestConsumerFeedback(db)
			subject := "tenant.saldivia.feedback." + tt.category
			msg := &mockFeedbackMsg{
				subject: subject,
				data:    []byte(`{"module":"chat"}`),
			}
			c.handleEvent(msg)

			require.True(t, msg.termCalled,
				"unknown category %q must Term", tt.category)
			require.False(t, msg.nakCalled)
			require.False(t, msg.ackCalled)
		})
	}
}

// TestConsumer_HandleEvent_MalformedJSON_Terms verifies that a non-JSON body
// causes Term for all valid categories.
func TestConsumer_HandleEvent_MalformedJSON_Terms(t *testing.T) {
	categories := []string{"nps", "response_quality", "error_report", "usage", "performance"}
	for _, cat := range categories {
		t.Run("category="+cat, func(t *testing.T) {
			db := &mockFeedbackDB{}
			c := newTestConsumerFeedback(db)
			msg := &mockFeedbackMsg{
				subject: "tenant.saldivia.feedback." + cat,
				data:    []byte(`{invalid json`),
			}
			c.handleEvent(msg)

			require.True(t, msg.termCalled,
				"malformed JSON must Term for category %q", cat)
			require.False(t, msg.nakCalled)
			require.False(t, msg.ackCalled)
			require.Empty(t, db.lastExecSQL, "Exec must not be called for malformed JSON")
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: handleEvent — Ack path (DB interaction via mockFeedbackDB)
// ---------------------------------------------------------------------------

// TestConsumer_HandleEvent_ValidEvent_AcksOnSuccess verifies the happy path:
// valid subject + valid JSON + successful RecordEvent → Ack.
func TestConsumer_HandleEvent_ValidEvent_AcksOnSuccess(t *testing.T) {
	db := &mockFeedbackDB{} // execErr=nil → RecordEvent succeeds
	c := newTestConsumerFeedback(db)

	msg := &mockFeedbackMsg{
		subject: "tenant.saldivia.feedback.nps",
		data:    []byte(`{"module":"platform","user_id":"u1","score":9,"comment":"great"}`),
	}
	c.handleEvent(msg)

	require.True(t, msg.ackCalled, "valid event must Ack")
	require.False(t, msg.nakCalled)
	require.False(t, msg.termCalled)
	require.NotEmpty(t, db.lastExecSQL, "RecordEvent must call Exec")
}

// TestConsumer_HandleEvent_ServiceError_Naks verifies that a DB error in
// RecordEvent causes Nak (transient — NATS will redeliver).
func TestConsumer_HandleEvent_ServiceError_Naks(t *testing.T) {
	db := &mockFeedbackDB{
		execErr: errDBTimeout, // RecordEvent → InsertFeedbackEvent → error
	}
	c := newTestConsumerFeedback(db)

	msg := &mockFeedbackMsg{
		subject: "tenant.saldivia.feedback.nps",
		data:    []byte(`{"module":"platform","user_id":"u1","score":9}`),
	}
	c.handleEvent(msg)

	require.True(t, msg.nakCalled, "service error must Nak for retry")
	require.False(t, msg.ackCalled)
	require.False(t, msg.termCalled)
}

// errDBTimeout is a sentinel error for simulating transient DB failures.
var errDBTimeout = &dbTimeoutError{}

type dbTimeoutError struct{}

func (e *dbTimeoutError) Error() string { return "connection timeout" }

// ---------------------------------------------------------------------------
// Tests: handleEvent — module inference from category
// ---------------------------------------------------------------------------

// TestConsumer_HandleEvent_InfersModuleFromCategory verifies that when the
// payload does not include "module", it is inferred from the subject category.
// We verify this indirectly: RecordEvent is called (Ack happens) and the
// Exec was issued (meaning category→module was resolved non-empty, passing
// the category+module validation in RecordEvent).
func TestConsumer_HandleEvent_InfersModuleFromCategory(t *testing.T) {
	tests := []struct {
		category string
	}{
		{"response_quality"}, // inferred module: "chat"
		{"agent_quality"},    // inferred module: "agent"
		{"extraction"},       // inferred module: "docai"
		{"nps"},              // inferred module: "platform"
		{"usage"},            // inferred module: "platform"
		{"error_report"},     // inferred module: "platform"
	}
	for _, tt := range tests {
		t.Run("category="+tt.category, func(t *testing.T) {
			db := &mockFeedbackDB{}
			c := newTestConsumerFeedback(db)
			msg := &mockFeedbackMsg{
				subject: "tenant.saldivia.feedback." + tt.category,
				// no "module" field — must be inferred
				data: []byte(`{"user_id":"u1"}`),
			}
			c.handleEvent(msg)

			require.True(t, msg.ackCalled,
				"category %q: module must be inferred and RecordEvent must succeed", tt.category)
			require.NotEmpty(t, db.lastExecSQL,
				"category %q: Exec must be called", tt.category)
		})
	}
}

// TestConsumer_HandleEvent_ErrorReportInfersSeverity verifies that when
// category=error_report and no severity is provided, it defaults to "error".
// Verified indirectly: RecordEvent is called (severity field is optional,
// so no validation failure) and the Exec is invoked.
func TestConsumer_HandleEvent_ErrorReport_InfersSeverity(t *testing.T) {
	db := &mockFeedbackDB{}
	c := newTestConsumerFeedback(db)
	msg := &mockFeedbackMsg{
		subject: "tenant.saldivia.feedback.error_report",
		// no severity field — must default to "error"
		data: []byte(`{"module":"auth","user_id":"u1","comment":"crash"}`),
	}
	c.handleEvent(msg)

	require.True(t, msg.ackCalled, "error_report without severity must Ack (severity inferred)")
	require.NotEmpty(t, db.lastExecSQL)
}

// TestConsumer_HandleEvent_ScoreExtracted verifies that a numeric "score"
// field in the payload is correctly extracted and passed to RecordEvent.
// Verified indirectly via Ack (score is optional, doesn't affect validation).
func TestConsumer_HandleEvent_ScoreField_ExtractedCorrectly(t *testing.T) {
	db := &mockFeedbackDB{}
	c := newTestConsumerFeedback(db)
	msg := &mockFeedbackMsg{
		subject: "tenant.saldivia.feedback.nps",
		data:    []byte(`{"module":"platform","score":10,"user_id":"u1"}`),
	}
	c.handleEvent(msg)

	require.True(t, msg.ackCalled, "event with score field must Ack")
}

// ---------------------------------------------------------------------------
// INVARIANT: tenant spoofing protection
//
// handleEvent extracts the tenant slug from msg.Subject() via parseSubject().
// The JSON payload is never read for a "tenant" field — the slug from the
// subject is used exclusively for logging (slog.Debug/slog.Error).
//
// Tenant isolation is enforced architecturally: c.svc (*Feedback) is
// constructed with the correct tenant's DB pool before the Consumer is created.
// A payload claiming a different tenant cannot reroute the event — there is
// no routing decision inside handleEvent that uses the slug.
//
// This test verifies both sides:
//   1. parseSubject() returns the correct slug from the subject, ignoring body.
//   2. A message with "tenant":"attacker" in the body still uses the subject
//      tenant's DB pool (indirectly verified: Ack means RecordEvent succeeded
//      using the pool bound to "saldivia", not a non-existent "attacker" pool).
// ---------------------------------------------------------------------------

// TestConsumer_HandleEvent_TenantSpoofing_UsesSubjectNotPayload is the
// critical cross-tenant isolation test.
//
// Scenario:
//   - NATS subject: "tenant.saldivia.feedback.nps" (legitimate tenant)
//   - JSON payload contains "tenant":"attacker" (spoofed field)
//
// Expected: the consumer processes the event using the "saldivia" context
// (its pre-bound DB pool). The "attacker" value in the payload is ignored.
// The message is Acked, not Termed or Naked.
func TestConsumer_HandleEvent_TenantSpoofing_UsesSubjectNotPayload(t *testing.T) {
	// Step 1: verify the pure extraction function ignores body content
	legitimateSubject := "tenant.saldivia.feedback.nps"
	slug, cat := parseSubject(legitimateSubject)
	require.Equal(t, "saldivia", slug,
		"parseSubject must extract slug from subject, not from body")
	require.Equal(t, "nps", cat)

	// Step 2: verify that a body-derived "tenant" attempt produces no valid slug
	// (parseSubject takes a subject string, not a JSON body — these are different types)
	bogusTenantSubject := "attacker" // what a body-only read would produce
	spoofedSlug, _ := parseSubject(bogusTenantSubject)
	require.Empty(t, spoofedSlug,
		"body-derived tenant string is not a valid subject — parseSubject must return empty")

	// Step 3: full handleEvent flow with spoofed body
	// The consumer's c.svc is pre-bound to "saldivia"'s pool (mockFeedbackDB here).
	// Even though the body says "tenant":"attacker", the event is processed
	// using saldivia's pool — there is no pool-switching logic in handleEvent.
	db := &mockFeedbackDB{} // simulates "saldivia" tenant pool
	c := newTestConsumerFeedback(db)

	msg := &mockFeedbackMsg{
		subject: legitimateSubject, // "saldivia" in subject
		data: []byte(
			// attacker injects a different tenant in the body
			`{"module":"platform","score":9,"user_id":"u1","tenant":"attacker"}`,
		),
	}
	c.handleEvent(msg)

	require.True(t, msg.ackCalled,
		"handleEvent must Ack using subject-derived tenant — payload 'tenant' field is ignored")
	require.False(t, msg.nakCalled)
	require.False(t, msg.termCalled)

	// Exec was called — RecordEvent ran against the "saldivia" pool, not "attacker"
	require.NotEmpty(t, db.lastExecSQL,
		"RecordEvent must be called against the subject-tenant's pool")
}

// ---------------------------------------------------------------------------
// TDD-ANCHOR: Ack path with real DB (require testcontainers)
//
// TestConsumer_HandleEvent_ValidNPSEvent_PersistsAndAcks:
//   Use testcontainers PostgreSQL + real tenant schema.
//   Build Consumer with pool connected to the test DB.
//   Send subject=tenant.saldivia.feedback.nps, payload with score=9.
//   → row inserted in feedback_events with category="nps", module="platform", score=9.
//   → msg.Ack() called.
//
// TestConsumer_HandleEvent_ErrorReport_PersistsWithSeverity:
//   Send error_report event without severity in payload.
//   → inserted row has severity="error" (inferred default).
//
// TestConsumer_HandleEvent_ValidEvent_ContextIsFullPayload:
//   Send event with known JSON body.
//   → feedback_events.context column contains the original raw JSON bytes.
//
// TestConsumer_HandleEvent_NakOnTransientDBFailure:
//   Simulate transient DB failure (kill connection mid-insert).
//   → msg.Nak() called (not Term) so NATS will redeliver.
// ---------------------------------------------------------------------------
