package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

// ServiceTokenConfig configures the service token endpoint.
type ServiceTokenConfig struct {
	// Key is the shared secret that service accounts present.
	// Must be at least 32 bytes.
	Key string
	// PlatformTenantID is the tenant UUID for the platform.
	PlatformTenantID string
	// PlatformSlug is the tenant slug for the platform.
	PlatformSlug string
}

// SetServiceTokenConfig configures the service token endpoint.
// If not called, ServiceToken returns 501 Not Implemented.
func (h *Auth) SetServiceTokenConfig(cfg ServiceTokenConfig) {
	h.serviceTokenCfg = &cfg
}

type serviceTokenRequest struct {
	Service string `json:"service"`
	Key     string `json:"key"`
}

type serviceTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// ServiceToken handles POST /v1/auth/service-token
// Issues a short-lived (5 min) platform admin JWT for machine-to-machine calls.
// Authenticated via a shared service account key, not via JWT.
func (h *Auth) ServiceToken(w http.ResponseWriter, r *http.Request) {
	if h.serviceTokenCfg == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal,
			"service token endpoint not configured", http.StatusNotImplemented))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req serviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Service == "" || req.Key == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("service and key are required"))
		return
	}

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(req.Key), []byte(h.serviceTokenCfg.Key)) != 1 {
		httperr.WriteError(w, r, httperr.Unauthorized("invalid service account key"))
		return
	}

	// Issue a short-lived token (5 min) with platform admin claims
	cfg := h.jwtCfg
	cfg.AccessExpiry = 5 * time.Minute

	expiresAt := time.Now().Add(cfg.AccessExpiry)
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID:   "svc:" + req.Service,
		Email:    req.Service + "@sda.internal",
		Name:     req.Service + " service account",
		TenantID: h.serviceTokenCfg.PlatformTenantID,
		Slug:     h.serviceTokenCfg.PlatformSlug,
		Role:     "admin",
	})
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	writeJSON(w, http.StatusOK, serviceTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}
