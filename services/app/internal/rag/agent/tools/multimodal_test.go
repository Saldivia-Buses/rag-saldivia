package tools

// Tests for MultimodalTool — the vision model client.
//
// MultimodalTool uses an injected *http.Client, which can be pointed at
// an httptest.Server to simulate SGLang's /v1/chat/completions endpoint.
// In white-box tests (same package), we can set the httpClient field directly.
//
// Key invariants tested:
//   - Valid base64 + question → correct OpenAI-compatible request payload
//   - image_url content part uses data:image/png;base64,{data} format
//   - messages array: single user message with two content parts (image + text)
//   - LLM non-200 → Result{Status:"error"}
//   - Invalid base64 → Result{Status:"error"} (no HTTP call made)
//   - Empty choices response → Result{Status:"error"}
//   - Definition() returns the tool definition for LLM tool calling

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalBase64PNG is a 1x1 PNG image encoded in base64 — small but valid.
// Used as test input without needing to load a file.
var minimalBase64PNG = base64.StdEncoding.EncodeToString([]byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG magic
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
	0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc,
	0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82, // IEND
})

// newTestMultimodalTool creates a MultimodalTool using an httptest.Server.
// The tool's httpClient is redirected to the test server via a custom RoundTripper
// so URL rewriting is not needed.
func newTestMultimodalTool(srv *httptest.Server) *MultimodalTool {
	// Create a client that always routes to srv regardless of the requested URL.
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// Replace host with test server — keep path and method.
		req2 := req.Clone(req.Context())
		req2.URL.Scheme = "http"
		req2.URL.Host = srv.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req2)
	})
	return &MultimodalTool{
		visionEndpoint: "http://sglang-vision:8000", // overridden by transport
		visionModel:    "Qwen/Qwen3.5-9B",
		httpClient:     &http.Client{Transport: transport},
	}
}

// roundTripperFunc adapts a function to the http.RoundTripper interface.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// capturedRequest captures the body of the last request sent to the server.
type capturedRequest struct {
	Body map[string]any
}

// captureServer returns an httptest.Server that captures the request body and
// returns a successful SGLang response with the given content string.
func captureServer(t *testing.T, cap *capturedRequest, content string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &cap.Body)

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": content}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// --- Tests ---

// TestMultimodalTool_Execute_BuildsCorrectPayload verifies that AnalyzeImage
// sends an OpenAI-compatible chat completions payload with:
//   - model field matching the configured model
//   - messages array with one user message containing two content parts
//   - first part: image_url with data:image/png;base64,{data}
//   - second part: text with the question
func TestMultimodalTool_Execute_BuildsCorrectPayload(t *testing.T) {
	t.Parallel()
	cap := &capturedRequest{}
	srv := captureServer(t, cap, "This is a truck fleet vehicle.")
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	result, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "What type of vehicle is this?",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "success", result.Status, "expected success result")

	// Verify model field
	assert.Equal(t, "Qwen/Qwen3.5-9B", cap.Body["model"])

	// Verify messages structure
	messages, ok := cap.Body["messages"].([]any)
	require.True(t, ok, "messages must be an array")
	require.Len(t, messages, 1, "must have exactly one user message")

	userMsg := messages[0].(map[string]any)
	assert.Equal(t, "user", userMsg["role"])

	contentParts, ok := userMsg["content"].([]any)
	require.True(t, ok, "content must be an array of parts")
	require.Len(t, contentParts, 2, "must have image_url and text parts")

	// First part: image_url
	imagePart := contentParts[0].(map[string]any)
	assert.Equal(t, "image_url", imagePart["type"])
	imageURL := imagePart["image_url"].(map[string]any)
	assert.Equal(t, "data:image/png;base64,"+minimalBase64PNG, imageURL["url"])

	// Second part: text question
	textPart := contentParts[1].(map[string]any)
	assert.Equal(t, "text", textPart["type"])
	assert.Equal(t, "What type of vehicle is this?", textPart["text"])
}

