// Package server provides bootstrap helpers for SDA microservices.
//
// Eliminates the ~30 lines of identical boilerplate in every cmd/main.go:
// logger init, signal context, OpenTelemetry setup, chi router with base
// middleware, HTTP server with standard timeouts, and graceful shutdown.
//
// Usage:
//
//	app := server.New("sda-auth", server.WithPort("AUTH_PORT", "8001"))
//	r := app.Router() // chi.Router pre-wired with RequestID, RealIP, Recoverer, SecureHeaders, Timeout
//	ctx := app.Context() // signal-aware context (SIGINT/SIGTERM)
//	// ... service-specific setup using ctx ...
//	r.Get("/health", hc.Handler())
//	r.Get("/v1/info", build.Handler("sda-auth"))
//	r.Post("/v1/auth/login", handler.Login)
//	app.Run() // blocks until signal, then graceful shutdown
//
// The package intentionally does NOT absorb DB, NATS, Redis, or gRPC setup.
// Those stay in each main.go because they vary across services.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Camionerou/rag-saldivia/pkg/build"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// App holds the shared resources for an SDA microservice.
type App struct {
	name    string
	port    string
	timeout time.Duration // 0 = no timeout middleware (for WebSocket/SSE)
	router  chi.Router
	ctx     context.Context
	cancel  context.CancelFunc
	cleanup []func()
}

// Option configures an App.
type Option func(*App)

// WithPort sets the port env var name and default value.
func WithPort(envVar, defaultPort string) Option {
	return func(a *App) {
		a.port = config.Env(envVar, defaultPort)
	}
}

// WithTimeout sets the request timeout. Default is 30s.
// Use 0 to disable the timeout middleware (required for WebSocket and SSE services).
func WithTimeout(d time.Duration) Option {
	return func(a *App) {
		a.timeout = d
	}
}

// New creates a new App with logger, signal context, OTel, and chi router.
// The router comes pre-wired with: RequestID, RealIP, Recoverer, SecureHeaders, Timeout(30s).
func New(name string, opts ...Option) *App {
	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Signal context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	app := &App{
		name:    name,
		port:    "8080",
		timeout: 30 * time.Second,
		ctx:     ctx,
		cancel:  cancel,
	}

	for _, opt := range opts {
		opt(app)
	}

	// OpenTelemetry
	version := build.ReadVersionFile("VERSION")
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    name,
		ServiceVersion: version,
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		app.cleanup = append(app.cleanup, func() { _ = otelShutdown(context.Background()) })
	}

	// Chi router with base middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())
	if app.timeout > 0 {
		r.Use(middleware.Timeout(app.timeout))
	}

	// /v1/info endpoint
	r.Get("/v1/info", build.Handler(name))

	app.router = r
	return app
}

// Context returns the signal-aware context. Use for DB connections, NATS, etc.
func (a *App) Context() context.Context {
	return a.ctx
}

// Router returns the chi.Router pre-wired with base middleware.
func (a *App) Router() chi.Router {
	return a.router
}

// Port returns the configured port.
func (a *App) Port() string {
	return a.port
}

// Name returns the service name.
func (a *App) Name() string {
	return a.name
}

// OnShutdown registers a cleanup function called during graceful shutdown.
func (a *App) OnShutdown(fn func()) {
	a.cleanup = append(a.cleanup, fn)
}

// Run starts the HTTP server and blocks until SIGINT/SIGTERM.
// Performs graceful shutdown with 30s timeout, then runs cleanup functions.
// WriteTimeout defaults to 30s. Use WithWriteTimeout(0) for WebSocket services.
func (a *App) Run() {
	a.run(30*time.Second, 60*time.Second)
}

// RunWithWriteTimeout is like Run but allows custom WriteTimeout.
// Use 0 for WebSocket services (long-lived connections).
func (a *App) RunWithWriteTimeout(writeTimeout time.Duration) {
	idleTimeout := 60 * time.Second
	if writeTimeout == 0 {
		idleTimeout = 120 * time.Second
	}
	a.run(writeTimeout, idleTimeout)
}

func (a *App) run(writeTimeout, idleTimeout time.Duration) {
	srv := &http.Server{
		Addr:              ":" + a.port,
		Handler:           otelhttp.NewHandler(a.router, a.name),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		slog.Info(a.name+" starting", "port", a.port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-a.ctx.Done()
	slog.Info(a.name + " shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	defer a.cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	for i := len(a.cleanup) - 1; i >= 0; i-- {
		a.cleanup[i]()
	}

	slog.Info(a.name + " stopped")
}
