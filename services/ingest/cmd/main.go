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
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/service"
)

func main() {
	app := server.New("sda-ingest", server.WithPort("INGEST_PORT", "8007"))
	ctx := app.Context()

	dbURL := config.Env("POSTGRES_TENANT_URL", "")
	natsURL := config.Env("NATS_URL", nats.DefaultURL)
	blueprintURL := config.Env("RAG_SERVER_URL", "http://localhost:8081")
	stagingDir := config.Env("INGEST_STAGING_DIR", "/tmp/ingest-staging")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// Database
	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	// NATS (required for async pipeline)
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err, "url", config.RedactURL(natsURL))
		os.Exit(1)
	}
	app.OnShutdown(func() { nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

	publisher := natspub.New(nc)

	cfg := service.Config{
		BlueprintURL: blueprintURL,
		StagingDir:   stagingDir,
		Timeout:      120 * time.Second,
	}

	// Service + Worker
	ingestSvc := service.New(pool, nc, publisher, cfg)

	worker := service.NewWorker(nc, ingestSvc, publisher, cfg)
	if err := worker.Start(ctx); err != nil {
		slog.Error("failed to start ingest worker", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(worker.Stop)

	// Health checker
	hc := health.New("ingest")
	hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}

	// HTTP
	ingestHandler := handler.NewIngest(ingestSvc)

	r := app.Router()

	// Health check outside auth middleware (monitoring needs unauthenticated access)
	r.Get("/health", hc.Handler())

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Mount("/v1/ingest", ingestHandler.Routes())
	})

	app.Run()
}
