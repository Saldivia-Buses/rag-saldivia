package main

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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := env("AUTH_PORT", "8001")
	jwtSecret := env("JWT_SECRET", "")
	dbURL := env("POSTGRES_TENANT_URL", "")
	tenantID := env("TENANT_ID", "dev")
	tenantSlug := env("TENANT_SLUG", "dev")

	if jwtSecret == "" {
		slog.Error("JWT_SECRET is required")
		os.Exit(1)
	}
	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Connect to tenant database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	// Connect to NATS (best-effort — retries in background, events fail gracefully)
	natsURL := env("NATS_URL", nats.DefaultURL)
	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			slog.Warn("NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			slog.Info("NATS reconnected")
		}),
	)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	publisher := natspub.New(nc)
	slog.Info("connected to NATS", "url", natsURL)

	// Initialize services
	jwtCfg := sdajwt.DefaultConfig(jwtSecret)
	authSvc := service.NewAuth(pool, jwtCfg, tenantID, tenantSlug, publisher)
	authHandler := handler.NewAuth(authSvc)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", authHandler.Health)
	r.Post("/v1/auth/login", authHandler.Login)
	r.Post("/v1/auth/refresh", authHandler.Refresh)
	r.Post("/v1/auth/logout", authHandler.Logout)

	// Protected routes — require valid access token
	r.Group(func(r chi.Router) {
		r.Use(sdamw.Auth(jwtSecret))
		r.Get("/v1/auth/me", authHandler.Me)
		r.Get("/v1/modules/enabled", authHandler.EnabledModules)
	})

	// Server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start
	go func() {
		slog.Info("auth service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	slog.Info("auth service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("auth service stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
