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
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	"github.com/Camionerou/rag-saldivia/pkg/build"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/ws/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/ws/internal/hub"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("WS_PORT", "8002")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	natsURL := config.Env("NATS_URL", nats.DefaultURL)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-ws",
		ServiceVersion: "1.0.0",
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// Connect to NATS
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err, "url", config.RedactURL(natsURL))
		os.Exit(1)
	}
	defer nc.Drain()
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

	// Create hub
	h := hub.New()
	go h.Run()

	// Start NATS bridge (forwards NATS events to WebSocket clients)
	bridge := hub.NewNATSBridge(h, nc)
	if err := bridge.Start(); err != nil {
		slog.Error("failed to start NATS bridge", "error", err)
		os.Exit(1)
	}
	defer bridge.Stop()

	// Wire mutations via gRPC to Chat service
	chatGRPC := config.Env("CHAT_GRPC_URL", "")
	if mutations := hub.NewMutationHandler(chatGRPC); mutations != nil {
		h.Mutations = mutations
		defer mutations.Close()
	}

	// Handlers
	wsHandler := handler.NewWS(h, publicKey)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())

	r.Get("/ws", wsHandler.Upgrade)
	hc := health.New("ws-hub")
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	hc.AddExtra(func() (string, any) { return "clients", h.ClientCount() })
	r.Get("/health", hc.Handler())
	r.Get("/v1/info", build.Handler("sda-ws"))

	// Server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-ws"),
		ReadTimeout:  10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 0,               // WebSocket connections are long-lived
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("ws-hub starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("ws-hub shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("ws-hub stopped")
}



