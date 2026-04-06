package main

import (
	"context"
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

	searchv1 "github.com/Camionerou/rag-saldivia/gen/go/search/v1"
	"github.com/Camionerou/rag-saldivia/pkg/audit"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	sdagrpc "github.com/Camionerou/rag-saldivia/pkg/grpc"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/search/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/search/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("SEARCH_PORT", "8010")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	tenantDBURL := config.Env("POSTGRES_TENANT_URL", "")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-search",
		ServiceVersion: "0.1.0",
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// Connect to tenant DB (read document_pages, document_trees)
	pool, err := pgxpool.New(ctx, tenantDBURL)
	if err != nil {
		slog.Error("failed to connect to tenant db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// LLM client for tree navigation
	llmEndpoint := config.Env("SGLANG_LLM_URL", "http://localhost:8102")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")

	searchSvc := service.New(pool, llmEndpoint, llmModel)
	auditWriter := audit.NewWriter(pool)
	searchHandler := handler.New(searchSvc, auditWriter)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(55 * time.Second))
	r.Use(sdamw.SecureHeaders())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	searchRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 30, Window: time.Minute, KeyFunc: sdamw.ByUser})

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Use(searchRL)
		r.Mount("/v1/search", searchHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-search"),
		ReadTimeout:  10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// gRPC server (internal inter-service communication)
	grpcPort := config.Env("SEARCH_GRPC_PORT", "50051")
	grpcSrv := sdagrpc.NewServer(sdagrpc.InterceptorConfig{PublicKey: publicKey, Blacklist: blacklist, FailOpen: true})
	searchv1.RegisterSearchServiceServer(grpcSrv, handler.NewGRPC(searchSvc))
	sdagrpc.RegisterHealthServer(grpcSrv)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			slog.Error("grpc listen failed", "error", err)
			os.Exit(1)
		}
		slog.Info("search grpc server starting", "port", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			slog.Error("grpc serve error", "error", err)
		}
	}()

	// HTTP server (frontend REST + health checks)
	go func() {
		slog.Info("search service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("search service shutting down")

	grpcSrv.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}



