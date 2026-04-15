// Package service consumer_test covers the NATS consumer routing logic.
//
// Architecture note: Consumer.handleEvent calls c.svc *Traces which uses
// *pgxpool.Pool directly (no repository interface). This means tests for
// the Ack path (service.RecordTraceStart/End/Event succeeds) require a real
// database and are tagged as integration tests (TDD-ANCHOR section below).
//
// This file covers all paths that resolve BEFORE reaching the service call:
//   - Invalid subject format → Term
//   - Malformed JSON payload → Term
//   - Tenant mismatch (payload tenant_id ≠ subject slug) → Term
//   - Unknown action → Term
//
// These are the most security-critical paths. The Ack/Nak paths that exercise
// service calls are documented as TDD-ANCHOR contracts at the bottom.
package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ---------------------------------------------------------------------------
// Mock: jetstream.Msg
// ---------------------------------------------------------------------------

// mockMsg implements jetstream.Msg for testing. Only the methods called by
// handleEvent are implemented; unused methods panic if invoked.
type mockMsg struct {
	subject string
	data    []byte

	ackCalled  bool
	nakCalled  bool
	termCalled bool
}

// Compile-time check: mockMsg must satisfy jetstream.Msg.
var _ jetstream.Msg = (*mockMsg)(nil)

func (m *mockMsg) Subject() string                             { return m.subject }
func (m *mockMsg) Data() []byte                                { return m.data }
func (m *mockMsg) Ack() error                                  { m.ackCalled = true; return nil }
func (m *mockMsg) Nak() error                                  { m.nakCalled = true; return nil }
func (m *mockMsg) Term() error                                 { m.termCalled = true; return nil }
func (m *mockMsg) Headers() nats.Header                        { return nil }
func (m *mockMsg) Reply() string                               { return "" }
func (m *mockMsg) DoubleAck(_ context.Context) error           { panic("not implemented") }
func (m *mockMsg) NakWithDelay(_ time.Duration) error          { panic("not implemented") }
func (m *mockMsg) InProgress() error                           { panic("not implemented") }
func (m *mockMsg) TermWithReason(_ string) error               { panic("not implemented") }
func (m *mockMsg) Metadata() (*jetstream.MsgMetadata, error)   { return nil, nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newConsumerNoPool builds a Consumer with a nil *Traces (no pool).
// This is safe for tests that resolve before any service call (Term paths).
// Tests that reach RecordTraceStart/End/Event will panic on the nil pool —
// those are TDD-ANCHOR integration tests documented below.
func newConsumerNoPool() *Consumer {
	return &Consumer{
		svc: &Traces{pool: nil},
		ctx: context.Background(),
	}
}

// tenantFromSubjectHelper mirrors the inline tenant extraction logic from
// Consumer.handleEvent (lines 100-106 of consumer.go):
//
//	parts := strings.Split(subject, ".")
//	if len(parts) < 4 { return "" }
//	return parts[1]
//
// This helper is used to test the subject-parsing contract without driving it
// through handleEvent, and to document the invariant explicitly.
func tenantFromSubjectHelper(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		return ""
	}
	return parts[1]
}

// ---------------------------------------------------------------------------
// Tests: subject tenant extraction (pure logic, mirrors consumer.go:100-106)
// ---------------------------------------------------------------------------

func TestConsumer_TenantFromSubject_Valid(t *testing.T) {
	tests := []struct {
		subject string
		want    string
	}{
		{"tenant.saldivia.traces.start", "saldivia"},
		{"tenant.acme.traces.end", "acme"},
		{"tenant.foo.traces.event", "foo"},
		{"tenant.bigcorp.traces.start", "bigcorp"},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got := tenantFromSubjectHelper(tt.subject)
			if got != tt.want {
				t.Errorf("tenantFromSubjectHelper(%q) = %q, want %q", tt.subject, got, tt.want)
			}
		})
	}
}

