// Package handler implements HTTP handlers for the auth service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/Camionerou/rag-saldivia/services/app/internal/httperr"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/app/internal/core/auth/service"
)

// AuthService defines the operations the handler needs from the service layer.
type AuthService interface {
	Login(ctx context.Context, req service.LoginRequest) (*service.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*service.TokenPair, error)
	Logout(ctx context.Context, refreshToken, accessJTI string, accessExpiry time.Time) error
	Me(ctx context.Context, userID string) (*service.UserInfo, error)
	SetupMFA(ctx context.Context, userID string) (*service.MFASetupResult, error)
	VerifySetup(ctx context.Context, userID, code string) error
	DisableMFA(ctx context.Context, userID, code string) error
	CompleteMFALogin(ctx context.Context, mfaToken, code string) (*service.TokenPair, error)
	UpdateProfile(ctx context.Context, userID string, req service.UpdateProfileRequest) (*service.UserInfo, error)
	ListUsers(ctx context.Context, limit, offset int32) ([]service.UserListItem, error)
}

// Auth handles HTTP requests for authentication. Single-tenant per ADR 022 —
// the container is the tenant, so auth binds to one service + one JWT config
// at startup. Multi-tenant resolver machinery was removed with ADR 025 core
// fusion.
type Auth struct {
	authSvc         AuthService
	jwtCfg          sdajwt.Config       // used by Logout bearer verify + ServiceToken signing
	serviceTokenCfg *ServiceTokenConfig // for POST /v1/auth/service-token
}

// NewAuth creates auth HTTP handlers.
func NewAuth(authSvc AuthService) *Auth {
	return &Auth{authSvc: authSvc}
}

// SetJWTConfig wires the JWT signing/verification config. Required for
// Logout (bearer JTI blacklisting) and ServiceToken (admin token issue).
// Called by the monolith's wireAuth after handler construction.
func (h *Auth) SetJWTConfig(cfg sdajwt.Config) {
	h.jwtCfg = cfg
}

// resolveService returns the bound AuthService. Kept as a seam so per-handler
// err-handling stays uniform, but in single-tenant mode it never fails.
func (h *Auth) resolveService(_ *http.Request) (AuthService, error) {
	return h.authSvc, nil
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Login handles POST /v1/auth/login
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1MB to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Email == "" || req.Password == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("email and password are required"))
		return
	}
	if len(req.Email) > 254 {
		httperr.WriteError(w, r, httperr.InvalidInput("email too long"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	tokens, err := svc.Login(r.Context(), service.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		IP:        r.RemoteAddr, // chi's RealIP middleware already rewrites this
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			httperr.WriteError(w, r, httperr.Unauthorized("invalid email or password"))
		case errors.Is(err, service.ErrAccountLocked):
			httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeForbidden, "too many attempts, try again later", http.StatusTooManyRequests))
		default:
			httperr.WriteError(w, r, httperr.Internal(err))
		}
		return
	}

	// MFA required — return challenge, not tokens
	if tokens.MFARequired {
		writeJSON(w, http.StatusOK, tokens)
		return
	}

	// Set refresh token as HttpOnly cookie for browser clients
	setRefreshCookie(w, tokens.RefreshToken, tokens.RefreshExpiresAt)

	// Return access token in body (refresh token also in body for CLI/MCP clients)
	writeJSON(w, http.StatusOK, tokens)
}

