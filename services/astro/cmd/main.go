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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/llm"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/pkg/security"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/business"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/cache"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/intelligence"
)

const serviceVersion = "0.1.0"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("ASTRO_PORT", "8011")
	ephePath := config.Env("EPHE_PATH", "/ephe")
	dbURL := config.Env("POSTGRES_TENANT_URL", "")
	llmEndpoint := config.Env("SGLANG_LLM_URL", "")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")
	llmAPIKey := config.Env("LLM_API_KEY", "")

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-astro",
		ServiceVersion: serviceVersion,
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	ephemeris.Init(ephePath)
	defer ephemeris.Close()

	var pool *pgxpool.Pool
	if dbURL != "" {
		pool, err = pgxpool.New(ctx, dbURL)
		if err != nil {
			slog.Error("database connection failed", "error", err)
			os.Exit(1)
		}
		defer pool.Close()
	}

	var llmClient llm.ChatClient
	if llmEndpoint != "" {
		llmClient = llm.NewClient(llmEndpoint, llmModel, llmAPIKey)
	}

	// Plan 12: Intelligence engine + chart cache
	intel, err := intelligence.NewEngine(logger)
	if err != nil {
		slog.Error("intelligence engine init failed", "error", err)
		os.Exit(1)
	}
	chartCache := cache.NewChartRegistry()
	bizService := business.NewService()

	astroHandler := handler.New(pool, llmClient, intel, chartCache, bizService)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())

	hc := health.New("astro")
	if pool != nil {
		hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
	}
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}
	r.Get("/health", hc.Handler())

	// Read endpoints: FailOpen true (available during Redis outage)
	authRead := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
		Blacklist: blacklist,
		FailOpen:  true,
	})
	// Write endpoints: FailOpen false (revoked tokens must be rejected)
	authWrite := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
		Blacklist: blacklist,
		FailOpen:  false,
	})
	rateMw := sdamw.RateLimit(sdamw.RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		KeyFunc:  sdamw.ByUser,
	})

	// Read-only calculation endpoints — FailOpen, astro.read permission
	r.Group(func(r chi.Router) {
		r.Use(authRead)
		r.Use(rateMw)
		r.Use(sdamw.RequirePermission("astro.read"))
		r.Use(middleware.Timeout(5 * time.Minute))

		r.Post("/v1/astro/natal", astroHandler.Natal)
		r.Post("/v1/astro/transits", astroHandler.Transits)
		r.Post("/v1/astro/solar-arc", astroHandler.SolarArc)
		r.Post("/v1/astro/directions", astroHandler.Directions)
		r.Post("/v1/astro/progressions", astroHandler.Progressions)
		r.Post("/v1/astro/returns", astroHandler.Returns)
		r.Post("/v1/astro/profections", astroHandler.Profections)
		r.Post("/v1/astro/firdaria", astroHandler.Firdaria)
		r.Post("/v1/astro/fixed-stars", astroHandler.FixedStars)
		r.Post("/v1/astro/brief", astroHandler.Brief)
		// Plan 12: new technique endpoints
		r.Post("/v1/astro/eclipses", astroHandler.Eclipses)
		r.Post("/v1/astro/zodiacal-releasing", astroHandler.ZodiacalReleasing)
		r.Post("/v1/astro/lunations", astroHandler.Lunations)
		r.Post("/v1/astro/lots", astroHandler.Lots)
		r.Post("/v1/astro/dignities", astroHandler.Dignities)
		r.Post("/v1/astro/midpoints", astroHandler.Midpoints)
		r.Post("/v1/astro/declinations", astroHandler.Declinations)
		r.Post("/v1/astro/fast-transits", astroHandler.FastTransits)
		r.Post("/v1/astro/wheel", astroHandler.Wheel)
		// Multi-chart endpoints
		r.Post("/v1/astro/synastry", astroHandler.Synastry)
		r.Post("/v1/astro/composite", astroHandler.Composite)
		// Contacts
		r.Get("/v1/astro/contacts", astroHandler.ListContacts)
		r.Get("/v1/astro/contacts/search", astroHandler.SearchContacts)
		// Sessions (read)
		r.Get("/v1/astro/sessions", astroHandler.ListSessions)
		r.Get("/v1/astro/sessions/{id}", astroHandler.GetSession)
		r.Get("/v1/astro/sessions/{id}/messages", astroHandler.GetMessages)
		// Quality (read)
		r.Get("/v1/astro/predictions", astroHandler.ListPredictions)
		r.Get("/v1/astro/predictions/stats", astroHandler.PredictionStats)
		r.Get("/v1/astro/usage", astroHandler.DailyUsage)
	})

	// Write endpoints — FailOpen false, astro.write permission
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(rateMw)
		r.Use(sdamw.RequirePermission("astro.write"))
		r.Use(middleware.Timeout(30 * time.Second))

		r.Post("/v1/astro/contacts", astroHandler.CreateContact)
		r.Put("/v1/astro/contacts/{id}", astroHandler.UpdateContact)
		r.Delete("/v1/astro/contacts/{id}", astroHandler.DeleteContact)
		// Sessions (write)
		r.Post("/v1/astro/sessions", astroHandler.CreateSession)
		r.Patch("/v1/astro/sessions/{id}", astroHandler.UpdateSession)
		r.Delete("/v1/astro/sessions/{id}", astroHandler.DeleteSession)
		// Predictions (write)
		r.Post("/v1/astro/predictions", astroHandler.CreatePrediction)
		r.Patch("/v1/astro/predictions/{id}/verify", astroHandler.VerifyPrediction)
		r.Post("/v1/astro/feedback", astroHandler.SubmitFeedback)
	})

	// SSE endpoint — no chi timeout, astro.read permission
	r.Group(func(r chi.Router) {
		r.Use(authRead)
		r.Use(sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 5, Window: time.Minute, KeyFunc: sdamw.ByUser}))
		r.Use(sdamw.RequirePermission("astro.read"))
		r.Post("/v1/astro/query", astroHandler.Query)
	})

	// Business intelligence — separate permission, lower rate limit
	r.Group(func(r chi.Router) {
		r.Use(authRead)
		r.Use(sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 10, Window: time.Minute, KeyFunc: sdamw.ByUser}))
		r.Use(sdamw.RequirePermission("astro.business"))
		r.Use(middleware.Timeout(2 * time.Minute))
		r.Get("/v1/astro/business/dashboard", astroHandler.Dashboard)
		r.Get("/v1/astro/business/cashflow", astroHandler.CashFlow)
		r.Get("/v1/astro/business/risk", astroHandler.RiskHeatmap)
		r.Get("/v1/astro/business/forecast", astroHandler.QuarterlyForecast)
		r.Get("/v1/astro/business/team", astroHandler.TeamCompatibility)
		r.Get("/v1/astro/business/hiring", astroHandler.HiringCalendar)
		r.Get("/v1/astro/business/mercury-rx", astroHandler.MercuryRx)
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           otelhttp.NewHandler(r, "astro"),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("astro service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("astro service shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
