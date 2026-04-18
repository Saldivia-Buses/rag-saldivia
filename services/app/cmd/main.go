// Command app is the consolidated monolith per ADR 025.
//
// The target shape is 5 domain modules in a single Go binary:
//
//	internal/core/     auth + platform + feedback
//	internal/rag/      ingest + search + agent
//	internal/realtime/ chat + ws + notification
//	internal/ops/      bigbrother + healthwatch + traces
//	internal/erp/      erp (isolated, last to fold)
//
// Fusions land one domain at a time. This binary grows accordingly; any
// module not yet folded still runs as a standalone service under its
// existing services/<svc>/cmd/main.go until its fusion session.
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/crypto"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/services/app/internal/guardrails"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	agentllm "github.com/Camionerou/rag-saldivia/services/app/internal/llm"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/services/app/internal/outbox"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"

	authhandler "github.com/Camionerou/rag-saldivia/services/app/internal/core/auth/handler"
	authservice "github.com/Camionerou/rag-saldivia/services/app/internal/core/auth/service"

	fbhandler "github.com/Camionerou/rag-saldivia/services/app/internal/core/feedback/handler"
	fbservice "github.com/Camionerou/rag-saldivia/services/app/internal/core/feedback/service"

	plhandler "github.com/Camionerou/rag-saldivia/services/app/internal/core/platform/handler"
	plservice "github.com/Camionerou/rag-saldivia/services/app/internal/core/platform/service"

	bbhandler "github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/handler"
	bbscanner "github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/scanner"
	bbservice "github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/service"

	hwhandler "github.com/Camionerou/rag-saldivia/services/app/internal/ops/healthwatch/handler"
	hwservice "github.com/Camionerou/rag-saldivia/services/app/internal/ops/healthwatch/service"
	hwcollector "github.com/Camionerou/rag-saldivia/services/app/internal/ops/healthwatch/collector"

	trhandler "github.com/Camionerou/rag-saldivia/services/app/internal/ops/traces/handler"
	trservice "github.com/Camionerou/rag-saldivia/services/app/internal/ops/traces/service"

	agenthandler "github.com/Camionerou/rag-saldivia/services/app/internal/rag/agent/handler"
	agentservice "github.com/Camionerou/rag-saldivia/services/app/internal/rag/agent/service"
	agenttools "github.com/Camionerou/rag-saldivia/services/app/internal/rag/agent/tools"

	ingesthandler "github.com/Camionerou/rag-saldivia/services/app/internal/rag/ingest/handler"
	ingestservice "github.com/Camionerou/rag-saldivia/services/app/internal/rag/ingest/service"

	searchhandler "github.com/Camionerou/rag-saldivia/services/app/internal/rag/search/handler"
	searchservice "github.com/Camionerou/rag-saldivia/services/app/internal/rag/search/service"

	chathandler "github.com/Camionerou/rag-saldivia/services/app/internal/realtime/chat/handler"
	chatservice "github.com/Camionerou/rag-saldivia/services/app/internal/realtime/chat/service"

	notifhandler "github.com/Camionerou/rag-saldivia/services/app/internal/realtime/notification/handler"
	notifservice "github.com/Camionerou/rag-saldivia/services/app/internal/realtime/notification/service"

	wshandler "github.com/Camionerou/rag-saldivia/services/app/internal/realtime/ws/handler"
	"github.com/Camionerou/rag-saldivia/services/app/internal/realtime/ws/hub"
)

