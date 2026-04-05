package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/guardrails"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/handler"
	agentllm "github.com/Camionerou/rag-saldivia/services/agent/internal/llm"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/service"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/tools"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := env("AGENT_PORT", "8004")
	publicKey := loadPublicKey()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// LLM adapter — resolves to SGLang or cloud via config
	llmEndpoint := env("SGLANG_LLM_URL", "http://localhost:8102")
	llmModel := env("SGLANG_LLM_MODEL", "")
	llmAPIKey := env("LLM_API_KEY", "")
	adapter := agentllm.NewAdapter(llmEndpoint, llmModel, llmAPIKey)

	// Tool definitions — hardcoded for now, will come from tool_registry in Phase 9
	searchURL := env("SEARCH_SERVICE_URL", "http://localhost:8010")
	toolDefs := []tools.Definition{
		{
			Name:        "search_documents",
			Service:     "search",
			Endpoint:    searchURL + "/v1/search/query",
			Method:      http.MethodPost,
			Description: "Search through document trees to find relevant sections. Returns text with citations.",
			Parameters:  json.RawMessage(`{"type":"object","required":["query"],"properties":{"query":{"type":"string","description":"the search query"},"collection_id":{"type":"string","description":"optional collection filter"},"max_nodes":{"type":"integer","description":"max tree nodes to select"}}}`),
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

	agentSvc := service.New(adapter, executor, schemas, service.Config{
		SystemPrompt:        env("SYSTEM_PROMPT", "Sos el asistente inteligente. Responde en espanol. Usa las tools disponibles para buscar informacion. Siempre cita la fuente."),
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

	r.Group(func(r chi.Router) {
		r.Use(sdamw.Auth(publicKey))
		r.Mount("/v1/agent", agentHandler.Routes())
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // no limit for streaming responses
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
	srv.Shutdown(shutdownCtx)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadPublicKey() ed25519.PublicKey {
	pubB64 := env("JWT_PUBLIC_KEY", "")
	if pubB64 == "" {
		slog.Error("JWT_PUBLIC_KEY is required")
		os.Exit(1)
	}
	key, err := sdajwt.ParsePublicKeyEnv(pubB64)
	if err != nil {
		slog.Error("failed to parse JWT_PUBLIC_KEY", "error", err)
		os.Exit(1)
	}
	return key
}
