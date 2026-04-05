// Package llm provides an OpenAI-compatible HTTP client for calling SGLang
// or any OpenAI API-compatible endpoint. Used by the tree generation pipeline.
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

// Client calls an OpenAI-compatible chat completions endpoint.
type Client struct {
	endpoint   string // e.g. "http://sglang-llm:8000"
	model      string // e.g. "google/gemma-4-26b-a4b-it"
	apiKey     string // optional, empty for local SGLang
	httpClient *http.Client
}

// NewClient creates an LLM client.
func NewClient(endpoint, model, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ChatRequest is the request body for /v1/chat/completions.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message is a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// ChatResult is the parsed response from a chat completion.
type ChatResult struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
}

// Chat sends a chat completion request and returns the response text.
func (c *Client) Chat(ctx context.Context, messages []Message, temperature float64, maxTokens int) (*ChatResult, error) {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("llm returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("llm returned empty choices")
	}

	return &ChatResult{
		Content:          chatResp.Choices[0].Message.Content,
		PromptTokens:     chatResp.Usage.PromptTokens,
		CompletionTokens: chatResp.Usage.CompletionTokens,
	}, nil
}

// SimplePrompt sends a single user message and returns the content.
func (c *Client) SimplePrompt(ctx context.Context, prompt string, temperature float64) (string, error) {
	result, err := c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	}, temperature, 4096)
	if err != nil {
		return "", err
	}
	return result.Content, nil
}
