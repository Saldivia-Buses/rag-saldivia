package tools

// White-box tests for GRPCSearchClient — same package for access to unexported fields.
//
// GRPCSearchClient.client is a searchv1.SearchServiceClient interface — injectable
// via struct literal in white-box tests. The conn field is *grpc.ClientConn;
// tests that don't call Close() leave it nil to avoid a nil-pointer deref.
//
// Key invariants tested:
//   - JWT is forwarded via gRPC metadata (ForwardJWT)
//   - CollectionID is optional (only set when non-empty)
//   - gRPC errors map to Result{Status:"error"}, not a Go error
//   - Successful response is marshalled to JSON and returned as Result.Data
//   - Invalid params JSON → Result{Status:"error"} (not a panic)

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	searchv1 "github.com/Camionerou/rag-saldivia/gen/go/search/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSearchClient is a test double for searchv1.SearchServiceClient.
type mockSearchClient struct {
	// capturedCtx is set when Query is called, for metadata inspection.
	capturedCtx context.Context
	// capturedReq is the SearchRequest sent by the client.
	capturedReq *searchv1.SearchRequest
	// resp is returned when Query is called.
	resp *searchv1.SearchResponse
	// err is returned by Query.
	err error
}

func (m *mockSearchClient) Query(ctx context.Context, in *searchv1.SearchRequest, _ ...grpc.CallOption) (*searchv1.SearchResponse, error) {
	m.capturedCtx = ctx
	m.capturedReq = in
	return m.resp, m.err
}

// newTestGRPCSearchClient creates a GRPCSearchClient with an injected mock.
// conn is left nil — tests must not call Close().
func newTestGRPCSearchClient(mock *mockSearchClient) *GRPCSearchClient {
	return &GRPCSearchClient{
		client: mock,
		conn:   nil, // do not call Close() in tests using this constructor
	}
}

// TestGRPCSearchTool_Execute_Success verifies the happy path: valid params,
// gRPC returns a response → Result{Status:"success", Data: json}.
func TestGRPCSearchTool_Execute_Success(t *testing.T) {
	t.Parallel()
	mock := &mockSearchClient{
		resp: &searchv1.SearchResponse{
			Query:      "test query",
			DurationMs: 42,
		},
	}
	c := newTestGRPCSearchClient(mock)

	params := json.RawMessage(`{"query":"test query","max_nodes":5}`)
	result, err := c.Execute(context.Background(), "test-jwt", params)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.Data, "result data must be non-empty protojson")
}

// TestGRPCSearchTool_Execute_GRPCError_ReturnsToolError verifies that a gRPC
// error is wrapped as a tool error result, not returned as a Go error.
// The caller (service.Agent) uses Result.Status to determine outcome.
func TestGRPCSearchTool_Execute_GRPCError_ReturnsToolError(t *testing.T) {
	t.Parallel()
	mock := &mockSearchClient{
		err: errors.New("rpc error: connection refused"),
	}
	c := newTestGRPCSearchClient(mock)

	params := json.RawMessage(`{"query":"test"}`)
	result, err := c.Execute(context.Background(), "tok", params)

	require.NoError(t, err, "Go error must be nil — gRPC errors are wrapped in Result")
	require.NotNil(t, result)
	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestGRPCSearchTool_Execute_InvalidJSON_ReturnsError verifies that invalid
// params JSON is handled gracefully without panicking.
func TestGRPCSearchTool_Execute_InvalidJSON_ReturnsError(t *testing.T) {
	t.Parallel()
	mock := &mockSearchClient{}
	c := newTestGRPCSearchClient(mock)

	result, err := c.Execute(context.Background(), "tok", json.RawMessage(`{invalid`))

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "error", result.Status)
	assert.NotEmpty(t, result.Error)
}

// TestGRPCSearchTool_Execute_PropagatesTenantID verifies that the JWT is
// forwarded via gRPC outgoing metadata (ForwardJWT wraps it in the
// "authorization" key). The mock captures the context and we inspect it.
func TestGRPCSearchTool_Execute_PropagatesTenantID(t *testing.T) {
	t.Parallel()
	mock := &mockSearchClient{
		resp: &searchv1.SearchResponse{},
	}
	c := newTestGRPCSearchClient(mock)

	const testJWT = "eyJhbGciOiJFZERTQSJ9.test.sig"
	params := json.RawMessage(`{"query":"what is the route?"}`)
	_, err := c.Execute(context.Background(), testJWT, params)
	require.NoError(t, err)

	// The mock captures the ctx passed to Query. ForwardJWT adds outgoing metadata.
	require.NotNil(t, mock.capturedCtx, "mock Query must have been called")
	md, ok := metadata.FromOutgoingContext(mock.capturedCtx)
	require.True(t, ok, "outgoing metadata must be present in the gRPC context")
	authValues := md.Get("authorization")
	require.NotEmpty(t, authValues, "authorization header must be present")
	assert.Equal(t, "Bearer "+testJWT, authValues[0], "JWT must be forwarded as Bearer token")
}

// TestGRPCSearchTool_Execute_OptionalCollectionID verifies that when
// collection_id is empty, the SearchRequest.CollectionId field is NOT set
// (optional field semantics).
func TestGRPCSearchTool_Execute_OptionalCollectionID(t *testing.T) {
	t.Parallel()
	mock := &mockSearchClient{resp: &searchv1.SearchResponse{}}
	c := newTestGRPCSearchClient(mock)

	// No collection_id in params
	params := json.RawMessage(`{"query":"fleet vehicles","max_nodes":10}`)
	_, err := c.Execute(context.Background(), "tok", params)
	require.NoError(t, err)

	require.NotNil(t, mock.capturedReq)
	assert.Nil(t, mock.capturedReq.CollectionId, "CollectionId must be nil when not provided")
	assert.Equal(t, "fleet vehicles", mock.capturedReq.Query)
	assert.Equal(t, int32(10), mock.capturedReq.MaxNodes)
}

// TestGRPCSearchTool_Execute_WithCollectionID verifies that collection_id is
// set on the request when provided in params.
func TestGRPCSearchTool_Execute_WithCollectionID(t *testing.T) {
	t.Parallel()
	mock := &mockSearchClient{resp: &searchv1.SearchResponse{}}
	c := newTestGRPCSearchClient(mock)

	params := json.RawMessage(`{"query":"invoices","collection_id":"col-123"}`)
	_, err := c.Execute(context.Background(), "tok", params)
	require.NoError(t, err)

	require.NotNil(t, mock.capturedReq)
	require.NotNil(t, mock.capturedReq.CollectionId, "CollectionId must be set when provided")
	assert.Equal(t, "col-123", *mock.capturedReq.CollectionId)
}

// TestGRPCSearchTool_Execute_EmptyParams_ReturnsError verifies that empty/nil
// params are handled without panicking.
func TestGRPCSearchTool_Execute_EmptyParams_ReturnsError(t *testing.T) {
	t.Parallel()
	mock := &mockSearchClient{}
	c := newTestGRPCSearchClient(mock)

	result, err := c.Execute(context.Background(), "tok", nil)
	// nil params → json.Unmarshal fails → error result
	require.NoError(t, err)
	require.NotNil(t, result)
	// Either an error (unmarshal failed) or the mock wasn't called
	assert.NotEqual(t, "success", result.Status, "nil params should not succeed")
}
