package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/build"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/crypto"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/scanner"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("BIGBROTHER_PORT", "8012")
	tenantDBURL := config.Env("POSTGRES_TENANT_URL", "")

	if tenantDBURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize OpenTelemetry
	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-bigbrother",
		ServiceVersion: "0.1.0",
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed, traces disabled", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// Token blacklist (shared Redis)
	redisURL := config.Env("REDIS_URL", "localhost:6379")
	blacklist := security.InitBlacklist(ctx, redisURL)

	// Connect to tenant database (single pool — RBAC via JWT, no platform queries)
	tenantPool, err := database.NewPool(ctx, tenantDBURL)
	if err != nil {
		slog.Error("failed to connect to tenant database", "error", err)
		os.Exit(1)
	}
	defer tenantPool.Close()
	if err := tenantPool.Ping(ctx); err != nil {
		slog.Error("failed to ping tenant database", "error", err)
		os.Exit(1)
	}

	// Connect to NATS
	natsURL := config.Env("NATS_URL", nats.DefaultURL)
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()
	slog.Info("connected to NATS", "url", config.RedactURL(natsURL))

	// Audit writer (both non-failing and fail-closed)
	auditWriter := audit.NewWriter(tenantPool)

	// Tenant slug for NATS subjects
	tenantSlug := config.Env("TENANT_SLUG", "dev")

	// Scanner — use StubScanner in dev (WSL2), ARPScanner on physical NIC
	scanMode := scanner.ScanMode(config.Env("SCAN_MODE", "passive"))
	var netScanner scanner.NetworkScanner
	lanIface := config.Env("LAN_INTERFACE", "")
	if lanIface != "" {
		arpScanner, err := scanner.NewARPScanner(lanIface, 10*time.Second)
		if err != nil {
			slog.Warn("ARP scanner init failed, using stub", "interface", lanIface, "error", err)
			netScanner = scanner.NewStubScanner()
		} else {
			netScanner = arpScanner
		}
	} else {
		slog.Info("no LAN_INTERFACE set, using stub scanner")
		netScanner = scanner.NewStubScanner()
	}

	scannerSvc := service.NewScanner(tenantPool, nc, tenantSlug)
	scanLoop := scanner.NewLoop(netScanner, scanMode, scannerSvc.ProcessResults)
	scanLoop.Start(ctx)
	defer scanLoop.Stop()

	// Envelope encryption (KEK from Docker secret)
	kekPath := config.Env("BB_KEK_PATH", "/run/secrets/bb_kek")
	encryptor, err := crypto.NewEncryptor(kekPath)
	if err != nil {
		slog.Warn("KEK not available, credential endpoints disabled", "path", kekPath, "error", err)
	}

	// Redis client for rate limiting
	rdb := redis.NewClient(&redis.Options{Addr: redisURL})

	// Health checker
	hc := health.New("bigbrother")
	hc.Add("postgres-tenant", func(ctx context.Context) error { return tenantPool.Ping(ctx) })
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	hc.Add("scanner", func(_ context.Context) error {
		if !scanLoop.IsAlive() {
			return fmt.Errorf("scanner goroutine dead")
		}
		return nil
	})
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())
	r.Use(middleware.Timeout(30 * time.Second))

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	r.Get("/health", hc.Handler())
	r.Get("/v1/info", build.Handler("sda-bigbrother"))

	// All BigBrother routes: FailOpen=false (security > availability)
	devicesHandler := handler.NewDevices(tenantPool, nc, auditWriter, tenantSlug)

	// Wire credential service if KEK is available
	if encryptor != nil {
		credSvc := service.NewCredentialService(tenantPool, encryptor, auditWriter, rdb, tenantSlug)
		devicesHandler.SetCredentialService(credSvc)
		slog.Info("credential vault enabled")
	}

	// Wire scan loop reference for scan endpoints
	devicesHandler.SetScanLoop(scanLoop)

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
			Blacklist: blacklist,
			FailOpen:  false,
		}))
		r.Mount("/v1/bigbrother", devicesHandler.Routes())
	})

	// Event retention cleanup goroutine
	retentionDays, _ := strconv.Atoi(config.Env("EVENT_RETENTION_DAYS", "90"))
	go runCleanup(ctx, tenantPool, tenantSlug, retentionDays)

	// Server
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           otelhttp.NewHandler(r, "sda-bigbrother"),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("bigbrother service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	slog.Info("bigbrother service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("bigbrother service stopped")
}

// runCleanup periodically deletes old events and expired pending writes.
func runCleanup(ctx context.Context, db *pgxpool.Pool, tenantSlug string, retentionDays int) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().AddDate(0, 0, -retentionDays)

			// Batched event deletion (1000 at a time)
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

			// Expire pending writes
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
