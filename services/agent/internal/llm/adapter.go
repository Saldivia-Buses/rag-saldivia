// Package llm re-exports pkg/llm types for backward compatibility.
// The actual client implementation lives in pkg/llm/.
// This file exists so internal/service and internal/handler don't need
// to change their import paths in this PR.
package llm

import (
	pkgllm "github.com/Camionerou/rag-saldivia/pkg/llm"
)

// Re-export all types from pkg/llm so existing code compiles unchanged.
type (
	Message        = pkgllm.Message
	ToolCall       = pkgllm.ToolCall
	FunctionCall   = pkgllm.FunctionCall
	ToolSchema     = pkgllm.ToolSchema
	ToolDefinition = pkgllm.ToolDefinition
	ChatResponse   = pkgllm.ChatResponse
	Adapter        = pkgllm.Client
)

// NewAdapter creates a pkg/llm.Client (re-exported as Adapter for compatibility).
func NewAdapter(endpoint, model, apiKey string) *Adapter {
	return pkgllm.NewClient(endpoint, model, apiKey)
}
