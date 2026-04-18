// Package handler implements HTTP handlers for the WebSocket service.
package handler

import (
	"crypto/ed25519"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/coder/websocket"

	"github.com/Camionerou/rag-saldivia/services/app/internal/httperr"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/services/app/internal/realtime/ws/hub"
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
// JWT is read from one of (in order):
//  1. Sec-WebSocket-Protocol: bearer, <token>   (browsers — only option, can't set Authorization on WS)
//  2. Authorization: Bearer <token>             (server-to-server, CLI, MCP)
//
// Query-param transport is intentionally not supported (logs would leak the token).
func (h *WS) Upgrade(w http.ResponseWriter, r *http.Request) {
	token, subprotocol := extractToken(r)
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

	// Echo the negotiated subprotocol back to the client. RFC 6455 requires
	// the server to pick one of the offered subprotocols; without this the
	// browser rejects the upgrade even though we authenticated the request.
	if subprotocol != "" {
		opts.Subprotocols = []string{subprotocol}
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

// extractToken pulls the JWT from one of the two transports the WS handler
// accepts. Returns the token plus the subprotocol the client negotiated (only
// non-empty for the Sec-WebSocket-Protocol path — the server must echo it back
// in the upgrade response or the browser drops the connection).
//
// Sec-WebSocket-Protocol carries comma-separated values; the frontend sends
// "bearer, <token>" so we accept "bearer" anywhere in the list and treat the
// remaining entry as the JWT. Only "bearer" is echoed back as the subprotocol —
// the token never appears in any header on the response.
func extractToken(r *http.Request) (token, subprotocol string) {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer "), ""
	}

	const scheme = "bearer"
	for _, h := range r.Header.Values("Sec-WebSocket-Protocol") {
		var sawScheme bool
		var candidate string
		for _, raw := range strings.Split(h, ",") {
			v := strings.TrimSpace(raw)
			if v == "" {
				continue
			}
			if strings.EqualFold(v, scheme) {
				sawScheme = true
				continue
			}
			if candidate == "" {
				candidate = v
			}
		}
		if sawScheme && candidate != "" {
			return candidate, scheme
		}
	}

	return "", ""
}

// extractBearerToken is kept for backward compatibility with tests that target
// only the Authorization header path.
func extractBearerToken(r *http.Request) string {
	t, _ := extractToken(r)
	return t
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
