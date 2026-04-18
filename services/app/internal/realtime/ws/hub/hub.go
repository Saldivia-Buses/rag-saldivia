package hub

import (
	"encoding/json"
	"log/slog"
	"sync"
)

const (
	defaultMaxClients          = 1000
	defaultMaxClientsPerTenant = 300
)

// Hub manages all WebSocket connections and channel subscriptions.
type Hub struct {
	clients    map[*Client]bool
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client

	MaxClients          int
	MaxClientsPerTenant int
	Mutations           *MutationHandler // nil = mutations disabled
}

// New creates a new Hub with default connection limits.
func New() *Hub {
	return &Hub{
		clients:             make(map[*Client]bool),
		register:            make(chan *Client, 32),
		unregister:          make(chan *Client, 32),
		MaxClients:          defaultMaxClients,
		MaxClientsPerTenant: defaultMaxClientsPerTenant,
	}
}

// Run starts the Hub event loop. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Check global limit
			if len(h.clients) >= h.MaxClients {
				h.mu.Unlock()
				client.SendMessage(Message{Type: Error, Error: "server at capacity"})
				client.markClosed()
				continue
			}
			// Check per-tenant limit
			tenantCount := 0
			for c := range h.clients {
				if c.Slug == client.Slug {
					tenantCount++
				}
			}
			if tenantCount >= h.MaxClientsPerTenant {
				h.mu.Unlock()
				client.SendMessage(Message{Type: Error, Error: "tenant connection limit reached"})
				client.markClosed()
				continue
			}
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
				client.markClosed()
			}
			h.mu.Unlock()
			slog.Info("client disconnected",
				"user_id", client.UserID,
				"tenant", client.Slug,
				"total_clients", h.ClientCount())
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

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.IsSubscribed(channel) {
			client.TrySend(data)
		}
	}
}

// BroadcastToTenant sends a message to all clients of a specific tenant
// that are subscribed to the channel. Uses TrySend to avoid write-after-close panic.
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
			client.TrySend(data)
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
		if !client.Subscribe(msg.Channel) {
			client.SendMessage(Message{Type: Error, ID: msg.ID, Error: "max subscriptions reached"})
			return
		}
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
		if h.Mutations == nil {
			client.SendMessage(Message{Type: Error, ID: msg.ID, Error: "mutations not available"})
			return
		}
		h.Mutations.Handle(client, msg)

	default:
		client.SendMessage(Message{
			Type:  Error,
			ID:    msg.ID,
			Error: "unknown message type: " + string(msg.Type),
		})
	}
}
