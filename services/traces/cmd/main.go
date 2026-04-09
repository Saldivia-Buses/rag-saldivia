package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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
	"github.com/Camionerou/rag-saldivia/services/traces/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/traces/internal/service"
)

func main() {
	app := server.New("sda-traces", server.WithPort("TRACES_PORT", "8009"))
	ctx := app.Context()

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	pool, err := database.NewPool(ctx, config.Env("POSTGRES_PLATFORM_URL", ""))
	if err != nil {
		slog.Error("failed to connect to platform db", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	tracesSvc := service.New(pool)
	tracesHandler := handler.New(tracesSvc)

	nc, err := natspub.Connect(config.Env("NATS_URL", nats.DefaultURL))
	if err != nil {
		slog.Error("failed to connect to nats", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(config.Env("NATS_URL", "")))

	tracesConsumer := service.NewConsumer(nc, tracesSvc)
	if err := tracesConsumer.Start(ctx); err != nil {
		slog.Error("failed to start traces consumer", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(tracesConsumer.Stop)

	hc := health.New("traces")
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

	r := app.Router()
	r.Get("/health", hc.Handler())

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Mount("/v1/traces", tracesHandler.Routes())
	})

	app.Run()
}
