package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSummary_NoAuth_Returns401(t *testing.T) {
	h := NewFeedback(nil) // DB not needed for auth check

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/summary", nil)
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestErrors_NoAuth_Returns401(t *testing.T) {
	h := NewFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/errors", nil)
	rec := httptest.NewRecorder()

	h.Errors(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestQuality_NoAuth_Returns401(t *testing.T) {
	h := NewFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/quality", nil)
	rec := httptest.NewRecorder()

	h.Quality(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUsage_NoAuth_Returns401(t *testing.T) {
	h := NewFeedback(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feedback/usage", nil)
	rec := httptest.NewRecorder()

	h.Usage(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

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

func TestParsePeriod(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"7d", 168},
		{"30d", 720},
		{"90d", 2160},
		{"", 720},       // default
		{"invalid", 720}, // default
	}

	for _, tc := range tests {
		got := parsePeriod(tc.input)
		if got != tc.expected {
			t.Errorf("parsePeriod(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}
