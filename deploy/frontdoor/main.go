// Command frontdoor is the per-tenant container's :80 entrypoint.
//
// It owns the public port of the all-in-one tenant image (ADR 023 + 024):
//
//	/readyz, /healthz    → answered in-process
//	/v1/<service>/...    → reverse-proxied to that Go service's localhost port
//	/ws, /ws/...         → reverse-proxied to the ws service (WebSocket upgrade)
//	/*                   → reverse-proxied to the Next.js standalone on :3000
//
// One small Go binary replaces what would otherwise be an in-container
// Traefik (explicitly rejected in ADR 024 alt 5). When the service
// consolidation campaign (ADR 021) lands and the 13 services collapse into
// a single monolith, this binary folds into that monolith's chi router and
// goes away entirely.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// upstreams is the service name → localhost URL map. Ports match the
// per-service conventions in the Makefile (dev-services target). When the
// consolidation campaign absorbs these into a monolith, this map is the one
// place that has to change.
var upstreams = map[string]string{
	"auth":         "http://127.0.0.1:8001",
	"ws":           "http://127.0.0.1:8002",
	"chat":         "http://127.0.0.1:8003",
	"agent":        "http://127.0.0.1:8004",
	"notification": "http://127.0.0.1:8005",
	"platform":     "http://127.0.0.1:8006",
	"ingest":       "http://127.0.0.1:8007",
	"feedback":     "http://127.0.0.1:8008",
	"traces":       "http://127.0.0.1:8009",
	"search":       "http://127.0.0.1:8010",
	"erp":          "http://127.0.0.1:8013",
	"bigbrother":   "http://127.0.0.1:8011",
	"healthwatch":  "http://127.0.0.1:8012",
}

const nextjsURL = "http://127.0.0.1:3000"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	addr := os.Getenv("FRONTDOOR_ADDR")
	if addr == "" {
		addr = ":80"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /readyz", ready)
	mux.HandleFunc("GET /healthz", ready)
	mux.Handle("/v1/", v1Proxy(logger))
	mux.Handle("/ws", wsProxy(logger))
	mux.Handle("/ws/", wsProxy(logger))
	mux.Handle("/", nextjsProxy(logger))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("frontdoor listening", "addr", addr, "tenant", os.Getenv("SDA_TENANT"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen failed", "err", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("frontdoor shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("shutdown error", "err", err.Error())
	}
}

func ready(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"tenant":   os.Getenv("SDA_TENANT"),
		"hostname": hostname(),
	})
}

// v1Proxy extracts the service name from /v1/<service>/... and proxies to it.
// It preserves the original path, so the upstream service sees the same
// /v1/<service>/... URL it would see behind Traefik today.
func v1Proxy(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /v1/<service>/<rest...>
		rest := strings.TrimPrefix(r.URL.Path, "/v1/")
		svc, _, _ := strings.Cut(rest, "/")
		if svc == "" {
			http.Error(w, "missing service in /v1/<service>/...", http.StatusNotFound)
			return
		}
		target, ok := upstreams[svc]
		if !ok {
			logger.Warn("unknown upstream", "service", svc, "path", r.URL.Path)
			http.Error(w, "unknown service: "+svc, http.StatusNotFound)
			return
		}
		proxy(w, r, target, logger)
	})
}

func wsProxy(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy(w, r, upstreams["ws"], logger)
	})
}

func nextjsProxy(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy(w, r, nextjsURL, logger)
	})
}

// proxy reverse-proxies a single request to `target`. httputil.ReverseProxy
// handles WebSocket upgrades transparently when the default transport is
// used — the hop-by-hop header `Connection: Upgrade` is preserved by the
// standard library's reverse proxy since Go 1.12.
func proxy(w http.ResponseWriter, r *http.Request, target string, logger *slog.Logger) {
	u, err := url.Parse(target)
	if err != nil {
		logger.Error("bad upstream url", "target", target, "err", err.Error())
		http.Error(w, "bad upstream", http.StatusInternalServerError)
		return
	}
	rp := httputil.NewSingleHostReverseProxy(u)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Warn("upstream failed", "target", target, "path", r.URL.Path, "err", err.Error())
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	// Rewrite to the backend host so Host header matches the upstream's
	// expectations (most Go services don't care, but be explicit).
	rp.Director = func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "sda-frontdoor")
		}
	}
	rp.ServeHTTP(w, r)
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}
