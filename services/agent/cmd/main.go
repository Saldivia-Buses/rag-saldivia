package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/guardrails"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	sdaotel "github.com/Camionerou/rag-saldivia/pkg/otel"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/handler"
	agentllm "github.com/Camionerou/rag-saldivia/pkg/llm"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/service"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/tools"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := config.Env("AGENT_PORT", "8004")
	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	otelShutdown, err := sdaotel.Setup(ctx, sdaotel.Config{
		ServiceName:    "sda-agent",
		ServiceVersion: "0.1.0",
		Endpoint:       config.Env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       true,
	})
	if err != nil {
		slog.Warn("otel init failed", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// NATS connection for trace + feedback event publishing
	natsURL := config.Env("NATS_URL", "nats://localhost:4222")
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Warn("nats connect failed, trace publishing disabled", "error", err)
	} else {
		defer nc.Drain()
		slog.Info("connected to nats", "url", natsURL)
	}
	tracePublisher := service.NewTracePublisher(nc)

	// LLM adapter — resolves to SGLang or cloud via config
	llmEndpoint := config.Env("SGLANG_LLM_URL", "http://localhost:8102")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")
	llmAPIKey := config.Env("LLM_API_KEY", "")
	adapter := agentllm.NewClient(llmEndpoint, llmModel, llmAPIKey)

	// Tool definitions — hardcoded for now, will come from tool_registry in Phase 9
	searchURL := config.Env("SEARCH_SERVICE_URL", "http://localhost:8010")
	ingestURL := config.Env("INGEST_SERVICE_URL", "http://localhost:8007")
	notificationURL := config.Env("NOTIFICATION_SERVICE_URL", "http://localhost:8005")

	toolDefs := []tools.Definition{
		// Read tools
		{
			Name:        "search_documents",
			Service:     "search",
			Endpoint:    searchURL + "/v1/search/query",
			Method:      http.MethodPost,
			Type:        "read",
			Description: "Search through document trees to find relevant sections. Returns text with citations.",
			Parameters:  json.RawMessage(`{"type":"object","required":["query"],"properties":{"query":{"type":"string","description":"the search query"},"collection_id":{"type":"string","description":"optional collection filter"},"max_nodes":{"type":"integer","description":"max tree nodes to select"}}}`),
		},
		// Action tools (require confirmation)
		{
			Name:                 "create_ingest_job",
			Service:              "ingest",
			Endpoint:             ingestURL + "/v1/ingest/upload",
			Method:               http.MethodPost,
			Type:                 "action",
			RequiresConfirmation: true,
			Description:          "Upload and process a new document into the knowledge base. Requires confirmation before execution.",
			Parameters:           json.RawMessage(`{"type":"object","required":["file_name","collection"],"properties":{"file_name":{"type":"string","description":"name of the file to ingest"},"collection":{"type":"string","description":"collection to add the document to"}}}`),
		},
		{
			Name:        "check_job_status",
			Service:     "ingest",
			Endpoint:    ingestURL + "/v1/ingest/jobs",
			Method:      http.MethodGet,
			Type:        "read",
			Description: "Check the status of a document ingestion job.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"job_id":{"type":"string","description":"the job ID to check"}}}`),
		},
		{
			Name:                 "send_notification",
			Service:              "notification",
			Endpoint:             notificationURL + "/v1/notifications/send",
			Method:               http.MethodPost,
			Type:                 "action",
			RequiresConfirmation: true,
			Description:          "Send a notification to a user or group. Requires confirmation before sending.",
			Parameters:           json.RawMessage(`{"type":"object","required":["message","recipients"],"properties":{"message":{"type":"string","description":"notification message"},"recipients":{"type":"array","description":"list of user IDs to notify","items":{"type":"string"}}}}`),
		},
	}
	executor := tools.NewExecutor(toolDefs)

	// Build tool schemas for LLM
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

	agentSvc := service.New(adapter, executor, schemas, tracePublisher, service.Config{
		SystemPrompt:        config.Env("SYSTEM_PROMPT", "Sos el asistente inteligente. Responde en espanol. Usa las tools disponibles para buscar informacion. Siempre cita la fuente."),
		MaxToolCallsPerTurn: 25,
		MaxLoopIterations:   10,
		Temperature:         0.2,
		MaxTokens:           8192,
		GuardrailsConfig: guardrails.InputConfig{
			MaxLength:     10000,
			BlockPatterns: []string{"ignora tus instrucciones", "ignore your instructions"},
		},
	})

	agentHandler := handler.New(agentSvc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(90 * time.Second))
	r.Use(sdamw.SecureHeaders())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	aiRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 30, Window: time.Minute, KeyFunc: sdamw.ByUser})

	r.Group(func(r chi.Router) {
		r.Use(sdamw.Auth(publicKey))
		r.Use(aiRL)
		r.Mount("/v1/agent", agentHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      otelhttp.NewHandler(r, "sda-agent"),
		ReadTimeout:  10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 5 * time.Minute, // long for LLM streaming, but prevents indefinite slowloris
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("agent runtime starting", "port", port, "model", llmModel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("agent runtime shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}



