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
	json.NewDecoder(rec.Body).Decode(&resp)
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
