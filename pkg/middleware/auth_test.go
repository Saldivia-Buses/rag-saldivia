package middleware

import (
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
)

var (
	testPub  ed25519.PublicKey
	testPriv ed25519.PrivateKey
)

func init() {
	testPub, testPriv, _ = ed25519.GenerateKey(nil)
}

func validToken(t *testing.T) string {
	t.Helper()
	cfg := sdajwt.DefaultConfig(testPriv, testPub)
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID:   "u-123",
		Email:    "admin@saldivia.com",
		Name:     "Admin",
		TenantID: "t-456",
		Slug:     "saldivia",
		Role:     "admin",
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	return token
}

// echoHandler writes identity headers back as JSON so tests can inspect them.
func echoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info, _ := tenant.FromContext(r.Context())
		resp := map[string]string{
			"user_id":     r.Header.Get("X-User-ID"),
			"user_email":  r.Header.Get("X-User-Email"),
			"user_role":   r.Header.Get("X-User-Role"),
			"tenant_id":   r.Header.Get("X-Tenant-ID"),
			"tenant_slug": r.Header.Get("X-Tenant-Slug"),
			"ctx_tenant":  info.Slug,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func TestAuth_ValidToken_InjectsHeaders(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got map[string]string
	json.NewDecoder(rec.Body).Decode(&got)

	if got["user_id"] != "u-123" {
		t.Errorf("expected user_id u-123, got %q", got["user_id"])
	}
	if got["user_email"] != "admin@saldivia.com" {
		t.Errorf("expected email, got %q", got["user_email"])
	}
	if got["user_role"] != "admin" {
		t.Errorf("expected role admin, got %q", got["user_role"])
	}
	if got["tenant_id"] != "t-456" {
		t.Errorf("expected tenant_id t-456, got %q", got["tenant_id"])
	}
	if got["tenant_slug"] != "saldivia" {
		t.Errorf("expected tenant_slug saldivia, got %q", got["tenant_slug"])
	}
	if got["ctx_tenant"] != "saldivia" {
		t.Errorf("expected context tenant slug saldivia, got %q", got["ctx_tenant"])
	}
}

func TestAuth_NoAuthHeader_Returns401(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_WrongKey_Returns401(t *testing.T) {
	// Token signed with testPriv, middleware configured with different public key
	otherPub, _, _ := ed25519.GenerateKey(nil)
	handler := Auth(otherPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_HealthBypass(t *testing.T) {
	called := false
	handler := Auth(testPub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// /health without token should pass
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /health, got %d", rec.Code)
	}
	if !called {
		t.Error("handler was not called for /health")
	}
}

func TestAuth_HealthBypass_TrailingSlash(t *testing.T) {
	handler := Auth(testPub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /health/, got %d", rec.Code)
	}
}

func TestAuth_SpoofedHeadersStripped(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	// Attacker tries to inject headers — middleware should overwrite with JWT claims
	req.Header.Set("X-User-ID", "attacker-id")
	req.Header.Set("X-User-Role", "platform_admin")
	req.Header.Set("X-Tenant-Slug", "victim-tenant")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var got map[string]string
	json.NewDecoder(rec.Body).Decode(&got)

	// Must be from JWT, not from spoofed headers
	if got["user_id"] != "u-123" {
		t.Errorf("spoofed X-User-ID leaked: got %q, want u-123", got["user_id"])
	}
	if got["user_role"] != "admin" {
		t.Errorf("spoofed X-User-Role leaked: got %q, want admin", got["user_role"])
	}
	if got["tenant_slug"] != "saldivia" {
		t.Errorf("spoofed X-Tenant-Slug leaked: got %q, want saldivia", got["tenant_slug"])
	}
}

func TestAuth_SpoofedHeaders_NoToken_NotLeaked(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("X-User-ID", "attacker-id")
	req.Header.Set("X-Tenant-Slug", "victim-tenant")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ExpiredToken_Returns401(t *testing.T) {
	cfg := sdajwt.Config{
		PrivateKey:   testPriv,
		PublicKey:    testPub,
		AccessExpiry: -1 * time.Hour,
		Issuer:       "sda",
	}
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID: "u-1", TenantID: "t-1", Slug: "test", Role: "user",
	})
	if err != nil {
		t.Fatalf("create expired token: %v", err)
	}

	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", rec.Code)
	}
}

func TestAuth_BearerPrefix_Required(t *testing.T) {
	handler := Auth(testPub)(echoHandler())

	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", validToken(t))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without Bearer prefix, got %d", rec.Code)
	}
}