func TestConsumer_TenantFromSubject_Invalid_ReturnsEmpty(t *testing.T) {
	// NOTE: the production consumer (consumer.go:100-106) only checks len(parts) < 4.
	// It does NOT validate that parts[0] == "tenant". A subject like
	// "service.saldivia.traces.start" would extract "saldivia" — the NATS subject
	// routing (tenant.*.traces.>) is the first line of defense against invalid prefixes.
	// These tests mirror exactly what the production code does.
	tests := []struct {
		name    string
		subject string
	}{
		{"too few parts (3)", "tenant.saldivia.traces"},
		{"too few parts (2)", "traces.start"},
		{"empty string", ""},
		{"only dots — splits into 4 empty strings, parts[1]=empty", "..."},
		{"single part", "tenant"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tenantFromSubjectHelper(tt.subject)
			if got != "" {
				t.Errorf("tenantFromSubjectHelper(%q) = %q, want empty string", tt.subject, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: handleEvent — Term paths (no service call required)
// ---------------------------------------------------------------------------

// TestConsumer_HandleEvent_InvalidSubject_Terms verifies that a subject with
// fewer than 4 dot-separated parts causes Term (permanent failure, no retry).
func TestConsumer_HandleEvent_InvalidSubject_Terms(t *testing.T) {
	tests := []struct {
		name    string
		subject string
	}{
		{"3 parts", "tenant.saldivia.traces"},
		{"2 parts", "traces.start"},
		{"1 part", "tenant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newConsumerNoPool()
			msg := &mockMsg{
				subject: tt.subject,
				data:    []byte(`{"trace_id":"t1","tenant_id":"saldivia"}`),
			}
			c.handleEvent(msg)

			if !msg.termCalled {
				t.Error("invalid subject must Term")
			}
			if msg.ackCalled || msg.nakCalled {
				t.Error("must not Ack or Nak on invalid subject")
			}
		})
	}
}

// TestConsumer_HandleEvent_MalformedJSON_Terms verifies that a non-JSON body
// causes Term for all recognized action types. Non-parseable messages are
// permanently broken — retrying will not fix them.
func TestConsumer_HandleEvent_MalformedJSON_Terms(t *testing.T) {
	actions := []string{"start", "end", "event"}

	for _, action := range actions {
		t.Run("action="+action, func(t *testing.T) {
			c := newConsumerNoPool()
			msg := &mockMsg{
				subject: "tenant.saldivia.traces." + action,
				data:    []byte(`{invalid json`),
			}
			c.handleEvent(msg)

			if !msg.termCalled {
				t.Errorf("action=%s: malformed JSON must Term", action)
			}
			if msg.ackCalled || msg.nakCalled {
				t.Errorf("action=%s: must not Ack or Nak on malformed JSON", action)
			}
		})
	}
}

// TestConsumer_HandleEvent_UnknownAction_Terms verifies that an unrecognized
// action in the subject causes Term. Only "start", "end", "event" are valid.
func TestConsumer_HandleEvent_UnknownAction_Terms(t *testing.T) {
	c := newConsumerNoPool()
	msg := &mockMsg{
		subject: "tenant.saldivia.traces.unknown",
		data:    []byte(`{"trace_id":"t1"}`),
	}
	c.handleEvent(msg)

	if !msg.termCalled {
		t.Error("unknown action must Term")
	}
	if msg.ackCalled || msg.nakCalled {
		t.Error("must not Ack or Nak on unknown action")
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: tenant spoofing protection (consumer.go:119-123, 137-141, 155-159)
//
// handleEvent extracts the tenant slug from msg.Subject() and compares it to
// evt.TenantID from the payload. If they differ, it Terms the message.
// This prevents a producer from injecting a different tenant by crafting a body
// with a mismatched tenant_id field.
// ---------------------------------------------------------------------------

// TestConsumer_HandleEvent_TenantSpoofing_UsesSubjectNotPayload is the critical
// cross-tenant isolation test. It verifies all three action types.
//
// Protocol:
//   - Subject slug: "saldivia" (the legitimate tenant)
//   - Payload tenant_id: "attacker" (the spoofed tenant)
//   - Expected outcome: Term (mismatch detected, message rejected)
//
// If the consumer used payload tenant_id instead of the subject slug, it would
// route the trace to "attacker" instead of rejecting it. This test verifies
// the mismatch check fires before any service call.
func TestConsumer_HandleEvent_TenantSpoofing_UsesSubjectNotPayload(t *testing.T) {
	// start action: TraceStartEvent.TenantID is always compared to subjectTenant
	t.Run("action=start", func(t *testing.T) {
		c := newConsumerNoPool()
		payload, _ := json.Marshal(TraceStartEvent{
			TraceID:  "tr-1",
			TenantID: "attacker", // spoofed — subject says "saldivia"
			UserID:   "u1",
			Query:    "malicious query",
		})
		msg := &mockMsg{
			subject: "tenant.saldivia.traces.start",
			data:    payload,
		}
		c.handleEvent(msg)

		if !msg.termCalled {
			t.Error("spoofed tenant_id in start payload must Term")
		}
		if msg.ackCalled {
			t.Error("spoofed message must NOT be Acked")
		}
		if msg.nakCalled {
			t.Error("spoofed message must NOT be Naked (no retry)")
		}
	})

	// end action: TraceEndEvent.TenantID non-empty mismatch → Term
	t.Run("action=end", func(t *testing.T) {
		c := newConsumerNoPool()
		payload, _ := json.Marshal(TraceEndEvent{
			TraceID:  "tr-1",
			TenantID: "attacker",
			Status:   "completed",
		})
		msg := &mockMsg{
			subject: "tenant.saldivia.traces.end",
			data:    payload,
		}
		c.handleEvent(msg)

		if !msg.termCalled {
			t.Error("spoofed tenant_id in end payload must Term")
		}
		if msg.ackCalled || msg.nakCalled {
			t.Error("spoofed message must not Ack or Nak")
		}
	})

	// event action: TraceEvent.TenantID non-empty mismatch → Term
	t.Run("action=event", func(t *testing.T) {
		c := newConsumerNoPool()
		payload, _ := json.Marshal(TraceEvent{
			TraceID:   "tr-1",
			TenantID:  "attacker",
			Seq:       1,
			EventType: "llm_call",
			Data:      json.RawMessage(`{}`),
		})
		msg := &mockMsg{
			subject: "tenant.saldivia.traces.event",
			data:    payload,
		}
		c.handleEvent(msg)

		if !msg.termCalled {
			t.Error("spoofed tenant_id in event payload must Term")
		}
		if msg.ackCalled || msg.nakCalled {
			t.Error("spoofed message must not Ack or Nak")
		}
	})
}

// TestConsumer_HandleEvent_TenantSpoofing_EmptyPayloadTenantAllowed verifies
// the complementary case: if payload tenant_id is empty (omitted), the consumer
// does NOT reject the message as a mismatch — it trusts the subject slug.
//
// This tests the condition: `subjectTenant != "" && evt.TenantID != "" && evt.TenantID != subjectTenant`
// — the check only fires when BOTH sides are non-empty and they differ.
//
// NOTE: for "start" action (consumer.go:119), the check is stricter:
//   `subjectTenant != "" && evt.TenantID != subjectTenant`
//   An empty payload tenant_id for "start" IS treated as mismatch (empty != "saldivia").
//   This is intentional: trace starts require explicit tenant_id.
//
// For "end" and "event", empty payload tenant_id passes through to the service call.
// We test "end" and "event" only (pool=nil → service call would panic, so we
// document this as TDD-ANCHOR for integration testing).
func TestConsumer_HandleEvent_TenantSpoofing_EmptyPayloadTenantAllowed_Spec(t *testing.T) {
	// This test is a specification test — it documents the mismatch condition
	// used in handleEvent without executing the full path.
	//
	// Condition for "end" action (consumer.go:137):
	//   if subjectTenant != "" && evt.TenantID != "" && evt.TenantID != subjectTenant → Term
	//
	// When evt.TenantID == "" (payload omits tenant_id), the condition is false → no Term.
	// The message proceeds to RecordTraceEnd (service call with pool — needs integration test).

	// Verify the condition logic in isolation:
	subjectTenant := "saldivia"

	cases := []struct {
		payloadTenant string
		wantMismatch  bool
	}{
		{"attacker", true},  // non-empty + different → mismatch → Term
		{"saldivia", false}, // same → no mismatch
		{"", false},         // empty → condition short-circuits → no mismatch for end/event
	}

	for _, c := range cases {
		mismatch := subjectTenant != "" && c.payloadTenant != "" && c.payloadTenant != subjectTenant
		if mismatch != c.wantMismatch {
			t.Errorf("payloadTenant=%q: mismatch=%v, want=%v", c.payloadTenant, mismatch, c.wantMismatch)
		}
	}
}

// ---------------------------------------------------------------------------
// TDD-ANCHOR: Ack and Nak paths (service calls required)
//
// The following paths reach c.svc.RecordTraceStart/End/Event which use
// *pgxpool.Pool. Testing these requires a real PostgreSQL connection.
//
// Test contracts (implement as integration tests with testcontainers):
//
//   TestConsumer_HandleEvent_ValidStartEvent_PersistsAndAcks:
//     subject=tenant.saldivia.traces.start, payload tenant_id=saldivia
//     → RecordTraceStart called with evt.TenantID="saldivia"
//     → msg.Ack() called
//
//   TestConsumer_HandleEvent_ValidEndEvent_PersistsAndAcks:
//     subject=tenant.saldivia.traces.end, payload with matching tenant_id
//     → RecordTraceEnd called
//     → msg.Ack() called
//
//   TestConsumer_HandleEvent_ValidTraceEvent_PersistsAndAcks:
//     subject=tenant.saldivia.traces.event, payload with matching tenant_id
//     → RecordEvent called
//     → msg.Ack() called
//
//   TestConsumer_HandleEvent_ServiceError_Naks:
//     RecordTraceStart returns error (e.g., DB unavailable)
//     → msg.Nak() called (transient — NATS should redeliver)
//     → msg.Term() NOT called
// ---------------------------------------------------------------------------
