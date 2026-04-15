package hub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

// newTestHub returns a hub with Run() already started.
// Caller must not close/stop; tests are short-lived.
func newTestHub() *Hub {
	h := New()
	go h.Run()
	// Give the goroutine a moment to enter the select loop
	time.Sleep(5 * time.Millisecond)
	return h
}

// newTestClient returns a buffered client registered in the hub for the given
// tenant slug and subscribed to the given channels.
func newTestClient(h *Hub, userID, slug string, channels ...string) *Client {
	c := &Client{
		hub:    h,
		send:   make(chan []byte, 64),
		subs:   make(map[string]bool),
		UserID: userID,
		Slug:   slug,
	}
	h.register <- c
	time.Sleep(10 * time.Millisecond)
	for _, ch := range channels {
		c.Subscribe(ch)
	}
	return c
}

// invokeNATSHandler calls the private handleNATSMessage directly.
// Since this test file is in the same package (package hub), this is allowed.
func invokeNATSHandler(b *NATSBridge, subject string, data []byte) {
	b.handleNATSMessage(&nats.Msg{
		Subject: subject,
		Data:    data,
	})
}

// drainClient reads all messages currently in the client's send buffer.
func drainClient(c *Client) []Message {
	var msgs []Message
	for {
		select {
		case raw := <-c.send:
			var m Message
			if err := json.Unmarshal(raw, &m); err == nil {
				msgs = append(msgs, m)
			}
		default:
			return msgs
		}
	}
}

// --- Subject parsing tests ---

// TestNATSBridge_SubjectParsing_ValidSubject verifies that a well-formed subject
// of the form tenant.{slug}.{channel} routes to the correct tenant.
func TestNATSBridge_SubjectParsing_ValidSubject(t *testing.T) {
	h := newTestHub()

	// One client in "saldivia", subscribed to "chat.messages"
	c := newTestClient(h, "u-1", "saldivia", "chat.messages")

	b := NewNATSBridge(h, nil) // conn not needed — we invoke handler directly

	payload, _ := json.Marshal(Message{
		Type:    Event,
		Channel: "chat.messages",
		Data:    json.RawMessage(`{"text":"hello"}`),
	})
	invokeNATSHandler(b, "tenant.saldivia.chat.messages", payload)
	time.Sleep(10 * time.Millisecond)

	msgs := drainClient(c)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message delivered to saldivia client, got %d", len(msgs))
	}
	if msgs[0].Channel != "chat.messages" {
		t.Errorf("expected channel 'chat.messages', got %q", msgs[0].Channel)
	}
}

// TestNATSBridge_SubjectParsing_InvalidSubject_Ignored verifies that a subject
// without the "tenant." prefix is silently ignored and causes no panic.
// This mirrors Invariant 3: ALL events must be tenant-namespaced.
func TestNATSBridge_SubjectParsing_InvalidSubject_Ignored(t *testing.T) {
	h := newTestHub()
	c := newTestClient(h, "u-1", "saldivia", "chat.messages")

	b := NewNATSBridge(h, nil)

	// Subject has only 2 parts — missing tenant prefix entirely.
	// SplitN("chat.messages", ".", 3) → ["chat", "messages"] → len < 3 → ignored.
	invokeNATSHandler(b, "chat.messages", []byte(`{"type":"event"}`))
	time.Sleep(10 * time.Millisecond)

	msgs := drainClient(c)
	if len(msgs) != 0 {
		t.Errorf("expected no messages for invalid subject, got %d", len(msgs))
	}
}

// TestNATSBridge_SubjectParsing_EmptySubject_Ignored verifies that an empty
// subject causes no panic and no delivery.
func TestNATSBridge_SubjectParsing_EmptySubject_Ignored(t *testing.T) {
	h := newTestHub()
	c := newTestClient(h, "u-1", "saldivia", "chat.messages")

	b := NewNATSBridge(h, nil)

	invokeNATSHandler(b, "", []byte(`{}`))
	time.Sleep(10 * time.Millisecond)

	msgs := drainClient(c)
	if len(msgs) != 0 {
		t.Errorf("expected no messages for empty subject, got %d", len(msgs))
	}
}

// TestNATSBridge_SubjectParsing_MissingChannelPart verifies a subject with
// only "tenant.slug" (no channel component) is ignored gracefully.
func TestNATSBridge_SubjectParsing_MissingChannelPart(t *testing.T) {
	h := newTestHub()
	c := newTestClient(h, "u-1", "saldivia")

	b := NewNATSBridge(h, nil)

	// SplitN("tenant.saldivia", ".", 3) → ["tenant", "saldivia"] → len < 3 → ignored.
	invokeNATSHandler(b, "tenant.saldivia", []byte(`{}`))
	time.Sleep(10 * time.Millisecond)

	msgs := drainClient(c)
	if len(msgs) != 0 {
		t.Errorf("expected no messages for subject missing channel, got %d", len(msgs))
	}
}

// --- Broadcast routing tests ---

