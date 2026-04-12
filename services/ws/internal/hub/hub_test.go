package hub

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	h := New()
	if h == nil {
		t.Fatal("expected non-nil hub")
	}
	if h.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", h.ClientCount())
	}
}

func TestHub_RegisterUnregister(t *testing.T) {
	h := New()
	go h.Run()

	client := &Client{
		hub:    h,
		send:   make(chan []byte, 16),
		subs:   make(map[string]bool),
		UserID: "u-1",
		Slug:   "saldivia",
	}

	h.register <- client
	time.Sleep(10 * time.Millisecond)

	if h.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", h.ClientCount())
	}

	h.unregister <- client
	time.Sleep(10 * time.Millisecond)

	if h.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after unregister, got %d", h.ClientCount())
	}
}

func TestHub_ClientCountByTenant(t *testing.T) {
	h := New()
	go h.Run()

	c1 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-1", Slug: "saldivia"}
	c2 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-2", Slug: "saldivia"}
	c3 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-3", Slug: "empresa2"}

	h.register <- c1
	h.register <- c2
	h.register <- c3
	time.Sleep(20 * time.Millisecond)

	if got := h.ClientCountByTenant("saldivia"); got != 2 {
		t.Errorf("expected 2 for saldivia, got %d", got)
	}
	if got := h.ClientCountByTenant("empresa2"); got != 1 {
		t.Errorf("expected 1 for empresa2, got %d", got)
	}
	if got := h.ClientCountByTenant("nobody"); got != 0 {
		t.Errorf("expected 0 for nobody, got %d", got)
	}
}

func TestHub_BroadcastToTenant(t *testing.T) {
	h := New()
	go h.Run()

	c1 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-1", Slug: "saldivia"}
	c2 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-2", Slug: "empresa2"}

	h.register <- c1
	h.register <- c2
	time.Sleep(10 * time.Millisecond)

	c1.Subscribe("notifications")
	c2.Subscribe("notifications")

	h.BroadcastToTenant("saldivia", "notifications", Message{
		Type:    Event,
		Channel: "notifications",
		Data:    json.RawMessage(`{"text":"hello saldivia"}`),
	})

	time.Sleep(10 * time.Millisecond)

	// c1 (saldivia) should receive
	select {
	case data := <-c1.send:
		var msg Message
		json.Unmarshal(data, &msg)
		if msg.Channel != "notifications" {
			t.Errorf("expected channel 'notifications', got %q", msg.Channel)
		}
	default:
		t.Error("c1 (saldivia) should have received a message")
	}

	// c2 (empresa2) should NOT receive
	select {
	case <-c2.send:
		t.Error("c2 (empresa2) should NOT have received a message")
	default:
		// good
	}
}

func TestHub_BroadcastToTenant_UnsubscribedClient_NotReceive(t *testing.T) {
	// Client in correct tenant but NOT subscribed to the channel must not receive.
	h := New()
	go h.Run()

	c := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-1", Slug: "saldivia"}
	h.register <- c
	time.Sleep(10 * time.Millisecond)

	// c is NOT subscribed to "notifications"

	h.BroadcastToTenant("saldivia", "notifications", Message{
		Type:    Event,
		Channel: "notifications",
		Data:    json.RawMessage(`{}`),
	})
	time.Sleep(10 * time.Millisecond)

	select {
	case <-c.send:
		t.Error("unsubscribed client should NOT receive broadcast")
	default:
		// correct
	}
}

func TestHub_Broadcast_AllSubscribedClients(t *testing.T) {
	h := New()
	go h.Run()

	c1 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-1", Slug: "t1"}
	c2 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-2", Slug: "t2"}
	c3 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-3", Slug: "t1"}

	h.register <- c1
	h.register <- c2
	h.register <- c3
	time.Sleep(10 * time.Millisecond)

	c1.Subscribe("global")
	c3.Subscribe("global")
	// c2 does NOT subscribe

	h.Broadcast("global", Message{Type: Event, Channel: "global", Data: json.RawMessage(`{}`)})
	time.Sleep(10 * time.Millisecond)

	for _, name := range []string{"c1", "c3"} {
		var c *Client
		if name == "c1" {
			c = c1
		} else {
			c = c3
		}
		select {
		case <-c.send:
			// correct — received
		default:
			t.Errorf("%s should have received global broadcast", name)
		}
	}

	select {
	case <-c2.send:
		t.Error("c2 (not subscribed) should NOT receive global broadcast")
	default:
		// correct
	}
}

