// Package llm re-exports pkg/llm for the ingest service.
// The actual implementation lives in pkg/llm/.
package llm

import (
	pkgllm "github.com/Camionerou/rag-saldivia/pkg/llm"
)

// Client is a re-export of pkg/llm.Client.
type Client = pkgllm.Client

// NewClient creates a pkg/llm.Client.
func NewClient(endpoint, model, apiKey string) *Client {
	return pkgllm.NewClient(endpoint, model, apiKey)
}
