package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	chatv1 "github.com/Camionerou/rag-saldivia/gen/go/chat/v1"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	sdagrpc "github.com/Camionerou/rag-saldivia/pkg/grpc"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/chat/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/chat/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("CHAT_PORT", "8003")
	dbURL := config.Env("POSTGRES_TENANT_URL", "")
	tenantSlug := config.Env("TENANT_SLUG", "dev")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	natsURL := config.Env("NATS_URL", nats.DefaultURL)

	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-chat",
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

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// NATS (best-effort — retries in background, events fail gracefully)
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()
	publisher := natspub.New(nc)
	slog.Info("connected to NATS", "url", natsURL)

	chatSvc := service.NewChat(pool, tenantSlug, publisher)
	chatHandler := handler.NewChat(chatSvc)

	// Health checker
	hc := health.New("chat")
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

	// All chat routes require authentication
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Mount("/v1/chat/sessions", chatHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-chat"),
		ReadTimeout:  10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC server (internal inter-service communication)
	grpcPort := config.Env("CHAT_GRPC_PORT", "50052")
	grpcSrv := sdagrpc.NewServer(sdagrpc.InterceptorConfig{PublicKey: publicKey, Blacklist: blacklist, FailOpen: true})
	chatv1.RegisterChatServiceServer(grpcSrv, handler.NewGRPC(chatSvc))
	sdagrpc.RegisterHealthServer(grpcSrv)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			slog.Error("grpc listen failed", "error", err)
			os.Exit(1)
		}
		slog.Info("chat grpc server starting", "port", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			slog.Error("grpc serve error", "error", err)
		}
	}()

	// HTTP server
	go func() {
		slog.Info("chat service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("chat service shutting down")

	grpcSrv.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("chat service stopped")
}