func TestHub_MaxClients_RejectsOverCapacity(t *testing.T) {
	h := New()
	h.MaxClients = 2
	go h.Run()

	c1 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-1", Slug: "t1"}
	c2 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-2", Slug: "t1"}
	c3 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-3", Slug: "t1"}

	h.register <- c1
	h.register <- c2
	time.Sleep(10 * time.Millisecond)

	if h.ClientCount() != 2 {
		t.Fatalf("expected 2 clients, got %d", h.ClientCount())
	}

	h.register <- c3
	time.Sleep(10 * time.Millisecond)

	// c3 should be rejected — hub at capacity
	if h.ClientCount() != 2 {
		t.Errorf("expected still 2 clients after cap exceeded, got %d", h.ClientCount())
	}

	// c3 should have received an error message
	select {
	case data := <-c3.send:
		var msg Message
		json.Unmarshal(data, &msg)
		if msg.Type != Error {
			t.Errorf("expected error type, got %s", msg.Type)
		}
	case <-time.After(50 * time.Millisecond):
		// c3 might be closed already without receiving (markClosed + close channel)
	}
}

func TestHub_MaxClientsPerTenant_RejectsOverLimit(t *testing.T) {
	h := New()
	h.MaxClientsPerTenant = 2
	go h.Run()

	c1 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-1", Slug: "saldivia"}
	c2 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-2", Slug: "saldivia"}
	c3 := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-3", Slug: "saldivia"}
	cOther := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool), UserID: "u-4", Slug: "empresa2"}

	h.register <- c1
	h.register <- c2
	h.register <- cOther
	time.Sleep(10 * time.Millisecond)

	if h.ClientCountByTenant("saldivia") != 2 {
		t.Fatalf("expected 2 clients for saldivia")
	}

	h.register <- c3
	time.Sleep(10 * time.Millisecond)

	// c3 should be rejected for saldivia — per-tenant limit reached
	if h.ClientCountByTenant("saldivia") != 2 {
		t.Errorf("expected 2 clients for saldivia after rejection, got %d", h.ClientCountByTenant("saldivia"))
	}
	// Other tenant unaffected
	if h.ClientCountByTenant("empresa2") != 1 {
		t.Errorf("expected 1 client for empresa2, got %d", h.ClientCountByTenant("empresa2"))
	}
}

func TestHub_Unregister_SendsDisconnectCleanup(t *testing.T) {
	h := New()
	go h.Run()

	c := &Client{hub: h, send: make(chan []byte, 64), subs: make(map[string]bool), UserID: "u-1", Slug: "t1"}
	h.register <- c
	time.Sleep(10 * time.Millisecond)

	if h.ClientCount() != 1 {
		t.Fatalf("expected 1 client after register, got %d", h.ClientCount())
	}

	h.unregister <- c
	time.Sleep(10 * time.Millisecond)

	if h.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after unregister, got %d", h.ClientCount())
	}
	if !c.closed.Load() {
		t.Error("expected client.closed = true after unregister")
	}
}

func TestClient_SubscribeUnsubscribe(t *testing.T) {
	c := &Client{subs: make(map[string]bool)}

	c.Subscribe("chat.messages:s-1")
	if !c.IsSubscribed("chat.messages:s-1") {
		t.Error("expected subscribed")
	}

	c.Unsubscribe("chat.messages:s-1")
	if c.IsSubscribed("chat.messages:s-1") {
		t.Error("expected unsubscribed")
	}
}

func TestClient_Channels(t *testing.T) {
	c := &Client{subs: make(map[string]bool)}
	c.Subscribe("a")
	c.Subscribe("b")

	chs := c.Channels()
	if len(chs) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(chs))
	}
}

func TestClient_MaxSubscriptions(t *testing.T) {
	c := &Client{subs: make(map[string]bool)}

	// Fill to the limit
	for i := 0; i < maxSubscriptions; i++ {
		ok := c.Subscribe("ch-" + string(rune('a'+i%26)) + "-" + string(rune('0'+i/26)))
		if !ok {
			t.Fatalf("subscribe %d failed before limit", i)
		}
	}

	// One more should fail
	ok := c.Subscribe("overflow-channel")
	if ok {
		t.Error("expected Subscribe to return false at max subscriptions")
	}
}

