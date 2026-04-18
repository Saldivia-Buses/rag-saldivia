// Package tools provides the tool execution layer for the Agent Runtime.
// Each tool is a wrapper around a service call (Search, Ingest, etc.).
//
// Inside the consolidated app binary (ADR 025), calls that used to cross
// HTTP or gRPC boundaries (search_documents, check_job_status) are served
// by in-process backends set via SetSearchBackend / SetIngestBackend.
// Tools that still target external services (notification, bigbrother, erp)
// keep going over HTTP via Definition.Endpoint.
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

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
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
	Endpoint             string          `json:"endpoint"` // full URL (unused when an in-process backend handles the tool)
	Method               string          `json:"method"`   // HTTP method
	Type                 string          `json:"type"`     // "read" or "action"
	RequiresConfirmation bool            `json:"requires_confirmation"`
	Description          string          `json:"description"`
	Parameters           json.RawMessage `json:"parameters"` // JSON schema
}

// SearchBackend serves the search_documents tool in-process. The return
// value must be JSON-marshalable (typically *search.SearchResult).
type SearchBackend interface {
	SearchDocuments(ctx context.Context, query, collectionID string, maxNodes int) (any, error)
}

// IngestBackend serves ingest-related tools in-process.
type IngestBackend interface {
	ListJobs(ctx context.Context, userID string, limit int) (any, error)
}

// Executor calls tools by name, routing to the correct backend.
// Inlined tools (search/ingest) go straight to the registered backend;
// everything else falls through to HTTP.
type Executor struct {
	tools  map[string]Definition
	client *http.Client
	search SearchBackend
	ingest IngestBackend
}

// NewExecutor creates a tool executor with the given tool definitions.
func NewExecutor(defs []Definition) *Executor {
	m := make(map[string]Definition, len(defs))
	for _, d := range defs {
		m[d.Name] = d
	}
	return &Executor{
		tools: m,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// SetSearchBackend wires the in-process search service for search_documents.
func (e *Executor) SetSearchBackend(s SearchBackend) { e.search = s }

// SetIngestBackend wires the in-process ingest service for ingest tools.
func (e *Executor) SetIngestBackend(i IngestBackend) { e.ingest = i }

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

	if def.RequiresConfirmation {
		return &Result{
			Status:               "pending_confirmation",
			RequiresConfirmation: true,
			ActionPlan:           fmt.Sprintf("Tool %q wants to: %s", def.Name, def.Description),
			Data:                 params,
		}, nil
	}

	if r, handled := e.executeInProcess(ctx, toolName, params); handled {
		return r, nil
	}
	return e.executeHTTP(ctx, jwt, def, params)
}

// ExecuteConfirmed runs a tool that was previously pending confirmation.
// The tool MUST have RequiresConfirmation=true — this prevents bypassing
// the confirmation step by calling ExecuteConfirmed directly on any tool.
func (e *Executor) ExecuteConfirmed(ctx context.Context, jwt, toolName string, params json.RawMessage) (*Result, error) {
	def, ok := e.tools[toolName]
	if !ok {
		return &Result{Status: "error", Error: fmt.Sprintf("unknown tool: %q", toolName)}, nil
	}
	if !def.RequiresConfirmation {
		return &Result{Status: "error", Error: fmt.Sprintf("tool %q does not require confirmation", toolName)}, nil
	}
	if r, handled := e.executeInProcess(ctx, toolName, params); handled {
		return r, nil
	}
	return e.executeHTTP(ctx, jwt, def, params)
}

// executeInProcess handles the subset of tools that have a registered backend.
// Returns (result, true) when the tool was dispatched in-process; (nil, false)
// otherwise so the caller can fall back to HTTP.
func (e *Executor) executeInProcess(ctx context.Context, toolName string, params json.RawMessage) (*Result, bool) {
	switch toolName {
	case "search_documents":
		if e.search == nil {
			return nil, false
		}
		var p struct {
			Query        string `json:"query"`
			CollectionID string `json:"collection_id"`
			MaxNodes     int    `json:"max_nodes"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return &Result{Status: "error", Error: "invalid search params"}, true
		}
		res, err := e.search.SearchDocuments(ctx, p.Query, p.CollectionID, p.MaxNodes)
		if err != nil {
			return &Result{Status: "error", Error: "search failed"}, true
		}
		data, err := json.Marshal(res)
		if err != nil {
			return &Result{Status: "error", Error: "search failed"}, true
		}
		return &Result{Status: "success", Data: data}, true

	case "check_job_status":
		if e.ingest == nil {
			return nil, false
		}
		userID := sdamw.UserIDFromContext(ctx)
		if userID == "" {
			return &Result{Status: "denied", Error: "missing user identity"}, true
		}
		var p struct {
			Limit int `json:"limit"`
		}
		_ = json.Unmarshal(params, &p) // params optional
		jobs, err := e.ingest.ListJobs(ctx, userID, p.Limit)
		if err != nil {
			return &Result{Status: "error", Error: "list jobs failed"}, true
		}
		data, err := json.Marshal(map[string]any{"jobs": jobs})
		if err != nil {
			return &Result{Status: "error", Error: "list jobs failed"}, true
		}
		return &Result{Status: "success", Data: data}, true
	}
	return nil, false
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
	defer func() { _ = resp.Body.Close() }()

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
