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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Camionerou/rag-saldivia/tools/pkg/admin"
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
		Description: "List all tenants in the platform. Returns slug, name, plan, enabled status.",
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
					"description": "Service name (auth, ws, chat, rag, notification, platform, ingest, feedback)",
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
	{
		Name:        "deploy",
		Description: "Deploy a service to production. Requires confirm=true to execute.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service": map[string]interface{}{
					"type":        "string",
					"description": "Service name to deploy",
				},
				"confirm": map[string]interface{}{
					"type":        "boolean",
					"description": "Must be true to execute the deploy",
				},
			},
			"required": []string{"service", "confirm"},
		},
	},
	{
		Name:        "rag_query",
		Description: "Query a RAG collection with a natural language question.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "RAG collection name",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Natural language query",
				},
			},
			"required": []string{"collection", "query"},
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
		_, _ = fmt.Fprintln(os.Stdout, string(out))
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

// envOrDefault reads an environment variable with a fallback.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getPlatformDBURL() string {
	url := envOrDefault("POSTGRES_PLATFORM_URL", "")
	if url == "" {
		url = envOrDefault("SDA_PLATFORM_DB", "")
	}
	if url == "" {
		url = "postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable"
	}
	return url
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
		return handleServiceHealth(req.ID)

	case "tenant_list":
		return handleTenantList(req.ID)

	case "service_logs":
		return handleServiceLogs(req.ID, params.Arguments)

	case "db_query":
		return handleDBQuery(req.ID, params.Arguments)

	case "deploy":
		return handleDeploy(req.ID, params.Arguments)

	case "rag_query":
		return handleRAGQuery(req.ID, params.Arguments)

	default:
		return errorResponse(req.ID, -32602, "unknown tool: "+params.Name)
	}
}

func handleServiceHealth(id any) jsonRPCResponse {
	baseHost := envOrDefault("SDA_HOST", "localhost")
	results := admin.ServiceHealth(baseHost)

	var sb strings.Builder
	sb.WriteString("SERVICE      PORT  STATUS  LATENCY\n")
	for _, s := range results {
		latency := "-"
		if s.Latency > 0 {
			latency = s.Latency.Round(time.Millisecond).String()
		}
		sb.WriteString(fmt.Sprintf("%-12s %s  %-6s  %s\n", s.Name, s.Port, s.Status, latency))
	}

	return textResult(id, sb.String())
}

func handleTenantList(id any) jsonRPCResponse {
	dbURL := getPlatformDBURL()
	tenants, err := admin.TenantList(dbURL)
	if err != nil {
		return textResult(id, fmt.Sprintf("Error listing tenants: %v", err))
	}

	if len(tenants) == 0 {
		return textResult(id, "No tenants found.")
	}

	var sb strings.Builder
	sb.WriteString("SLUG         NAME                  PLAN       ENABLED  CREATED\n")
	for _, t := range tenants {
		sb.WriteString(fmt.Sprintf("%-12s %-21s %-10s %-7v  %s\n",
			t.Slug, t.Name, t.PlanID, t.Enabled, t.CreatedAt.Format("2006-01-02")))
	}

	return textResult(id, sb.String())
}

func handleServiceLogs(id any, argsRaw json.RawMessage) jsonRPCResponse {
	var args struct {
		Service string `json:"service"`
		Lines   int    `json:"lines"`
	}
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return errorResponse(id, -32602, "invalid arguments for service_logs")
	}

	if args.Lines <= 0 {
		args.Lines = 50
	}

	output, err := admin.ServiceLogs(args.Service, args.Lines)
	if err != nil {
		return textResult(id, fmt.Sprintf("Error getting logs: %v", err))
	}

	return textResult(id, output)
}

func handleDBQuery(id any, argsRaw json.RawMessage) jsonRPCResponse {
	var args struct {
		Tenant string `json:"tenant"`
		Query  string `json:"query"`
	}
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return errorResponse(id, -32602, "invalid arguments for db_query")
	}

	dbURL := getPlatformDBURL()
	rows, err := admin.DBQuery(dbURL, args.Tenant, args.Query)
	if err != nil {
		return textResult(id, fmt.Sprintf("Error: %v", err))
	}
	if len(rows) == 0 {
		return textResult(id, "No rows returned.")
	}

	out, _ := json.MarshalIndent(rows, "", "  ")
	return textResult(id, fmt.Sprintf("%d rows returned:\n%s", len(rows), string(out)))
}

func handleDeploy(id any, argsRaw json.RawMessage) jsonRPCResponse {
	var args struct {
		Service string `json:"service"`
		Confirm bool   `json:"confirm"`
	}
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return errorResponse(id, -32602, "invalid arguments for deploy")
	}
	if !args.Confirm {
		return textResult(id, fmt.Sprintf("Deploy %q requires confirm=true. This will restart the service in production.", args.Service))
	}

	output, err := admin.Deploy(args.Service)
	if err != nil {
		return textResult(id, fmt.Sprintf("Deploy failed: %v\nOutput: %s", err, output))
	}
	return textResult(id, fmt.Sprintf("Service %q deployed successfully.\n%s", args.Service, output))
}

func handleRAGQuery(id any, argsRaw json.RawMessage) jsonRPCResponse {
	var args struct {
		Collection string `json:"collection"`
		Query      string `json:"query"`
	}
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return errorResponse(id, -32602, "invalid arguments for rag_query")
	}

	// Call /v1/search on the consolidated app binary (ADR 025 rag fusion).
	// Override with SEARCH_SERVICE_URL when pointing at a remote workstation.
	searchURL := envOrDefault("SEARCH_SERVICE_URL", "http://localhost:8020")

	payload := map[string]string{"collection": args.Collection, "query": args.Query}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return errorResponse(id, -32602, "failed to marshal request")
	}
	resp, err := http.Post(searchURL+"/v1/search/query", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return textResult(id, fmt.Sprintf("Search service unreachable: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	var result json.RawMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(&result); err != nil {
		return textResult(id, fmt.Sprintf("Failed to parse search response: %v", err))
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return textResult(id, string(out))
}

func textResult(id any, text string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": text},
			},
		},
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
