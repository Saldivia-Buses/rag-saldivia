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

	"crypto/ed25519"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := env("AUTH_PORT", "8001")
	tenantDBURL := env("POSTGRES_TENANT_URL", "")
	platformDBURL := env("POSTGRES_PLATFORM_URL", "")

	privateKey, publicKey := loadJWTKeys()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-auth",
		ServiceVersion: "1.0.0",
		Endpoint:       env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// Connect to NATS
	natsURL := env("NATS_URL", nats.DefaultURL)
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
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	publisher := natspub.New(nc)
	slog.Info("connected to NATS", "url", natsURL)

	jwtCfg := sdajwt.DefaultConfig(privateKey, publicKey)

	// Multi-tenant mode: platform DB available → use Resolver
	// Single-tenant mode: only tenant DB URL → legacy mode (dev)
	var authHandler *handler.Auth

	if platformDBURL != "" {
		// Multi-tenant: resolve tenant DB per request from platform DB
		platformPool, err := pgxpool.New(ctx, platformDBURL)
		if err != nil {
			slog.Error("failed to connect to platform database", "error", err)
			os.Exit(1)
		}
		defer platformPool.Close()

		if err := platformPool.Ping(ctx); err != nil {
			slog.Error("failed to ping platform database", "error", err)
			os.Exit(1)
		}

		resolver := tenant.NewResolver(platformPool, nil) // nil = no credential encryption (yet)
		defer resolver.Close()

		authHandler = handler.NewMultiTenantAuth(resolver, jwtCfg, publisher)
		slog.Info("auth service starting in multi-tenant mode")
	} else if tenantDBURL != "" {
		// Single-tenant: direct connection (dev mode)
		pool, err := pgxpool.New(ctx, tenantDBURL)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer pool.Close()

		if err := pool.Ping(ctx); err != nil {
			slog.Error("failed to ping database", "error", err)
			os.Exit(1)
		}

		tenantID := env("TENANT_ID", "dev")
		tenantSlug := env("TENANT_SLUG", "dev")
		authSvc := service.NewAuth(pool, jwtCfg, tenantID, tenantSlug, publisher)
		authHandler = handler.NewAuth(authSvc)
		slog.Info("auth service starting in single-tenant mode", "tenant_slug", tenantSlug)
	} else {
		slog.Error("either POSTGRES_TENANT_URL or POSTGRES_PLATFORM_URL is required")
		os.Exit(1)
	}

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", authHandler.Health)
	r.Post("/v1/auth/login", authHandler.Login)
	r.Post("/v1/auth/refresh", authHandler.Refresh)
	r.Post("/v1/auth/logout", authHandler.Logout)

	// Protected routes — require valid access token
	r.Group(func(r chi.Router) {
		r.Use(sdamw.Auth(publicKey))
		r.Get("/v1/auth/me", authHandler.Me)
		r.Patch("/v1/auth/me", authHandler.UpdateMe)
		r.With(sdamw.RequirePermission("users.read")).Get("/v1/auth/users", authHandler.ListUsers)
		r.Get("/v1/modules/enabled", authHandler.EnabledModules)
		// MFA management (requires authenticated user)
		r.Post("/v1/auth/mfa/setup", authHandler.SetupMFA)
		r.Post("/v1/auth/mfa/verify-setup", authHandler.VerifySetup)
		r.Post("/v1/auth/mfa/disable", authHandler.DisableMFA)
	})

	// MFA login verification (uses temp mfa_token, not regular access token)
	r.Post("/v1/auth/mfa/verify", authHandler.VerifyMFALogin)

	// Server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-auth"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("auth service listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("auth service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("auth service stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadJWTKeys loads Ed25519 keys from env vars (base64-encoded PEM).
// Auth service needs both private (signing) and public (verification).
func loadJWTKeys() (ed25519.PrivateKey, ed25519.PublicKey) {
	privB64 := env("JWT_PRIVATE_KEY", "")
	pubB64 := env("JWT_PUBLIC_KEY", "")
	if privB64 == "" || pubB64 == "" {
		slog.Error("JWT_PRIVATE_KEY and JWT_PUBLIC_KEY are required")
		os.Exit(1)
	}

	privateKey, err := sdajwt.ParsePrivateKeyEnv(privB64)
	if err != nil {
		slog.Error("failed to parse JWT_PRIVATE_KEY", "error", err)
		os.Exit(1)
	}

	publicKey, err := sdajwt.ParsePublicKeyEnv(pubB64)
	if err != nil {
		slog.Error("failed to parse JWT_PUBLIC_KEY", "error", err)
		os.Exit(1)
	}

	return privateKey, publicKey
}
