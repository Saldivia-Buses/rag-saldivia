package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeCompletionResponse builds a minimal OpenAI-compatible JSON response.
func fakeCompletionResponse(content string, toolCalls []ToolCall, promptTok, completionTok int) string {
	msg := struct {
		Content   string     `json:"content"`
		ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	}{Content: content, ToolCalls: toolCalls}

	resp := struct {
		Choices []struct {
			Message any `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}{
		Choices: []struct {
			Message any `json:"message"`
		}{{Message: msg}},
	}
	resp.Usage.PromptTokens = promptTok
	resp.Usage.CompletionTokens = completionTok

	b, _ := json.Marshal(resp)
	return string(b)
}

// newTestServer creates an httptest.Server that records the last request body
// and responds with the given status and body.
func newTestServer(t *testing.T, status int, respBody string) (*httptest.Server, *[]byte) {
	t.Helper()
	var lastBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		lastBody = b
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(srv.Close)

	return srv, &lastBody
}

// newTestClient creates a Client pointing at the given test server,
// bypassing otelhttp transport (plain DefaultTransport for tests).
func newTestClient(endpoint, model, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}
}

// --- Chat tests ---

func TestChat_RequestBodyFormat(t *testing.T) {
	resp := fakeCompletionResponse("ok", nil, 10, 5)
	srv, lastBody := newTestServer(t, http.StatusOK, resp)

	tools := []ToolSchema{
		{
			Type: "function",
			Function: ToolDefinition{
				Name:        "get_weather",
				Description: "Get the weather",
				Parameters:  json.RawMessage(`{"type":"object"}`),
			},
		},
	}

	c := newTestClient(srv.URL, "test-model", "sk-test")
	msgs := []Message{{Role: "user", Content: "hello"}}
	_, err := c.Chat(t.Context(), msgs, tools, 0.7, 1024)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(*lastBody, &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}

	// model
	if body["model"] != "test-model" {
		t.Errorf("model = %v, want test-model", body["model"])
	}

	// temperature
	if temp, ok := body["temperature"].(float64); !ok || temp != 0.7 {
		t.Errorf("temperature = %v, want 0.7", body["temperature"])
	}

	// max_tokens
	if mt, ok := body["max_tokens"].(float64); !ok || int(mt) != 1024 {
		t.Errorf("max_tokens = %v, want 1024", body["max_tokens"])
	}

	// messages
	rawMsgs, ok := body["messages"].([]any)
	if !ok || len(rawMsgs) != 1 {
		t.Fatalf("messages count = %d, want 1", len(rawMsgs))
	}
	msg0 := rawMsgs[0].(map[string]any)
	if msg0["role"] != "user" || msg0["content"] != "hello" {
		t.Errorf("message[0] = %v, want {role:user, content:hello}", msg0)
	}

	// tools
	rawTools, ok := body["tools"].([]any)
	if !ok || len(rawTools) != 1 {
		t.Fatalf("tools count = %d, want 1", len(rawTools))
	}
}

func TestChat_MaxTokensOmittedWhenZero(t *testing.T) {
	resp := fakeCompletionResponse("ok", nil, 1, 1)
	srv, lastBody := newTestServer(t, http.StatusOK, resp)

	c := newTestClient(srv.URL, "m", "")
	_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0.5, 0)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	var body map[string]any
	_ = json.Unmarshal(*lastBody, &body)

	if _, exists := body["max_tokens"]; exists {
		t.Error("max_tokens should be omitted when value is 0")
	}
}

func TestChat_ToolsOmittedWhenNil(t *testing.T) {
	resp := fakeCompletionResponse("ok", nil, 1, 1)
	srv, lastBody := newTestServer(t, http.StatusOK, resp)

	c := newTestClient(srv.URL, "m", "")
	_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 100)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	var body map[string]any
	_ = json.Unmarshal(*lastBody, &body)

	if _, exists := body["tools"]; exists {
		t.Error("tools should be omitted when nil")
	}
}

func TestChat_ResponseParsing(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		toolCalls    []ToolCall
		promptTok    int
		completeTok  int
		wantContent  string
		wantTools    int
		wantInput    int
		wantOutput   int
	}{
		{
			name:        "text response",
			content:     "The weather is sunny.",
			promptTok:   15,
			completeTok: 8,
			wantContent: "The weather is sunny.",
			wantTools:   0,
			wantInput:   15,
			wantOutput:  8,
		},
		{
			name: "tool call response",
			toolCalls: []ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location":"Buenos Aires"}`,
					},
				},
			},
			promptTok:   20,
			completeTok: 12,
			wantContent: "",
			wantTools:   1,
			wantInput:   20,
			wantOutput:  12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := fakeCompletionResponse(tt.content, tt.toolCalls, tt.promptTok, tt.completeTok)
			srv, _ := newTestServer(t, http.StatusOK, resp)

			c := newTestClient(srv.URL, "m", "")
			got, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "q"}}, nil, 0, 0)
			if err != nil {
				t.Fatalf("Chat: %v", err)
			}

			if got.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q", got.Content, tt.wantContent)
			}
			if len(got.ToolCalls) != tt.wantTools {
				t.Errorf("ToolCalls count = %d, want %d", len(got.ToolCalls), tt.wantTools)
			}
			if got.InputTokens != tt.wantInput {
				t.Errorf("InputTokens = %d, want %d", got.InputTokens, tt.wantInput)
			}
			if got.OutputTokens != tt.wantOutput {
				t.Errorf("OutputTokens = %d, want %d", got.OutputTokens, tt.wantOutput)
			}
		})
	}
}

