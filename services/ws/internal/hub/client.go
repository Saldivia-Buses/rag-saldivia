package hub

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	writeTimeout = 10 * time.Second
	sendBufSize  = 64
)

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	subs   map[string]bool // channels this client is subscribed to
	mu     sync.RWMutex

	// Identity (set after JWT verification)
	UserID   string
	Email    string
	TenantID string
	Slug     string
	Role     string
}

// NewClientWithIdentity creates a client with pre-set identity from JWT claims.
func NewClientWithIdentity(hub *Hub, conn *websocket.Conn, userID, email, tenantID, slug, role string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBufSize),
		subs:     make(map[string]bool),
		UserID:   userID,
		Email:    email,
		TenantID: tenantID,
		Slug:     slug,
		Role:     role,
	}
}

// Subscribe adds a channel subscription.
func (c *Client) Subscribe(channel string) {
	c.mu.Lock()
	c.subs[channel] = true
	c.mu.Unlock()
}

// Unsubscribe removes a channel subscription.
func (c *Client) Unsubscribe(channel string) {
	c.mu.Lock()
	delete(c.subs, channel)
	c.mu.Unlock()
}

// IsSubscribed checks if the client is subscribed to a channel.
func (c *Client) IsSubscribed(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subs[channel]
}

// Channels returns a copy of the client's subscribed channels.
func (c *Client) Channels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	chs := make([]string, 0, len(c.subs))
	for ch := range c.subs {
		chs = append(chs, ch)
	}
	return chs
}

// SendMessage sends a message to this client. Non-blocking — drops if buffer full.
func (c *Client) SendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal message", "error", err)
		return
	}

	select {
	case c.send <- data:
	default:
		slog.Warn("client send buffer full, dropping message",
			"user_id", c.UserID, "channel", msg.Channel)
	}
}

// ReadPump reads messages from the WebSocket connection and dispatches them.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				return
			}
			slog.Debug("read error", "error", err, "user_id", c.UserID)
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.SendMessage(Message{
				Type:  Error,
				Error: "invalid message format",
			})
			continue
		}

		c.hub.handleMessage(c, msg)
	}
}

// WritePump sends messages from the send channel to the WebSocket connection.
func (c *Client) WritePump(ctx context.Context) {
	defer c.conn.Close(websocket.StatusNormalClosure, "")

	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(writeCtx, websocket.MessageText, data)
			cancel()
			if err != nil {
				slog.Debug("write error", "error", err, "user_id", c.UserID)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
