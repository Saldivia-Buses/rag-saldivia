package main

import (
	"context"
	"crypto/ed25519"
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
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := env("INGEST_PORT", "8007")
	dbURL := env("POSTGRES_TENANT_URL", "")
	natsURL := env("NATS_URL", nats.DefaultURL)
	blueprintURL := env("RAG_SERVER_URL", "http://localhost:8081")
	stagingDir := env("INGEST_STAGING_DIR", "/tmp/ingest-staging")
	publicKey := loadPublicKey()

	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-ingest",
		ServiceVersion: "1.0.0",
		Endpoint:       env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// Database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// NATS (required for async pipeline)
	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			slog.Warn("NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			slog.Info("NATS reconnected")
		}),
	)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err, "url", natsURL)
		os.Exit(1)
	}
	defer nc.Close()
	slog.Info("connected to NATS", "url", natsURL)

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
	defer worker.Stop()

	// HTTP
	ingestHandler := handler.NewIngest(ingestSvc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check outside auth middleware (monitoring needs unauthenticated access)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"ingest"}`))
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(sdamw.Auth(publicKey))
		r.Mount("/v1/ingest", ingestHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-ingest"),
		ReadTimeout:  120 * time.Second, // large uploads
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("ingest service starting", "port", port, "blueprint", blueprintURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("ingest service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
	slog.Info("ingest service stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadPublicKey() ed25519.PublicKey {
	pubB64 := env("JWT_PUBLIC_KEY", "")
	if pubB64 == "" {
		slog.Error("JWT_PUBLIC_KEY is required")
		os.Exit(1)
	}
	key, err := sdajwt.ParsePublicKeyEnv(pubB64)
	if err != nil {
		slog.Error("failed to parse JWT_PUBLIC_KEY", "error", err)
		os.Exit(1)
	}
	return key
}
