package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	searchv1 "github.com/Camionerou/rag-saldivia/gen/go/search/v1"
	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	sdagrpc "github.com/Camionerou/rag-saldivia/pkg/grpc"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/services/search/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/search/internal/service"
)

func main() {
	app := server.New("sda-search", server.WithPort("SEARCH_PORT", "8010"))
	ctx := app.Context()

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	tenantDBURL := config.Env("POSTGRES_TENANT_URL", "")

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// Connect to tenant DB (read document_pages, document_trees)
	pool, err := database.NewPool(ctx, tenantDBURL)
	if err != nil {
		slog.Error("failed to connect to tenant db", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	// LLM client for tree navigation
	llmEndpoint := config.Env("SGLANG_LLM_URL", "http://localhost:8102")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")

	searchSvc := service.New(pool, llmEndpoint, llmModel)
	auditWriter := audit.NewWriter(pool)
	searchHandler := handler.New(searchSvc, auditWriter)

	hc := health.New("search")
	hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}

	r := app.Router()
	r.Get("/health", hc.Handler())

	searchRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 30, Window: time.Minute, KeyFunc: sdamw.ByUser})

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Use(searchRL)
		r.Mount("/v1/search", searchHandler.Routes())
	})

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
	app.OnShutdown(grpcSrv.GracefulStop)

	app.Run()
}
