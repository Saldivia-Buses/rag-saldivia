package handler

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

func testKeypair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keys: %v", err)
	}
	return pub, priv
}

func setupServiceTokenRouter(t *testing.T, key string) (*chi.Mux, ed25519.PublicKey) {
	t.Helper()
	pub, priv := testKeypair(t)
	jwtCfg := sdajwt.DefaultConfig(priv, pub)

	h := &Auth{jwtCfg: jwtCfg}
	if key != "" {
		h.SetServiceTokenConfig(ServiceTokenConfig{
			Key:              key,
			PlatformTenantID: "test-platform-id",
			PlatformSlug:     "platform",
		})
	}

	r := chi.NewRouter()
	r.Post("/v1/auth/service-token", h.ServiceToken)
	return r, pub
}

func TestServiceToken_ValidKey(t *testing.T) {
	r, pub := setupServiceTokenRouter(t, "test-secret-key-minimum-32-bytes!")

	body := `{"service":"healthwatch","key":"test-secret-key-minimum-32-bytes!"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/service-token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp serviceTokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.ExpiresAt == "" {
		t.Fatal("expected non-empty expires_at")
	}

	// Verify the issued token is valid and has correct claims
	claims, err := sdajwt.Verify(pub, resp.Token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.UserID != "svc:healthwatch" {
		t.Errorf("expected uid svc:healthwatch, got %s", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
	if claims.Slug != "platform" {
		t.Errorf("expected slug platform, got %s", claims.Slug)
	}
	if claims.TenantID != "test-platform-id" {
		t.Errorf("expected tid test-platform-id, got %s", claims.TenantID)
	}
}

func TestServiceToken_InvalidKey(t *testing.T) {
	r, _ := setupServiceTokenRouter(t, "correct-secret-key-minimum-32-bytes!")

	body := `{"service":"healthwatch","key":"wrong-key"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/service-token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestServiceToken_MissingFields(t *testing.T) {
	r, _ := setupServiceTokenRouter(t, "test-secret-key-minimum-32-bytes!")

	tests := []struct {
		name string
		body string
	}{
		{"missing service", `{"key":"test-secret-key-minimum-32-bytes!"}`},
		{"missing key", `{"service":"healthwatch"}`},
		{"empty body", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/service-token", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestServiceToken_NotConfigured(t *testing.T) {
	r, _ := setupServiceTokenRouter(t, "") // no key = not configured

	body := `{"service":"healthwatch","key":"anything"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/service-token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}
