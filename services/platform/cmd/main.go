package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	"github.com/Camionerou/rag-saldivia/pkg/build"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/services/platform/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/platform/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("PLATFORM_PORT", "8006")
	dbURL := config.Env("POSTGRES_PLATFORM_URL", "")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	if dbURL == "" {
		slog.Error("POSTGRES_PLATFORM_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-platform",
		ServiceVersion: "1.0.0",
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to platform database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// NATS for lifecycle event publishing
	natsURL := config.Env("NATS_URL", "nats://localhost:4222")
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Warn("nats connect failed, lifecycle events disabled", "error", err)
	} else {
		defer nc.Drain()
	}
	publisher := natspub.New(nc)

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	platformSvc := service.New(pool, publisher)
	platformSlug := config.Env("PLATFORM_TENANT_SLUG", "platform")
	platformHandler := handler.NewPlatform(platformSvc, publicKey, platformSlug, blacklist)

	// Health checker
	hc := health.New("platform")
	hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
	if nc != nil {
		hc.Add("nats", func(ctx context.Context) error {
			if !nc.IsConnected() {
				return fmt.Errorf("nats disconnected")
			}
			return nil
		})
	}
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", hc.Handler())
	r.Get("/v1/info", build.Handler("sda-platform"))
	r.Mount("/v1/platform", platformHandler.Routes())

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-platform"),
		ReadTimeout:  10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("platform service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("platform service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("platform service stopped")
}



