package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Camionerou/rag-saldivia/services/auth/internal/service"
)

// --- mock ---

type mockAuthService struct {
	tokens      *service.TokenPair
	err         error
	lastReq     service.LoginRequest
	refreshErr  error
	logoutErr   error
	userInfo    *service.UserInfo
	meErr       error
}

func (m *mockAuthService) Login(_ context.Context, req service.LoginRequest) (*service.TokenPair, error) {
	m.lastReq = req
	if m.err != nil {
		return nil, m.err
	}
	return m.tokens, nil
}

func (m *mockAuthService) Refresh(_ context.Context, _ string) (*service.TokenPair, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.tokens, nil
}

func (m *mockAuthService) Logout(_ context.Context, _, _ string, _ time.Time) error {
	return m.logoutErr
}

func (m *mockAuthService) Me(_ context.Context, _ string) (*service.UserInfo, error) {
	if m.meErr != nil {
		return nil, m.meErr
	}
	return m.userInfo, nil
}

func (m *mockAuthService) SetupMFA(_ context.Context, _ string) (*service.MFASetupResult, error) {
	return &service.MFASetupResult{Secret: "TESTSECRET", URI: "otpauth://totp/test"}, nil
}
func (m *mockAuthService) VerifySetup(_ context.Context, _, _ string) error { return nil }
func (m *mockAuthService) DisableMFA(_ context.Context, _, _ string) error  { return nil }
func (m *mockAuthService) CompleteMFALogin(_ context.Context, _, _ string) (*service.TokenPair, error) {
	return m.tokens, m.err
}
func (m *mockAuthService) UpdateProfile(_ context.Context, _ string, _ service.UpdateProfileRequest) (*service.UserInfo, error) {
	if m.meErr != nil {
		return nil, m.meErr
	}
	return m.userInfo, nil
}

func (m *mockAuthService) ListUsers(_ context.Context, _, _ int32) ([]service.UserListItem, error) {
	return nil, nil
}

// --- tests ---

func TestLogin_Success(t *testing.T) {
	mock := &mockAuthService{
		tokens: &service.TokenPair{
			AccessToken:  "access.jwt.token",
			RefreshToken: "refresh.jwt.token",
			ExpiresIn:    900,
		},
	}
	h := NewAuth(mock)

	body := `{"email":"admin@test.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var tokens service.TokenPair
	json.NewDecoder(rec.Body).Decode(&tokens)
	if tokens.AccessToken != "access.jwt.token" {
		t.Errorf("expected access token, got %q", tokens.AccessToken)
	}
	if tokens.ExpiresIn != 900 {
		t.Errorf("expected expires_in 900, got %d", tokens.ExpiresIn)
	}
}

func TestLogin_MissingEmail_Returns400(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	body := `{"email":"","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing email, got %d", rec.Code)
	}
}

func TestLogin_MissingPassword_Returns400(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	body := `{"email":"user@test.com","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing password, got %d", rec.Code)
	}
}

func TestLogin_InvalidJSON_Returns400(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_InvalidCredentials_Returns401(t *testing.T) {
	mock := &mockAuthService{err: service.ErrInvalidCredentials}
	h := NewAuth(mock)

	body := `{"email":"user@test.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var resp errorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != "invalid email or password" {
		t.Errorf("expected generic credential error, got %q", resp.Error)
	}
}

func TestLogin_AccountLocked_Returns429(t *testing.T) {
	mock := &mockAuthService{err: service.ErrAccountLocked}
	h := NewAuth(mock)

	body := `{"email":"locked@test.com","password":"anything"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestLogin_InternalError_Returns500_Generic(t *testing.T) {
	mock := &mockAuthService{err: errors.New("database exploded")}
	h := NewAuth(mock)

	body := `{"email":"user@test.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp errorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != "internal error" {
		t.Errorf("expected generic error, got %q — internals may be leaking", resp.Error)
	}
}

func TestLogin_PropagatesIPAndUserAgent(t *testing.T) {
	mock := &mockAuthService{
		tokens: &service.TokenPair{AccessToken: "t", RefreshToken: "r", ExpiresIn: 900},
	}
	h := NewAuth(mock)

	body := `{"email":"user@test.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:54321"
	req.Header.Set("User-Agent", "SDA-Client/1.0")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if mock.lastReq.IP != "192.168.1.100:54321" {
		t.Errorf("expected IP propagated, got %q", mock.lastReq.IP)
	}
	if mock.lastReq.UserAgent != "SDA-Client/1.0" {
		t.Errorf("expected UserAgent propagated, got %q", mock.lastReq.UserAgent)
	}
	if mock.lastReq.Email != "user@test.com" {
		t.Errorf("expected email propagated, got %q", mock.lastReq.Email)
	}
}

func TestHealth_Returns200(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- Refresh tests ---

func TestRefresh_Success_WithBody(t *testing.T) {
	mock := &mockAuthService{
		tokens: &service.TokenPair{
			AccessToken:  "new.access.token",
			RefreshToken: "new.refresh.token",
			ExpiresIn:    900,
		},
	}
	h := NewAuth(mock)

	body := `{"refresh_token":"old.refresh.token"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Should set refresh cookie
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "sda_refresh" {
			found = true
			if !c.HttpOnly {
				t.Error("refresh cookie must be HttpOnly")
			}
			if !c.Secure {
				t.Error("refresh cookie must be Secure")
			}
		}
	}
	if !found {
		t.Error("expected sda_refresh cookie to be set")
	}
}

func TestRefresh_Success_WithCookie(t *testing.T) {
	mock := &mockAuthService{
		tokens: &service.TokenPair{
			AccessToken:  "new.access.token",
			RefreshToken: "new.refresh.token",
			ExpiresIn:    900,
		},
	}
	h := NewAuth(mock)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "sda_refresh", Value: "old.refresh.token"})
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRefresh_MissingToken_Returns400(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRefresh_InvalidToken_Returns401(t *testing.T) {
	mock := &mockAuthService{refreshErr: service.ErrInvalidRefreshToken}
	h := NewAuth(mock)

	body := `{"refresh_token":"bad.token"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// --- Logout tests ---

func TestLogout_ClearsCookie(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "sda_refresh", Value: "some.token"})
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "sda_refresh" && c.MaxAge != -1 {
			t.Error("expected refresh cookie to be cleared (MaxAge -1)")
		}
	}
}

// --- Me tests ---

func TestMe_Success(t *testing.T) {
	mock := &mockAuthService{
		userInfo: &service.UserInfo{
			ID:         "user-123",
			Email:      "enzo@saldivia.com",
			Name:       "Enzo",
			Role:       "admin",
			TenantID:   "tenant-1",
			TenantSlug: "saldivia",
		},
	}
	h := NewAuth(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Header.Set("X-User-ID", "user-123")
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var user service.UserInfo
	json.NewDecoder(rec.Body).Decode(&user)
	if user.Email != "enzo@saldivia.com" {
		t.Errorf("expected email enzo@saldivia.com, got %q", user.Email)
	}
}

func TestMe_NoUserID_Returns401(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMe_UserNotFound_Returns404(t *testing.T) {
	mock := &mockAuthService{meErr: service.ErrUserNotFound}
	h := NewAuth(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Header.Set("X-User-ID", "nonexistent")
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
