package hub

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// Hub manages all WebSocket connections and channel subscriptions.
// It routes messages between clients and NATS.
type Hub struct {
	clients    map[*Client]bool
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

type broadcastMsg struct {
	channel string
	data    []byte
}

// New creates a new Hub.
func New() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client, 32),
		unregister: make(chan *Client, 32),
		broadcast:  make(chan broadcastMsg, 256),
	}
}

// Run starts the Hub event loop. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			slog.Info("client connected",
				"user_id", client.UserID,
				"tenant", client.Slug,
				"total_clients", h.ClientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			slog.Info("client disconnected",
				"user_id", client.UserID,
				"tenant", client.Slug,
				"total_clients", h.ClientCount())

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.IsSubscribed(msg.channel) {
					select {
					case client.send <- msg.data:
					default:
						// Buffer full — will be cleaned up
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Broadcast sends a message to all clients subscribed to a channel.
func (h *Hub) Broadcast(channel string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal broadcast", "error", err, "channel", channel)
		return
	}
	h.broadcast <- broadcastMsg{channel: channel, data: data}
}

// BroadcastToTenant sends a message to all clients of a specific tenant
// that are subscribed to the channel.
func (h *Hub) BroadcastToTenant(tenantSlug, channel string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal broadcast", "error", err, "channel", channel)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.Slug == tenantSlug && client.IsSubscribed(channel) {
			select {
			case client.send <- data:
			default:
			}
		}
	}
}

// ClientCount returns the total number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ClientCountByTenant returns connected clients for a specific tenant.
func (h *Hub) ClientCountByTenant(slug string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for client := range h.clients {
		if client.Slug == slug {
			count++
		}
	}
	return count
}

// handleMessage processes an incoming message from a client.
func (h *Hub) handleMessage(client *Client, msg Message) {
	switch msg.Type {
	case Subscribe:
		if msg.Channel == "" {
			client.SendMessage(Message{Type: Error, ID: msg.ID, Error: "channel is required"})
			return
		}
		client.Subscribe(msg.Channel)
		client.SendMessage(Message{
			Type:    Event,
			Channel: msg.Channel,
			ID:      msg.ID,
			Data:    json.RawMessage(`{"subscribed":true}`),
		})
		slog.Debug("client subscribed", "user_id", client.UserID, "channel", msg.Channel)

	case Unsubscribe:
		client.Unsubscribe(msg.Channel)
		client.SendMessage(Message{
			Type:    Event,
			Channel: msg.Channel,
			ID:      msg.ID,
			Data:    json.RawMessage(`{"subscribed":false}`),
		})

	case Mutation:
		// Mutations are forwarded to the appropriate service via gRPC/HTTP.
		// For now, return an error — will be implemented per service.
		client.SendMessage(Message{
			Type:  Error,
			ID:    msg.ID,
			Error: "mutations not yet implemented",
		})

	default:
		client.SendMessage(Message{
			Type:  Error,
			ID:    msg.ID,
			Error: "unknown message type: " + string(msg.Type),
		})
	}
}
