package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"localhost:3000", 1},
		{"*.sda.app,localhost:3000", 2},
		{"*.sda.app, localhost:3000 , http://example.com", 3},
		{" , , ", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			origins := parseOrigins(tt.input)
			if len(origins) != tt.want {
				t.Errorf("parseOrigins(%q) = %d origins, want %d", tt.input, len(origins), tt.want)
			}
		})
	}
}

func TestUpgrade_NoToken_Returns401(t *testing.T) {
	h := &WS{jwtSecret: "test-secret-at-least-32-chars-long!!"}
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	h.Upgrade(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", rec.Code)
	}
}

func TestUpgrade_InvalidToken_Returns401(t *testing.T) {
	h := &WS{jwtSecret: "test-secret-at-least-32-chars-long!!"}
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()

	h.Upgrade(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", rec.Code)
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid bearer", "Bearer abc.jwt.token", "abc.jwt.token"},
		{"missing prefix", "abc.jwt.token", ""},
		{"empty", "", ""},
		{"basic auth", "Basic dXNlcjpwYXNz", ""},
		{"bearer lowercase", "bearer abc", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/ws", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			got := extractBearerToken(req)
			if got != tt.want {
				t.Errorf("extractBearerToken() = %q, want %q", got, tt.want)
			}
		})
	}
}