func main() {
	// Distroless has no shell/wget, so the container healthcheck runs the
	// binary itself with --healthcheck. Must happen before server.New()
	// so the probe doesn't spin up the full stack.
	server.RunHealthcheckAndExit("APP_PORT", "8020")

	// WithTimeout(0) disables the request-timeout middleware so the hijacked
	// /ws connections survive past the default 30s. Handlers that need a
	// deadline (ingest upload, agent streaming, auth login) set their own.
	app := server.New("sda-app",
		server.WithPort("APP_PORT", "8020"),
		server.WithTimeout(0))
	ctx := app.Context()

	// Shared dependencies. Any of these failing is fatal — the monolith
	// cannot partially boot. (Old per-service main.go had the same rule.)
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	privateKey := sdajwt.MustLoadPrivateKey("JWT_PRIVATE_KEY")
	jwtCfg := sdajwt.DefaultConfig(privateKey, publicKey)

	tenantID := config.Env("TENANT_ID", "dev")
	tenantSlug := config.Env("TENANT_SLUG", "dev")

	redisURL := config.Env("REDIS_URL", "localhost:6379")
	redisOpts := &redis.Options{
		Addr:     redisURL,
		Password: config.EnvOrFile("REDIS_PASSWORD"),
	}
	blacklist := security.InitBlacklist(ctx, redisOpts)
	if blacklist == nil {
		// healthwatch admin routes fail-closed on blacklist unavailability.
		slog.Error("redis is required for token revocation")
		os.Exit(1)
	}

	nc, err := natspub.Connect(config.Env("NATS_URL", nats.DefaultURL))
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { _ = nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(config.Env("NATS_URL", "")))

	tenantPool := mustConnectDB(ctx, "tenant",
		config.Env("POSTGRES_TENANT_URL", ""))
	app.OnShutdown(tenantPool.Close)

	platformPool := mustConnectDB(ctx, "platform",
		loadSecret("/run/secrets/db_platform_url",
			config.Env("POSTGRES_PLATFORM_URL", "")))
	app.OnShutdown(platformPool.Close)

	rdb := redis.NewClient(redisOpts)
	app.OnShutdown(func() { _ = rdb.Close() })

	hc := health.New("app")
	hc.Add("postgres-tenant", func(ctx context.Context) error { return tenantPool.Ping(ctx) })
	hc.Add("postgres-platform", func(ctx context.Context) error { return platformPool.Ping(ctx) })
	hc.Add("nats", func(_ context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })

	r := app.Router()
	r.Get("/health", hc.Handler())

	wireAuth(ctx, app, r, tenantPool, nc, jwtCfg, publicKey, blacklist, tenantID, tenantSlug)
	wirePlatform(app, r, platformPool, nc, publicKey, blacklist)
	wireFeedback(ctx, app, r, tenantPool, platformPool, nc, publicKey, blacklist, tenantID, tenantSlug)

	wireBigBrother(app, r, hc, tenantPool, nc, rdb, publicKey, blacklist)
	wireHealthwatch(ctx, app, r, platformPool, nc, publicKey, blacklist)
	wireTraces(ctx, app, r, platformPool, nc, publicKey, blacklist)

	ingestSvc := wireIngest(ctx, app, r, tenantPool, nc, publicKey, blacklist, tenantSlug)
	searchSvc := wireSearch(r, tenantPool, publicKey, blacklist)
	wireAgent(r, nc, tenantPool, publicKey, blacklist, searchSvc, ingestSvc)

	wireNotification(ctx, app, r, tenantPool, platformPool, nc, publicKey, blacklist)
	chatSvc := wireChat(r, tenantPool, publicKey, blacklist, tenantSlug)
	wireWS(app, r, nc, publicKey, blacklist, chatSvc)

	// WriteTimeout(0) keeps long-lived /ws connections alive — the only
	// other streaming endpoint (/v1/agent) self-limits via its own ctx.
	app.RunWithWriteTimeout(0)
}

func mustConnectDB(ctx context.Context, name, url string) *pgxpool.Pool {
	if url == "" {
		slog.Error("database URL is required", "pool", name)
		os.Exit(1)
	}
	pool, err := database.NewPool(ctx, url)
	if err != nil {
		slog.Error("failed to connect to database", "pool", name, "error", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "pool", name, "error", err)
		os.Exit(1)
	}
	return pool
}

// loadSecret reads a Docker secret file, falling back to a default value.
func loadSecret(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(data))
}

