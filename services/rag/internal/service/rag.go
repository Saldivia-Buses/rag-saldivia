// Package service implements the RAG proxy business logic.
// The RAG Service is a thin proxy to the NVIDIA RAG Blueprint (:8081).
// It adds tenant context (collection namespacing), permission checks,
// and transforms the SSE stream for WebSocket delivery.
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config holds RAG service configuration.
type Config struct {
	BlueprintURL string        // http://localhost:8081 or https://openrouter.ai/api
	Timeout      time.Duration // request timeout
	APIKey       string        // optional: API key for OpenRouter/external providers
	Model        string        // optional: model ID for OpenRouter (e.g. "anthropic/claude-sonnet-4")
}

// RAG proxies requests to the NVIDIA Blueprint.
type RAG struct {
	cfg    Config
	client *http.Client
}

// NewRAG creates a RAG proxy service.
func NewRAG(cfg Config) *RAG {
	return &RAG{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// GenerateRequest holds the input for a RAG query.
type GenerateRequest struct {
	Messages         []ChatMessage  `json:"messages"`
	Model            string         `json:"model,omitempty"`
	CollectionName   string         `json:"collection_name,omitempty"`
	Stream           bool           `json:"stream"`
	Temperature      float64        `json:"temperature,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	VdbTopK          int            `json:"vdb_top_k,omitempty"`
	RerankerTopK     int            `json:"reranker_top_k,omitempty"`
	UseKnowledgeBase bool           `json:"use_knowledge_base"`
	Reasoning        *ReasoningCfg  `json:"reasoning,omitempty"`
}

// ReasoningCfg enables extended thinking for supported models.
type ReasoningCfg struct {
	Effort string `json:"effort,omitempty"` // "low", "medium", "high"
}

// ChatMessage is a single message in the conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GenerateStream sends a streaming request to the Blueprint and returns the raw SSE body.
// The caller is responsible for closing the response body.
func (r *RAG) GenerateStream(ctx context.Context, tenantSlug string, req GenerateRequest) (io.ReadCloser, string, error) {
	// Namespace collection by tenant (only for Blueprint mode, not OpenRouter)
	if r.cfg.APIKey == "" {
		collection := req.CollectionName
		if collection != "" && !strings.HasPrefix(collection, tenantSlug+"-") {
			collection = tenantSlug + "-" + collection
		}
		req.CollectionName = collection
	} else {
		// OpenRouter/external mode: inject model if configured
		if r.cfg.Model != "" && req.Model == "" {
			req.Model = r.cfg.Model
		}
		// Clear Blueprint-specific fields
		req.CollectionName = ""
		req.UseKnowledgeBase = false
		req.VdbTopK = 0
		req.RerankerTopK = 0
	}
	req.Stream = true

	body, err := marshalJSON(req)
	if err != nil {
		return nil, "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.cfg.BlueprintURL+"/v1/chat/completions", body)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// Add API key for external providers (OpenRouter, OpenAI, etc.)
	if r.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)
	}

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("blueprint request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, "", fmt.Errorf("blueprint returned %d: %s", resp.StatusCode, string(respBody))
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
}

// ListCollections returns all collections from the Blueprint, filtered by tenant prefix.
// In OpenRouter mode (no Blueprint), returns an empty list.
func (r *RAG) ListCollections(ctx context.Context, tenantSlug string) ([]string, error) {
	if r.cfg.APIKey != "" {
		return []string{}, nil // no collections in OpenRouter mode
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		r.cfg.BlueprintURL+"/v1/collections", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blueprint returned %d", resp.StatusCode)
	}

	// Parse response and filter by tenant prefix
	type collectionsResponse struct {
		Collections []struct {
			Name string `json:"name"`
		} `json:"collections"`
	}

	var cr collectionsResponse
	if err := decodeJSON(resp.Body, &cr); err != nil {
		return nil, fmt.Errorf("decode collections: %w", err)
	}

	prefix := tenantSlug + "-"
	var filtered []string
	for _, c := range cr.Collections {
		if strings.HasPrefix(c.Name, prefix) {
			// Return without prefix for the tenant
			filtered = append(filtered, strings.TrimPrefix(c.Name, prefix))
		}
	}
	if filtered == nil {
		filtered = []string{}
	}
	return filtered, nil
}

// Health checks if the upstream LLM provider is reachable.
func (r *RAG) Health(ctx context.Context) error {
	if r.cfg.APIKey != "" {
		// OpenRouter mode — no health endpoint, just verify config exists
		return nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		r.cfg.BlueprintURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("blueprint unreachable: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blueprint unhealthy: %d", resp.StatusCode)
	}
	return nil
}
