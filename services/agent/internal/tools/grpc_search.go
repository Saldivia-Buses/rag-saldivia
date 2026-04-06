package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	searchv1 "github.com/Camionerou/rag-saldivia/gen/go/search/v1"
	sdagrpc "github.com/Camionerou/rag-saldivia/pkg/grpc"
	"google.golang.org/grpc"
)

// GRPCSearchClient wraps the search gRPC client for tool execution.
type GRPCSearchClient struct {
	client searchv1.SearchServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCSearchClient creates a gRPC client for the search service.
// Connection is lazy — does not fail if search is unreachable at startup.
func NewGRPCSearchClient(target string) (*GRPCSearchClient, error) {
	conn, err := sdagrpc.Dial(target)
	if err != nil {
		return nil, fmt.Errorf("dial search grpc: %w", err)
	}
	return &GRPCSearchClient{
		client: searchv1.NewSearchServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *GRPCSearchClient) Close() error {
	return c.conn.Close()
}

// Execute calls the search service via gRPC, returning a tool Result.
// params is the JSON from the LLM tool call: {"query": "...", "collection_id": "...", "max_nodes": N}
func (c *GRPCSearchClient) Execute(ctx context.Context, jwt string, params json.RawMessage) (*Result, error) {
	var p struct {
		Query        string `json:"query"`
		CollectionID string `json:"collection_id"`
		MaxNodes     int32  `json:"max_nodes"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return &Result{Status: "error", Error: "invalid search params: " + err.Error()}, nil
	}

	// Forward user JWT via gRPC metadata
	grpcCtx := sdagrpc.ForwardJWT(ctx, jwt)

	req := &searchv1.SearchRequest{
		Query:    p.Query,
		MaxNodes: p.MaxNodes,
	}
	if p.CollectionID != "" {
		req.CollectionId = &p.CollectionID
	}

	resp, err := c.client.Query(grpcCtx, req)
	if err != nil {
		slog.Warn("grpc search call failed, result as error", "error", err)
		return &Result{Status: "error", Error: "search failed"}, nil
	}

	data, _ := json.Marshal(resp)
	return &Result{Status: "success", Data: data}, nil
}
