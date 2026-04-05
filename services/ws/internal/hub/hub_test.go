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
