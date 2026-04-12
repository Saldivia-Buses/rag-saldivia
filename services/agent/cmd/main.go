package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/guardrails"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	agentllm "github.com/Camionerou/rag-saldivia/pkg/llm"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/service"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/tools"
)

func main() {
	app := server.New("sda-agent", server.WithPort("AGENT_PORT", "8004"))
	ctx := app.Context()

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")

	// Token blacklist (shared Redis)
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	// NATS connection for trace + feedback event publishing
	natsURL := config.Env("NATS_URL", "nats://localhost:4222")
	nc, err := natspub.Connect(natsURL)
	if err != nil {
		slog.Warn("nats connect failed, trace publishing disabled", "error", err)
	} else {
		app.OnShutdown(func() { nc.Drain() })
		slog.Info("connected to nats", "url", config.RedactURL(natsURL))
	}
	tracePublisher := service.NewTracePublisher(nc)

	// LLM adapter — resolves to SGLang or cloud via config
	llmEndpoint := config.Env("SGLANG_LLM_URL", "http://localhost:8102")
	llmModel := config.Env("SGLANG_LLM_MODEL", "")
	llmAPIKey := config.Env("LLM_API_KEY", "")
	adapter := agentllm.NewClient(llmEndpoint, llmModel, llmAPIKey)

	// Tool definitions — loaded from YAML manifests + hardcoded core tools
	searchURL := config.Env("SEARCH_SERVICE_URL", "http://localhost:8010")
	ingestURL := config.Env("INGEST_SERVICE_URL", "http://localhost:8007")
	notificationURL := config.Env("NOTIFICATION_SERVICE_URL", "http://localhost:8005")
	astroURL := config.Env("ASTRO_SERVICE_URL", "http://localhost:8011")
	bigbrotherURL := config.Env("BIGBROTHER_SERVICE_URL", "http://localhost:8012")
	erpURL := config.Env("ERP_SERVICE_URL", "http://localhost:8013")

	// Core tools always available (not module-dependent)
	toolDefs := []tools.Definition{
		{Name: "search_documents", Service: "search", Endpoint: searchURL + "/v1/search/query", Method: http.MethodPost, Type: "read",
			Description: "Search through document trees to find relevant sections. Returns text with citations.",
			Parameters:  json.RawMessage(`{"type":"object","required":["query"],"properties":{"query":{"type":"string","description":"the search query"},"collection_id":{"type":"string","description":"optional collection filter"},"max_nodes":{"type":"integer","description":"max tree nodes to select"}}}`)},
		{Name: "create_ingest_job", Service: "ingest", Endpoint: ingestURL + "/v1/ingest/upload", Method: http.MethodPost, Type: "action", RequiresConfirmation: true,
			Description: "Upload and process a new document into the knowledge base.",
			Parameters:  json.RawMessage(`{"type":"object","required":["file_name","collection"],"properties":{"file_name":{"type":"string","description":"name of the file"},"collection":{"type":"string","description":"target collection"}}}`)},
		{Name: "check_job_status", Service: "ingest", Endpoint: ingestURL + "/v1/ingest/jobs", Method: http.MethodGet, Type: "read",
			Description: "List document ingestion jobs and their statuses.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"job_id":{"type":"string","description":"optional job ID filter"}}}`)},
		// send_notification removed: notification service consumes NATS events, not HTTP POST.
		// Notifications are triggered by NATS publish from other services, not by the agent directly.
	}

	// Load module tools from YAML manifests (extends core tools)
	modulesDir := config.Env("MODULES_DIR", "modules")
	serviceURLs := map[string]string{
		"search": searchURL, "ingest": ingestURL, "notification": notificationURL,
		"astro": astroURL, "bigbrother": bigbrotherURL, "erp": erpURL,
	}
	// TODO: enabledModules should come from Platform DB per-tenant.
	// For now, load all modules' tools as available.
	moduleDefs, err := tools.LoadModuleTools(modulesDir, map[string]bool{"fleet": true, "astro": true, "bigbrother": true, "erp": true}, serviceURLs)
	if err != nil {
		slog.Warn("failed to load module tools", "error", err)
	} else if len(moduleDefs) > 0 {
		toolDefs = append(toolDefs, moduleDefs...)
		slog.Info("loaded module tools", "count", len(moduleDefs))
	}

	executor := tools.NewExecutor(toolDefs)

	// Wire gRPC for search (falls back to HTTP if unavailable)
	searchGRPC := config.Env("SEARCH_GRPC_URL", "")
	if searchGRPC != "" {
		grpcClient, err := tools.NewGRPCSearchClient(searchGRPC)
		if err != nil {
			slog.Warn("grpc search client failed, using http fallback", "error", err)
		} else {
			executor.SetGRPCSearch(grpcClient)
			app.OnShutdown(func() { grpcClient.Close() })
			slog.Info("search via grpc", "target", searchGRPC)
		}
	}

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
		GuardrailsConfig:    guardrails.DefaultInputConfig(10000),
	})

	agentHandler := handler.New(agentSvc)

	r := app.Router()

	hc := health.New("agent")
	if nc != nil {
		hc.Add("nats", func(ctx context.Context) error {
			if !nc.IsConnected() {
				return fmt.Errorf("nats disconnected")
			}
			return nil
		})
	}
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}
	r.Get("/health", hc.Handler())

	aiRL := sdamw.RateLimit(sdamw.RateLimitConfig{Requests: 30, Window: time.Minute, KeyFunc: sdamw.ByUser})

	r.Group(func(r chi.Router) {
		r.Use(sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false}))
		r.Use(aiRL)
		r.Mount("/v1/agent", agentHandler.Routes())
	})

	app.Run()
}