func TestChat_ToolCallFields(t *testing.T) {
	tc := []ToolCall{
		{
			ID:   "call_abc",
			Type: "function",
			Function: FunctionCall{
				Name:      "search_docs",
				Arguments: `{"query":"bolt size","collection":"manuales"}`,
			},
		},
	}
	resp := fakeCompletionResponse("", tc, 5, 3)
	srv, _ := newTestServer(t, http.StatusOK, resp)

	c := newTestClient(srv.URL, "m", "")
	got, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "q"}}, nil, 0, 0)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if len(got.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(got.ToolCalls))
	}

	call := got.ToolCalls[0]
	if call.ID != "call_abc" {
		t.Errorf("ID = %q, want call_abc", call.ID)
	}
	if call.Type != "function" {
		t.Errorf("Type = %q, want function", call.Type)
	}
	if call.Function.Name != "search_docs" {
		t.Errorf("Function.Name = %q, want search_docs", call.Function.Name)
	}
	if call.Function.Arguments != `{"query":"bolt size","collection":"manuales"}` {
		t.Errorf("Function.Arguments = %q", call.Function.Arguments)
	}
}

// --- Authorization header ---

func TestChat_AuthorizationHeader(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		wantHeader string
	}{
		{
			name:       "with api key",
			apiKey:     "sk-secret-key",
			wantHeader: "Bearer sk-secret-key",
		},
		{
			name:       "empty api key omits header",
			apiKey:     "",
			wantHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAuth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth = r.Header.Get("Authorization")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(fakeCompletionResponse("ok", nil, 1, 1)))
			}))
			t.Cleanup(srv.Close)

			c := newTestClient(srv.URL, "m", tt.apiKey)
			_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)
			if err != nil {
				t.Fatalf("Chat: %v", err)
			}

			if gotAuth != tt.wantHeader {
				t.Errorf("Authorization = %q, want %q", gotAuth, tt.wantHeader)
			}
		})
	}
}

func TestChat_ContentTypeHeader(t *testing.T) {
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fakeCompletionResponse("ok", nil, 1, 1)))
	}))
	t.Cleanup(srv.Close)

	c := newTestClient(srv.URL, "m", "")
	_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
}

func TestChat_EndpointPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fakeCompletionResponse("ok", nil, 1, 1)))
	}))
	t.Cleanup(srv.Close)

	c := newTestClient(srv.URL, "m", "")
	_, _ = c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)

	if gotPath != "/v1/chat/completions" {
		t.Errorf("path = %q, want /v1/chat/completions", gotPath)
	}
}

// --- Error handling ---

func TestChat_ServerError(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantSub  string
	}{
		{
			name:    "500 internal server error",
			status:  http.StatusInternalServerError,
			body:    `{"error":"model overloaded"}`,
			wantSub: "500",
		},
		{
			name:    "429 rate limited",
			status:  http.StatusTooManyRequests,
			body:    `{"error":"rate limit exceeded"}`,
			wantSub: "429",
		},
		{
			name:    "401 unauthorized",
			status:  http.StatusUnauthorized,
			body:    `{"error":"invalid api key"}`,
			wantSub: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t, tt.status, tt.body)

			c := newTestClient(srv.URL, "m", "")
			_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)
			if err == nil {
				t.Fatal("expected error for non-200 status")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestChat_ErrorIncludesResponseBody(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusInternalServerError, `{"error":"gpu_oom"}`)

	c := newTestClient(srv.URL, "m", "")
	_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "gpu_oom") {
		t.Errorf("error should include response body, got: %q", err.Error())
	}
}