// wireBigBrother mounts /v1/bigbrother and starts the scanner loop +
// event/pending-writes retention goroutine. FailOpen=false (security > availability).
func wireBigBrother(
	app *server.App, r chi.Router, hc *health.Checker,
	pool *pgxpool.Pool, nc *nats.Conn, rdb *redis.Client,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
) {
	auditWriter := audit.NewWriter(pool)
	tenantSlug := config.Env("TENANT_SLUG", "dev")

	scanMode := bbscanner.ScanMode(config.Env("SCAN_MODE", "passive"))
	var netScanner bbscanner.NetworkScanner
	if lanIface := config.Env("LAN_INTERFACE", ""); lanIface != "" {
		arp, err := bbscanner.NewARPScanner(lanIface, 10*time.Second)
		if err != nil {
			slog.Warn("ARP scanner init failed, using stub",
				"interface", lanIface, "error", err)
			netScanner = bbscanner.NewStubScanner()
		} else {
			netScanner = arp
		}
	} else {
		slog.Info("no LAN_INTERFACE set, using stub scanner")
		netScanner = bbscanner.NewStubScanner()
	}

	scannerSvc := bbservice.NewScanner(pool, nc, tenantSlug)
	scanLoop := bbscanner.NewLoop(netScanner, scanMode, scannerSvc.ProcessResults)
	scanLoop.Start(app.Context())
	app.OnShutdown(scanLoop.Stop)
	hc.Add("bigbrother-scanner", func(_ context.Context) error {
		if !scanLoop.IsAlive() {
			return fmt.Errorf("scanner goroutine dead")
		}
		return nil
	})

	devices := bbhandler.NewDevices(pool, nc, auditWriter, tenantSlug)
	if enc, err := crypto.NewEncryptor(config.Env("BB_KEK_PATH", "/run/secrets/bb_kek")); err == nil {
		credSvc := bbservice.NewCredentialService(pool, enc, auditWriter, rdb, tenantSlug)
		devices.SetCredentialService(credSvc)
		slog.Info("credential vault enabled")
	} else {
		slog.Warn("KEK not available, credential endpoints disabled", "error", err)
	}
	devices.SetScanLoop(scanLoop)

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  false,
		}))
		r.Mount("/v1/bigbrother", devices.Routes())
	})

	retentionDays, _ := strconv.Atoi(config.Env("EVENT_RETENTION_DAYS", "90"))
	go runBigBrotherCleanup(app.Context(), pool, tenantSlug, retentionDays)
}

// runBigBrotherCleanup periodically deletes old events and expired pending writes.
func runBigBrotherCleanup(ctx context.Context, db *pgxpool.Pool, tenantSlug string, retentionDays int) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().AddDate(0, 0, -retentionDays)

			for {
				tag, err := db.Exec(ctx,
					`DELETE FROM bb_events WHERE id IN (
						SELECT id FROM bb_events
						WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
						  AND created_at < $2
						LIMIT 1000
					)`, tenantSlug, cutoff)
				if err != nil {
					slog.Error("event cleanup failed", "error", err)
					break
				}
				if tag.RowsAffected() == 0 {
					break
				}
				slog.Info("events cleaned up", "deleted", tag.RowsAffected())
				time.Sleep(100 * time.Millisecond) // prevent lock contention
			}

			tag, err := db.Exec(ctx,
				`UPDATE bb_pending_writes SET status = 'expired'
				 WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
				   AND status = 'pending' AND expires_at < now()`, tenantSlug)
			if err != nil {
				slog.Error("pending writes cleanup failed", "error", err)
			} else if tag.RowsAffected() > 0 {
				slog.Info("pending writes expired", "count", tag.RowsAffected())
			}
		}
	}
}

// wireHealthwatch mounts /v1/healthwatch and its admin DLQ routes, starts
// the DLQ consumer and retention-cleanup scheduler. Auth flows through the
// handler's own requirePlatformAdmin middleware (role + blacklist check).
func wireHealthwatch(
	ctx context.Context, app *server.App, r chi.Router,
	pool *pgxpool.Pool, nc *nats.Conn,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
) {
	prom := hwcollector.NewPrometheus(
		config.Env("PROMETHEUS_URL", "http://prometheus:9090"))
	docker := hwcollector.NewDocker(
		config.Env("DOCKER_PROXY_URL", "http://docker-socket-proxy:2375"))
	svcCol := hwcollector.NewService()

	svc := hwservice.New(pool, prom, docker, svcCol)
	svc.StartRetentionCleanup(ctx)

	dlqConsumer := hwservice.NewDLQConsumer(pool, nc)
	if err := dlqConsumer.Start(ctx); err != nil {
		slog.Error("failed to start DLQ consumer", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(dlqConsumer.Stop)

	platformSlug := config.Env("PLATFORM_TENANT_SLUG", "platform")
	hw := hwhandler.New(svc, publicKey, platformSlug, blacklist)
	dlqHandler := hwhandler.NewDLQ(pool, nc)

	r.Mount("/v1/healthwatch", hw.Routes())
	hw.AdminRoutes(r, dlqHandler)
}

// wireTraces mounts /v1/traces and starts the NATS consumer that persists
// trace events into the platform DB. FailOpen=true on auth (availability wins
// for telemetry ingest).
func wireTraces(
	ctx context.Context, app *server.App, r chi.Router,
	pool *pgxpool.Pool, nc *nats.Conn,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
) {
	svc := trservice.New(pool)
	handler := trhandler.New(svc)

	consumer := trservice.NewConsumer(nc, svc)
	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start traces consumer", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(consumer.Stop)

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  true,
		}))
		r.Mount("/v1/traces", handler.Routes())
	})
}

