package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/app/internal/core/auth/service"
)

// --- Login edge cases ---

func TestLogin_EmailTooLong_Returns400(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	// 255-character local part — exceeds RFC 5321 limit of 254 total
	longEmail := strings.Repeat("a", 246) + "@test.com" // 246 + 9 = 255 chars
	body := `{"email":"` + longEmail + `","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for email >254 chars, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_WrongCredentials_Returns401_GenericMessage(t *testing.T) {
	// Verify the error message does not leak whether the email exists
	// or whether it was the password that was wrong.
	mock := &mockAuthService{err: service.ErrInvalidCredentials}
	h := NewAuth(mock)

	body := `{"email":"user@test.com","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// Must not mention "password", "wrong", "not found", "exist", "user" specifically
	lower := strings.ToLower(resp.Error)
	for _, forbidden := range []string{"wrong password", "not found", "does not exist", "no user"} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("error message leaks information: %q contains %q", resp.Error, forbidden)
		}
	}
	if resp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestLogin_TOTPRequired_Returns200_WithChallenge(t *testing.T) {
	// When MFA is required, the handler must return 200 with mfa_required:true
	// and an mfa_token, NOT an access/refresh token pair.
	mock := &mockAuthService{
		tokens: &service.TokenPair{
			MFARequired: true,
			MFAToken:    "mfa.temp.jwt",
		},
	}
	h := NewAuth(mock)

	body := `{"email":"mfa@test.com","password":"correctpass"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for MFA challenge, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp service.TokenPair
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.MFARequired {
		t.Error("expected mfa_required: true in response")
	}
	if resp.MFAToken == "" {
		t.Error("expected non-empty mfa_token for challenge")
	}
	// Real access/refresh tokens must NOT be present
	if resp.AccessToken != "" {
		t.Errorf("access_token must not be returned when MFA is pending, got %q", resp.AccessToken)
	}
	if resp.RefreshToken != "" {
		t.Errorf("refresh_token must not be returned when MFA is pending, got %q", resp.RefreshToken)
	}

	// No refresh cookie should be set
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sda_refresh" {
			t.Error("sda_refresh cookie must not be set when MFA is pending")
		}
	}
}

// --- Refresh edge cases ---

func TestRefresh_ExpiredRefreshToken_Returns401(t *testing.T) {
	// An expired (or revoked) refresh token gets ErrInvalidRefreshToken from the service.
	mock := &mockAuthService{refreshErr: service.ErrInvalidRefreshToken}
	h := NewAuth(mock)

	body := `{"refresh_token":"expired.refresh.token"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired refresh token, got %d", rec.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestRefresh_MalformedToken_Returns400(t *testing.T) {
	// An empty body (no refresh_token field at all) should return 400.
	h := NewAuth(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty refresh_token field, got %d", rec.Code)
	}
}

func TestRefresh_WithAccessTokenInstead_Returns401(t *testing.T) {
	// Using an access token where a refresh token is expected must fail.
	// The service will not find its hash in the DB → ErrInvalidRefreshToken.
	mock := &mockAuthService{refreshErr: service.ErrInvalidRefreshToken}
	h := NewAuth(mock)

	// Simulate passing an access token as if it were a refresh token
	body := `{"refresh_token":"access.token.used.as.refresh"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when access token used as refresh, got %d", rec.Code)
	}
}

// --- Cookie security ---

func TestRefresh_InvalidToken_ClearsCookie(t *testing.T) {
	// When refresh fails the stale cookie should be cleared, not left intact.
	mock := &mockAuthService{refreshErr: service.ErrInvalidRefreshToken}
	h := NewAuth(mock)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "sda_refresh", Value: "stale.token"})
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	// Cookie must be cleared (MaxAge -1 or expired)
	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sda_refresh" && c.MaxAge == -1 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("expected stale sda_refresh cookie to be cleared on invalid refresh")
	}
}
