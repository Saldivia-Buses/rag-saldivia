package hub

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
)

// NATSBridge subscribes to NATS subjects and forwards events to WebSocket clients.
// NATS subjects are namespaced by tenant: tenant.{slug}.{service}.{event}
// The bridge extracts the tenant slug and channel from the subject, then
// broadcasts to clients of that tenant subscribed to that channel.
type NATSBridge struct {
	hub  *Hub
	conn *nats.Conn
	subs []*nats.Subscription
}

// NewNATSBridge creates a bridge between NATS and the WebSocket Hub.
func NewNATSBridge(hub *Hub, nc *nats.Conn) *NATSBridge {
	return &NATSBridge{
		hub:  hub,
		conn: nc,
	}
}

// Start subscribes to all tenant event subjects.
// Pattern: tenant.*.> matches all tenant events.
func (b *NATSBridge) Start() error {
	// Subscribe to all tenant events
	sub, err := b.conn.Subscribe("tenant.*.>", b.handleNATSMessage)
	if err != nil {
		return err
	}
	b.subs = append(b.subs, sub)

	slog.Info("NATS bridge started", "subject", "tenant.*.>")
	return nil
}

// Stop unsubscribes from all NATS subjects.
func (b *NATSBridge) Stop() {
	for _, sub := range b.subs {
		_ = sub.Unsubscribe()
	}
}

// handleNATSMessage receives a NATS message and broadcasts to WebSocket clients.
// Subject format: tenant.{slug}.{channel} (e.g., tenant.saldivia.chat.messages)
//
// Three wire shapes are supported for backwards compatibility:
//  1. Spine envelope — detected by the presence of id + schema_version.
//     The payload is extracted and forwarded as Data.
//  2. WS Message — a fully-formed Message is forwarded as-is (channel filled in).
//  3. Raw opaque JSON — wrapped as Event with the data untouched.
func (b *NATSBridge) handleNATSMessage(msg *nats.Msg) {
	parts := strings.SplitN(msg.Subject, ".", 3)
	if len(parts) < 3 {
		slog.Warn("invalid NATS subject format", "subject", msg.Subject)
		return
	}

	tenantSlug := parts[1]
	channel := parts[2]

	if wsMsg, ok := unwrapEnvelope(msg.Data, channel); ok {
		b.hub.BroadcastToTenant(tenantSlug, channel, wsMsg)
		return
	}

	var wsMsg Message
	if err := json.Unmarshal(msg.Data, &wsMsg); err != nil {
		wsMsg = Message{Type: Event, Channel: channel, Data: msg.Data}
	} else {
		if wsMsg.Channel == "" {
			wsMsg.Channel = channel
		}
		if wsMsg.Type == "" {
			wsMsg.Type = Event
		}
	}
	b.hub.BroadcastToTenant(tenantSlug, channel, wsMsg)
}

// unwrapEnvelope detects a spine envelope by the presence of id and
// schema_version fields and returns a WS Message with the payload as Data.
// Returns (_, false) if the message is not a spine envelope.
//
// The detection is intentionally lenient — we do not import pkg/spine here
// to keep the ws service's dependency graph minimal. The contract (id,
// tenant_id, type, schema_version, payload) is documented in
// docs/packages/spine.md.
func unwrapEnvelope(data []byte, fallbackChannel string) (Message, bool) {
	var probe struct {
		ID            string          `json:"id"`
		SchemaVersion uint8           `json:"schema_version"`
		Type          string          `json:"type"`
		Payload       json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return Message{}, false
	}
	if probe.ID == "" || probe.SchemaVersion == 0 || len(probe.Payload) == 0 {
		return Message{}, false
	}
	return Message{
		Type:    Event,
		Channel: fallbackChannel,
		Action:  probe.Type,
		Data:    probe.Payload,
	}, true
}
