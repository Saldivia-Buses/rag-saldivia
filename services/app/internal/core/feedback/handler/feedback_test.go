package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- auth / no-auth guard tests ---

func TestSummary_NoAuth_Returns401(t *testing.T) {
	h := NewFeedback(nil, nil) // DB not needed for auth check

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/summary", nil)
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestErrors_NoAuth_Returns401(t *testing.T) {
	h := NewFeedback(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/errors", nil)
	rec := httptest.NewRecorder()

	h.Errors(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestQuality_NoAuth_Returns401(t *testing.T) {
	// NewFeedback requires two arguments: repo + platformDB
	h := NewFeedback(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/quality", nil)
	rec := httptest.NewRecorder()

	h.Quality(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUsage_NoAuth_Returns401(t *testing.T) {
	h := NewFeedback(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/usage", nil)
	rec := httptest.NewRecorder()

	h.Usage(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestHealthScore_NoTenantID_Returns401(t *testing.T) {
	h := NewFeedback(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/health-score", nil)
	rec := httptest.NewRecorder()

	h.HealthScore(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestHealthScore_NilPlatformDB_Returns503(t *testing.T) {
	// platformDB is nil — service unavailable
	h := NewFeedback(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/health-score", nil)
	req.Header.Set("X-Tenant-ID", "t-1")
	rec := httptest.NewRecorder()

	h.HealthScore(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when platformDB nil, got %d", rec.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Error("expected non-empty error in response body")
	}
}

// --- platform feedback auth tests ---

func TestPlatformTenants_NonAdmin_Returns403(t *testing.T) {
	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/tenants", nil)
	req.Header.Set("X-User-Role", "user")
	rec := httptest.NewRecorder()

	h.Tenants(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestPlatformAlerts_NonAdmin_Returns403(t *testing.T) {
	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/alerts", nil)
	req.Header.Set("X-User-Role", "user")
	rec := httptest.NewRecorder()

	h.Alerts(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestPlatformQuality_NoRole_Returns403(t *testing.T) {
	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/quality", nil)
	rec := httptest.NewRecorder()

	h.Quality(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// Wrong tenant slug with admin role must still be rejected.
func TestPlatformTenants_AdminWrongSlug_Returns403(t *testing.T) {
	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/tenants", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Tenant-Slug", "saldivia") // not "platform"
	rec := httptest.NewRecorder()

	h.Tenants(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for admin with wrong slug, got %d", rec.Code)
	}
}

func TestPlatformAlerts_AdminWrongSlug_Returns403(t *testing.T) {
	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/alerts", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()

	h.Alerts(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestPlatformQuality_AdminWrongSlug_Returns403(t *testing.T) {
	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/quality", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()

	h.Quality(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// --- parsePeriod ---

func TestParsePeriod(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"7d", 168},
		{"30d", 720},
		{"90d", 2160},
		{"", 720},        // default
		{"invalid", 720}, // default
	}

	for _, tc := range tests {
		got := parsePeriod(tc.input)
		if got != tc.expected {
			t.Errorf("parsePeriod(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}

// ── Tenant-scoped handler: nil repo → panics on DB call, but we can test
// the HealthScore path that guards platformDB before touching repo. ──────────

func TestHealthScore_MissingTenantID_Returns401_ErrorFieldPresent(t *testing.T) {
	h := NewFeedback(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/health-score", nil)
	rec := httptest.NewRecorder()

	h.HealthScore(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Error("expected non-empty 'error' field in 401 response")
	}
}

// ── Platform feedback: admin with correct slug is allowed ────────────────────

func TestPlatformTenants_AdminPlatformSlug_Allowed_NilDB_Returns500(t *testing.T) {
	// Admin role + "platform" slug passes the auth check.
	// With nil platformDB, the handler panics or returns 500 on the Query call.
	// We only test that the auth gate passes — the nil DB will cause a panic
	// which we recover here, OR the handler returns 500.
	// This test documents that slug="platform" + role="admin" is the only
	// combination that proceeds past the guard.
	defer func() {
		if r := recover(); r != nil {
			// panic from nil DB Query is acceptable — auth guard was passed
		}
	}()

	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/tenants", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Tenant-Slug", "platform")
	rec := httptest.NewRecorder()

	h.Tenants(rec, req)

	// If we get here without a panic, the handler must have returned a non-403.
	if rec.Code == http.StatusForbidden {
		t.Fatalf("admin+platform slug must not get 403, got %d", rec.Code)
	}
}

func TestPlatformAlerts_AdminPlatformSlug_Allowed_NilDB_Returns500(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// acceptable — auth passed, DB nil caused panic
		}
	}()

	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/alerts", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Tenant-Slug", "platform")
	rec := httptest.NewRecorder()

	h.Alerts(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("admin+platform slug must not get 403, got %d", rec.Code)
	}
}

func TestPlatformQuality_AdminPlatformSlug_Allowed_NilDB_Returns500(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// acceptable — auth passed, DB nil caused panic
		}
	}()

	h := NewPlatformFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/feedback/quality", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Tenant-Slug", "platform")
	rec := httptest.NewRecorder()

	h.Quality(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("admin+platform slug must not get 403, got %d", rec.Code)
	}
}

// ── Tenant feedback handler: auth check is required on every endpoint ─────────

func TestSummary_WithUserID_NilRepo_Panics_NotAuth(t *testing.T) {
	// Once userID is present the handler calls repo.GetSummaryAIQuality which panics
	// on a nil repo. We verify it does NOT return 401 (auth passed).
	defer func() {
		if r := recover(); r != nil {
			// Acceptable — auth guard passed, nil repo caused panic.
		}
	}()

	h := NewFeedback(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/summary", nil)
	req.Header.Set("X-User-ID", "u-1")
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

	// If we reach here without panic, ensure it wasn't a 401.
	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("present X-User-ID must not yield 401, got %d", rec.Code)
	}
}

// --- parseIntParam ---

func TestParseIntParam(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback int
		max      int
		want     int
	}{
		{"empty uses fallback", "", 50, 200, 50},
		{"valid value", "10", 50, 200, 10},
		{"zero uses fallback", "0", 50, 200, 50},
		{"negative uses fallback", "-5", 50, 200, 50},
		{"non-numeric uses fallback", "abc", 50, 200, 50},
		{"over max capped", "999", 50, 200, 200},
		{"exactly max", "200", 50, 200, 200},
		{"one below max", "199", 50, 200, 199},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseIntParam(tc.input, tc.fallback, tc.max)
			if got != tc.want {
				t.Errorf("parseIntParam(%q, %d, %d) = %d, want %d",
					tc.input, tc.fallback, tc.max, got, tc.want)
			}
		})
	}
}