// wireAuth mounts /v1/auth/* and /v1/modules/enabled and starts the outbox
// drainer for the silo tenant. Single-tenant per ADR 022 — the container is
// the tenant, so one authservice/pool/drainer covers the whole binary.
func wireAuth(
	ctx context.Context, app *server.App, r chi.Router,
	tenantPool *pgxpool.Pool, nc *nats.Conn,
	jwtCfg sdajwt.Config, publicKey ed25519.PublicKey,
	blacklist *security.TokenBlacklist,
	tenantID, tenantSlug string,
) {
	authSvc := authservice.NewAuth(tenantPool, jwtCfg, tenantID, tenantSlug)
	authSvc.SetBlacklist(blacklist)

	// Transactional outbox drainer — publishes events written by Login/
	// UpdateProfile/password-reset to NATS. One drainer per tenant pool
	// (here, the silo's single tenant).
	drainer := outbox.NewDrainer(tenantPool, nc, tenantSlug)
	drainerCtx, drainerCancel := context.WithCancel(ctx)
	go drainer.Run(drainerCtx)
	app.OnShutdown(drainerCancel)

	h := authhandler.NewAuth(authSvc)
	h.SetJWTConfig(jwtCfg)

	if key := loadSecret("/run/secrets/service_account_key",
		config.Env("SERVICE_ACCOUNT_KEY", "")); key != "" {
		h.SetServiceTokenConfig(authhandler.ServiceTokenConfig{
			Key:              key,
			PlatformTenantID: config.Env("PLATFORM_TENANT_ID", "platform"),
			PlatformSlug:     config.Env("PLATFORM_TENANT_SLUG", "platform"),
		})
		slog.Info("auth service token endpoint enabled")
	}

	loginRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 5, Window: time.Minute, KeyFunc: sdamw.ByIP})
	refreshRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 10, Window: time.Minute, KeyFunc: sdamw.ByIP})
	mfaRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 5, Window: time.Minute, KeyFunc: sdamw.ByIP})
	svcTokenRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 5, Window: time.Minute, KeyFunc: sdamw.ByIP})

	r.With(loginRL).Post("/v1/auth/login", h.Login)
	r.With(refreshRL).Post("/v1/auth/refresh", h.Refresh)
	r.Post("/v1/auth/logout", h.Logout)
	r.With(svcTokenRL).Post("/v1/auth/service-token", h.ServiceToken)
	r.With(mfaRL).Post("/v1/auth/mfa/verify", h.VerifyMFALogin)

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  false,
		}))
		r.Get("/v1/auth/me", h.Me)
		r.Patch("/v1/auth/me", h.UpdateMe)
		r.With(sdamw.RequirePermission("users.read")).Get("/v1/auth/users", h.ListUsers)
		r.Get("/v1/modules/enabled", h.EnabledModules)
		r.Post("/v1/auth/mfa/setup", h.SetupMFA)
		r.Post("/v1/auth/mfa/verify-setup", h.VerifySetup)
		r.Post("/v1/auth/mfa/disable", h.DisableMFA)
	})
}

// wirePlatform mounts /v1/platform/* (tenant + deploy admin) and /v1/flags/*
// (feature flag evaluation). The handler owns its own admin middleware via
// requirePlatformAdmin, so routes mount directly under r (no outer Group).
func wirePlatform(
	app *server.App, r chi.Router,
	pool *pgxpool.Pool, nc *nats.Conn,
	publicKey ed25519.PublicKey,
	blacklist *security.TokenBlacklist,
) {
	_ = app // shutdown hooks wired by caller's pools; platform has nothing of its own
	publisher := natspub.New(nc)
	svc := plservice.New(pool, publisher)
	platformSlug := config.Env("PLATFORM_TENANT_SLUG", "platform")
	h := plhandler.NewPlatform(svc, publicKey, platformSlug, blacklist)

	r.Mount("/v1/platform", h.Routes())
	r.Mount("/v1/flags", h.FlagsRoutes())
}

