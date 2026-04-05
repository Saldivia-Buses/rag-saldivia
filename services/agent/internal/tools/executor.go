// Package tools provides the tool execution layer for the Agent Runtime.
// Each tool is a wrapper around a service call (Search, Ingest, etc.).
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Result is the output of a tool execution.
type Result struct {
	Data                 json.RawMessage `json:"data"`
	Error                string          `json:"error,omitempty"`
	Status               string          `json:"status"` // success, error, timeout, denied, pending_confirmation
	RequiresConfirmation bool            `json:"requires_confirmation,omitempty"`
	ActionPlan           string          `json:"action_plan,omitempty"` // description for user
}

// Definition describes a registered tool.
type Definition struct {
	Name                 string          `json:"name"`
	Service              string          `json:"service"`
	Endpoint             string          `json:"endpoint"` // full URL
	Method               string          `json:"method"`   // HTTP method
	Type                 string          `json:"type"`     // "read" or "action"
	RequiresConfirmation bool            `json:"requires_confirmation"`
	Description          string          `json:"description"`
	Parameters           json.RawMessage `json:"parameters"` // JSON schema
}

// Executor calls tools by name, routing to the correct service.
type Executor struct {
	tools  map[string]Definition
	client *http.Client
}

// NewExecutor creates a tool executor with the given tool definitions.
func NewExecutor(defs []Definition) *Executor {
	m := make(map[string]Definition, len(defs))
	for _, d := range defs {
		m[d.Name] = d
	}
	return &Executor{
		tools:  m,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// Execute calls a tool by name with the given parameters and JWT.
// If the tool requires confirmation, returns a pending_confirmation result
// instead of executing. Call ExecuteConfirmed after user approves.
func (e *Executor) Execute(ctx context.Context, jwt, toolName string, params json.RawMessage) (*Result, error) {
	def, ok := e.tools[toolName]
	if !ok {
		return &Result{
			Status: "error",
			Error:  fmt.Sprintf("unknown tool: %q", toolName),
		}, nil
	}

	// Check if tool requires confirmation
	if def.RequiresConfirmation {
		return &Result{
			Status:               "pending_confirmation",
			RequiresConfirmation: true,
			ActionPlan:           fmt.Sprintf("Tool %q wants to: %s", def.Name, def.Description),
			Data:                 params,
		}, nil
	}

	return e.executeHTTP(ctx, jwt, def, params)
}

// ExecuteConfirmed runs a tool that was previously pending confirmation.
func (e *Executor) ExecuteConfirmed(ctx context.Context, jwt, toolName string, params json.RawMessage) (*Result, error) {
	def, ok := e.tools[toolName]
	if !ok {
		return &Result{Status: "error", Error: fmt.Sprintf("unknown tool: %q", toolName)}, nil
	}
	return e.executeHTTP(ctx, jwt, def, params)
}

func (e *Executor) executeHTTP(ctx context.Context, jwt string, def Definition, params json.RawMessage) (*Result, error) {
	req, err := http.NewRequestWithContext(ctx, def.Method, def.Endpoint, bytes.NewReader(params))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := e.client.Do(req)
	if err != nil {
		return &Result{Status: "timeout", Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB max
	if err != nil {
		return &Result{Status: "error", Error: "read response: " + err.Error()}, nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return &Result{Status: "denied", Error: "permission denied"}, nil
	}

	if resp.StatusCode >= 400 {
		return &Result{Status: "error", Error: fmt.Sprintf("service returned %d: %s", resp.StatusCode, string(body))}, nil
	}

	return &Result{Status: "success", Data: body}, nil
}

// GetDefinition returns a tool definition by name.
func (e *Executor) GetDefinition(name string) (Definition, bool) {
	d, ok := e.tools[name]
	return d, ok
}

// ListTools returns all registered tool names.
func (e *Executor) ListTools() []string {
	names := make([]string, 0, len(e.tools))
	for name := range e.tools {
		names = append(names, name)
	}
	return names
}
