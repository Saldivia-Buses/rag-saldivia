// Command frontdoor is the per-tenant container's :80 entrypoint.
//
// It owns the public port of the all-in-one tenant image (ADR 023 + 024):
//
//	/readyz              → in-process, probes app /health on :8020
//	/healthz             → in-process liveness
//	/v1/erp/...          → erp standalone on :8013 (until the ERP fusion)
//	/v1/<anything-else>  → app monolith on :8020
//	/ws                  → app monolith on :8020 (realtime fusion absorbed ws)
//	/*                   → Next.js standalone on :3000
//
// One small Go binary replaces what would otherwise be an in-container
// Traefik (explicitly rejected in ADR 024 alt 5). When the ERP fusion
// lands and erp folds into the monolith, this binary folds into app's
// chi router and goes away entirely.
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

// upstreamAddrs is the service name → localhost URL map for the minority
// of /v1/<svc>/... prefixes that still belong to standalone services.
// Anything not listed here falls through to appURL.
var upstreamAddrs = map[string]string{
	"erp": "http://127.0.0.1:8013",
}

const (
	appURL    = "http://127.0.0.1:8020"
	nextjsURL = "http://127.0.0.1:3000"
)

// appHealthURL is what /readyz probes to prove the monolith is serving.
// Must point at an endpoint that doesn't need auth — /health is wired
// unprotected by services/app/cmd/main.go.
const appHealthURL = appURL + "/health"

// router holds pre-built reverse proxies so the hot path does no allocation.
type router struct {
	logger    *slog.Logger
	upstreams map[string]*httputil.ReverseProxy
	app       *httputil.ReverseProxy
	nextjs    *httputil.ReverseProxy
}

func newRouter(logger *slog.Logger) (*router, error) {
	return newRouterWithAddrs(logger, upstreamAddrs, appURL, nextjsURL)
}

// newRouterWithAddrs lets tests inject fake upstream + app + nextjs URLs
// without touching the production defaults.
func newRouterWithAddrs(logger *slog.Logger, addrs map[string]string, app, nextjs string) (*router, error) {
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
	ap, err := buildProxy(logger, app)
	if err != nil {
		return nil, err
	}
	r.app = ap
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
	if rp, ok := r.upstreams[svc]; ok {
		rp.ServeHTTP(w, req)
		return
	}
	// Default: the app monolith owns everything else under /v1.
	r.app.ServeHTTP(w, req)
}

func (r *router) ws(w http.ResponseWriter, req *http.Request) {
	r.app.ServeHTTP(w, req)
}

func (r *router) next(w http.ResponseWriter, req *http.Request) {
	r.nextjs.ServeHTTP(w, req)
}

// probeApp returns nil if app's /health replies 200 within the timeout.
// Called from /readyz on every request — keep it cheap.
var probeApp = func(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appHealthURL, nil)
	if err != nil {
		return err
	}
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

// readyz returns 503 unless the app monolith is also reporting /health.
// frontdoor alive + app alive is the contract the container HEALTHCHECK
// relies on; degrading one of them should mark the container unhealthy.
func readyz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := "ok"
	code := http.StatusOK
	if err := probeApp(req.Context()); err != nil {
		status = "app-unhealthy"
		code = http.StatusServiceUnavailable
	}
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   status,
		"tenant":   os.Getenv("SDA_TENANT"),
		"hostname": hostname(),
	})
}

// healthz is liveness only — 200 as long as the frontdoor process is
// accepting requests. Does NOT probe app (that's readyz's job).
func healthz(w http.ResponseWriter, _ *http.Request) {
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
	mux.HandleFunc("GET /readyz", readyz)
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("/ws", r.ws)
	mux.HandleFunc("/v1/", r.v1)
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
