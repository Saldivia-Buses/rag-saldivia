package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/services/healthwatch/internal/collector"
	"github.com/Camionerou/rag-saldivia/services/healthwatch/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/healthwatch/internal/service"
)

func main() {
	app := server.New("sda-healthwatch", server.WithPort("HEALTHWATCH_PORT", "8014"))
	ctx := app.Context()

	// Platform DB for health_snapshots + triage_records persistence
	dbURL := loadSecret("/run/secrets/db_platform_url",
		config.Env("POSTGRES_PLATFORM_URL", ""))
	if dbURL == "" {
		slog.Error("POSTGRES_PLATFORM_URL or db_platform_url secret is required")
		os.Exit(1)
	}

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))
	if blacklist == nil {
		slog.Error("redis is required for token revocation on admin endpoints")
		os.Exit(1)
	}

	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to platform database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	// Collectors
	promCollector := collector.NewPrometheus(
		config.Env("PROMETHEUS_URL", "http://prometheus:9090"),
	)
	dockerCollector := collector.NewDocker(
		config.Env("DOCKER_PROXY_URL", "http://docker-socket-proxy:2375"),
	)
	svcCollector := collector.NewService()

	// NATS — for DLQ consumer + replay
	natsURL := config.Env("NATS_URL", nats.DefaultURL)
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { _ = nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

	// Service layer
	svc := service.New(pool, promCollector, dockerCollector, svcCollector)

	// Start retention cleanup scheduler
	svc.StartRetentionCleanup(ctx)

	// DLQ consumer — reads dlq.> and persists to dead_events
	dlqConsumer := service.NewDLQConsumer(pool, nc)
	if err := dlqConsumer.Start(ctx); err != nil {
		slog.Error("failed to start DLQ consumer", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(dlqConsumer.Stop)

	// Health checker
	hc := health.New("healthwatch")
	hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})

	// Handler + routes
	platformSlug := config.Env("PLATFORM_TENANT_SLUG", "platform")
	hw := handler.New(svc, publicKey, platformSlug, blacklist)
	dlqHandler := handler.NewDLQ(pool, nc)

	r := app.Router()
	r.Get("/health", hc.Handler())
	r.Mount("/v1/healthwatch", hw.Routes())
	// DLQ admin routes — guarded by requirePlatformAdmin via hw.AdminRoutes
	hw.AdminRoutes(r, dlqHandler)

	app.Run()
}

// loadSecret reads a Docker secret file, falling back to a default value.
func loadSecret(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(data))
}
