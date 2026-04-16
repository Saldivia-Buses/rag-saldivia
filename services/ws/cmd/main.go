package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/services/ws/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/ws/internal/hub"
)

func main() {
	app := server.New("sda-ws", server.WithPort("WS_PORT", "8002"), server.WithTimeout(0))
	ctx := app.Context()

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	natsURL := config.Env("NATS_URL", nats.DefaultURL)

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// Connect to NATS
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err, "url", config.RedactURL(natsURL))
		os.Exit(1)
	}
	app.OnShutdown(func() { _ = nc.Drain() })
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
	app.OnShutdown(bridge.Stop)

	// Wire mutations via gRPC to Chat service
	chatGRPC := config.Env("CHAT_GRPC_URL", "")
	if mutations := hub.NewMutationHandler(chatGRPC); mutations != nil {
		h.Mutations = mutations
		app.OnShutdown(mutations.Close)
	}

	// Handlers
	wsHandler := handler.NewWS(h, publicKey, blacklist)

	// Router
	r := app.Router()
	r.Get("/ws", wsHandler.Upgrade)

	hc := health.New("ws-hub")
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}
	hc.AddExtra(func() (string, any) { return "clients", h.ClientCount() })
	r.Get("/health", hc.Handler())

	// WebSocket connections are long-lived — need 0 write timeout
	app.RunWithWriteTimeout(0)
}
