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
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/llm"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/pkg/security"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/handler"
)

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
		ServiceVersion: "0.1.0",
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

	astroHandler := handler.New(pool, llmClient)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sdamw.SecureHeaders())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"astro"}`))
	})

	authMw := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
		Blacklist: blacklist,
		FailOpen:  true,
	})
	rateMw := sdamw.RateLimit(sdamw.RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		KeyFunc:  sdamw.ByUser,
	})

	// JSON endpoints — with request timeout
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(rateMw)
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

		r.Get("/v1/astro/contacts", astroHandler.ListContacts)
		r.Post("/v1/astro/contacts", astroHandler.CreateContact)
	})

	// SSE endpoint — no chi timeout (WriteTimeout on http.Server is the safety net)
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(rateMw)
		r.Post("/v1/astro/query", astroHandler.Query)
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
