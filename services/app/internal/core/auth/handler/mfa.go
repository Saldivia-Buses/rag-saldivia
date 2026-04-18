package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Camionerou/rag-saldivia/services/app/internal/httperr"
	"github.com/Camionerou/rag-saldivia/services/app/internal/core/auth/service"
)

type mfaCodeRequest struct {
	Code string `json:"code"`
}

type mfaVerifyLoginRequest struct {
	MFAToken string `json:"mfa_token"`
	Code     string `json:"code"`
}

// SetupMFA handles POST /v1/auth/mfa/setup
// Returns the TOTP secret and URI for QR code generation.
// Requires: valid access token (user must be logged in).
func (h *Auth) SetupMFA(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("authentication required"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	result, err := svc.SetupMFA(r.Context(), userID)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// VerifySetup handles POST /v1/auth/mfa/verify-setup
// Activates MFA after the user proves they can generate valid codes.
// Requires: valid access token + valid TOTP code.
func (h *Auth) VerifySetup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("authentication required"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req mfaCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("code is required"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	if err := svc.VerifySetup(r.Context(), userID, req.Code); err != nil {
		if errors.Is(err, service.ErrInvalidMFACode) {
			httperr.WriteError(w, r, httperr.Unauthorized("invalid code"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "mfa_enabled"})
}

// VerifyMFALogin handles POST /v1/auth/mfa/verify
// Completes a MFA-gated login by verifying the TOTP code.
// Requires: mfa_token (temp JWT from login) + valid TOTP code.
// Returns: access_token + refresh_token.
func (h *Auth) VerifyMFALogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req mfaVerifyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MFAToken == "" || req.Code == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("mfa_token and code are required"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	tokens, err := svc.CompleteMFALogin(r.Context(), req.MFAToken, req.Code)
	if err != nil {
		if errors.Is(err, service.ErrInvalidMFACode) {
			httperr.WriteError(w, r, httperr.Unauthorized("invalid MFA code"))
			return
		}
		httperr.WriteError(w, r, httperr.Unauthorized("invalid or expired MFA token"))
		return
	}

	setRefreshCookie(w, tokens.RefreshToken, tokens.RefreshExpiresAt)
	writeJSON(w, http.StatusOK, tokens)
}

// DisableMFA handles POST /v1/auth/mfa/disable
// Disables MFA for the user. Requires valid access token + valid TOTP code as confirmation.
func (h *Auth) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("authentication required"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req mfaCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("code is required"))
		return
	}

	svc, err := h.resolveService(r)
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "tenant not available", http.StatusBadGateway))
		return
	}

	if err := svc.DisableMFA(r.Context(), userID, req.Code); err != nil {
		if errors.Is(err, service.ErrInvalidMFACode) {
			httperr.WriteError(w, r, httperr.Unauthorized("invalid code"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "mfa_disabled"})
}
