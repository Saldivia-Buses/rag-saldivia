package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// buildFakes spins up one httptest.Server per production upstream name
// plus an app fake and a Next.js fake. Each fake writes two response
// headers so tests can assert which upstream saw the request and on
// which path.
func buildFakes(t *testing.T) (svrs map[string]*httptest.Server, app *httptest.Server, nextjs *httptest.Server) {
	t.Helper()
	svrs = make(map[string]*httptest.Server, len(upstreamAddrs))
	for name := range upstreamAddrs {
		name := name
		svrs[name] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Fake-Upstream", name)
			w.Header().Set("X-Fake-Path", req.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, name)
		}))
	}
	app = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Fake-Upstream", "app")
		w.Header().Set("X-Fake-Path", req.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "app")
	}))
	nextjs = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Fake-Upstream", "nextjs")
		w.Header().Set("X-Fake-Path", req.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "nextjs")
	}))
	return svrs, app, nextjs
}

func closeAll(svrs map[string]*httptest.Server, extra ...*httptest.Server) {
	for _, s := range svrs {
		s.Close()
	}
	for _, s := range extra {
		s.Close()
	}
}

func newTestServer(t *testing.T, svrs map[string]*httptest.Server, app, nextjs *httptest.Server) *httptest.Server {
	t.Helper()
	addrs := make(map[string]string, len(svrs))
	for name, s := range svrs {
		addrs[name] = s.URL
	}
	r, err := newRouterWithAddrs(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		addrs,
		app.URL,
		nextjs.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(newMux(r))
}

// withProbeApp swaps out the /readyz probe for a fake and restores the
// original on cleanup. Tests that hit /readyz have to control this — the
// real probe calls the hardcoded appHealthURL (127.0.0.1:8020) which is
// not reachable in unit tests.
func withProbeApp(t *testing.T, fn func(context.Context) error) {
	t.Helper()
	orig := probeApp
	probeApp = fn
	t.Cleanup(func() { probeApp = orig })
}

func TestRouting(t *testing.T) {
	withProbeApp(t, func(context.Context) error { return nil })

	svrs, app, nextjs := buildFakes(t)
	defer closeAll(svrs, app, nextjs)
	srv := newTestServer(t, svrs, app, nextjs)
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
		{"v1 erp", "GET", "/v1/erp/orders", http.StatusOK, "erp"},
		{"v1 with no service is 404", "GET", "/v1/", http.StatusNotFound, ""},
		{"v1 info goes to app", "GET", "/v1/info", http.StatusOK, "app"},
		{"v1 auth goes to app", "POST", "/v1/auth/login", http.StatusOK, "app"},
		{"v1 agent goes to app", "GET", "/v1/agent/sessions", http.StatusOK, "app"},
		{"ws goes to app", "GET", "/ws", http.StatusOK, "app"},
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

// TestReadyzReflectsAppHealth proves the /readyz aggregate: green when
// the app probe succeeds, 503 when it fails. The container HEALTHCHECK
// depends on this — if readyz always said OK, a dead app would still
// look healthy to Docker.
func TestReadyzReflectsAppHealth(t *testing.T) {
	svrs, app, nextjs := buildFakes(t)
	defer closeAll(svrs, app, nextjs)
	srv := newTestServer(t, svrs, app, nextjs)
	defer srv.Close()

	t.Run("ok when probe succeeds", func(t *testing.T) {
		withProbeApp(t, func(context.Context) error { return nil })
		resp, err := srv.Client().Get(srv.URL + "/readyz")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status: got %d want 200", resp.StatusCode)
		}
	})

	t.Run("503 when probe fails", func(t *testing.T) {
		withProbeApp(t, func(context.Context) error { return errors.New("app down") })
		resp, err := srv.Client().Get(srv.URL + "/readyz")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("status: got %d want 503", resp.StatusCode)
		}
	})

	t.Run("healthz stays 200 even with probe failing", func(t *testing.T) {
		withProbeApp(t, func(context.Context) error { return errors.New("app down") })
		resp, err := srv.Client().Get(srv.URL + "/healthz")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("healthz should be liveness-only: got %d want 200", resp.StatusCode)
		}
	})
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
		map[string]string{"erp": up.URL},
		up.URL,
		up.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(newMux(r))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/v1/erp/probe")
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

// TestEmptyV1Path confirms /v1/ (no service) returns the explicit
// "missing service" message — the fallback to app must not swallow
// this case, it's a real client bug worth surfacing.
func TestEmptyV1Path(t *testing.T) {
	svrs, app, nextjs := buildFakes(t)
	defer closeAll(svrs, app, nextjs)
	srv := newTestServer(t, svrs, app, nextjs)
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/v1/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("empty /v1/: got %d want 404", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "missing service") {
		t.Fatalf("expected body to flag missing service; got %q", body)
	}
}

// TestUpstreamMapCoverage guards against the H1-style bug where the
// frontdoor map drifts from the real services/*/cmd/main.go ports.
// app is NOT in the map — it's the fallback target — so the map only
// carries the still-standalone services.
func TestUpstreamMapCoverage(t *testing.T) {
	// Mirrors the standalone service list compiled by the all-in-one's
	// go-services-builder stage — must exclude anything absorbed into the
	// app monolith (ops + core + rag + realtime per ADR 025). Only erp
	// is still a standalone at this point.
	want := []string{"erp"}
	for _, name := range want {
		if _, ok := upstreamAddrs[name]; !ok {
			t.Errorf("upstreamAddrs missing entry for %q", name)
		}
	}
	if _, ok := upstreamAddrs["app"]; ok {
		t.Errorf("app must not be in upstreamAddrs — it is the /v1 fallback, not a named upstream")
	}
}
