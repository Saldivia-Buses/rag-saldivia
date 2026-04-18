// Package hub implements the WebSocket Hub — the real-time backbone of SDA.
// All client-server communication flows through the Hub. Services push events
// via NATS, the Hub forwards them to connected clients.
package hub

import "encoding/json"

// MessageType defines the three types of WebSocket messages.
type MessageType string

const (
	// Subscribe — client asks to receive live updates for a channel.
	Subscribe MessageType = "subscribe"
	// Unsubscribe — client stops receiving updates for a channel.
	Unsubscribe MessageType = "unsubscribe"
	// Mutation — client asks to execute an action (create, update, delete).
	Mutation MessageType = "mutation"
	// Event — server pushes data or notifications to the client.
	Event MessageType = "event"
	// Error — server reports an error to the client.
	Error MessageType = "error"
)

// Message is the envelope for all WebSocket communication.
type Message struct {
	Type    MessageType     `json:"type"`
	Channel string          `json:"channel,omitempty"` // e.g., "chat.messages:session-123"
	Action  string          `json:"action,omitempty"`  // for mutations: "create_session", "send_message"
	ID      string          `json:"id,omitempty"`      // correlation ID for request/response matching
	Data    json.RawMessage `json:"data,omitempty"`    // payload
	Error   string          `json:"error,omitempty"`   // error message (only for Error type)
}

// Common channel prefixes.
const (
	ChannelSessions      = "sessions"
	ChannelChatMessages  = "chat.messages"     // + ":session_id"
	ChannelNotifications = "notifications"
	ChannelAdminStats    = "admin.stats"
	ChannelIngestJobs    = "ingest.jobs"
	ChannelPresence      = "presence"
	ChannelCollections   = "collections"
	ChannelModules       = "modules"
	ChannelFleet         = "fleet.vehicles"
	ChannelMaintenance   = "fleet.maintenance"
)