func TestHub_HandleMessage_Subscribe(t *testing.T) {
	h := New()
	go h.Run()

	c := &Client{
		hub:    h,
		send:   make(chan []byte, 16),
		subs:   make(map[string]bool),
		UserID: "u-1",
		Slug:   "test",
	}
	h.register <- c
	time.Sleep(10 * time.Millisecond)

	h.handleMessage(c, Message{
		Type:    Subscribe,
		Channel: "chat.messages:s-1",
		ID:      "req-1",
	})

	if !c.IsSubscribed("chat.messages:s-1") {
		t.Error("expected client subscribed to chat.messages:s-1")
	}

	// Should have received confirmation
	select {
	case data := <-c.send:
		var msg Message
		json.Unmarshal(data, &msg)
		if msg.ID != "req-1" {
			t.Errorf("expected correlation ID 'req-1', got %q", msg.ID)
		}
	default:
		t.Error("expected subscription confirmation message")
	}
}

func TestHub_HandleMessage_SubscribeEmpty(t *testing.T) {
	h := New()
	c := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool)}

	h.handleMessage(c, Message{Type: Subscribe, Channel: ""})

	select {
	case data := <-c.send:
		var msg Message
		json.Unmarshal(data, &msg)
		if msg.Type != Error {
			t.Errorf("expected error type, got %s", msg.Type)
		}
	default:
		t.Error("expected error message for empty channel")
	}
}

func TestHub_HandleMessage_Unsubscribe(t *testing.T) {
	h := New()
	c := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool)}

	c.Subscribe("ch-1")

	h.handleMessage(c, Message{
		Type:    Unsubscribe,
		Channel: "ch-1",
		ID:      "req-2",
	})

	if c.IsSubscribed("ch-1") {
		t.Error("expected client unsubscribed from ch-1")
	}

	// Should receive confirmation
	select {
	case data := <-c.send:
		var msg Message
		json.Unmarshal(data, &msg)
		if msg.Channel != "ch-1" {
			t.Errorf("expected channel 'ch-1' in confirmation, got %q", msg.Channel)
		}
	default:
		t.Error("expected unsubscribe confirmation message")
	}
}

func TestHub_HandleMessage_MutationsDisabled(t *testing.T) {
	h := New() // Mutations is nil by default
	c := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool)}

	h.handleMessage(c, Message{Type: Mutation, ID: "req-3"})

	select {
	case data := <-c.send:
		var msg Message
		json.Unmarshal(data, &msg)
		if msg.Type != Error {
			t.Errorf("expected error type, got %s", msg.Type)
		}
	default:
		t.Error("expected error message when mutations disabled")
	}
}

func TestHub_HandleMessage_UnknownType(t *testing.T) {
	h := New()
	c := &Client{hub: h, send: make(chan []byte, 16), subs: make(map[string]bool)}

	h.handleMessage(c, Message{Type: "unknown_type", ID: "req-4"})

	select {
	case data := <-c.send:
		var msg Message
		json.Unmarshal(data, &msg)
		if msg.Type != Error {
			t.Errorf("expected error type for unknown message type, got %s", msg.Type)
		}
	default:
		t.Error("expected error message for unknown type")
	}
}

func TestProtocol_MessageTypes(t *testing.T) {
	if Subscribe != "subscribe" {
		t.Error("wrong subscribe constant")
	}
	if Mutation != "mutation" {
		t.Error("wrong mutation constant")
	}
	if Event != "event" {
		t.Error("wrong event constant")
	}
}

func TestClient_TrySend_ClosedClient_ReturnsFalse(t *testing.T) {
	c := &Client{send: make(chan []byte, 4), subs: make(map[string]bool)}
	c.markClosed()

	ok := c.TrySend([]byte("data"))
	if ok {
		t.Error("expected TrySend to return false on closed client")
	}
}

func TestClient_TrySend_FullBuffer_ReturnsFalse(t *testing.T) {
	// Buffer size 1 — fill it, then try another send
	c := &Client{send: make(chan []byte, 1), subs: make(map[string]bool)}
	c.TrySend([]byte("fill"))

	ok := c.TrySend([]byte("overflow"))
	if ok {
		t.Error("expected TrySend to return false when buffer full")
	}
}
