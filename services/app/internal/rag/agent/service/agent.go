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
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/pkg/llm"
	"github.com/Camionerou/rag-saldivia/services/app/internal/rag/agent/tools"
)

// Agent orchestrates chat queries: guardrails → LLM → tools → response.
type Agent struct {
	llmAdapter     *llm.Client
	toolExecutor   *tools.Executor
	toolSchemas    []llm.ToolSchema
	tracePublisher *TracePublisher
	config         Config
}

// Config holds agent runtime configuration (from agent_config).
type Config struct {
	SystemPrompt        string
	MaxToolCallsPerTurn int
	MaxLoopIterations   int
	Temperature         float64
	MaxTokens           int
	GuardrailsConfig    guardrails.InputConfig
}

// New creates an Agent service.
func New(adapter *llm.Client, executor *tools.Executor, schemas []llm.ToolSchema, tp *TracePublisher, cfg Config) *Agent {
	if cfg.MaxToolCallsPerTurn <= 0 {
		cfg.MaxToolCallsPerTurn = 25
	}
	if cfg.MaxLoopIterations <= 0 {
		cfg.MaxLoopIterations = 10
	}
	return &Agent{
		llmAdapter:     adapter,
		tracePublisher: tp,
		toolExecutor: executor,
		toolSchemas:  schemas,
		config:       cfg,
	}
}

// QueryResult is the output of an agent query.
type QueryResult struct {
	Response             string              `json:"response"`
	ToolCalls            []ToolCallLog       `json:"tool_calls"`
	InputTokens          int                 `json:"input_tokens"`
	OutputTokens         int                 `json:"output_tokens"`
	DurationMS           int                 `json:"duration_ms"`
	Model                string              `json:"model"`
	PendingConfirmation  *PendingConfirmation `json:"pending_confirmation,omitempty"`
}

// PendingConfirmation is returned when a tool needs user approval before executing.
type PendingConfirmation struct {
	Tool       string          `json:"tool"`
	ActionPlan string          `json:"action_plan"`
	Params     json.RawMessage `json:"params"`
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
func (a *Agent) Query(ctx context.Context, jwt, userID, userMessage string, history []llm.Message) (*QueryResult, error) {
	start := time.Now()

	// B1: extract tenant from context
	ti, _ := tenant.FromContext(ctx)
	tenantSlug := ti.Slug
	if tenantSlug == "" {
		tenantSlug = ti.ID
	}
	traceID := a.tracePublisher.TraceStart(tenantSlug, "", userID, userMessage)

	// Input guardrails
	sanitized, err := guardrails.ValidateInput(ctx, userMessage, a.config.GuardrailsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("guardrails blocked: %w", err)
	}

	// B1: filter history — only allow user and assistant roles, validate content
	safeHistory := filterHistory(history, a.config.GuardrailsConfig)

	// Build messages
	messages := make([]llm.Message, 0, len(safeHistory)+2)
	messages = append(messages, llm.Message{Role: "system", Content: a.config.SystemPrompt})
	messages = append(messages, safeHistory...)
	messages = append(messages, llm.Message{Role: "user", Content: sanitized})

	var allToolCalls []ToolCallLog
	var totalInput, totalOutput int
	var loopHistory []guardrails.ToolCallRecord

	// Output guardrails config with system prompt fragments
	outputCfg := guardrails.OutputConfig{
		SystemPromptFragments: []string{a.config.SystemPrompt},
	}

	// Agent loop: LLM → tool calls → LLM → ... → text response
	for i := 0; i < a.config.MaxLoopIterations; i++ {
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

		// No tool calls — final text response
		if len(resp.ToolCalls) == 0 {
			// B5: output guardrails with system prompt leak detection
			output := guardrails.ValidateOutput(resp.Content, outputCfg)
			r := &QueryResult{
				Response:     output,
				ToolCalls:    allToolCalls,
				InputTokens:  totalInput,
				OutputTokens: totalOutput,
				DurationMS:   int(time.Since(start).Milliseconds()),
				Model:        a.llmAdapter.Model(),
			}
			return a.publishTraceEnd(tenantSlug, traceID, r, "completed"), nil
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

			// B4: validate tool params against schema
			if def, ok := a.toolExecutor.GetDefinition(tc.Function.Name); ok {
				if err := guardrails.ValidateToolParams(
					json.RawMessage(tc.Function.Arguments),
					def.Parameters,
				); err != nil {
					slog.Warn("invalid tool params", "tool", tc.Function.Name, "error", err)
					messages = append(messages, llm.Message{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf(`{"error":"invalid parameters: %s"}`, err.Error()),
					})
					allToolCalls = append(allToolCalls, ToolCallLog{
						Tool:   tc.Function.Name,
						Input:  json.RawMessage(tc.Function.Arguments),
						Status: "error",
					})
					continue
				}
			}

			tcStart := time.Now()
			result, err := a.toolExecutor.Execute(ctx, jwt, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			tcDuration := int(time.Since(tcStart).Milliseconds())

			tcLog := ToolCallLog{
				Tool:       tc.Function.Name,
				Input:      json.RawMessage(tc.Function.Arguments),
				DurationMS: tcDuration,
			}

			if err != nil {
				tcLog.Status = "error"
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    `{"error":"tool execution failed"}`,
				})
			} else if result.Status == "pending_confirmation" {
				// Tool needs user approval — pause the loop and return
				tcLog.Status = "pending_confirmation"
				allToolCalls = append(allToolCalls, tcLog)
				r := &QueryResult{
					Response:     result.ActionPlan,
					ToolCalls:    allToolCalls,
					InputTokens:  totalInput,
					OutputTokens: totalOutput,
					DurationMS:   int(time.Since(start).Milliseconds()),
					Model:        a.llmAdapter.Model(),
					PendingConfirmation: &PendingConfirmation{
						Tool:       tc.Function.Name,
						ActionPlan: result.ActionPlan,
						Params:     json.RawMessage(tc.Function.Arguments),
					},
				}
				// B3: publish trace end even for pending_confirmation
				return a.publishTraceEnd(tenantSlug, traceID, r, "pending_confirmation"), nil
			} else {
				tcLog.Status = result.Status
				if result.Status == "success" {
					tcLog.Output = result.Data
					messages = append(messages, llm.Message{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    string(result.Data),
					})
				} else {
					messages = append(messages, llm.Message{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf(`{"error":"tool %s: %s"}`, result.Status, sanitizeError(result.Error)),
					})
				}
			}

			allToolCalls = append(allToolCalls, tcLog)
			loopHistory = append(loopHistory, guardrails.ToolCallRecord{
				Tool:   tc.Function.Name,
				Params: tc.Function.Arguments,
			})
		}
	}

	r := &QueryResult{
		Response:     "Lo siento, no pude completar la consulta en el tiempo permitido.",
		ToolCalls:    allToolCalls,
		InputTokens:  totalInput,
		OutputTokens: totalOutput,
		DurationMS:   int(time.Since(start).Milliseconds()),
		Model:        a.llmAdapter.Model(),
	}
	return a.publishTraceEnd(tenantSlug, traceID, r, "timeout"), nil
}