// TestMultimodalTool_Execute_Success_ReturnsAnalysis verifies that the
// response content is extracted and returned in Result.Data as JSON.
func TestMultimodalTool_Execute_Success_ReturnsAnalysis(t *testing.T) {
	t.Parallel()
	cap := &capturedRequest{}
	srv := captureServer(t, cap, "A large truck with the Saldivia logo visible.")
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	result, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "Describe the vehicle.",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "success", result.Status)

	var data map[string]string
	require.NoError(t, json.Unmarshal(result.Data, &data))
	assert.Equal(t, "A large truck with the Saldivia logo visible.", data["analysis"])
}

// TestMultimodalTool_Execute_InvalidBase64_ReturnsError verifies that invalid
// base64 is rejected without making an HTTP call.
func TestMultimodalTool_Execute_InvalidBase64_ReturnsError(t *testing.T) {
	t.Parallel()
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.Error(w, "should not be called", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	result, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: "not-valid-base64!!!",
		Question:    "What is this?",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
	assert.False(t, called, "HTTP must not be called for invalid base64")
}

// TestMultimodalTool_Execute_LLMError_ReturnsToolError verifies that when
// the vision model returns a non-200 status, the result is an error.
func TestMultimodalTool_Execute_LLMError_ReturnsToolError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"model overloaded"}`, http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	result, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "Describe the image.",
	})

	require.NoError(t, err, "Go error must be nil — LLM errors are wrapped in Result")
	require.NotNil(t, result)
	assert.Equal(t, "error", result.Status)
	assert.Contains(t, result.Error, "503", "error must include the HTTP status code")
}

// TestMultimodalTool_Execute_EmptyChoices_ReturnsError verifies that when
// the LLM returns an empty choices array, the tool returns an error result.
func TestMultimodalTool_Execute_EmptyChoices_ReturnsError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{}})
	}))
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	result, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "Describe.",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "error", result.Status)
}

// TestMultimodalTool_Execute_ContextCancelled_ReturnsError verifies that
// a cancelled context causes the request to fail gracefully.
func TestMultimodalTool_Execute_ContextCancelled_ReturnsError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // block until cancelled
	}))
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before making the request

	result, err := tool.AnalyzeImage(ctx, AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "Describe.",
	})

	// cancelled context → http.Client.Do returns an error → wrapped in Result
	require.NoError(t, err, "Go error must be nil — transport errors are wrapped in Result")
	require.NotNil(t, result)
	assert.Equal(t, "error", result.Status)
	assert.Contains(t, result.Error, "unreachable")
}

// TestMultimodalTool_Definition verifies that Definition() returns the
// expected tool name and required parameter fields.
func TestMultimodalTool_Definition(t *testing.T) {
	t.Parallel()
	tool := NewMultimodalTool("http://sglang:8000", "Qwen/Qwen3.5-9B")
	def := tool.Definition()

	assert.Equal(t, "analyze_image", def.Name)
	assert.Equal(t, "agent", def.Service)
	assert.NotEmpty(t, def.Description)
	assert.True(t, json.Valid(def.Parameters), "Parameters must be valid JSON schema")

	// Verify required fields are in the schema
	var schema map[string]any
	require.NoError(t, json.Unmarshal(def.Parameters, &schema))
	required, ok := schema["required"].([]any)
	require.True(t, ok, "schema must have required field")
	var requiredFields []string
	for _, r := range required {
		requiredFields = append(requiredFields, r.(string))
	}
	assert.Contains(t, requiredFields, "image_base64")
	assert.Contains(t, requiredFields, "question")
}

// TestMultimodalTool_Execute_MaxTokensAndTemperature verifies that the
// payload includes max_tokens and temperature fields (model control knobs).
func TestMultimodalTool_Execute_MaxTokensAndTemperature(t *testing.T) {
	t.Parallel()
	cap := &capturedRequest{}
	srv := captureServer(t, cap, "analysis result")
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	_, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "Describe.",
	})
	require.NoError(t, err)

	// Verify control knobs are present in the payload
	maxTokens, hasMaxTokens := cap.Body["max_tokens"]
	assert.True(t, hasMaxTokens, "payload must include max_tokens")
	assert.EqualValues(t, 1024, maxTokens, "max_tokens must match the hardcoded value")

	temperature, hasTemp := cap.Body["temperature"]
	assert.True(t, hasTemp, "payload must include temperature")
	assert.EqualValues(t, 0.2, temperature, "temperature must match the hardcoded value")
}

// TestMultimodalTool_Execute_SendsToCorrectPath verifies that the HTTP
// request targets /v1/chat/completions (OpenAI-compatible path).
func TestMultimodalTool_Execute_SendsToCorrectPath(t *testing.T) {
	t.Parallel()
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "ok"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	_, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "Describe.",
	})
	require.NoError(t, err)
	assert.Equal(t, "/v1/chat/completions", capturedPath)
}

// TestMultimodalTool_Execute_SendsCorrectContentType verifies that the
// request Content-Type is application/json.
func TestMultimodalTool_Execute_SendsCorrectContentType(t *testing.T) {
	t.Parallel()
	var capturedCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"choices": []map[string]any{{"message": map[string]any{"content": "ok"}}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	_, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: minimalBase64PNG,
		Question:    "Describe.",
	})
	require.NoError(t, err)
	assert.Equal(t, "application/json", capturedCT)
}

// TestNewMultimodalTool_ConfigurationPreserved verifies that NewMultimodalTool
// stores the endpoint and model correctly.
func TestNewMultimodalTool_ConfigurationPreserved(t *testing.T) {
	t.Parallel()
	tool := NewMultimodalTool("http://sglang-vision:9000", "Qwen/Qwen-VL-7B")
	assert.Equal(t, "http://sglang-vision:9000", tool.visionEndpoint)
	assert.Equal(t, "Qwen/Qwen-VL-7B", tool.visionModel)
	assert.NotNil(t, tool.httpClient)
}

// TestMultimodalTool_AnalyzeImage_SetsCorrectBase64URL verifies that the
// image is embedded as data URI in the request (not a URL pointer).
// This is critical for SGLang vision models that require inline data.
func TestMultimodalTool_AnalyzeImage_SetsCorrectBase64URL(t *testing.T) {
	t.Parallel()
	// Use a very small but unique base64 payload to verify exact embedding.
	testImg := base64.StdEncoding.EncodeToString([]byte("fake-png-data"))

	cap := &capturedRequest{}
	srv := captureServer(t, cap, "something")
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	_, _ = tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: testImg,
		Question:    "describe",
	})

	messages := cap.Body["messages"].([]any)
	userMsg := messages[0].(map[string]any)
	parts := userMsg["content"].([]any)
	imgPart := parts[0].(map[string]any)
	imgURL := imgPart["image_url"].(map[string]any)

	expected := "data:image/png;base64," + testImg
	assert.Equal(t, expected, imgURL["url"],
		"image must be embedded as data URI — SGLang vision requires inline data, not a URL")
}

// TestMultimodalTool_Execute_LargeImageBase64_Accepted verifies that large
// base64 inputs are not rejected by the validation step.
func TestMultimodalTool_Execute_LargeImageBase64_Accepted(t *testing.T) {
	t.Parallel()
	// Create a 10KB payload — should pass base64 validation.
	raw := bytes.Repeat([]byte("A"), 10*1024)
	largeBase64 := base64.StdEncoding.EncodeToString(raw)

	cap := &capturedRequest{}
	srv := captureServer(t, cap, "large image analyzed")
	t.Cleanup(srv.Close)

	tool := newTestMultimodalTool(srv)
	result, err := tool.AnalyzeImage(context.Background(), AnalyzeImageParams{
		ImageBase64: largeBase64,
		Question:    "Analyze this image.",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "success", result.Status, "large but valid base64 must be accepted")
}