func TestChat_MalformedJSON(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{not valid json`)

	c := newTestClient(srv.URL, "m", "")
	_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)
	if err == nil {
		t.Fatal("expected error for malformed JSON response")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("error = %q, want 'decode' in message", err.Error())
	}
}

func TestChat_EmptyChoices(t *testing.T) {
	emptyResp := `{"choices":[],"usage":{"prompt_tokens":5,"completion_tokens":0}}`
	srv, _ := newTestServer(t, http.StatusOK, emptyResp)

	c := newTestClient(srv.URL, "m", "")
	_, err := c.Chat(t.Context(), []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !strings.Contains(err.Error(), "empty choices") {
		t.Errorf("error = %q, want 'empty choices'", err.Error())
	}
}

func TestChat_CancelledContext(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, fakeCompletionResponse("ok", nil, 1, 1))

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately

	c := newTestClient(srv.URL, "m", "")
	_, err := c.Chat(ctx, []Message{{Role: "user", Content: "hi"}}, nil, 0, 0)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// --- SimplePrompt tests ---

func TestSimplePrompt_SendsSingleUserMessage(t *testing.T) {
	srv, lastBody := newTestServer(t, http.StatusOK, fakeCompletionResponse("42", nil, 5, 2))

	c := newTestClient(srv.URL, "m", "")
	got, err := c.SimplePrompt(t.Context(), "what is 6*7?", 0.0)
	if err != nil {
		t.Fatalf("SimplePrompt: %v", err)
	}

	if got != "42" {
		t.Errorf("result = %q, want 42", got)
	}

	var body map[string]any
	_ = json.Unmarshal(*lastBody, &body)

	msgs := body["messages"].([]any)
	if len(msgs) != 1 {
		t.Fatalf("messages count = %d, want 1", len(msgs))
	}

	msg := msgs[0].(map[string]any)
	if msg["role"] != "user" {
		t.Errorf("role = %v, want user", msg["role"])
	}
	if msg["content"] != "what is 6*7?" {
		t.Errorf("content = %v, want 'what is 6*7?'", msg["content"])
	}

	// no tools
	if _, exists := body["tools"]; exists {
		t.Error("SimplePrompt should not send tools")
	}
}

func TestSimplePrompt_DefaultMaxTokens(t *testing.T) {
	srv, lastBody := newTestServer(t, http.StatusOK, fakeCompletionResponse("ok", nil, 1, 1))

	c := newTestClient(srv.URL, "m", "")
	_, err := c.SimplePrompt(t.Context(), "hi", 0.5)
	if err != nil {
		t.Fatalf("SimplePrompt: %v", err)
	}

	var body map[string]any
	_ = json.Unmarshal(*lastBody, &body)

	mt := int(body["max_tokens"].(float64))
	if mt != 4096 {
		t.Errorf("default max_tokens = %d, want 4096", mt)
	}
}

func TestSimplePrompt_CustomMaxTokens(t *testing.T) {
	srv, lastBody := newTestServer(t, http.StatusOK, fakeCompletionResponse("ok", nil, 1, 1))

	c := newTestClient(srv.URL, "m", "")
	_, err := c.SimplePrompt(t.Context(), "hi", 0.5, 256)
	if err != nil {
		t.Fatalf("SimplePrompt: %v", err)
	}

	var body map[string]any
	_ = json.Unmarshal(*lastBody, &body)

	mt := int(body["max_tokens"].(float64))
	if mt != 256 {
		t.Errorf("custom max_tokens = %d, want 256", mt)
	}
}

func TestSimplePrompt_ZeroMaxTokensFallsBackToDefault(t *testing.T) {
	srv, lastBody := newTestServer(t, http.StatusOK, fakeCompletionResponse("ok", nil, 1, 1))

	c := newTestClient(srv.URL, "m", "")
	_, err := c.SimplePrompt(t.Context(), "hi", 0.5, 0)
	if err != nil {
		t.Fatalf("SimplePrompt: %v", err)
	}

	var body map[string]any
	_ = json.Unmarshal(*lastBody, &body)

	mt := int(body["max_tokens"].(float64))
	if mt != 4096 {
		t.Errorf("max_tokens with 0 arg = %d, want 4096 (default fallback)", mt)
	}
}

func TestSimplePrompt_PropagatesError(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusInternalServerError, `{"error":"fail"}`)

	c := newTestClient(srv.URL, "m", "")
	_, err := c.SimplePrompt(t.Context(), "hi", 0)
	if err == nil {
		t.Fatal("expected error to propagate from Chat")
	}
}

// --- Accessor methods ---

func TestModel(t *testing.T) {
	c := NewClient("http://localhost:8080", "qwen3-8b", "key")
	if c.Model() != "qwen3-8b" {
		t.Errorf("Model() = %q, want qwen3-8b", c.Model())
	}
}

func TestEndpoint(t *testing.T) {
	c := NewClient("http://localhost:8080", "m", "")
	if c.Endpoint() != "http://localhost:8080" {
		t.Errorf("Endpoint() = %q, want http://localhost:8080", c.Endpoint())
	}
}
