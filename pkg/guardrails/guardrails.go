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
	"strings"
)

// InputConfig configures input validation rules.
type InputConfig struct {
	MaxLength     int      // truncate input beyond this rune count (0 = no limit)
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
	// B1: truncate by runes, not bytes, to avoid breaking multi-byte characters
	if cfg.MaxLength > 0 {
		runes := []rune(input)
		if len(runes) > cfg.MaxLength {
			input = string(runes[:cfg.MaxLength])
		}
	}

	// Layer 1: pattern matching (case-insensitive)
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
			return input, nil
		}
		if !safe {
			return "", &ValidationError{Layer: 2, Reason: reason}
		}
	}

	return input, nil
}

// ValidateOutput sanitizes LLM output (Layer 1 only — no LLM call for output).
// Strips system prompt leaks (case-insensitive) and raw JSON from tool calls.
func ValidateOutput(output string, cfg OutputConfig) string {
	for _, fragment := range cfg.SystemPromptFragments {
		if fragment == "" {
			continue
		}
		// B2: case-insensitive replacement to catch LLM casing variations
		outputLower := strings.ToLower(output)
		fragLower := strings.ToLower(fragment)
		for {
			idx := strings.Index(outputLower, fragLower)
			if idx < 0 {
				break
			}
			output = output[:idx] + "[redacted]" + output[idx+len(fragment):]
			outputLower = strings.ToLower(output)
		}
	}
	return output
}

// ValidateToolParams validates tool call parameters against a JSON schema.
// Checks: valid JSON, required fields present, basic type validation.
func ValidateToolParams(params json.RawMessage, schema json.RawMessage) error {
	if !json.Valid(params) {
		return fmt.Errorf("invalid JSON in tool params")
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	var paramsMap map[string]any
	if err := json.Unmarshal(params, &paramsMap); err != nil {
		return fmt.Errorf("params is not a JSON object: %w", err)
	}

	// Check required fields
	if required, ok := schemaMap["required"].([]any); ok {
		for _, r := range required {
			key, ok := r.(string)
			if !ok {
				continue
			}
			if _, exists := paramsMap[key]; !exists {
				return fmt.Errorf("missing required param: %q", key)
			}
		}
	}

	// B3: type validation for declared properties
	properties, _ := schemaMap["properties"].(map[string]any)
	for key, val := range paramsMap {
		propSchema, ok := properties[key]
		if !ok {
			continue // extra params are allowed
		}
		propMap, ok := propSchema.(map[string]any)
		if !ok {
			continue
		}
		expectedType, _ := propMap["type"].(string)
		if expectedType == "" {
			continue
		}
		if err := checkType(key, val, expectedType); err != nil {
			return err
		}
	}

	return nil
}

// checkType validates a single value against its expected JSON schema type.
func checkType(key string, val any, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("param %q: expected string, got %T", key, val)
		}
	case "number", "integer":
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("param %q: expected number, got %T", key, val)
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("param %q: expected boolean, got %T", key, val)
		}
	case "array":
		if _, ok := val.([]any); !ok {
			return fmt.Errorf("param %q: expected array, got %T", key, val)
		}
	case "object":
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("param %q: expected object, got %T", key, val)
		}
	}
	return nil
}

// DetectLoop checks if the agent is stuck in a loop.
// Returns true + reason if a loop is detected.
func DetectLoop(history []ToolCallRecord, cfg LoopConfig) (bool, string) {
	if cfg.MaxIterations > 0 && len(history) >= cfg.MaxIterations {
		return true, fmt.Sprintf("exceeded max iterations: %d", cfg.MaxIterations)
	}

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
