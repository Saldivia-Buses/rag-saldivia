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
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"

	bbhandler "github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/handler"
	bbscanner "github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/scanner"
	bbservice "github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/service"

	hwhandler "github.com/Camionerou/rag-saldivia/services/app/internal/ops/healthwatch/handler"
	hwservice "github.com/Camionerou/rag-saldivia/services/app/internal/ops/healthwatch/service"
	hwcollector "github.com/Camionerou/rag-saldivia/services/app/internal/ops/healthwatch/collector"

	trhandler "github.com/Camionerou/rag-saldivia/services/app/internal/ops/traces/handler"
	trservice "github.com/Camionerou/rag-saldivia/services/app/internal/ops/traces/service"
)

func main() {
	app := server.New("sda-app", server.WithPort("APP_PORT", "8020"))
	ctx := app.Context()

	// Shared dependencies. Any of these failing is fatal — the monolith
	// cannot partially boot. (Old per-service main.go had the same rule.)
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	redisURL := config.Env("REDIS_URL", "localhost:6379")
	blacklist := security.InitBlacklist(ctx, redisURL)
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

	rdb := redis.NewClient(&redis.Options{Addr: redisURL})
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

	wireBigBrother(app, r, hc, tenantPool, nc, rdb, publicKey, blacklist)
	wireHealthwatch(ctx, app, r, platformPool, nc, publicKey, blacklist)
	wireTraces(ctx, app, r, platformPool, nc, publicKey, blacklist)

	app.Run()
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
