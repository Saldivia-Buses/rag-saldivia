package main

import (
	"context"
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

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := env("INGEST_PORT", "8007")
	dbURL := env("POSTGRES_TENANT_URL", "")
	natsURL := env("NATS_URL", nats.DefaultURL)
	blueprintURL := env("RAG_SERVER_URL", "http://localhost:8081")
	stagingDir := env("INGEST_STAGING_DIR", "/tmp/ingest-staging")
	jwtSecret := env("JWT_SECRET", "")

	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}
	if jwtSecret == "" {
		slog.Error("JWT_SECRET is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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
	r.Use(sdamw.Auth(jwtSecret))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"ingest"}`))
	})
	r.Mount("/v1/ingest", ingestHandler.Routes())

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
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
