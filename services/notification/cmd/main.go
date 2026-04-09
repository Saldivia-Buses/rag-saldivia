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
	"github.com/Camionerou/rag-saldivia/services/notification/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/notification/internal/service"
)

func main() {
	app := server.New("sda-notification", server.WithPort("NOTIFICATION_PORT", "8005"))
	ctx := app.Context()

	dbURL := config.Env("POSTGRES_TENANT_URL", "")
	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	nc, err := natspub.Connect(config.Env("NATS_URL", nats.DefaultURL))
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { nc.Drain() })

	mailer := service.NewSMTPMailer(
		config.Env("SMTP_HOST", "localhost"),
		config.Env("SMTP_PORT", "1025"),
		config.Env("SMTP_FROM", "noreply@sda.local"),
	)
	notifSvc := service.New(pool)
	consumer := service.NewConsumer(nc, notifSvc, mailer)
	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start NATS consumer", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(consumer.Stop)

	notifHandler := handler.NewNotification(notifSvc)

	hc := health.New("notification")
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
		r.Mount("/v1/notifications", notifHandler.Routes())
	})

	app.Run()
}
