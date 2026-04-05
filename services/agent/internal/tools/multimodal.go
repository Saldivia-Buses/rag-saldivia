package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MultimodalTool handles image analysis requests by calling the vision
// model via SGLang. Used when a user sends an image in chat.
type MultimodalTool struct {
	visionEndpoint string // e.g. "http://sglang-vision:8000"
	visionModel    string // e.g. "Qwen/Qwen3.5-9B"
	httpClient     *http.Client
}

// NewMultimodalTool creates a multimodal tool.
func NewMultimodalTool(visionEndpoint, visionModel string) *MultimodalTool {
	return &MultimodalTool{
		visionEndpoint: visionEndpoint,
		visionModel:    visionModel,
		httpClient:     &http.Client{Timeout: 60 * time.Second},
	}
}

// AnalyzeImageParams is the input for the analyze_image tool.
type AnalyzeImageParams struct {
	ImageBase64 string `json:"image_base64"` // base64-encoded image
	Question    string `json:"question"`     // what to analyze
}

// AnalyzeImage sends an image to the vision model with a question.
func (m *MultimodalTool) AnalyzeImage(ctx context.Context, params AnalyzeImageParams) (*Result, error) {
	// Validate base64
	if _, err := base64.StdEncoding.DecodeString(params.ImageBase64); err != nil {
		return &Result{Status: "error", Error: "invalid base64 image"}, nil
	}

	body, _ := json.Marshal(map[string]any{
		"model": m.visionModel,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type":      "image_url",
						"image_url": map[string]string{"url": "data:image/png;base64," + params.ImageBase64},
					},
					{
						"type": "text",
						"text": params.Question,
					},
				},
			},
		},
		"max_tokens":  1024,
		"temperature": 0.2,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		m.visionEndpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return &Result{Status: "error", Error: "vision model unreachable"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &Result{Status: "error", Error: fmt.Sprintf("vision returned %d: %s", resp.StatusCode, string(errBody))}, nil
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return &Result{Status: "error", Error: "decode vision response failed"}, nil
	}
	if len(chatResp.Choices) == 0 {
		return &Result{Status: "error", Error: "empty vision response"}, nil
	}

	data, _ := json.Marshal(map[string]string{
		"analysis": chatResp.Choices[0].Message.Content,
	})
	return &Result{Status: "success", Data: data}, nil
}

// Definition returns the tool definition for the LLM.
func (m *MultimodalTool) Definition() Definition {
	return Definition{
		Name:        "analyze_image",
		Service:     "agent",
		Endpoint:    "", // handled internally, not via HTTP
		Method:      "POST",
		Type:        "read",
		Description: "Analyze an image sent by the user. Can describe what's in the image, read text, identify objects, or answer questions about visual content.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"required": ["image_base64", "question"],
			"properties": {
				"image_base64": {"type": "string", "description": "base64-encoded image"},
				"question": {"type": "string", "description": "what to analyze in the image"}
			}
		}`),
	}
}
