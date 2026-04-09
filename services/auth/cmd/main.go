package main

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/auth/internal/service"
)

func main() {
	app := server.New("sda-auth", server.WithPort("AUTH_PORT", "8001"))
	ctx := app.Context()

	tenantDBURL := config.Env("POSTGRES_TENANT_URL", "")
	platformDBURL := config.Env("POSTGRES_PLATFORM_URL", "")

	privateKey, publicKey := loadJWTKeys()

	// Connect to NATS
	natsURL := config.Env("NATS_URL", nats.DefaultURL)
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { nc.Drain() })
	publisher := natspub.New(nc)
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// Health checker — dependency checks added per mode below
	hc := health.New("auth")
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}

	jwtCfg := sdajwt.DefaultConfig(privateKey, publicKey)

	// Multi-tenant mode: platform DB available → use Resolver
	// Single-tenant mode: only tenant DB URL → legacy mode (dev)
	var authHandler *handler.Auth

	if platformDBURL != "" {
		// Multi-tenant: resolve tenant DB per request from platform DB
		platformPool, err := database.NewPool(ctx, platformDBURL)
		if err != nil {
			slog.Error("failed to connect to platform database", "error", err)
			os.Exit(1)
		}
		app.OnShutdown(platformPool.Close)

		if err := platformPool.Ping(ctx); err != nil {
			slog.Error("failed to ping platform database", "error", err)
			os.Exit(1)
		}

		resolver := tenant.NewResolver(platformPool, nil) // nil = no credential encryption (yet)
		app.OnShutdown(resolver.Close)

		authHandler = handler.NewMultiTenantAuth(resolver, jwtCfg, publisher, blacklist)
		hc.Add("postgres", func(ctx context.Context) error { return platformPool.Ping(ctx) })
		slog.Info("auth service starting in multi-tenant mode")
	} else if tenantDBURL != "" {
		// Single-tenant: direct connection (dev mode)
		pool, err := database.NewPool(ctx, tenantDBURL)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		app.OnShutdown(pool.Close)

		if err := pool.Ping(ctx); err != nil {
			slog.Error("failed to ping database", "error", err)
			os.Exit(1)
		}

		tenantID := config.Env("TENANT_ID", "dev")
		tenantSlug := config.Env("TENANT_SLUG", "dev")
		authSvc := service.NewAuth(pool, jwtCfg, tenantID, tenantSlug, publisher)
		authSvc.SetBlacklist(blacklist)
		authHandler = handler.NewAuth(authSvc)
		hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
		slog.Info("auth service starting in single-tenant mode", "tenant_slug", tenantSlug)
	} else {
		slog.Error("either POSTGRES_TENANT_URL or POSTGRES_PLATFORM_URL is required")
		os.Exit(1)
	}

	// Rate limiters for sensitive endpoints
	loginRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 5, Window: time.Minute, KeyFunc: sdamw.ByIP})
	refreshRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 10, Window: time.Minute, KeyFunc: sdamw.ByIP})
	mfaRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 5, Window: time.Minute, KeyFunc: sdamw.ByIP})

	r := app.Router()
	r.Get("/health", hc.Handler())
	r.With(loginRL).Post("/v1/auth/login", authHandler.Login)
	r.With(refreshRL).Post("/v1/auth/refresh", authHandler.Refresh)
	r.Post("/v1/auth/logout", authHandler.Logout)

	// Protected routes — require valid access token + blacklist check
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false}))
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
	r.With(mfaRL).Post("/v1/auth/mfa/verify", authHandler.VerifyMFALogin)

	app.Run()
}

// loadJWTKeys loads Ed25519 keys from env vars (base64-encoded PEM).
// Auth service needs both private (signing) and public (verification).
func loadJWTKeys() (ed25519.PrivateKey, ed25519.PublicKey) {
	privB64 := config.Env("JWT_PRIVATE_KEY", "")
	pubB64 := config.Env("JWT_PUBLIC_KEY", "")
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
