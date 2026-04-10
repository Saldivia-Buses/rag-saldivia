package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

func main() {
	app := server.New("sda-erp", server.WithPort("ERP_PORT", "8013"))
	ctx := app.Context()

	tenantDBURL := config.Env("POSTGRES_TENANT_URL", "")
	if tenantDBURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	pool, err := database.NewPool(ctx, tenantDBURL)
	if err != nil {
		slog.Error("failed to connect to tenant database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	nc, err := natspub.Connect(config.Env("NATS_URL", nats.DefaultURL))
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(config.Env("NATS_URL", "")))

	// Core dependencies
	auditWriter := audit.NewWriter(pool)
	publisher := traces.NewPublisher(nc)
	tenantSlug := config.Env("TENANT_SLUG", "dev")

	// Repository + Service + Handler
	repo := repository.New(pool)
	suggestionsSvc := service.NewSuggestions(repo, auditWriter, publisher, tenantSlug)
	suggestionsHandler := handler.NewSuggestions(suggestionsSvc, tenantSlug)

	// Health
	hc := health.New("erp")
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

	// Routes
	r := app.Router()
	r.Get("/health", hc.Handler())

	authRead := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true})
	authWrite := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false})

	r.Group(func(r chi.Router) {
		r.Use(authRead) // default for mount, write endpoints override below
		r.Mount("/v1/erp/suggestions", suggestionsHandler.Routes(authWrite))
		// Future ERP modules mount here:
		// r.Mount("/v1/erp/reclamos", reclamosHandler.Routes())
		// r.Mount("/v1/erp/entregas", entregasHandler.Routes())
	})

	app.Run()
}
