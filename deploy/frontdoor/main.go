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

// upstreamAddrs is the service name → localhost URL map. Ports match each
// service's `server.WithPort(...)` default in services/<svc>/cmd/main.go.
// When the consolidation campaign absorbs these into a monolith, this map
// is the one place that has to change.
var upstreamAddrs = map[string]string{
	"ws":           "http://127.0.0.1:8002",
	"chat":         "http://127.0.0.1:8003",
	"notification": "http://127.0.0.1:8005",
	"erp":          "http://127.0.0.1:8013",
}

const nextjsURL = "http://127.0.0.1:3000"

// router holds pre-built reverse proxies so the hot path does no allocation.
type router struct {
	logger    *slog.Logger
	upstreams map[string]*httputil.ReverseProxy
	nextjs    *httputil.ReverseProxy
}

func newRouter(logger *slog.Logger) (*router, error) {
	return newRouterWithAddrs(logger, upstreamAddrs, nextjsURL)
}

// newRouterWithAddrs lets tests inject a fake upstream map + nextjs URL
// without touching the production defaults.
func newRouterWithAddrs(logger *slog.Logger, addrs map[string]string, nextjs string) (*router, error) {
	r := &router{
		logger:    logger,
		upstreams: make(map[string]*httputil.ReverseProxy, len(addrs)),
	}
	for name, addr := range addrs {
		rp, err := buildProxy(logger, addr)
		if err != nil {
			return nil, err
		}
		r.upstreams[name] = rp
	}
	nx, err := buildProxy(logger, nextjs)
	if err != nil {
		return nil, err
	}
	r.nextjs = nx
	return r, nil
}

// buildProxy constructs a ReverseProxy for a single upstream. Uses the
// modern Rewrite API (Go 1.20+): SetXForwarded populates X-Forwarded-For,
// X-Forwarded-Host and X-Forwarded-Proto, which the legacy Director API
// only partially did. WebSocket upgrades flow through unchanged since the
// default Transport preserves the Connection: Upgrade header.
func buildProxy(logger *slog.Logger, addr string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)
			r.SetXForwarded()
			if _, ok := r.Out.Header["User-Agent"]; !ok {
				r.Out.Header.Set("User-Agent", "sda-frontdoor")
			}
		},
		ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
			logger.Warn("upstream failed",
				"target", addr,
				"path", req.URL.Path,
				"err", err.Error())
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
		},
	}
	return rp, nil
}

func (r *router) v1(w http.ResponseWriter, req *http.Request) {
	rest := strings.TrimPrefix(req.URL.Path, "/v1/")
	svc, _, _ := strings.Cut(rest, "/")
	if svc == "" {
		http.Error(w, "missing service in /v1/<service>/...", http.StatusNotFound)
		return
	}
	rp, ok := r.upstreams[svc]
	if !ok {
		r.logger.Warn("unknown upstream", "service", svc, "path", req.URL.Path)
		http.Error(w, "unknown service: "+svc, http.StatusNotFound)
		return
	}
	rp.ServeHTTP(w, req)
}

func (r *router) ws(w http.ResponseWriter, req *http.Request) {
	r.upstreams["ws"].ServeHTTP(w, req)
}

func (r *router) next(w http.ResponseWriter, req *http.Request) {
	r.nextjs.ServeHTTP(w, req)
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

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func newMux(r *router) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /readyz", ready)
	mux.HandleFunc("GET /healthz", ready)
	mux.HandleFunc("/v1/", r.v1)
	mux.HandleFunc("/ws", r.ws)
	mux.HandleFunc("/ws/", r.ws)
	mux.HandleFunc("/", r.next)
	return mux
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	addr := os.Getenv("FRONTDOOR_ADDR")
	if addr == "" {
		addr = ":80"
	}

	r, err := newRouter(logger)
	if err != nil {
		logger.Error("router init failed", "err", err.Error())
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           newMux(r),
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