// TestNATSBridge_BroadcastRouting_CorrectTenant verifies that a NATS message
// on tenant.saldivia.* is delivered ONLY to clients of tenant "saldivia".
// This is Invariant 3 at the NATS bridge layer.
func TestNATSBridge_BroadcastRouting_CorrectTenant(t *testing.T) {
	h := newTestHub()

	cSaldivia := newTestClient(h, "u-1", "saldivia", "notifications")
	cOther := newTestClient(h, "u-2", "empresa2", "notifications")

	b := NewNATSBridge(h, nil)

	payload, _ := json.Marshal(Message{
		Type:    Event,
		Channel: "notifications",
		Data:    json.RawMessage(`{"msg":"for saldivia"}`),
	})
	invokeNATSHandler(b, "tenant.saldivia.notifications", payload)
	time.Sleep(10 * time.Millisecond)

	// saldivia client must receive
	saldivMsgs := drainClient(cSaldivia)
	if len(saldivMsgs) != 1 {
		t.Errorf("saldivia client: expected 1 message, got %d", len(saldivMsgs))
	}

	// empresa2 client must NOT receive
	otherMsgs := drainClient(cOther)
	if len(otherMsgs) != 0 {
		t.Errorf("empresa2 client: expected 0 messages, got %d — tenant isolation breach", len(otherMsgs))
	}
}

// TestNATSBridge_BroadcastRouting_WrongTenant_NotDelivered is the CRITICAL
// tenant isolation test. It verifies that a NATS message addressed to tenantA
// is NEVER delivered to a client of tenantB, even if both are subscribed to the
// same channel.
//
// SECURITY NOTE: This test catches the specific regression where a developer
// reads the tenant slug from the NATS message *payload* (e.g. msg.Data.tenant)
// instead of the *subject*. If the bridge ever used the payload slug for routing,
// a malicious publisher could inject events into any tenant's channel by crafting
// a payload with a different tenant_id. The subject must be the authoritative
// source of the tenant identity — it is set by the publishing service and verified
// by NATS subject ACLs, whereas the payload is opaque user-controlled data.
func TestNATSBridge_BroadcastRouting_WrongTenant_NotDelivered(t *testing.T) {
	h := newTestHub()

	cTenantA := newTestClient(h, "u-1", "tenantA", "chat.messages")
	cTenantB := newTestClient(h, "u-2", "tenantB", "chat.messages")

	b := NewNATSBridge(h, nil)

	// Publish on tenantA's subject. The payload contains a "tenant" field that
	// could trick a naive implementation into routing to tenantB.
	// The bridge MUST use the subject ("tenantA") — not the payload field.
	payload := json.RawMessage(`{"type":"event","channel":"chat.messages","tenant":"tenantB","data":{}}`)
	invokeNATSHandler(b, "tenant.tenantA.chat.messages", payload)
	time.Sleep(10 * time.Millisecond)

	// tenantA client should receive (subject slug matches)
	aMsgs := drainClient(cTenantA)
	if len(aMsgs) != 1 {
		t.Errorf("tenantA client: expected 1 message, got %d", len(aMsgs))
	}

	// tenantB client must NOT receive — subject says tenantA, regardless of payload
	bMsgs := drainClient(cTenantB)
	if len(bMsgs) != 0 {
		t.Fatalf("TENANT ISOLATION BREACH: tenantB received %d message(s) addressed to tenantA — "+
			"bridge must route using subject slug, not payload tenant field", len(bMsgs))
	}
}

// TestNATSBridge_RawPayload_WrappedAsEvent verifies that when the NATS payload
// is not a valid Message JSON, the bridge wraps it as a raw event and still
// routes it correctly to the tenant.
func TestNATSBridge_RawPayload_WrappedAsEvent(t *testing.T) {
	h := newTestHub()

	c := newTestClient(h, "u-1", "saldivia", "fleet.update")

	b := NewNATSBridge(h, nil)

	// Raw payload — not a Message struct, just arbitrary JSON
	rawPayload := []byte(`{"vehicle_id":"v-42","status":"moving"}`)
	invokeNATSHandler(b, "tenant.saldivia.fleet.update", rawPayload)
	time.Sleep(10 * time.Millisecond)

	msgs := drainClient(c)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message for raw payload, got %d", len(msgs))
	}
	if msgs[0].Type != Event {
		t.Errorf("expected type 'event' for wrapped raw payload, got %q", msgs[0].Type)
	}
	if msgs[0].Channel != "fleet.update" {
		t.Errorf("expected channel 'fleet.update', got %q", msgs[0].Channel)
	}
}

// TestNATSBridge_ChannelFromSubject_MultiDotChannel verifies that a channel
// with multiple dot components (e.g. "chat.messages") is extracted correctly
// from subjects like "tenant.saldivia.chat.messages".
// SplitN(..., ".", 3) must yield parts[2] = "chat.messages" (the full rest).
func TestNATSBridge_ChannelFromSubject_MultiDotChannel(t *testing.T) {
	h := newTestHub()

	// Subscribe to the multi-dot channel
	c := newTestClient(h, "u-1", "saldivia", "chat.messages")

	b := NewNATSBridge(h, nil)

	payload, _ := json.Marshal(Message{
		Type:    Event,
		Channel: "chat.messages",
		Data:    json.RawMessage(`{}`),
	})
	invokeNATSHandler(b, "tenant.saldivia.chat.messages", payload)
	time.Sleep(10 * time.Millisecond)

	msgs := drainClient(c)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message for multi-dot channel, got %d", len(msgs))
	}
}
