package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// buildFakes spins up one httptest.Server per production upstream name
// plus a Next.js fake. Each fake writes two response headers so tests can
// assert which upstream saw the request and on which path.
func buildFakes(t *testing.T) (map[string]*httptest.Server, *httptest.Server) {
	t.Helper()
	svrs := make(map[string]*httptest.Server, len(upstreamAddrs))
	for name := range upstreamAddrs {
		name := name
		svrs[name] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Fake-Upstream", name)
			w.Header().Set("X-Fake-Path", req.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, name)
		}))
	}
	nextjs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Fake-Upstream", "nextjs")
		w.Header().Set("X-Fake-Path", req.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "nextjs")
	}))
	return svrs, nextjs
}

func closeAll(svrs map[string]*httptest.Server, extra ...*httptest.Server) {
	for _, s := range svrs {
		s.Close()
	}
	for _, s := range extra {
		s.Close()
	}
}

func newTestServer(t *testing.T, svrs map[string]*httptest.Server, nextjs *httptest.Server) *httptest.Server {
	t.Helper()
	addrs := make(map[string]string, len(svrs))
	for name, s := range svrs {
		addrs[name] = s.URL
	}
	r, err := newRouterWithAddrs(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		addrs,
		nextjs.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(newMux(r))
}

func TestRouting(t *testing.T) {
	svrs, nextjs := buildFakes(t)
	defer closeAll(svrs, nextjs)
	srv := newTestServer(t, svrs, nextjs)
	defer srv.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantUp     string // empty = in-process response or error before proxy
	}{
		{"readyz is in-process", "GET", "/readyz", http.StatusOK, ""},
		{"healthz is in-process", "GET", "/healthz", http.StatusOK, ""},
		{"root goes to nextjs", "GET", "/", http.StatusOK, "nextjs"},
		{"static goes to nextjs", "GET", "/favicon.ico", http.StatusOK, "nextjs"},
		{"deep path goes to nextjs", "GET", "/dashboard/chats/42", http.StatusOK, "nextjs"},
		{"v1 chat", "POST", "/v1/chat/messages", http.StatusOK, "chat"},
		{"v1 erp", "GET", "/v1/erp/orders", http.StatusOK, "erp"},
		{"v1 with no service is 404", "GET", "/v1/", http.StatusNotFound, ""},
		{"v1 unknown service is 404", "GET", "/v1/doesnotexist/foo", http.StatusNotFound, ""},
		{"ws root goes to ws", "GET", "/ws", http.StatusOK, "ws"},
		{"ws sub-path goes to ws", "GET", "/ws/chat/42", http.StatusOK, "ws"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, srv.URL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("status: got %d want %d — body: %s", resp.StatusCode, tc.wantStatus, body)
			}
			if got := resp.Header.Get("X-Fake-Upstream"); got != tc.wantUp {
				t.Fatalf("upstream: got %q want %q", got, tc.wantUp)
			}
			if tc.wantUp != "" {
				// Upstream must see the full original path (frontdoor
				// preserves it — no prefix stripping).
				if got := resp.Header.Get("X-Fake-Path"); got != tc.path {
					t.Fatalf("path at upstream: got %q want %q", got, tc.path)
				}
			}
		})
	}
}

// TestXForwardedHeaders pins the Rewrite API behaviour — these headers
// are the reason we migrated off the legacy Director API.
func TestXForwardedHeaders(t *testing.T) {
	var seen http.Header
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		seen = req.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	r, err := newRouterWithAddrs(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		map[string]string{"chat": up.URL},
		up.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(newMux(r))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/v1/chat/probe")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	for _, h := range []string{"X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto"} {
		if v := seen.Get(h); v == "" {
			t.Errorf("%s missing at upstream; got headers: %v", h, seen)
		}
	}
}

// TestUnknownUpstreamBody confirms the 404 body names the offending
// service — makes ops debugging fast.
func TestUnknownUpstreamBody(t *testing.T) {
	svrs, nextjs := buildFakes(t)
	defer closeAll(svrs, nextjs)
	srv := newTestServer(t, svrs, nextjs)
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/v1/xyzzy/foo")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "xyzzy") {
		t.Fatalf("expected body to name the unknown service; got %q", body)
	}
}

// TestUpstreamsCoverServices guards against the H1-style bug where the
// frontdoor map drifts from the real services/*/cmd/main.go ports. This
// test just proves the map has an entry for every service name the
// Makefile's build target compiles.
func TestUpstreamMapCoverage(t *testing.T) {
	// Mirrors the standalone service list compiled by the all-in-one's
	// go-services-builder stage — must exclude anything absorbed into the
	// app monolith (ops + core, per ADR 025).
	want := []string{
		"chat", "erp",
		"notification", "ws",
	}
	for _, name := range want {
		if _, ok := upstreamAddrs[name]; !ok {
			t.Errorf("upstreamAddrs missing entry for %q", name)
		}
	}
}
