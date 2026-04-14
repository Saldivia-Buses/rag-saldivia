package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/services/platform/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/platform/internal/service"
)

func main() {
	app := server.New("sda-platform", server.WithPort("PLATFORM_PORT", "8006"))
	ctx := app.Context()

	dbURL := config.Env("POSTGRES_PLATFORM_URL", "")
	if dbURL == "" {
		slog.Error("POSTGRES_PLATFORM_URL is required")
		os.Exit(1)
	}

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to platform database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	// NATS for lifecycle event publishing
	natsURL := config.Env("NATS_URL", "nats://localhost:4222")
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Warn("nats connect failed, lifecycle events disabled", "error", err)
	} else {
		app.OnShutdown(func() { nc.Drain() })
	}
	publisher := natspub.New(nc)

	// Token blacklist (shared Redis) — required for admin endpoints
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))
	if blacklist == nil {
		slog.Error("redis is required for token revocation on admin endpoints")
		os.Exit(1)
	}

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

	r := app.Router()
	r.Get("/health", hc.Handler())
	r.Mount("/v1/platform", platformHandler.Routes())
	r.Mount("/v1/flags", platformHandler.FlagsRoutes())

	app.Run()
}
