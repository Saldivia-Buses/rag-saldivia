// Package handler implements HTTP handlers for the WebSocket service.
package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/coder/websocket"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/services/ws/internal/hub"
)

// WS handles WebSocket upgrade requests.
type WS struct {
	hub       *hub.Hub
	jwtSecret string
}

// NewWS creates a WebSocket handler.
func NewWS(h *hub.Hub, jwtSecret string) *WS {
	return &WS{hub: h, jwtSecret: jwtSecret}
}

// Upgrade handles GET /ws — upgrades HTTP to WebSocket.
// The client must provide a valid JWT either as:
//   - Authorization: Bearer <token> header
//   - ?token=<token> query parameter (for browsers that can't set headers on WS)
func (h *WS) Upgrade(w http.ResponseWriter, r *http.Request) {
	// Extract JWT
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing authentication token", http.StatusUnauthorized)
		return
	}

	// Verify JWT
	claims, err := sdajwt.Verify(h.jwtSecret, token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Accept WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// In production, set InsecureSkipVerify to false and configure allowed origins.
		// For dev, accept all origins.
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}

	// Create client with identity from JWT
	client := hub.NewClientWithIdentity(h.hub, conn, claims.UserID, claims.Email, claims.TenantID, claims.Slug, claims.Role)

	// Register with hub
	h.hub.Register(client)

	// Auto-subscribe to tenant-wide channels
	client.Subscribe(hub.ChannelNotifications)
	client.Subscribe(hub.ChannelModules)
	client.Subscribe(hub.ChannelPresence)

	// Start read/write pumps
	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx) // blocks until disconnect
}

func extractToken(r *http.Request) string {
	// Try Authorization header first
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Fallback to query parameter (for browser WebSocket API)
	return r.URL.Query().Get("token")
}
