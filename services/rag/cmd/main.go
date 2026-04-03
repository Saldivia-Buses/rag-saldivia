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

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/rag/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/rag/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := env("RAG_PORT", "8004")
	jwtSecret := env("JWT_SECRET", "")
	blueprintURL := env("RAG_SERVER_URL", "http://localhost:8081")
	timeoutMs := env("RAG_TIMEOUT_MS", "120000")

	timeout, _ := time.ParseDuration(timeoutMs + "ms")
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-rag",
		ServiceVersion: "1.0.0",
		Endpoint:       env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	ragSvc := service.NewRAG(service.Config{
		BlueprintURL: blueprintURL,
		Timeout:      timeout,
		APIKey:       env("RAG_API_KEY", ""),
		Model:        env("RAG_MODEL", ""),
	})

	ragHandler := handler.NewRAG(ragSvc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", ragHandler.Health)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.Auth(jwtSecret))
		r.Mount("/v1/rag", ragHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-rag"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // no limit for SSE streaming
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("rag service starting", "port", port, "blueprint", blueprintURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("rag service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
	slog.Info("rag service stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
