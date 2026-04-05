// Package service implements the Agent Runtime business logic.
// Orchestrates the LLM → tool call → response loop.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Camionerou/rag-saldivia/pkg/guardrails"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/llm"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/tools"
)

// Agent orchestrates chat queries: guardrails → LLM → tools → response.
type Agent struct {
	llmAdapter   *llm.Adapter
	toolExecutor *tools.Executor
	toolSchemas  []llm.ToolSchema
	config       Config
}

// Config holds agent runtime configuration (from agent_config).
type Config struct {
	SystemPrompt       string
	MaxToolCallsPerTurn int
	MaxLoopIterations   int
	Temperature         float64
	MaxTokens           int
	GuardrailsConfig    guardrails.InputConfig
}

// New creates an Agent service.
func New(adapter *llm.Adapter, executor *tools.Executor, schemas []llm.ToolSchema, cfg Config) *Agent {
	if cfg.MaxToolCallsPerTurn <= 0 {
		cfg.MaxToolCallsPerTurn = 25
	}
	if cfg.MaxLoopIterations <= 0 {
		cfg.MaxLoopIterations = 10
	}
	return &Agent{
		llmAdapter:   adapter,
		toolExecutor: executor,
		toolSchemas:  schemas,
		config:       cfg,
	}
}

// QueryResult is the output of an agent query.
type QueryResult struct {
	Response      string          `json:"response"`
	ToolCalls     []ToolCallLog   `json:"tool_calls"`
	InputTokens   int             `json:"input_tokens"`
	OutputTokens  int             `json:"output_tokens"`
	DurationMS    int             `json:"duration_ms"`
	Model         string          `json:"model"`
}

// ToolCallLog records a single tool call for tracing.
type ToolCallLog struct {
	Tool       string          `json:"tool"`
	Input      json.RawMessage `json:"input"`
	Output     json.RawMessage `json:"output,omitempty"`
	Status     string          `json:"status"`
	DurationMS int             `json:"duration_ms"`
}

// Query runs a user query through the agent loop.
func (a *Agent) Query(ctx context.Context, jwt, userMessage string, history []llm.Message) (*QueryResult, error) {
	start := time.Now()

	// Input guardrails
	sanitized, err := guardrails.ValidateInput(ctx, userMessage, a.config.GuardrailsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("guardrails blocked: %w", err)
	}

	// Build messages
	messages := make([]llm.Message, 0, len(history)+2)
	messages = append(messages, llm.Message{Role: "system", Content: a.config.SystemPrompt})
	messages = append(messages, history...)
	messages = append(messages, llm.Message{Role: "user", Content: sanitized})

	var allToolCalls []ToolCallLog
	var totalInput, totalOutput int
	var loopHistory []guardrails.ToolCallRecord

	// Agent loop: LLM → tool calls → LLM → ... → text response
	for i := 0; i < a.config.MaxLoopIterations; i++ {
		// Loop detection
		if loop, reason := guardrails.DetectLoop(loopHistory, guardrails.LoopConfig{
			MaxIterations:         a.config.MaxLoopIterations,
			MaxIdenticalToolCalls: 3,
		}); loop {
			slog.Warn("loop detected, breaking", "reason", reason)
			break
		}

		resp, err := a.llmAdapter.Chat(ctx, messages, a.toolSchemas, a.config.Temperature, a.config.MaxTokens)
		if err != nil {
			return nil, fmt.Errorf("llm call: %w", err)
		}

		totalInput += resp.InputTokens
		totalOutput += resp.OutputTokens

		// No tool calls — we have a final text response
		if len(resp.ToolCalls) == 0 {
			output := guardrails.ValidateOutput(resp.Content, guardrails.OutputConfig{})
			return &QueryResult{
				Response:     output,
				ToolCalls:    allToolCalls,
				InputTokens:  totalInput,
				OutputTokens: totalOutput,
				DurationMS:   int(time.Since(start).Milliseconds()),
				Model:        a.llmAdapter.Model(),
			}, nil
		}

		// Process tool calls
		messages = append(messages, llm.Message{
			Role:      "assistant",
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			if len(allToolCalls) >= a.config.MaxToolCallsPerTurn {
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    `{"error":"max tool calls exceeded"}`,
				})
				continue
			}

			tcStart := time.Now()
			result, err := a.toolExecutor.Execute(ctx, jwt, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			tcDuration := int(time.Since(tcStart).Milliseconds())

			var tcLog ToolCallLog
			tcLog.Tool = tc.Function.Name
			tcLog.Input = json.RawMessage(tc.Function.Arguments)
			tcLog.DurationMS = tcDuration

			if err != nil {
				tcLog.Status = "error"
				tcLog.Output = json.RawMessage(fmt.Sprintf(`{"error":%q}`, err.Error()))
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf(`{"error":%q}`, err.Error()),
				})
			} else {
				tcLog.Status = result.Status
				tcLog.Output = result.Data
				// Feed result back to LLM
				content := string(result.Data)
				if result.Status != "success" {
					content = fmt.Sprintf(`{"error":%q,"status":%q}`, result.Error, result.Status)
				}
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    content,
				})
			}

			allToolCalls = append(allToolCalls, tcLog)
			loopHistory = append(loopHistory, guardrails.ToolCallRecord{
				Tool:   tc.Function.Name,
				Params: tc.Function.Arguments,
			})
		}
	}

	// Max loops reached — return what we have
	return &QueryResult{
		Response:     "Lo siento, no pude completar la consulta en el tiempo permitido.",
		ToolCalls:    allToolCalls,
		InputTokens:  totalInput,
		OutputTokens: totalOutput,
		DurationMS:   int(time.Since(start).Milliseconds()),
		Model:        a.llmAdapter.Model(),
	}, nil
}
