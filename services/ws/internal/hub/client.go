package hub

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

const (
	writeTimeout   = 10 * time.Second
	sendBufSize    = 64
	maxSubscriptions = 64
)

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	subs   map[string]bool // channels this client is subscribed to
	mu     sync.RWMutex
	closed atomic.Bool // set when send channel is closed, prevents write-after-close panic

	// Identity (set after JWT verification)
	UserID   string
	Email    string
	TenantID string
	Slug     string
	Role     string
	JWT      string // raw token for forwarding to gRPC services
}

// NewClientWithIdentity creates a client with pre-set identity from JWT claims.
func NewClientWithIdentity(hub *Hub, conn *websocket.Conn, userID, email, tenantID, slug, role, jwt string) *Client {
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
		JWT:      jwt,
	}
}

// Subscribe adds a channel subscription. Returns false if max subscriptions reached.
func (c *Client) Subscribe(channel string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.subs) >= maxSubscriptions && !c.subs[channel] {
		return false
	}
	c.subs[channel] = true
	return true
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

// TrySend attempts to send data to the client. Returns false if the client
// is closed or the buffer is full. Safe to call concurrently with Close.
func (c *Client) TrySend(data []byte) bool {
	if c.closed.Load() {
		return false
	}
	select {
	case c.send <- data:
		return true
	default:
		slog.Warn("client send buffer full, dropping message", "user_id", c.UserID)
		return false
	}
}

// SendMessage sends a structured message to this client. Non-blocking.
func (c *Client) SendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal message", "error", err)
		return
	}
	c.TrySend(data)
}

// markClosed marks the client as closed and closes the send channel.
// Must only be called once (by the hub's unregister handler).
func (c *Client) markClosed() {
	if c.closed.CompareAndSwap(false, true) {
		close(c.send)
	}
}

// ReadPump reads messages from the WebSocket connection and dispatches them.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close(websocket.StatusNormalClosure, "")
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
	defer func() { _ = c.conn.Close(websocket.StatusNormalClosure, "") }()

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
