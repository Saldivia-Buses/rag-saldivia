package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecureHeaders_SetsAllHeaders(t *testing.T) {
	handler := SecureHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	checks := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "DENY",
		"X-XSS-Protection":      "0",
		"Referrer-Policy":       "strict-origin-when-cross-origin",
		"Permissions-Policy":    "camera=(), microphone=(), geolocation=()",
	}
	for header, want := range checks {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
}

func TestSecureHeaders_HSTS_OnlyWithTLS(t *testing.T) {
	handler := SecureHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Plain HTTP — no HSTS
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS should not be set on plain HTTP, got %q", got)
	}

	// With X-Forwarded-Proto: https — HSTS should be set
	req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
	req2.Header.Set("X-Forwarded-Proto", "https")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if got := rec2.Header().Get("Strict-Transport-Security"); got == "" {
		t.Error("HSTS should be set when X-Forwarded-Proto is https")
	}
}
