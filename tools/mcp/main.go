// Package main implements the SDA Framework MCP Server.
// Exposes the system as tools for Claude and other AI agents.
//
// Tools:
//   - tenant_list — list all tenants
//   - tenant_status — status of a specific tenant
//   - service_health — health check all services
//   - service_logs — recent logs from a service
//   - db_query — read-only SQL query against a tenant DB
//   - deploy — deploy a service (with confirmation)
//   - rag_query — query a RAG collection
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// MCP JSON-RPC types

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type toolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Available tools
var tools = []toolDef{
	{
		Name:        "service_health",
		Description: "Check health status of all SDA services. Returns each service name, port, and status (UP/DOWN).",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name:        "tenant_list",
		Description: "List all tenants in the platform. Returns slug, name, enabled status.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name:        "service_logs",
		Description: "Get recent logs from a specific service container.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service": map[string]interface{}{
					"type":        "string",
					"description": "Service name (auth, ws, chat, rag, notification, platform, ingest)",
				},
				"lines": map[string]interface{}{
					"type":        "integer",
					"description": "Number of log lines to return (default 50)",
				},
			},
			"required": []string{"service"},
		},
	},
	{
		Name:        "db_query",
		Description: "Execute a read-only SQL query against a tenant database. Only SELECT queries are allowed.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant": map[string]interface{}{
					"type":        "string",
					"description": "Tenant slug",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "SQL SELECT query",
				},
			},
			"required": []string{"tenant", "query"},
		},
	},
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	slog.Info("SDA MCP Server starting", "tools", len(tools))

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Bytes()

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			slog.Error("invalid JSON-RPC request", "error", err)
			continue
		}

		resp := handleRequest(req)
		out, _ := json.Marshal(resp)
		fmt.Fprintln(os.Stdout, string(out))
	}
}

func handleRequest(req jsonRPCRequest) jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "sda-mcp",
					"version": "1.0.0",
				},
			},
		}

	case "tools/list":
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": tools,
			},
		}

	case "tools/call":
		return handleToolCall(req)

	default:
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}

func handleToolCall(req jsonRPCRequest) jsonRPCResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, -32602, "invalid params")
	}

	switch params.Name {
	case "service_health":
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "TODO: implement service health check via tools/pkg/admin"},
				},
			},
		}

	case "tenant_list":
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "TODO: implement tenant list via tools/pkg/admin"},
				},
			},
		}

	default:
		return errorResponse(req.ID, -32602, "unknown tool: "+params.Name)
	}
}

func errorResponse(id any, code int, msg string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{Code: code, Message: msg},
	}
}