// ExecuteConfirmed runs a previously-pending tool after user approval.
func (a *Agent) ExecuteConfirmed(ctx context.Context, jwt, toolName string, params json.RawMessage) (*tools.Result, error) {
	return a.toolExecutor.ExecuteConfirmed(ctx, jwt, toolName, params)
}

// publishTraceEnd publishes trace end + feedback events, returns result.
func (a *Agent) publishTraceEnd(tenantSlug, traceID string, result *QueryResult, status string) *QueryResult {
	a.tracePublisher.TraceEnd(
		tenantSlug, traceID, status,
		[]string{result.Model}, result.DurationMS,
		result.InputTokens, result.OutputTokens,
		len(result.ToolCalls), 0,
	)
	// Publish feedback events for the feedback service
	a.tracePublisher.PublishFeedback(tenantSlug, "usage", map[string]any{
		"trace_id":      traceID,
		"model":         result.Model,
		"input_tokens":  result.InputTokens,
		"output_tokens": result.OutputTokens,
		"duration_ms":   result.DurationMS,
		"tool_calls":    len(result.ToolCalls),
		"status":        status,
	})
	if status != "completed" {
		a.tracePublisher.PublishFeedback(tenantSlug, "error_report", map[string]any{
			"trace_id": traceID,
			"status":   status,
		})
	}
	return result
}

// filterHistory allows only user and assistant messages, validates content.
func filterHistory(history []llm.Message, cfg guardrails.InputConfig) []llm.Message {
	safe := make([]llm.Message, 0, len(history))
	for _, m := range history {
		if m.Role != "user" && m.Role != "assistant" {
			continue // B1: reject system, tool, and other roles
		}
		// Truncate content with same guardrails config
		if cfg.MaxLength > 0 {
			runes := []rune(m.Content)
			if len(runes) > cfg.MaxLength {
				m.Content = string(runes[:cfg.MaxLength])
			}
		}
		safe = append(safe, llm.Message{Role: m.Role, Content: m.Content})
	}
	return safe
}

// sanitizeError removes internal details from error messages.
func sanitizeError(err string) string {
	if len(err) > 200 {
		err = err[:200]
	}
	return err
}
