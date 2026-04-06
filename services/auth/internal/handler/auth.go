// Package handler implements HTTP handlers for the auth service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/service"
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

// EventPublisher can publish notification events via NATS.
type EventPublisher interface {
	Notify(tenantSlug string, evt any) error
	Broadcast(tenantSlug, channel string, data any) error
}

// Auth handles HTTP requests for authentication.
type Auth struct {
	authSvc   AuthService     // static service (single-tenant mode)
	resolver  *tenant.Resolver // per-request resolution (multi-tenant mode)
	jwtCfg    sdajwt.Config
	publisher EventPublisher
}

// NewAuth creates auth HTTP handlers in single-tenant mode.
func NewAuth(authSvc AuthService) *Auth {
	return &Auth{authSvc: authSvc}
}

// NewMultiTenantAuth creates auth HTTP handlers that resolve the tenant DB per request.
func NewMultiTenantAuth(resolver *tenant.Resolver, jwtCfg sdajwt.Config, publisher EventPublisher) *Auth {
	return &Auth{
		resolver:  resolver,
		jwtCfg:    jwtCfg,
		publisher: publisher,
	}
}

// resolveService returns the AuthService for the current request's tenant.
// In single-tenant mode, returns the static service. In multi-tenant mode,
// resolves the tenant DB from the X-Tenant-Slug header.
func (h *Auth) resolveService(r *http.Request) (AuthService, error) {
	if h.authSvc != nil {
		return h.authSvc, nil
	}

	slug := r.Header.Get("X-Tenant-Slug")
	if slug == "" {
		return nil, errors.New("missing tenant context")
	}

	pool, err := h.resolver.PostgresPool(r.Context(), slug)
	if err != nil {
		return nil, err
	}

	tenantID, err := h.resolver.TenantID(r.Context(), slug)
	if err != nil {
		return nil, err
	}

	return service.NewAuth(pool, h.jwtCfg, tenantID, slug, h.publisher), nil
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
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password are required"})
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		slog.Error("failed to resolve tenant for login", "error", err)
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "tenant not available"})
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
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid email or password"})
		case errors.Is(err, service.ErrAccountLocked):
			writeJSON(w, http.StatusTooManyRequests, errorResponse{Error: "too many attempts, try again later"})
		default:
			reqID := middleware.GetReqID(r.Context())
			slog.Error("login failed", "error", err, "request_id", reqID)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
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
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "refresh token is required"})
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "tenant not available"})
		return
	}

	tokens, err := svc.Refresh(r.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			clearRefreshCookie(w)
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid or expired refresh token"})
		default:
			reqID := middleware.GetReqID(r.Context())
			slog.Error("refresh failed", "error", err, "request_id", reqID)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
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
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "not authenticated"})
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "tenant not available"})
		return
	}

	user, err := svc.Me(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
		default:
			reqID := middleware.GetReqID(r.Context())
			slog.Error("me failed", "error", err, "request_id", reqID)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// EnabledModules handles GET /v1/modules/enabled
// Returns the list of modules enabled for the current tenant.
// TODO: Read from Platform DB via tenant_modules table. Currently returns
// core modules as a baseline until Platform Service integration.
func (h *Auth) EnabledModules(w http.ResponseWriter, r *http.Request) {
	type moduleEntry struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
	}

	// Core modules — always enabled for all tenants
	modules := []moduleEntry{
		{ID: "chat", Name: "Chat", Category: "core"},
		{ID: "rag", Name: "RAG", Category: "core"},
		{ID: "notifications", Name: "Notificaciones", Category: "core"},
		{ID: "ingest", Name: "Ingesta", Category: "core"},
	}

	writeJSON(w, http.StatusOK, modules)
}

// Health handles GET /health
func (h *Auth) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "auth"})
}

// setRefreshCookie sets the refresh token as an HttpOnly secure cookie.
func setRefreshCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sda_refresh",
		Value:    token,
		Path:     "/v1/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expiresAt,
	})
}

// clearRefreshCookie removes the refresh cookie.
func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sda_refresh",
		Value:    "",
		Path:     "/v1/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// UpdateMe handles PATCH /v1/auth/me
func (h *Auth) UpdateMe(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "not authenticated"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "tenant not available"})
		return
	}

	user, err := svc.UpdateProfile(r.Context(), userID, service.UpdateProfileRequest{Name: req.Name})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
		case errors.Is(err, service.ErrValidation):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		default:
			slog.Error("update profile failed", "error", err, "user_id", userID)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ListUsers handles GET /v1/auth/users — returns active users for the tenant (paginated).
func (h *Auth) ListUsers(w http.ResponseWriter, r *http.Request) {
	svc, err := h.resolveService(r)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "tenant not available"})
		return
	}

	pg := pagination.Parse(r)
	users, err := svc.ListUsers(r.Context(), int32(pg.Limit()), int32(pg.Offset()))
	if err != nil {
		reqID := middleware.GetReqID(r.Context())
		slog.Error("list users failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
