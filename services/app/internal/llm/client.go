// Package llm provides a single OpenAI-compatible HTTP client used by the
// app binary (agent, ingest, search). Supports chat, tool calling, token
// counting, trace propagation (otelhttp), API key auth, and SimplePrompt.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	StreamChat(ctx context.Context, messages []Message, temperature float64, maxTokens int) (<-chan StreamDelta, error)
	Model() string
	Endpoint() string
}

// StreamDelta is a single token or error from a streaming LLM response.
type StreamDelta struct {
	Text         string // token text (empty on final/error)
	InputTokens  int    // set on final chunk only
	OutputTokens int    // set on final chunk only
	Done         bool   // true on final chunk
	Err          error  // non-nil on error
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
	defer func() { _ = resp.Body.Close() }()

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
	_, _ = io.Copy(io.Discard, resp.Body) // drain for connection reuse
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

// StreamChat sends a streaming chat completion request and returns a channel
// of token deltas. The channel is closed when the stream ends or errors.
// Callers must drain the channel.
func (c *Client) StreamChat(ctx context.Context, messages []Message, temperature float64, maxTokens int) (<-chan StreamDelta, error) {
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	body := map[string]any{
		"model":       c.model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      true,
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

	// Use a separate client without timeout for streaming
	streamClient := &http.Client{Transport: c.httpClient.Transport}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm stream request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("llm returned %d: %s", resp.StatusCode, string(errBody))
	}

	ch := make(chan StreamDelta, 64)
	go func() {
		defer close(ch)
		defer func() { _ = resp.Body.Close() }()
		parseSSEStream(resp.Body, ch)
	}()

	return ch, nil
}

// parseSSEStream reads an OpenAI-compatible SSE stream and sends deltas to ch.
func parseSSEStream(body io.Reader, ch chan<- StreamDelta) {
	buf := make([]byte, 4096)
	var partial string

	scanner := func(data string) {
		// Each SSE line starts with "data: "
		for _, line := range splitLines(data) {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				ch <- StreamDelta{Done: true}
				return
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
				Usage *struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
				} `json:"usage"`
			}
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				ch <- StreamDelta{Text: chunk.Choices[0].Delta.Content}
			}
			if chunk.Usage != nil {
				ch <- StreamDelta{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
				}
			}
		}
	}

	for {
		n, err := body.Read(buf)
		if n > 0 {
			partial += string(buf[:n])
			// Process complete lines
			if idx := strings.LastIndex(partial, "\n"); idx >= 0 {
				scanner(partial[:idx+1])
				partial = partial[idx+1:]
			}
		}
		if err != nil {
			if partial != "" {
				scanner(partial)
			}
			if err != io.EOF {
				ch <- StreamDelta{Err: err}
			}
			return
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	for s != "" {
		idx := strings.Index(s, "\n")
		if idx < 0 {
			lines = append(lines, s)
			break
		}
		lines = append(lines, s[:idx])
		s = s[idx+1:]
	}
	return lines
}

// Model returns the model ID.
func (c *Client) Model() string { return c.model }

// Endpoint returns the endpoint URL.
func (c *Client) Endpoint() string { return c.endpoint }
