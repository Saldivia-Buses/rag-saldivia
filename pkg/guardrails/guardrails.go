// Package guardrails provides input/output validation for LLM interactions.
//
// Two layers of protection:
//   - Layer 1: Deterministic rules (microseconds, zero cost) — pattern matching,
//     length limits, schema validation, loop detection.
//   - Layer 2: LLM classification (100-500ms) — semantic prompt injection detection.
//     Only runs on input if Layer 1 passes. The LLM adapter is injected as a
//     dependency so this package doesn't depend on any specific provider.
package guardrails

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// InputConfig configures input validation rules.
type InputConfig struct {
	MaxLength     int      // truncate input beyond this length (0 = no limit)
	BlockPatterns []string // reject input matching any of these (case-insensitive)
}

// OutputConfig configures output validation rules.
type OutputConfig struct {
	SystemPromptFragments []string // strip these from output if leaked
}

// LoopConfig configures loop detection.
type LoopConfig struct {
	MaxIterations         int // max total LLM→tool loops
	MaxIdenticalToolCalls int // max consecutive identical tool calls
}

// ToolCallRecord represents a single tool call for loop detection.
type ToolCallRecord struct {
	Tool   string
	Params string // serialized params for comparison
}

// LLMClassifier is the interface for semantic prompt injection detection.
// The caller implements this with their LLM adapter — this package doesn't
// know what model or API is used.
type LLMClassifier interface {
	Classify(ctx context.Context, prompt string) (safe bool, reason string, err error)
}

// ValidationError is returned when input or output fails validation.
type ValidationError struct {
	Layer  int    // 1 = deterministic, 2 = LLM classifier
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("guardrails layer %d: %s", e.Layer, e.Reason)
}

// ValidateInput runs both layers of input validation.
// Layer 1: deterministic rules (patterns, length).
// Layer 2: LLM classification (prompt injection detection).
// If llm is nil, only Layer 1 runs.
// Returns the (possibly truncated) input or an error.
func ValidateInput(ctx context.Context, input string, cfg InputConfig, llm LLMClassifier) (string, error) {
	// Layer 1: truncate
	if cfg.MaxLength > 0 && len(input) > cfg.MaxLength {
		input = input[:cfg.MaxLength]
	}

	// Layer 1: pattern matching
	lower := strings.ToLower(input)
	for _, pattern := range cfg.BlockPatterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return "", &ValidationError{Layer: 1, Reason: fmt.Sprintf("blocked pattern: %q", pattern)}
		}
	}

	// Layer 2: LLM classifier (optional)
	if llm != nil {
		safe, reason, err := llm.Classify(ctx, input)
		if err != nil {
			// fail-open: if classifier is down, continue with Layer 1 only
			// (configurable via guardrails.classifier_fail_open)
			return input, nil
		}
		if !safe {
			return "", &ValidationError{Layer: 2, Reason: reason}
		}
	}

	return input, nil
}

// ValidateOutput sanitizes LLM output (Layer 1 only — no LLM call for output).
// Strips system prompt leaks and raw JSON from tool calls.
func ValidateOutput(output string, cfg OutputConfig) string {
	for _, fragment := range cfg.SystemPromptFragments {
		if fragment != "" {
			output = strings.ReplaceAll(output, fragment, "[redacted]")
		}
	}
	return output
}

// ValidateToolParams validates tool call parameters against a JSON schema.
// Returns nil if valid, error with details if invalid.
func ValidateToolParams(params json.RawMessage, schema json.RawMessage) error {
	// Basic validation: ensure params is valid JSON
	if !json.Valid(params) {
		return fmt.Errorf("invalid JSON in tool params")
	}

	// Schema validation: check required fields from schema
	var schemaMap map[string]any
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	required, _ := schemaMap["required"].([]any)
	if len(required) == 0 {
		return nil
	}

	var paramsMap map[string]any
	if err := json.Unmarshal(params, &paramsMap); err != nil {
		return fmt.Errorf("params is not a JSON object: %w", err)
	}

	for _, r := range required {
		key, ok := r.(string)
		if !ok {
			continue
		}
		if _, exists := paramsMap[key]; !exists {
			return fmt.Errorf("missing required param: %q", key)
		}
	}

	return nil
}

// DetectLoop checks if the agent is stuck in a loop.
// Returns true + reason if a loop is detected.
func DetectLoop(history []ToolCallRecord, cfg LoopConfig) (bool, string) {
	// Check max iterations
	if cfg.MaxIterations > 0 && len(history) >= cfg.MaxIterations {
		return true, fmt.Sprintf("exceeded max iterations: %d", cfg.MaxIterations)
	}

	// Check consecutive identical calls
	if cfg.MaxIdenticalToolCalls > 0 && len(history) >= cfg.MaxIdenticalToolCalls {
		last := history[len(history)-1]
		identical := 0
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].Tool == last.Tool && history[i].Params == last.Params {
				identical++
			} else {
				break
			}
		}
		if identical >= cfg.MaxIdenticalToolCalls {
			return true, fmt.Sprintf("tool %q called %d times with identical params", last.Tool, identical)
		}
	}

	return false, ""
}

// compile-time check: ensure patterns are valid (called during init in production)
func CompilePatterns(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(p))
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}
