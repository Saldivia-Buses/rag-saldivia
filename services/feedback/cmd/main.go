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
	"github.com/Camionerou/rag-saldivia/pkg/config"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/feedback/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/feedback/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("FEEDBACK_PORT", "8008")
	tenantDBURL := config.Env("POSTGRES_TENANT_URL", "")
	platformDBURL := config.Env("POSTGRES_PLATFORM_URL", "")

	if tenantDBURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}
	if platformDBURL == "" {
		slog.Error("POSTGRES_PLATFORM_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-feedback",
		ServiceVersion: "0.1.0",
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// Connect to tenant database
	tenantPool, err := pgxpool.New(ctx, tenantDBURL)
	if err != nil {
		slog.Error("failed to connect to tenant database", "error", err)
		os.Exit(1)
	}
	defer tenantPool.Close()
	if err := tenantPool.Ping(ctx); err != nil {
		slog.Error("failed to ping tenant database", "error", err)
		os.Exit(1)
	}

	// Connect to platform database
	platformPool, err := pgxpool.New(ctx, platformDBURL)
	if err != nil {
		slog.Error("failed to connect to platform database", "error", err)
		os.Exit(1)
	}
	defer platformPool.Close()
	if err := platformPool.Ping(ctx); err != nil {
		slog.Error("failed to ping platform database", "error", err)
		os.Exit(1)
	}

	// Connect to NATS
	natsURL := config.Env("NATS_URL", nats.DefaultURL)
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()
	slog.Info("connected to NATS", "url", natsURL)

	// Initialize services
	publisher := natspub.New(nc)
	feedbackSvc := service.NewFeedback(tenantPool, platformPool)
	alerter := service.NewAlerter(platformPool, publisher)

	// Start NATS consumer
	consumer := service.NewConsumer(nc, feedbackSvc)
	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start feedback consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Stop()

	// Start aggregator
	aggInterval := 1 * time.Hour
	if v := config.Env("AGGREGATION_INTERVAL", ""); v == "1m" {
		aggInterval = 1 * time.Minute // for testing
	}
	tenantID := config.Env("TENANT_ID", "dev")
	tenantSlug := config.Env("TENANT_SLUG", "dev")
	aggregator := service.NewAggregator(tenantPool, platformPool, feedbackSvc, alerter, aggInterval)
	aggregator.Start(ctx, tenantID, tenantSlug)
	defer aggregator.Stop()

	// Router — health endpoint only for now (REST handlers in Fase 4)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())
	r.Use(middleware.Timeout(30 * time.Second))

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"feedback"}`))
	})

	// Tenant-scoped feedback endpoints (require auth)
	feedbackHandler := handler.NewFeedback(feedbackSvc.Repo(), platformPool)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Mount("/v1/feedback", feedbackHandler.Routes())
	})

	// Platform admin feedback endpoints (require admin JWT)
	platformFeedbackHandler := handler.NewPlatformFeedback(platformPool)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Mount("/v1/platform/feedback", platformFeedbackHandler.Routes())
	})

	// Server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-feedback"),
		ReadTimeout:  10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("feedback service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	slog.Info("feedback service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("feedback service stopped")
}