// wireFeedback mounts /v1/feedback/* (tenant-scoped, FailOpen=true so
// feedback submission keeps working if the blacklist is down) and
// /v1/platform/feedback/* (admin, FailOpen=false), and starts the NATS
// consumer + hourly aggregator that produces feedback metrics and alerts.
func wireFeedback(
	ctx context.Context, app *server.App, r chi.Router,
	tenantPool, platformPool *pgxpool.Pool, nc *nats.Conn,
	publicKey ed25519.PublicKey,
	blacklist *security.TokenBlacklist,
	tenantID, tenantSlug string,
) {
	publisher := natspub.New(nc)
	feedbackSvc := fbservice.NewFeedback(tenantPool, platformPool)
	alerter := fbservice.NewAlerter(platformPool, publisher)

	consumer := fbservice.NewConsumer(nc, feedbackSvc)
	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start feedback consumer", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(consumer.Stop)

	aggInterval := 1 * time.Hour
	if config.Env("AGGREGATION_INTERVAL", "") == "1m" {
		aggInterval = 1 * time.Minute // test override
	}
	aggregator := fbservice.NewAggregator(tenantPool, platformPool, feedbackSvc, alerter, aggInterval)
	aggregator.Start(ctx, tenantID, tenantSlug)
	app.OnShutdown(aggregator.Stop)

	feedbackH := fbhandler.NewFeedback(feedbackSvc.Repo(), platformPool)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  true,
		}))
		r.Mount("/v1/feedback", feedbackH.Routes())
	})

	platformFbH := fbhandler.NewPlatformFeedback(platformPool)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  false, // admin routes: security > availability
		}))
		r.Mount("/v1/platform/feedback", platformFbH.Routes())
	})
}

// wireIngest mounts /v1/ingest (FailOpen=false — uploads are state-changing)
// and starts the NATS-driven ingest worker. The outbox drainer is shared with
// wireAuth's (one per tenant pool, single-tenant silo).
func wireIngest(
	ctx context.Context, app *server.App, r chi.Router,
	tenantPool *pgxpool.Pool, nc *nats.Conn,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
	tenantSlug string,
) *ingestservice.Ingest {
	cfg := ingestservice.Config{
		BlueprintURL: config.Env("RAG_SERVER_URL", "http://localhost:8081"),
		StagingDir:   config.Env("INGEST_STAGING_DIR", "/tmp/ingest-staging"),
		Timeout:      120 * time.Second,
	}

	svc := ingestservice.New(tenantPool, nc, cfg)

	worker := ingestservice.NewWorker(nc, tenantPool, tenantSlug, svc, cfg)
	if err := worker.Start(ctx); err != nil {
		slog.Error("failed to start ingest worker", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(worker.Stop)

	h := ingesthandler.NewIngest(svc)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  false,
		}))
		r.Mount("/v1/ingest", h.Routes())
	})
	return svc
}

// wireSearch mounts /v1/search (FailOpen=true — read-only, availability wins)
// with a per-user rate limit. No gRPC server any more — search's only
// consumer (agent) now calls SearchDocuments in-process.
func wireSearch(
	r chi.Router,
	tenantPool *pgxpool.Pool,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
) *searchservice.Search {
	llmEndpoint := config.Env("SGLANG_LLM_URL", "http://localhost:8102")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")

	svc := searchservice.New(tenantPool, llmEndpoint, llmModel)
	auditWriter := audit.NewWriter(tenantPool)
	h := searchhandler.New(svc, auditWriter)

	searchRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 30, Window: time.Minute, KeyFunc: sdamw.ByUser})
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: true}))
		r.Use(searchRL)
		r.Mount("/v1/search", h.Routes())
	})
	return svc
}

// searchBackend adapts *searchservice.Search to agenttools.SearchBackend.
// The interface returns `any` so tools/ doesn't have to import rag/search.
type searchBackend struct{ svc *searchservice.Search }

func (b searchBackend) SearchDocuments(ctx context.Context, query, collectionID string, maxNodes int) (any, error) {
	return b.svc.SearchDocuments(ctx, query, collectionID, maxNodes)
}

