package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/services/feedback/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/feedback/internal/service"
)

func main() {
	app := server.New("sda-feedback", server.WithPort("FEEDBACK_PORT", "8008"))
	ctx := app.Context()

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

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// Connect to tenant database
	tenantPool, err := database.NewPool(ctx, tenantDBURL)
	if err != nil {
		slog.Error("failed to connect to tenant database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(tenantPool.Close)
	if err := tenantPool.Ping(ctx); err != nil {
		slog.Error("failed to ping tenant database", "error", err)
		os.Exit(1)
	}

	// Connect to platform database
	platformPool, err := database.NewPool(ctx, platformDBURL)
	if err != nil {
		slog.Error("failed to connect to platform database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(platformPool.Close)
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
	app.OnShutdown(func() { nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

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
	app.OnShutdown(consumer.Stop)

	// Start aggregator
	aggInterval := 1 * time.Hour
	if v := config.Env("AGGREGATION_INTERVAL", ""); v == "1m" {
		aggInterval = 1 * time.Minute // for testing
	}
	tenantID := config.Env("TENANT_ID", "dev")
	tenantSlug := config.Env("TENANT_SLUG", "dev")
	aggregator := service.NewAggregator(tenantPool, platformPool, feedbackSvc, alerter, aggInterval)
	aggregator.Start(ctx, tenantID, tenantSlug)
	app.OnShutdown(aggregator.Stop)

	// Health checker
	hc := health.New("feedback")
	hc.Add("postgres-tenant", func(ctx context.Context) error { return tenantPool.Ping(ctx) })
	hc.Add("postgres-platform", func(ctx context.Context) error { return platformPool.Ping(ctx) })
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}

	// Router
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	r := app.Router()
	r.Get("/health", hc.Handler())

	// Tenant-scoped feedback endpoints (require auth)
	feedbackHandler := handler.NewFeedback(feedbackSvc.Repo(), platformPool)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Mount("/v1/feedback", feedbackHandler.Routes())
	})

	// Platform admin feedback endpoints (require admin JWT)
	platformFeedbackHandler := handler.NewPlatformFeedback(platformPool)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false})) // admin routes: security > availability
		r.Mount("/v1/platform/feedback", platformFeedbackHandler.Routes())
	})

	app.Run()
}
