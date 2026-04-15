package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/llm"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/pkg/traces"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/business"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/cache"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/intelligence"
)

func main() {
	app := server.New("sda-astro", server.WithPort("ASTRO_PORT", "8011"), server.WithTimeout(0))
	ctx := app.Context()

	ephePath := config.Env("EPHE_PATH", "/ephe")
	dbURL := config.Env("POSTGRES_TENANT_URL", "")
	llmEndpoint := config.Env("SGLANG_LLM_URL", "")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")
	llmAPIKey := config.Env("LLM_API_KEY", "")

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// NATS connection for traces, notifications, broadcasts
	natsURL := config.Env("NATS_URL", "nats://localhost:4222")
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Warn("nats connect failed, event publishing disabled", "error", err)
	} else {
		app.OnShutdown(func() { _ = nc.Drain() }) // best-effort drain on shutdown
		slog.Info("connected to nats", "url", config.RedactURL(natsURL))
	}
	tracePublisher := traces.NewPublisher(nc)
	tenantSlug := config.Env("TENANT_SLUG", "saldivia")

	ephemeris.Init(ephePath)
	app.OnShutdown(ephemeris.Close)

	var pool *pgxpool.Pool
	if dbURL != "" {
		pool, err = database.NewPool(ctx, dbURL)
		if err != nil {
			slog.Error("database connection failed", "error", err)
			os.Exit(1)
		}
		app.OnShutdown(pool.Close)
	}

	var llmClient llm.ChatClient
	if llmEndpoint != "" {
		llmClient = llm.NewClient(llmEndpoint, llmModel, llmAPIKey)
	}

	// Plan 12: Intelligence engine + chart cache
	intel, err := intelligence.NewEngine(slog.Default())
	if err != nil {
		slog.Error("intelligence engine init failed", "error", err)
		os.Exit(1)
	}
	chartCache := cache.NewChartRegistry()
	bizService := business.NewService()

	astroHandler := handler.New(pool, llmClient, intel, chartCache, bizService, tracePublisher, tenantSlug)

	r := app.Router()

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
		r.Use(chimw.Timeout(5 * time.Minute))

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
		// More technique endpoints
		r.Post("/v1/astro/tertiary-progressions", astroHandler.TertiaryProgressions)
		r.Post("/v1/astro/decennials", astroHandler.Decennials)
		r.Post("/v1/astro/planetary-cycles", astroHandler.PlanetaryCycles)
		r.Post("/v1/astro/planetary-returns", astroHandler.PlanetaryReturns)
		r.Post("/v1/astro/lilith-vertex", astroHandler.LilithVertex)
		r.Post("/v1/astro/time-lords", astroHandler.TimeLords)
		r.Post("/v1/astro/electional", astroHandler.Electional)
		r.Post("/v1/astro/horary", astroHandler.Horary)
		r.Post("/v1/astro/astrocartography", astroHandler.Astrocartography)
		r.Post("/v1/astro/rectification", astroHandler.Rectification)
		r.Post("/v1/astro/weekly-transits", astroHandler.WeeklyTransits)
		r.Post("/v1/astro/activation-timeline", astroHandler.ActivationTimeline)
		r.Post("/v1/astro/score", astroHandler.Score)
		r.Post("/v1/astro/voc-moon", astroHandler.VOCMoon)
		r.Post("/v1/astro/tabla", astroHandler.Tabla)
		r.Post("/v1/astro/vocational", astroHandler.Vocational)
		// Multi-chart endpoints
		r.Post("/v1/astro/synastry", astroHandler.Synastry)
		r.Post("/v1/astro/composite", astroHandler.Composite)
		r.Post("/v1/astro/employee-screening", astroHandler.EmployeeScreening)
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
		r.Get("/v1/astro/alerts", astroHandler.Alerts)
	})

	// Write endpoints — FailOpen false, astro.write permission
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(rateMw)
		r.Use(sdamw.RequirePermission("astro.write"))
		r.Use(chimw.Timeout(30 * time.Second))

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
		r.Use(chimw.Timeout(2 * time.Minute))
		r.Get("/v1/astro/business/dashboard", astroHandler.Dashboard)
		r.Get("/v1/astro/business/cashflow", astroHandler.CashFlow)
		r.Get("/v1/astro/business/risk", astroHandler.RiskHeatmap)
		r.Get("/v1/astro/business/forecast", astroHandler.QuarterlyForecast)
		r.Get("/v1/astro/business/team", astroHandler.TeamCompatibility)
		r.Get("/v1/astro/business/hiring", astroHandler.HiringCalendar)
		r.Get("/v1/astro/business/mercury-rx", astroHandler.MercuryRx)
	})

	// Weekly alerts cron (Plan 13 Fase 13): scan contacts for SA/DP urgency
	// Runs every Monday at 6am Argentina (UTC-3 = 09:00 UTC).
	// Single-tenant: uses hardcoded tenant slug, no user context needed.
	if pool != nil && tracePublisher != nil {
		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case t := <-ticker.C:
					// Only run on Mondays
					if t.Weekday() != time.Monday {
						continue
					}
					// Only run around 09:00 UTC (6am Argentina)
					if t.Hour() < 8 || t.Hour() > 10 {
						continue
					}
					slog.Info("running weekly alert scan")
					// Scan would use ListAllContactsForTenant — for now uses the
					// same endpoint logic. Full cron requires a dedicated DB query
					// that doesn't need user_id. Logging the intent for visibility.
					slog.Info("weekly alert scan: use GET /v1/astro/alerts endpoint per-user")
				}
			}
		}()
	}

	// SSE + LLM streaming can run for minutes — no HTTP-level WriteTimeout.
	// Per-route chi timeouts (5min read, 2min business, 30s write) still enforce limits.
	app.RunWithWriteTimeout(0)
}