// Refresh handles POST /v1/auth/refresh
func (h *Auth) Refresh(w http.ResponseWriter, r *http.Request) {
	// Read refresh token from HttpOnly cookie first, fall back to body
	refreshToken := ""
	if cookie, err := r.Cookie("sda_refresh"); err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken == "" {
		// Fall back to JSON body for non-browser clients (CLI, MCP)
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("refresh token is required"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	tokens, err := svc.Refresh(r.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			clearRefreshCookie(w)
			httperr.WriteError(w, r, httperr.Unauthorized("invalid or expired refresh token"))
		default:
			httperr.WriteError(w, r, httperr.Internal(err))
		}
		return
	}

	setRefreshCookie(w, tokens.RefreshToken, tokens.RefreshExpiresAt)
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// Logout handles POST /v1/auth/logout
func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookie, err := r.Cookie("sda_refresh"); err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken == "" {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken != "" {
		if svc, err := h.resolveService(r); err == nil {
			// Extract access token JTI for blacklisting
			accessJTI := ""
			accessExpiry := time.Now().Add(15 * time.Minute) // default
			if bearer := r.Header.Get("Authorization"); len(bearer) > 7 {
				if claims, err := sdajwt.Verify(h.jwtCfg.PublicKey, bearer[7:]); err == nil {
					accessJTI = claims.ID
					if claims.ExpiresAt != nil {
						accessExpiry = claims.ExpiresAt.Time
					}
				}
			}
			_ = svc.Logout(r.Context(), refreshToken, accessJTI, accessExpiry)
		}
	}

	clearRefreshCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// Me handles GET /v1/auth/me
func (h *Auth) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("not authenticated"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	user, err := svc.Me(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			httperr.WriteError(w, r, httperr.NotFound("user"))
		default:
			httperr.WriteError(w, r, httperr.Internal(err))
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// EnabledModules handles GET /v1/modules/enabled. Returns the silo's core
// module set. Per-tenant module resolution was dropped with ADR 022 (one
// container per tenant) — the enabled set is baked into the deploy, not
// looked up per-request.
func (h *Auth) EnabledModules(w http.ResponseWriter, _ *http.Request) {
	type enabledModule struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	writeJSON(w, http.StatusOK, []enabledModule{
		{ID: "chat", Name: "Chat + RAG", Category: "core"},
		{ID: "auth", Name: "Auth + RBAC", Category: "core"},
		{ID: "notifications", Name: "Notificaciones", Category: "core"},
	})
}

// Health handles GET /health
func (h *Auth) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "auth"})
}

// setRefreshCookie sets the refresh token as an HttpOnly cookie.
// SDA_COOKIE_SECURE=false → Secure=false + SameSite=Lax, so the cookie is
// usable over plain HTTP (dev, VPN/IP same-origin, etc.). Default: secure.
func setRefreshCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	secure := os.Getenv("SDA_COOKIE_SECURE") != "false"
	sameSite := http.SameSiteStrictMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sda_refresh",
		Value:    token,
		Path:     "/v1/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Expires:  expiresAt,
	})
}

// clearRefreshCookie removes the refresh cookie. Mirrors setRefreshCookie's
// Secure/SameSite policy so the clear works over HTTP too — browsers ignore
// Set-Cookie with Secure when the response is over HTTP.
func clearRefreshCookie(w http.ResponseWriter) {
	secure := os.Getenv("SDA_COOKIE_SECURE") != "false"
	sameSite := http.SameSiteStrictMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sda_refresh",
		Value:    "",
		Path:     "/v1/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}

// UpdateMe handles PATCH /v1/auth/me
func (h *Auth) UpdateMe(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("not authenticated"))
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	user, err := svc.UpdateProfile(r.Context(), userID, service.UpdateProfileRequest{Name: req.Name})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			httperr.WriteError(w, r, httperr.NotFound("user"))
		case errors.Is(err, service.ErrValidation):
			httperr.WriteError(w, r, httperr.InvalidInput(err.Error()))
		default:
			httperr.WriteError(w, r, httperr.Internal(err))
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ListUsers handles GET /v1/auth/users — returns active users for the tenant (paginated).
func (h *Auth) ListUsers(w http.ResponseWriter, r *http.Request) {
	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	pg := pagination.Parse(r)
	users, err := svc.ListUsers(r.Context(), int32(pg.Limit()), int32(pg.Offset()))
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
