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
		sub.Unsubscribe()
	}
}

// handleNATSMessage receives a NATS message and broadcasts to WebSocket clients.
// Subject format: tenant.{slug}.{channel} (e.g., tenant.saldivia.chat.messages)
func (b *NATSBridge) handleNATSMessage(msg *nats.Msg) {
	// Parse subject: tenant.{slug}.{rest...}
	parts := strings.SplitN(msg.Subject, ".", 3)
	if len(parts) < 3 {
		slog.Warn("invalid NATS subject format", "subject", msg.Subject)
		return
	}

	tenantSlug := parts[1]
	channel := parts[2] // e.g., "chat.messages", "notifications", "module.enabled"

	// Parse the NATS payload as a WebSocket message
	var wsMsg Message
	if err := json.Unmarshal(msg.Data, &wsMsg); err != nil {
		// If not a Message, wrap the raw data as an event
		wsMsg = Message{
			Type:    Event,
			Channel: channel,
			Data:    msg.Data,
		}
	} else {
		// Ensure channel is set
		if wsMsg.Channel == "" {
			wsMsg.Channel = channel
		}
		if wsMsg.Type == "" {
			wsMsg.Type = Event
		}
	}

	b.hub.BroadcastToTenant(tenantSlug, channel, wsMsg)
}
