package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
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
	"github.com/Camionerou/rag-saldivia/services/traces/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/traces/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := env("TRACES_PORT", "8009")
	publicKey := loadPublicKey()
	platformDBURL := env("POSTGRES_PLATFORM_URL", "")
	natsURL := env("NATS_URL", "nats://localhost:4222")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, platformDBURL)
	if err != nil {
		slog.Error("failed to connect to platform db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tracesSvc := service.New(pool)
	tracesHandler := handler.New(tracesSvc)

	// NATS subscriber for trace events
	nc, err := nats.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to nats", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		slog.Error("failed to get jetstream", "error", err)
		os.Exit(1)
	}

	// Ensure stream
	js.AddStream(&nats.StreamConfig{
		Name:     "TRACES",
		Subjects: []string{"traces.>"},
	})

	// Subscribe to trace events
	js.Subscribe("traces.start.*", func(msg *nats.Msg) {
		var evt service.TraceStartEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			slog.Error("invalid trace start event", "error", err)
			msg.Ack()
			return
		}
		if err := tracesSvc.RecordTraceStart(ctx, evt); err != nil {
			slog.Error("record trace start failed", "error", err, "trace_id", evt.TraceID)
		}
		msg.Ack()
	}, nats.Durable("traces-start"), nats.ManualAck())

	js.Subscribe("traces.end.*", func(msg *nats.Msg) {
		var evt service.TraceEndEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			slog.Error("invalid trace end event", "error", err)
			msg.Ack()
			return
		}
		if err := tracesSvc.RecordTraceEnd(ctx, evt); err != nil {
			slog.Error("record trace end failed", "error", err, "trace_id", evt.TraceID)
		}
		msg.Ack()
	}, nats.Durable("traces-end"), nats.ManualAck())

	js.Subscribe("traces.event.*", func(msg *nats.Msg) {
		var evt service.TraceEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			slog.Error("invalid trace event", "error", err)
			msg.Ack()
			return
		}
		if err := tracesSvc.RecordEvent(ctx, evt); err != nil {
			slog.Error("record event failed", "error", err, "trace_id", evt.TraceID)
		}
		msg.Ack()
	}, nats.Durable("traces-events"), nats.ManualAck())

	slog.Info("nats trace subscribers active")

	// HTTP server
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(sdamw.SecureHeaders())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Group(func(r chi.Router) {
		r.Use(sdamw.Auth(publicKey))
		r.Mount("/v1/traces", tracesHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("traces service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("traces service shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
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
