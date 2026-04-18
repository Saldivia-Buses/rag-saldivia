// Package service consumer_test covers the NATS consumer routing logic.
//
// Consumer.svc is a TraceRecorder interface (see consumer.go), so tests can
// substitute a recording spy instead of a real *Traces + pgxpool. This file
// covers all handleEvent paths:
//   - Invalid subject format → Term
//   - Malformed JSON payload → Term
//   - Tenant mismatch (payload tenant_id ≠ subject slug) → Term
//   - Unknown action → Term
//   - Valid events (start/end/event) → spy called + Ack
//   - Spy returns error → Nak (transient, NATS redelivers)
package service

import (
	"context"
	"encoding/json"
	"errors"
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

// traceSpy implements TraceRecorder and records every call for assertions.
// StartErr / EndErr / EventErr let a test force the error path.
type traceSpy struct {
	Starts    []TraceStartEvent
	Ends      []TraceEndEvent
	Events    []TraceEvent
	StartErr  error
	EndErr    error
	EventErr  error
}

func (s *traceSpy) RecordTraceStart(_ context.Context, evt TraceStartEvent) error {
	s.Starts = append(s.Starts, evt)
	return s.StartErr
}

func (s *traceSpy) RecordTraceEnd(_ context.Context, evt TraceEndEvent) error {
	s.Ends = append(s.Ends, evt)
	return s.EndErr
}

func (s *traceSpy) RecordEvent(_ context.Context, evt TraceEvent) error {
	s.Events = append(s.Events, evt)
	return s.EventErr
}

// newConsumerNoPool builds a Consumer wired to a fresh traceSpy. Kept under
// this name to preserve call sites in existing tests (they exercise Term paths
// that never reach the spy, so the fresh spy is harmless).
func newConsumerNoPool() *Consumer {
	return &Consumer{
		svc: &traceSpy{},
		ctx: context.Background(),
	}
}

// newConsumerWithSpy builds a Consumer wired to the given spy. Use this when
// a test needs to assert on what was recorded or force an error.
func newConsumerWithSpy(spy *traceSpy) *Consumer {
	return &Consumer{
		svc: spy,
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
// This test drives handleEvent with a spy to verify the actual routing behavior
// for all three (attacker / match / empty) combinations.
func TestConsumer_HandleEvent_EndEvent_TenantMatchMatrix(t *testing.T) {
	cases := []struct {
		name          string
		payloadTenant string
		wantTerm      bool
		wantAck       bool
		wantSpyCalls  int
	}{
		{"mismatch", "attacker", true, false, 0},
		{"match", "saldivia", false, true, 1},
		{"empty_payload", "", false, true, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			spy := &traceSpy{}
			cons := newConsumerWithSpy(spy)

			evt := TraceEndEvent{TraceID: "t-1", TenantID: c.payloadTenant}
			data, _ := json.Marshal(evt)
			msg := &mockMsg{subject: "tenant.saldivia.traces.end", data: data}

			cons.handleEvent(msg)

			if msg.termCalled != c.wantTerm {
				t.Errorf("termCalled=%v, want %v", msg.termCalled, c.wantTerm)
			}
			if msg.ackCalled != c.wantAck {
				t.Errorf("ackCalled=%v, want %v", msg.ackCalled, c.wantAck)
			}
			if len(spy.Ends) != c.wantSpyCalls {
				t.Errorf("spy calls=%d, want %d", len(spy.Ends), c.wantSpyCalls)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Ack and Nak paths (use TraceRecorder spy)
// ---------------------------------------------------------------------------

func TestConsumer_HandleEvent_ValidStartEvent_PersistsAndAcks(t *testing.T) {
	spy := &traceSpy{}
	c := newConsumerWithSpy(spy)

	evt := TraceStartEvent{TraceID: "t-1", TenantID: "saldivia"}
	data, _ := json.Marshal(evt)
	msg := &mockMsg{subject: "tenant.saldivia.traces.start", data: data}

	c.handleEvent(msg)

	if len(spy.Starts) != 1 {
		t.Fatalf("expected 1 RecordTraceStart call, got %d", len(spy.Starts))
	}
	if spy.Starts[0].TenantID != "saldivia" {
		t.Errorf("TenantID: want saldivia, got %q", spy.Starts[0].TenantID)
	}
	if !msg.ackCalled {
		t.Error("Ack not called on valid start event")
	}
	if msg.nakCalled || msg.termCalled {
		t.Error("Nak/Term should not be called on success")
	}
}

func TestConsumer_HandleEvent_ValidEndEvent_PersistsAndAcks(t *testing.T) {
	spy := &traceSpy{}
	c := newConsumerWithSpy(spy)

	evt := TraceEndEvent{TraceID: "t-1", TenantID: "saldivia"}
	data, _ := json.Marshal(evt)
	msg := &mockMsg{subject: "tenant.saldivia.traces.end", data: data}

	c.handleEvent(msg)

	if len(spy.Ends) != 1 {
		t.Fatalf("expected 1 RecordTraceEnd call, got %d", len(spy.Ends))
	}
	if !msg.ackCalled {
		t.Error("Ack not called on valid end event")
	}
}

func TestConsumer_HandleEvent_ValidTraceEvent_PersistsAndAcks(t *testing.T) {
	spy := &traceSpy{}
	c := newConsumerWithSpy(spy)

	evt := TraceEvent{TraceID: "t-1", TenantID: "saldivia"}
	data, _ := json.Marshal(evt)
	msg := &mockMsg{subject: "tenant.saldivia.traces.event", data: data}

	c.handleEvent(msg)

	if len(spy.Events) != 1 {
		t.Fatalf("expected 1 RecordEvent call, got %d", len(spy.Events))
	}
	if !msg.ackCalled {
		t.Error("Ack not called on valid trace event")
	}
}

func TestConsumer_HandleEvent_ServiceError_Naks(t *testing.T) {
	// When RecordTraceStart returns error (DB unavailable, etc.), handleEvent
	// must Nak for transient retry — never Term (which would drop the message).
	spy := &traceSpy{StartErr: errors.New("connection refused")}
	c := newConsumerWithSpy(spy)

	evt := TraceStartEvent{TraceID: "t-1", TenantID: "saldivia"}
	data, _ := json.Marshal(evt)
	msg := &mockMsg{subject: "tenant.saldivia.traces.start", data: data}

	c.handleEvent(msg)

	if !msg.nakCalled {
		t.Error("Nak not called on service error")
	}
	if msg.ackCalled {
		t.Error("Ack should not fire when service returns error")
	}
	if msg.termCalled {
		t.Error("Term must not fire on transient service error — would lose the message")
	}
}

func TestConsumer_HandleEvent_EndServiceError_Naks(t *testing.T) {
	spy := &traceSpy{EndErr: errors.New("db timeout")}
	c := newConsumerWithSpy(spy)

	evt := TraceEndEvent{TraceID: "t-1", TenantID: "saldivia"}
	data, _ := json.Marshal(evt)
	msg := &mockMsg{subject: "tenant.saldivia.traces.end", data: data}

	c.handleEvent(msg)

	if !msg.nakCalled {
		t.Error("Nak not called on RecordTraceEnd error")
	}
}

func TestConsumer_HandleEvent_EventServiceError_Naks(t *testing.T) {
	spy := &traceSpy{EventErr: errors.New("db timeout")}
	c := newConsumerWithSpy(spy)

	evt := TraceEvent{TraceID: "t-1", TenantID: "saldivia"}
	data, _ := json.Marshal(evt)
	msg := &mockMsg{subject: "tenant.saldivia.traces.event", data: data}

	c.handleEvent(msg)

	if !msg.nakCalled {
		t.Error("Nak not called on RecordEvent error")
	}
}
