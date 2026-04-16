package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	chatv1 "github.com/Camionerou/rag-saldivia/gen/go/chat/v1"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	sdagrpc "github.com/Camionerou/rag-saldivia/pkg/grpc"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/outbox"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/services/chat/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/chat/internal/service"
)

func main() {
	app := server.New("sda-chat", server.WithPort("CHAT_PORT", "8003"))
	ctx := app.Context()

	dbURL := config.Env("POSTGRES_TENANT_URL", "")
	tenantSlug := config.Env("TENANT_SLUG", "dev")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	natsURL := config.Env("NATS_URL", nats.DefaultURL)

	if dbURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	// NATS (best-effort — retries in background, events fail gracefully)
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { _ = nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

	// Outbox drainer — publishes spine events from event_outbox to NATS.
	drainer := outbox.NewDrainer(pool, nc, tenantSlug)
	drainerCtx, drainerCancel := context.WithCancel(ctx)
	go drainer.Run(drainerCtx)
	app.OnShutdown(drainerCancel)

	chatSvc := service.NewChat(pool, tenantSlug)
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

	r := app.Router()
	r.Get("/health", hc.Handler())

	// All chat routes require authentication
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false}))
		r.Mount("/v1/chat/sessions", chatHandler.Routes())
	})

	// gRPC server (internal inter-service communication)
	grpcPort := config.Env("CHAT_GRPC_PORT", "50052")
	grpcSrv := sdagrpc.NewServer(sdagrpc.InterceptorConfig{PublicKey: publicKey, Blacklist: blacklist, FailOpen: false})
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
	app.OnShutdown(grpcSrv.GracefulStop)

	app.Run()
}
