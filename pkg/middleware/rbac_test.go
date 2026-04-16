package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

func TestRequirePermission_Allowed(t *testing.T) {
	handler := RequirePermission("chat.read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	ctx := WithRole(req.Context(), "user")
	ctx = WithPermissions(ctx, []string{"chat.read", "chat.write", "collections.read"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequirePermission_Denied(t *testing.T) {
	handler := RequirePermission("ingest.write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", nil)
	ctx := WithRole(req.Context(), "viewer")
	ctx = WithPermissions(ctx, []string{"chat.read", "collections.read", "docs.read"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequirePermission_AdminBypass(t *testing.T) {
	handler := RequirePermission("ingest.write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", nil)
	// Admin with no explicit permissions — should still pass
	ctx := WithRole(req.Context(), "admin")
	ctx = WithPermissions(ctx, nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (admin bypass), got %d", rec.Code)
	}
}

func TestRequirePermission_NoPermissions(t *testing.T) {
	handler := RequirePermission("chat.read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	ctx := WithRole(req.Context(), "user")
	// No permissions set at all
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 with no permissions, got %d", rec.Code)
	}
}

func TestRequirePermission_ViewerCantDelete(t *testing.T) {
	handler := RequirePermission("chat.write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("viewer should not reach delete handler")
	}))

	req := httptest.NewRequest(http.MethodDelete, "/v1/chat/sessions/123", nil)
	ctx := WithRole(req.Context(), "viewer")
	ctx = WithPermissions(ctx, []string{"chat.read", "collections.read", "docs.read"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequirePermission_WildcardMatch(t *testing.T) {
	handler := RequirePermission("erp.accounting.read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/accounting/entries", nil)
	ctx := WithRole(req.Context(), "user")
	ctx = WithPermissions(ctx, []string{"erp.accounting.*"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (wildcard match), got %d", rec.Code)
	}
}

func TestRequirePermission_TopLevelWildcard(t *testing.T) {
	handler := RequirePermission("erp.stock.write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/erp/stock/articles", nil)
	ctx := WithRole(req.Context(), "manager")
	ctx = WithPermissions(ctx, []string{"erp.*"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (erp.* wildcard), got %d", rec.Code)
	}
}

func TestRequirePermission_WildcardNoFalsePositive(t *testing.T) {
	handler := RequirePermission("chat.read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not match")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	ctx := WithRole(req.Context(), "user")
	ctx = WithPermissions(ctx, []string{"erp.*"}) // erp.* should NOT match chat.read
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 (erp.* should not match chat.read), got %d", rec.Code)
	}
}

func TestRequireModule_Allowed(t *testing.T) {
	handler := RequireModule("erp")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs", nil)
	ctx := WithEnabledModules(req.Context(), []string{"chat", "erp", "fleet"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireModule_Denied(t *testing.T) {
	handler := RequireModule("erp")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs", nil)
	ctx := WithEnabledModules(req.Context(), []string{"chat", "fleet"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestMatchPermission(t *testing.T) {
	tests := []struct {
		userPerm string
		required string
		want     bool
	}{
		{"chat.read", "chat.read", true},
		{"chat.read", "chat.write", false},
		{"erp.*", "erp.stock.read", true},
		{"erp.*", "erp.accounting.write", true},
		{"erp.*", "chat.read", false},
		{"erp.accounting.*", "erp.accounting.read", true},
		{"erp.accounting.*", "erp.accounting.reverse", true},
		{"erp.accounting.*", "erp.stock.read", false},
		{"*", "anything", true},
	}
	for _, tt := range tests {
		got := matchPermission(tt.userPerm, tt.required)
		if got != tt.want {
			t.Errorf("matchPermission(%q, %q) = %v, want %v", tt.userPerm, tt.required, got, tt.want)
		}
	}
}

// TestRequirePermission_NoJWTClaims_Returns403 simulates a route that is missing
// the Auth middleware in its chain — no role or permissions are in context.
// RequirePermission must return 403, not panic or 401.
func TestRequirePermission_NoJWTClaims_Returns403(t *testing.T) {
	handler := RequirePermission("chat.read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler must not be called when no claims in context")
	}))

	// Plain request — no Auth middleware, no context values
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when no claims in context, got %d", rec.Code)
	}
}

// TestRequirePermission_Returns403WithJSONBody verifies that the 403 response
// has Content-Type: application/json and a JSON body with an "error" key.
// INVARIANT #7: all API error responses must be JSON.
func TestRequirePermission_Returns403WithJSONBody(t *testing.T) {
	handler := RequirePermission("erp.write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler must not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/erp/entries", nil)
	ctx := WithRole(req.Context(), "viewer")
	ctx = WithPermissions(ctx, []string{"erp.read"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Errorf("JSON body must contain 'error' key, got %v", body)
	}
}

func TestRequirePermission_ContextFromAuth(t *testing.T) {
	// Simulate the full flow: Auth middleware sets context, RBAC reads it
	inner := RequirePermission("chat.read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context values are accessible
		role := RoleFromContext(r.Context())
		perms := PermissionsFromContext(r.Context())
		if role != "user" {
			t.Errorf("expected role 'user', got %q", role)
		}
		if len(perms) != 2 {
			t.Errorf("expected 2 perms, got %d", len(perms))
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Auth middleware wraps the chain and sets context
	outer := Auth(testPub)(inner)

	// Create a valid token with permissions
	cfg := sdajwt.Config{
		PrivateKey:   testPriv,
		PublicKey:    testPub,
		AccessExpiry: 15 * time.Minute,
		Issuer:       "sda",
	}
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID:      "u-1",
		TenantID:    "t-1",
		Slug:        "test",
		Role:        "user",
		Permissions: []string{"chat.read", "collections.read"},
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	outer.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
