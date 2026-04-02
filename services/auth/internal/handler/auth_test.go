package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/auth/internal/service"
)

// --- mock ---

type mockAuthService struct {
	tokens *service.TokenPair
	err    error
}

func (m *mockAuthService) Login(_ context.Context, req service.LoginRequest) (*service.TokenPair, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tokens, nil
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

func TestHealth_Returns200(t *testing.T) {
	h := NewAuth(&mockAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
