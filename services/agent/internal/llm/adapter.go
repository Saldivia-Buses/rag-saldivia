// Package llm provides the LLM adapter layer for the Agent Runtime.
// All model calls go through this interface — SGLang local and cloud
// providers use the same OpenAI-compatible API.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Adapter calls an OpenAI-compatible chat completions endpoint.
type Adapter struct {
	endpoint   string
	model      string
	apiKey     string
	httpClient *http.Client
}

// NewAdapter creates an LLM adapter. Works with SGLang and any OpenAI-compatible API.
func NewAdapter(endpoint, model, apiKey string) *Adapter {
	return &Adapter{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Message is a single chat message.
type Message struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	Name       string          `json:"name,omitempty"`
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
	Content    string     // text response (if no tool calls)
	ToolCalls  []ToolCall // tool calls (if any)
	InputTokens  int
	OutputTokens int
}

// Chat sends a chat completion request with optional tools.
func (a *Adapter) Chat(ctx context.Context, messages []Message, tools []ToolSchema, temperature float64, maxTokens int) (*ChatResponse, error) {
	body := map[string]any{
		"model":       a.model,
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
		a.endpoint+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	resp, err := a.httpClient.Do(req)
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

// Model returns the model ID.
func (a *Adapter) Model() string { return a.model }
