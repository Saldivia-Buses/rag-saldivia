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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/notification/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/notification/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("NOTIFICATION_PORT", "8005")
	dbURL := config.Env("POSTGRES_TENANT_URL", "")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	natsURL := config.Env("NATS_URL", nats.DefaultURL)
	smtpHost := config.Env("SMTP_HOST", "localhost")
	smtpPort := config.Env("SMTP_PORT", "1025")
	smtpFrom := config.Env("SMTP_FROM", "noreply@sda.local")

	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-notification",
		ServiceVersion: "1.0.0",
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

	// Database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// NATS
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err, "url", config.RedactURL(natsURL))
		os.Exit(1)
	}
	defer nc.Drain()
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

	// Services
	mailer := service.NewSMTPMailer(smtpHost, smtpPort, smtpFrom)
	notifSvc := service.New(pool)
	consumer := service.NewConsumer(nc, notifSvc, mailer)

	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start NATS consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Stop()

	// HTTP
	notifHandler := handler.NewNotification(notifSvc)

	// Health checker
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

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", hc.Handler())
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Mount("/v1/notifications", notifHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-notification"),
		ReadTimeout:  10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("notification service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("notification service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("notification service stopped")
}



