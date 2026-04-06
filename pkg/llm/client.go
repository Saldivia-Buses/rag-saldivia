// Package llm provides a single OpenAI-compatible HTTP client for the entire
// SDA Framework. All services that need to call an LLM (SGLang, OpenAI, any
// provider) import this package instead of implementing their own client.
//
// Supports: chat, tool calling, token counting, trace propagation (otelhttp),
// API key auth, and a SimplePrompt convenience method.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Client calls an OpenAI-compatible chat completions endpoint.
// Works with SGLang (local) and any cloud provider (OpenAI, Anthropic, etc).
// Thread-safe: the underlying http.Client is safe for concurrent use.
type Client struct {
	endpoint   string
	model      string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates an LLM client with trace propagation via otelhttp.
func NewClient(endpoint, model, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout:   120 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// Message is a single chat message.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall is a tool invocation from the LLM.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall is the function name + arguments.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// ToolSchema describes a tool available to the LLM.
type ToolSchema struct {
	Type     string         `json:"type"` // "function"
	Function ToolDefinition `json:"function"`
}

// ToolDefinition is the function definition for a tool.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ChatResponse is the parsed response from the LLM.
type ChatResponse struct {
	Content      string     // text response (if no tool calls)
	ToolCalls    []ToolCall // tool calls (if any)
	InputTokens  int
	OutputTokens int
}

// ChatClient is the interface for LLM interaction. Services should depend on
// this interface, not on *Client directly, to enable mocking in tests.
type ChatClient interface {
	Chat(ctx context.Context, messages []Message, tools []ToolSchema, temperature float64, maxTokens int) (*ChatResponse, error)
	SimplePrompt(ctx context.Context, prompt string, temperature float64, maxTokens ...int) (string, error)
	Model() string
	Endpoint() string
}

// Ensure Client implements ChatClient at compile time.
var _ ChatClient = (*Client)(nil)

// Chat sends a chat completion request with optional tools.
func (c *Client) Chat(ctx context.Context, messages []Message, tools []ToolSchema, temperature float64, maxTokens int) (*ChatResponse, error) {
	body := map[string]any{
		"model":       c.model,
		"messages":    messages,
		"temperature": temperature,
	}
	if maxTokens > 0 {
		body["max_tokens"] = maxTokens
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("llm returned %d: %s", resp.StatusCode, string(errBody))
	}

	var raw struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	io.Copy(io.Discard, resp.Body) // drain for connection reuse
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("empty choices")
	}

	return &ChatResponse{
		Content:      raw.Choices[0].Message.Content,
		ToolCalls:    raw.Choices[0].Message.ToolCalls,
		InputTokens:  raw.Usage.PromptTokens,
		OutputTokens: raw.Usage.CompletionTokens,
	}, nil
}

// SimplePrompt sends a single user message and returns the text content.
// Convenience for services that don't need tool calling (ingest, search).
// maxTokens of 0 defaults to 4096.
func (c *Client) SimplePrompt(ctx context.Context, prompt string, temperature float64, maxTokens ...int) (string, error) {
	mt := 4096
	if len(maxTokens) > 0 && maxTokens[0] > 0 {
		mt = maxTokens[0]
	}
	resp, err := c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	}, nil, temperature, mt)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Model returns the model ID.
func (c *Client) Model() string { return c.model }

// Endpoint returns the endpoint URL.
func (c *Client) Endpoint() string { return c.endpoint }
