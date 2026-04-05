package natspub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

// startTestNATS starts an embedded NATS server for testing.
// Returns the connection and a cleanup function.
func startTestNATS(t *testing.T) *nats.Conn {
	t.Helper()
	// Use a test NATS server if available, otherwise skip
	nc, err := nats.Connect(nats.DefaultURL,
		nats.Timeout(1*time.Second),
		nats.MaxReconnects(0),
	)
	if err != nil {
		t.Skipf("NATS not available, skipping: %v", err)
	}
	t.Cleanup(func() { nc.Close() })
	return nc
}

func TestPublisher_Notify_ValidEvent(t *testing.T) {
	nc := startTestNATS(t)
	pub := New(nc)

	// Subscribe before publishing
	sub, err := nc.SubscribeSync("tenant.saldivia.notify.chat.new_message")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	evt := Event{
		Type:   "chat.new_message",
		UserID: "u-123",
		Title:  "Nuevo mensaje",
		Body:   "Hola mundo",
	}

	if err := pub.Notify("saldivia", evt); err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	msg, err := sub.NextMsg(2 * time.Second)
	if err != nil {
		t.Fatalf("no message received: %v", err)
	}

	var received Event
	if err := json.Unmarshal(msg.Data, &received); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if received.Type != "chat.new_message" {
		t.Errorf("expected type chat.new_message, got %q", received.Type)
	}
	if received.UserID != "u-123" {
		t.Errorf("expected user_id u-123, got %q", received.UserID)
	}
	if received.Title != "Nuevo mensaje" {
		t.Errorf("expected title, got %q", received.Title)
	}
}

func TestPublisher_Notify_EmptyType_Returns_Error(t *testing.T) {
	nc := startTestNATS(t)
	pub := New(nc)

	err := pub.Notify("saldivia", Event{UserID: "u-1"})
	if err == nil {
		t.Fatal("expected error for empty event type")
	}
}

func TestPublisher_Notify_InvalidSlug_Returns_Error(t *testing.T) {
	nc := startTestNATS(t)
	pub := New(nc)

	evt := Event{Type: "test.event", UserID: "u-1", Title: "t", Body: "b"}

	tests := []struct {
		name string
		slug string
	}{
		{"empty", ""},
		{"dots", "saldivia.buses"},
		{"wildcard star", "saldivia*"},
		{"wildcard gt", "saldivia>"},
		{"spaces", "saldivia buses"},
		{"tabs", "saldivia\tbuses"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pub.Notify(tt.slug, evt)
			if err == nil {
				t.Errorf("expected error for slug %q", tt.slug)
			}
		})
	}
}

func TestPublisher_Notify_ValidSlug_Accepted(t *testing.T) {
	nc := startTestNATS(t)
	pub := New(nc)

	evt := Event{Type: "test.event", UserID: "u-1", Title: "t", Body: "b"}

	slugs := []string{"saldivia", "empresa-123", "a", "my-long-tenant-slug"}
	for _, slug := range slugs {
		t.Run(slug, func(t *testing.T) {
			// Subscribe to the expected subject
			sub, _ := nc.SubscribeSync("tenant." + slug + ".notify.test.event")
			if err := pub.Notify(slug, evt); err != nil {
				t.Errorf("Notify failed for slug %q: %v", slug, err)
			}
			if _, err := sub.NextMsg(1 * time.Second); err != nil {
				t.Errorf("no message for slug %q: %v", slug, err)
			}
		})
	}
}

func TestPublisher_Broadcast_ValidEvent(t *testing.T) {
	nc := startTestNATS(t)
	pub := New(nc)

	sub, err := nc.SubscribeSync("tenant.saldivia.notifications")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	data := map[string]string{"id": "n-1", "title": "Test"}
	if err := pub.Broadcast("saldivia", "notifications", data); err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}

	msg, err := sub.NextMsg(2 * time.Second)
	if err != nil {
		t.Fatalf("no message received: %v", err)
	}

	var received map[string]any
	json.Unmarshal(msg.Data, &received)

	if received["type"] != "event" {
		t.Errorf("expected type 'event', got %v", received["type"])
	}
	if received["channel"] != "notifications" {
		t.Errorf("expected channel 'notifications', got %v", received["channel"])
	}
}

func TestPublisher_Broadcast_InvalidSlug_Returns_Error(t *testing.T) {
	nc := startTestNATS(t)
	pub := New(nc)

	err := pub.Broadcast("bad.slug", "notifications", "data")
	if err == nil {
		t.Fatal("expected error for slug with dots")
	}
}

func TestPublisher_Broadcast_InvalidChannel_Returns_Error(t *testing.T) {
	nc := startTestNATS(t)
	pub := New(nc)

	tests := []struct {
		name    string
		channel string
	}{
		{"empty", ""},
		{"dots", "some.channel"},
		{"wildcard", "notifications*"},
		{"gt", "notifications>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pub.Broadcast("saldivia", tt.channel, "data")
			if err == nil {
				t.Errorf("expected error for channel %q", tt.channel)
			}
		})
	}
}

func TestIsValidSubjectToken(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"saldivia", true},
		{"my-tenant", true},
		{"a", true},
		{"", false},
		{"has.dot", false},
		{"has*star", false},
		{"has>gt", false},
		{"has space", false},
		{"has\ttab", false},
		{"has\nnewline", false},
		{"has\rcarriage", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isValidSubjectToken(tt.input); got != tt.want {
				t.Errorf("isValidSubjectToken(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