// ingestBackend adapts *ingestservice.Ingest to agenttools.IngestBackend.
type ingestBackend struct{ svc *ingestservice.Ingest }

func (b ingestBackend) ListJobs(ctx context.Context, userID string, limit int) (any, error) {
	return b.svc.ListJobs(ctx, userID, limit)
}

// wireAgent mounts /v1/agent with FailOpen=false + per-user rate limit.
// Core tools (search_documents, check_job_status) dispatch in-process via
// the SearchBackend / IngestBackend wired here; module tools and cross-module
// tools (notification, bigbrother, erp) still ride HTTP.
func wireAgent(
	r chi.Router, nc *nats.Conn, tenantPool *pgxpool.Pool,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
	searchSvc *searchservice.Search, ingestSvc *ingestservice.Ingest,
) {
	tracePublisher := agentservice.NewTracePublisher(nc)

	llmEndpoint := config.Env("SGLANG_LLM_URL", "http://localhost:8102")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")
	llmAPIKey := config.Env("LLM_API_KEY", "")
	adapter := agentllm.NewClient(llmEndpoint, llmModel, llmAPIKey)

	notificationURL := config.Env("NOTIFICATION_SERVICE_URL", "http://localhost:8005")
	bigbrotherURL := config.Env("BIGBROTHER_SERVICE_URL", "http://localhost:8020")
	erpURL := config.Env("ERP_SERVICE_URL", "http://localhost:8013")

	// Core tools. Endpoint is a placeholder for the inlined ones — the
	// executor dispatches to SearchBackend / IngestBackend instead.
	// Capability is mandatory (ADR 027 Phase 0 item 4); "authed" means
	// "any authed user" (search + job listing), ingest.write gates the
	// upload action tool.
	toolDefs := []agenttools.Definition{
		{Name: "search_documents", Service: "search", Type: "read",
			Capability:  agenttools.CapabilityAuthed,
			Description: "Search through document trees to find relevant sections. Returns text with citations.",
			Parameters:  json.RawMessage(`{"type":"object","required":["query"],"properties":{"query":{"type":"string","description":"the search query"},"collection_id":{"type":"string","description":"optional collection filter"},"max_nodes":{"type":"integer","description":"max tree nodes to select"}}}`)},
		{Name: "create_ingest_job", Service: "ingest", Type: "action", RequiresConfirmation: true,
			Capability:  "ingest.write",
			Description: "Upload and process a new document into the knowledge base.",
			Parameters:  json.RawMessage(`{"type":"object","required":["file_name","collection"],"properties":{"file_name":{"type":"string","description":"name of the file"},"collection":{"type":"string","description":"target collection"}}}`)},
		{Name: "check_job_status", Service: "ingest", Type: "read",
			Capability:  agenttools.CapabilityAuthed,
			Description: "List document ingestion jobs and their statuses.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"job_id":{"type":"string","description":"optional job ID filter"}}}`)},
	}

	modulesDir := config.Env("MODULES_DIR", "modules")
	serviceURLs := map[string]string{
		"notification": notificationURL,
		"bigbrother":   bigbrotherURL,
		"erp":          erpURL,
	}
	enabledModules := agenttools.ParseEnabledModules(config.Env("ENABLED_MODULES", ""))
	slog.Info("enabled module set", "modules", enabledModules)
	if moduleDefs, err := agenttools.LoadModuleTools(modulesDir, enabledModules, serviceURLs); err != nil {
		slog.Warn("failed to load module tools", "error", err)
	} else if len(moduleDefs) > 0 {
		toolDefs = append(toolDefs, moduleDefs...)
		slog.Info("loaded module tools", "count", len(moduleDefs))
	}

	executor := agenttools.NewExecutor(toolDefs)
	executor.SetSearchBackend(searchBackend{svc: searchSvc})
	executor.SetIngestBackend(ingestBackend{svc: ingestSvc})
	executor.SetAuditLogger(audit.NewWriter(tenantPool))

	schemas := make([]agentllm.ToolSchema, len(toolDefs))
	for i, d := range toolDefs {
		schemas[i] = agentllm.ToolSchema{
			Type: "function",
			Function: agentllm.ToolDefinition{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  d.Parameters,
			},
		}
	}

	svc := agentservice.New(adapter, executor, schemas, tracePublisher, agentservice.Config{
		SystemPrompt:        config.Env("SYSTEM_PROMPT", "Sos el asistente inteligente. Responde en espanol. Usa las tools disponibles para buscar informacion. Siempre cita la fuente."),
		MaxToolCallsPerTurn: 25,
		MaxLoopIterations:   10,
		Temperature:         0.2,
		MaxTokens:           8192,
		GuardrailsConfig:    guardrails.DefaultInputConfig(10000),
	})

	h := agenthandler.New(svc)
	aiRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 30, Window: time.Minute, KeyFunc: sdamw.ByUser})
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false}))
		r.Use(aiRL)
		r.Mount("/v1/agent", h.Routes())
	})
}

