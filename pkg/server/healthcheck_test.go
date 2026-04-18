package server

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// startLocalhostServer runs an http.Server on a freshly-picked loopback port
// (matching the interface healthcheckStatus probes) and returns the port.
func startLocalhostServer(t *testing.T, h http.HandlerFunc) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: h}
	go srv.Serve(ln)
	t.Cleanup(func() { _ = srv.Close() })
	return strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
}

func TestHealthcheckStatus_OK(t *testing.T) {
	port := startLocalhostServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/health") {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	if got := healthcheckStatus(port); got != 0 {
		t.Fatalf("expected 0 on 200 OK, got %d", got)
	}
}

func TestHealthcheckStatus_ServiceDegraded(t *testing.T) {
	port := startLocalhostServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	if got := healthcheckStatus(port); got != 1 {
		t.Fatalf("expected 1 on 503, got %d", got)
	}
}

func TestHealthcheckStatus_ConnectionRefused(t *testing.T) {
	// Bind + close to capture a free port, then probe it — no server listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	ln.Close()
	if got := healthcheckStatus(port); got != 1 {
		t.Fatalf("expected 1 on connection refused, got %d", got)
	}
}
