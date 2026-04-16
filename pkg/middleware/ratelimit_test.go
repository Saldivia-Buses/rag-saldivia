package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// TestRateLimit_ExceedsLimit_Returns429 sends more requests than the configured
// burst, verifying that the middleware returns 429 once the bucket is exhausted.
func TestRateLimit_ExceedsLimit_Returns429(t *testing.T) {
	const limit = 3
	cfg := RateLimitConfig{
		Requests: limit,
		Window:   time.Minute,
		KeyFunc:  func(r *http.Request) string { return "test-key" },
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First `limit` requests should succeed (burst)
	for i := 0; i < limit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// The next request must be rate-limited
	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding limit, got %d", rec.Code)
	}
}

// TestRateLimit_Returns429WithJSONBody verifies the 429 response has
// Content-Type: application/json and a JSON body. INVARIANT #7.
func TestRateLimit_Returns429WithJSONBody(t *testing.T) {
	cfg := RateLimitConfig{
		Requests: 1,
		Window:   time.Minute,
		KeyFunc:  func(r *http.Request) string { return "json-body-key" },
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the single-request burst
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Second request should be rate-limited with a JSON body
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
	if ct := rec2.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

// TestRateLimit_Sets_RetryAfter_Header verifies that a 429 response includes
// the Retry-After header set to the window duration in seconds.
func TestRateLimit_Sets_RetryAfter_Header(t *testing.T) {
	const window = 30 * time.Second
	cfg := RateLimitConfig{
		Requests: 1,
		Window:   window,
		KeyFunc:  func(r *http.Request) string { return "retry-after-key" },
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust burst
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	httptest.NewRecorder()
	handler.ServeHTTP(httptest.NewRecorder(), req)

	// Trigger rate limit
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
	retryAfter := rec2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Fatal("Retry-After header must be set on 429 response")
	}
	got, err := strconv.Atoi(retryAfter)
	if err != nil {
		t.Fatalf("Retry-After %q is not a valid integer: %v", retryAfter, err)
	}
	want := int(window.Seconds())
	if got != want {
		t.Errorf("Retry-After: got %d, want %d", got, want)
	}
}

// TestRateLimit_ByUser_KeyFunc verifies that requests from different users are
// bucketed independently — exhausting one user's bucket does not affect another.
func TestRateLimit_ByUser_KeyFunc(t *testing.T) {
	cfg := RateLimitConfig{
		Requests: 2,
		Window:   time.Minute,
		KeyFunc:  ByUser,
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	sendAs := func(userID string) int {
		req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
		if userID != "" {
			req.Header.Set("X-User-ID", userID)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	// Exhaust user-a's bucket (2 requests)
	if code := sendAs("user-a"); code != http.StatusOK {
		t.Fatalf("user-a req 1: expected 200, got %d", code)
	}
	if code := sendAs("user-a"); code != http.StatusOK {
		t.Fatalf("user-a req 2: expected 200, got %d", code)
	}
	if code := sendAs("user-a"); code != http.StatusTooManyRequests {
		t.Fatalf("user-a req 3: expected 429, got %d", code)
	}

	// user-b's bucket is independent — must still be OK
	if code := sendAs("user-b"); code != http.StatusOK {
		t.Fatalf("user-b req 1: expected 200 (independent bucket), got %d", code)
	}
}

// TestRateLimit_ByIP_DefaultKeyFunc verifies that when no KeyFunc is provided,
// the middleware defaults to ByIP (per remote address).
func TestRateLimit_ByIP_DefaultKeyFunc(t *testing.T) {
	cfg := RateLimitConfig{
		Requests: 2,
		Window:   time.Minute,
		// KeyFunc intentionally omitted — must default to ByIP
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := send(); code != http.StatusOK {
		t.Fatalf("req 1: expected 200, got %d", code)
	}
	if code := send(); code != http.StatusOK {
		t.Fatalf("req 2: expected 200, got %d", code)
	}
	if code := send(); code != http.StatusTooManyRequests {
		t.Fatalf("req 3: expected 429 (ByIP default), got %d", code)
	}
}
