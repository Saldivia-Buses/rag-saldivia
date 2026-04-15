package spine_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/spine"
)

func TestPublish_NilPublisher(t *testing.T) {
	env, _ := spine.New("t", "x", 1, fakePayload{})
	err := spine.Publish(context.Background(), nil, "tenant.t.notify.x", env)
	if err == nil || !strings.Contains(err.Error(), "nil publisher") {
		t.Errorf("expected nil publisher error, got %v", err)
	}
}

func TestPublish_InvalidSubject(t *testing.T) {
	// Nil nc: ValidateSubject runs first and fails, never reaches the PublishMsg call.
	pub := spine.NewPublisher(nil)
	env, _ := spine.New("t", "x", 1, fakePayload{})
	err := spine.Publish(context.Background(), pub, "bad subject", env)
	if err == nil {
		t.Fatal("expected error for invalid subject")
	}
}

// optionalNATS returns a live NATS connection or skips the test if none is
// running locally. Mirrors the pattern used by pkg/nats/publisher_test.go.
func optionalNATS(t *testing.T) *nats.Conn {
	t.Helper()
	nc, err := nats.Connect(nats.DefaultURL,
		nats.Timeout(1*time.Second),
		nats.MaxReconnects(0),
	)
	if err != nil {
		t.Skipf("NATS not available: %v", err)
	}
	t.Cleanup(nc.Close)
	return nc
}

func TestPublish_RoundTrip(t *testing.T) {
	nc := optionalNATS(t)
	pub := spine.NewPublisher(nc)

	received := make(chan []byte, 1)
	sub, err := nc.Subscribe("tenant.test_t.notify.chat.new_message", func(m *nats.Msg) {
		received <- m.Data
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer func() { _ = sub.Unsubscribe() }()

	env, err := spine.New("test_t", "chat.new_message", 1, fakePayload{UserID: "u1", Msg: "hi"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := spine.Publish(context.Background(), pub, "tenant.test_t.notify.chat.new_message", env); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case data := <-received:
		decoded, err := spine.Decode[fakePayload](data)
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		if decoded.Payload.UserID != "u1" {
			t.Errorf("roundtrip payload mismatch: %+v", decoded.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}