// wireNotification mounts /v1/notifications (FailOpen=false), starts the NATS
// consumer that persists incoming notify.* events and sends emails, and
// exposes the internal /internal/webhook/alert endpoint (token-authenticated,
// for Alertmanager). Platform DB is the already-connected shared pool, so
// this function does not open its own; the alert webhook is disabled if the
// token is absent.
func wireNotification(
	ctx context.Context, app *server.App, r chi.Router,
	tenantPool, platformPool *pgxpool.Pool, nc *nats.Conn,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
) {
	mailer := notifservice.NewSMTPMailer(
		config.Env("SMTP_HOST", "localhost"),
		config.Env("SMTP_PORT", "1025"),
		config.Env("SMTP_FROM", "noreply@sda.local"),
	)
	svc := notifservice.New(tenantPool)
	svc.SetMailer(mailer)

	consumer := notifservice.NewConsumer(nc, svc, mailer)
	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start notification consumer", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(consumer.Stop)

	h := notifhandler.NewNotification(svc)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  false,
		}))
		r.Mount("/v1/notifications", h.Routes())
	})

	// Internal alert webhook: not JWT — token-authenticated per ADR 022
	// (only reachable via the Docker internal network). Disabled if the
	// token is unset or too short.
	webhookToken := loadSecret("/run/secrets/alertmanager_webhook_token",
		config.Env("ALERTMANAGER_WEBHOOK_TOKEN", ""))
	if webhookToken == "" {
		slog.Info("alert webhook disabled (no token configured)")
		return
	}
	if len(webhookToken) < 32 {
		slog.Error("alert webhook token too short, minimum 32 bytes required")
		return
	}
	alertStore := notifservice.NewPgAlertStore(platformPool)
	alertTo := config.Env("ALERT_CRITICAL_EMAIL", "")
	webhook := notifhandler.NewAlertWebhook(webhookToken, alertStore, mailer, alertTo)
	r.Post("/internal/webhook/alert", webhook.HandleAlertWebhook)
	slog.Info("alert webhook enabled", "path", "/internal/webhook/alert")
}

// wireChat mounts /v1/chat/sessions (FailOpen=false — chat is state-changing)
// and returns the chat service so wireWS can hand it to the hub mutation
// handler. The outbox drainer is not started here — wireAuth already runs one
// per tenant pool.
func wireChat(
	r chi.Router,
	tenantPool *pgxpool.Pool,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
	tenantSlug string,
) *chatservice.Chat {
	svc := chatservice.NewChat(tenantPool, tenantSlug)

	h := chathandler.NewChat(svc)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  false,
		}))
		r.Mount("/v1/chat/sessions", h.Routes())
	})
	return svc
}

// wireWS mounts /ws (JWT validated on upgrade by the handler — no chi auth
// middleware), starts the hub event loop and the NATS→WS bridge, and wires
// the mutation handler so WS clients can invoke chat operations in-process
// (ws→chat gRPC gone as of the ADR 025 realtime fusion).
func wireWS(
	app *server.App, r chi.Router, nc *nats.Conn,
	publicKey ed25519.PublicKey, blacklist *security.TokenBlacklist,
	chatSvc *chatservice.Chat,
) {
	h := hub.New()
	go h.Run()

	bridge := hub.NewNATSBridge(h, nc)
	if err := bridge.Start(); err != nil {
		slog.Error("failed to start NATS bridge", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(bridge.Stop)

	if mutations := hub.NewMutationHandler(chatSvc); mutations != nil {
		h.Mutations = mutations
		app.OnShutdown(mutations.Close)
	}

	wsHandler := wshandler.NewWS(h, publicKey, blacklist)
	r.Get("/ws", wsHandler.Upgrade)
}
