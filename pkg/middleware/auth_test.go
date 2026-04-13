package middleware

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/redis/go-redis/v9"
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
	// Attacker tries to inject identity headers — non-slug headers overwritten with JWT claims
	req.Header.Set("X-User-ID", "attacker-id")
	req.Header.Set("X-User-Role", "platform_admin")
	// Slug matches JWT so cross-validation passes; identity headers are still tested
	req.Header.Set("X-Tenant-Slug", "saldivia")
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

func TestAuth_SpoofedSlug_DifferentTenant_Returns403(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	// Attacker has JWT for "saldivia" but tries to access "victim-tenant"
	req.Header.Set("X-Tenant-Slug", "victim-tenant")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Cross-validation blocks: JWT slug != header slug
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for slug mismatch attack, got %d", rec.Code)
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

func TestAuth_SlugCrossValidation_Mismatch_Returns403(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	// Simulate Traefik injecting a different tenant slug (token has "saldivia")
	req.Header.Set("X-Tenant-Slug", "other-tenant")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for slug mismatch, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuth_SlugCrossValidation_Match_Passes(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	// Traefik slug matches JWT slug — should pass
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for matching slug, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuth_SlugCrossValidation_NoHeader_Passes(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	// No X-Tenant-Slug header at all — should pass (no cross-validation)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when no slug header, got %d: %s", rec.Code, rec.Body.String())
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

// TestAuth_EmptyBearerToken_Returns401 covers the edge case where the
// Authorization header is present but the token after "Bearer " is empty.
func TestAuth_EmptyBearerToken_Returns401(t *testing.T) {
	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for empty bearer token, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("error response must be JSON, got Content-Type %q", ct)
	}
}

// TestAuth_MFAPending_Returns401 verifies that a token with role "mfa_pending"
// is blocked by the middleware. MFA-pending tokens are only valid for
// /v1/auth/mfa/verify — using one elsewhere must fail with 401.
func TestAuth_MFAPending_Returns401(t *testing.T) {
	cfg := sdajwt.Config{
		PrivateKey:   testPriv,
		PublicKey:    testPub,
		AccessExpiry: 15 * time.Minute,
		Issuer:       "sda",
	}
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID: "u-mfa", TenantID: "t-1", Slug: "test", Role: "mfa_pending",
	})
	if err != nil {
		t.Fatalf("create mfa_pending token: %v", err)
	}

	handler := Auth(testPub)(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for mfa_pending role, got %d", rec.Code)
	}
}

// TestAuth_ValidToken_InjectsClaimsInContext verifies that auth middleware stores
// all JWT-derived values in the request context (not just headers).
func TestAuth_ValidToken_InjectsClaimsInContext(t *testing.T) {
	var capturedUserID, capturedEmail, capturedRole string
	var capturedTenant tenant.Info

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		capturedEmail = UserEmailFromContext(r.Context())
		capturedRole = RoleFromContext(r.Context())
		capturedTenant, _ = tenant.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth(testPub)(inner)
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedUserID != "u-123" {
		t.Errorf("context UserID: got %q, want u-123", capturedUserID)
	}
	if capturedEmail != "admin@saldivia.com" {
		t.Errorf("context Email: got %q, want admin@saldivia.com", capturedEmail)
	}
	if capturedRole != "admin" {
		t.Errorf("context Role: got %q, want admin", capturedRole)
	}
	if capturedTenant.Slug != "saldivia" {
		t.Errorf("context TenantSlug: got %q, want saldivia", capturedTenant.Slug)
	}
	if capturedTenant.ID != "t-456" {
		t.Errorf("context TenantID: got %q, want t-456", capturedTenant.ID)
	}
}

// TestAuth_SpoofedXUserID_IsOverwritten confirms that a client-injected X-User-ID
// header is replaced with the value from the verified JWT, not the spoofed value.
// This is INVARIANT #2: JWT is the single source of identity.
func TestAuth_SpoofedXUserID_IsOverwritten(t *testing.T) {
	var headerUserID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerUserID = r.Header.Get("X-User-ID")
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth(testPub)(inner)
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	req.Header.Set("X-User-ID", "attacker-injected-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if headerUserID != "u-123" {
		t.Errorf("X-User-ID must come from JWT, got %q (spoofed value leaked)", headerUserID)
	}
}

// TestAuth_SpoofedXTenantID_IsOverwritten mirrors the above for X-Tenant-ID.
func TestAuth_SpoofedXTenantID_IsOverwritten(t *testing.T) {
	var headerTenantID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerTenantID = r.Header.Get("X-Tenant-ID")
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth(testPub)(inner)
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	req.Header.Set("X-Tenant-ID", "attacker-tenant-uuid")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if headerTenantID != "t-456" {
		t.Errorf("X-Tenant-ID must come from JWT, got %q (spoofed value leaked)", headerTenantID)
	}
}

// TestAuth_InjectsTenantSlugInContext verifies the slug is available via
// tenant.FromContext after the auth middleware runs.
func TestAuth_InjectsTenantSlugInContext(t *testing.T) {
	var slug string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, _ := tenant.FromContext(r.Context())
		slug = info.Slug
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth(testPub)(inner)
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if slug != "saldivia" {
		t.Errorf("tenant slug in context: got %q, want saldivia", slug)
	}
}

// --- Blacklist integration tests (require Redis, skip if unavailable) ---

// newTestRedisBlacklist creates a TokenBlacklist backed by a local Redis instance.
// Skips the test if Redis is not reachable.
func newTestRedisBlacklist(t *testing.T) *security.TokenBlacklist {
	t.Helper()
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis not available (%v) — skipping blacklist middleware test", err)
	}
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
		rdb.Close()
	})
	return security.NewTokenBlacklist(rdb)
}

// TestAuth_BlacklistedToken_Returns401 verifies that a valid but revoked token
// (present in the blacklist) is rejected with 401.
func TestAuth_BlacklistedToken_Returns401(t *testing.T) {
	bl := newTestRedisBlacklist(t)

	cfg := sdajwt.Config{
		PrivateKey:   testPriv,
		PublicKey:    testPub,
		AccessExpiry: 15 * time.Minute,
		Issuer:       "sda",
	}
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID: "u-revoked", TenantID: "t-1", Slug: "test", Role: "user",
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	// Verify the token to extract its JTI
	claims, err := sdajwt.Verify(testPub, token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}

	// Revoke the token
	if err := bl.Revoke(context.Background(), claims.ID, time.Now().Add(15*time.Minute)); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	handler := AuthWithConfig(testPub, AuthConfig{Blacklist: bl})(echoHandler())
	req := httptest.NewRequest(http.MethodGet, "/v1/something", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for blacklisted token, got %d: %s", rec.Code, rec.Body.String())
	}
}
