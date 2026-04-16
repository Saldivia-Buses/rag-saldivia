package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllHealthy(t *testing.T) {
	hc := New("test-service")
	hc.Add("db", func(ctx context.Context) error { return nil })
	hc.Add("redis", func(ctx context.Context) error { return nil })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	hc.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Service != "test-service" {
		t.Errorf("expected service test-service, got %s", resp.Service)
	}
	if resp.Dependencies["db"].Status != "up" {
		t.Errorf("expected db up, got %s", resp.Dependencies["db"].Status)
	}
	if resp.Dependencies["redis"].Status != "up" {
		t.Errorf("expected redis up, got %s", resp.Dependencies["redis"].Status)
	}
}

func TestDegradedOnFailure(t *testing.T) {
	hc := New("test-service")
	hc.Add("db", func(ctx context.Context) error { return nil })
	hc.Add("redis", func(ctx context.Context) error {
		return fmt.Errorf("connection refused")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	hc.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status degraded, got %s", resp.Status)
	}
	if resp.Dependencies["db"].Status != "up" {
		t.Errorf("expected db up")
	}
	if resp.Dependencies["redis"].Status != "down" {
		t.Errorf("expected redis down")
	}
	// Error should be generic, not the real error message
	if resp.Dependencies["redis"].Error != "unavailable" {
		t.Errorf("expected generic error, got %q", resp.Dependencies["redis"].Error)
	}
}

func TestNoDependencies(t *testing.T) {
	hc := New("minimal")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	hc.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected ok, got %s", resp.Status)
	}
}

func TestExtras(t *testing.T) {
	hc := New("ws-hub")
	hc.AddExtra(func() (string, any) { return "clients", 42 })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	hc.Handler().ServeHTTP(w, r)

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Extra["clients"] != float64(42) {
		t.Errorf("expected clients=42, got %v", resp.Extra["clients"])
	}
}

func TestTimeoutOnSlowCheck(t *testing.T) {
	hc := New("slow")
	hc.Add("slow-dep", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	hc.Handler().ServeHTTP(w, r)

	// Should complete within ~3 seconds (handler timeout), not 10
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on timeout, got %d", w.Code)
	}
}

func TestCacheControlHeader(t *testing.T) {
	hc := New("test")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	hc.Handler().ServeHTTP(w, r)

	if w.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("expected Cache-Control: no-store, got %q", w.Header().Get("Cache-Control"))
	}
}

func TestLatencyTracked(t *testing.T) {
	hc := New("test")
	hc.Add("fast", func(ctx context.Context) error { return nil })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	hc.Handler().ServeHTTP(w, r)

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Dependencies["fast"].LatencyMs < 0 {
		t.Errorf("expected non-negative latency, got %d", resp.Dependencies["fast"].LatencyMs)
	}
}
