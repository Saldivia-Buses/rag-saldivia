package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

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
	app.OnShutdown(func() { _ = nc.Drain() })

	mailer := service.NewSMTPMailer(
		config.Env("SMTP_HOST", "localhost"),
		config.Env("SMTP_PORT", "1025"),
		config.Env("SMTP_FROM", "noreply@sda.local"),
	)
	notifSvc := service.New(pool)
	notifSvc.SetMailer(mailer)
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
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false}))
		r.Mount("/v1/notifications", notifHandler.Routes())
	})

	// ── Internal routes (no Traefik, no JWT — token-based auth) ────────
	// The alert webhook is authenticated via a shared secret from Docker
	// secrets, not JWT. It's only reachable from the Docker internal
	// network (DS3).
	setupAlertWebhook(ctx, app, r, mailer)

	app.Run()
}

// setupAlertWebhook configures the internal alert webhook endpoint.
// Requires POSTGRES_PLATFORM_URL for alert persistence and
// ALERTMANAGER_WEBHOOK_TOKEN for auth. Both are optional — if not
// configured, the webhook is disabled (useful for dev environments).
func setupAlertWebhook(ctx context.Context, app *server.App, r chi.Router, mailer *service.SMTPMailer) {
	platformURL := loadSecret("/run/secrets/db_platform_url",
		config.Env("POSTGRES_PLATFORM_URL", ""))
	webhookToken := loadSecret("/run/secrets/alertmanager_webhook_token",
		config.Env("ALERTMANAGER_WEBHOOK_TOKEN", ""))

	if platformURL == "" || webhookToken == "" {
		slog.Info("alert webhook disabled (platform DB URL or webhook token not configured)")
		return
	}

	if len(webhookToken) < 32 {
		slog.Error("alert webhook token too short, minimum 32 bytes required")
		return
	}

	platformPool, err := database.NewPool(ctx, platformURL)
	if err != nil {
		slog.Error("failed to connect to platform database for alerts", "error", err)
		return
	}
	app.OnShutdown(platformPool.Close)

	alertStore := service.NewPgAlertStore(platformPool)
	alertTo := config.Env("ALERT_CRITICAL_EMAIL", "")

	webhookHandler := handler.NewAlertWebhook(webhookToken, alertStore, mailer, alertTo)
	r.Post("/internal/webhook/alert", webhookHandler.HandleAlertWebhook)

	slog.Info("alert webhook enabled", "path", "/internal/webhook/alert")
}

// loadSecret reads a Docker secret file, falling back to a default value.
func loadSecret(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(data))
}
