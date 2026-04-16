// Package handler implements HTTP handlers for the WebSocket service.
package handler

import (
	"crypto/ed25519"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/coder/websocket"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/services/ws/internal/hub"
)

// WS handles WebSocket upgrade requests.
type WS struct {
	hub       *hub.Hub
	publicKey ed25519.PublicKey
	blacklist *security.TokenBlacklist
	origins   []string // allowed origins (e.g., "*.sda.app", "localhost:*")
}

// NewWS creates a WebSocket handler. Blacklist can be nil (revocation disabled).
func NewWS(h *hub.Hub, publicKey ed25519.PublicKey, bl *security.TokenBlacklist) *WS {
	origins := parseOrigins(os.Getenv("WS_ALLOWED_ORIGINS"))
	return &WS{hub: h, publicKey: publicKey, blacklist: bl, origins: origins}
}

// Upgrade handles GET /ws — upgrades HTTP to WebSocket.
// The client must provide a valid JWT as Authorization: Bearer <token> header.
func (h *WS) Upgrade(w http.ResponseWriter, r *http.Request) {
	// Extract JWT from Authorization header only (not query param, to avoid log leakage)
	token := extractBearerToken(r)
	if token == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("missing authorization token"))
		return
	}

	// Verify JWT
	claims, err := sdajwt.Verify(h.publicKey, token)
	if err != nil {
		httperr.WriteError(w, r, httperr.Unauthorized("invalid token"))
		return
	}

	// Check token blacklist (revoked tokens from logout/password change)
	if h.blacklist != nil && claims.ID != "" {
		if revoked, _ := h.blacklist.IsRevoked(r.Context(), claims.ID); revoked {
			httperr.WriteError(w, r, httperr.Unauthorized("token revoked"))
			return
		}
	}

	// Accept WebSocket connection with origin check
	opts := &websocket.AcceptOptions{}
	if len(h.origins) > 0 {
		opts.OriginPatterns = h.origins
	} else {
		// Dev mode: no origins configured → accept all (log warning)
		opts.InsecureSkipVerify = true
		slog.Warn("WS_ALLOWED_ORIGINS not set, accepting all origins (dev mode)")
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}

	// Create client with identity from JWT
	client := hub.NewClientWithIdentity(h.hub, conn, claims.UserID, claims.Email, claims.TenantID, claims.Slug, claims.Role, token)

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

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// parseOrigins splits a comma-separated string of origin patterns.
// e.g., "*.sda.app,localhost:3000" → ["*.sda.app", "localhost:3000"]
func parseOrigins(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}
