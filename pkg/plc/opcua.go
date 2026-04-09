package plc

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

// OPCUAClient wraps an OPC-UA connection to a single PLC/server.
// Thread-safe: the underlying client is protected by a mutex.
type OPCUAClient struct {
	client   *opcua.Client
	mu       sync.Mutex
	endpoint string
}

// OPCUAConfig holds connection parameters for an OPC-UA client.
type OPCUAConfig struct {
	Endpoint string // "opc.tcp://host:4840"

	// Security — SignAndEncrypt is default. Set SecurityMode to ua.MessageSecurityModeNone
	// for explicitly insecure connections (requires opt-in).
	SecurityMode ua.MessageSecurityMode
	SecurityPolicy string

	// Client certificate for mutual TLS. Pre-generated and stored as Docker secret.
	CertFile string
	KeyFile  string
}

// NewOPCUAClient creates a new OPC-UA client. Does not connect immediately.
func NewOPCUAClient(cfg OPCUAConfig) (*OPCUAClient, error) {
	opts := []opcua.Option{
		opcua.SecurityModeString("SignAndEncrypt"),
	}

	if cfg.SecurityMode == ua.MessageSecurityModeNone {
		slog.Warn("OPC-UA: using insecure mode (MessageSecurityModeNone)", "endpoint", cfg.Endpoint)
		opts = []opcua.Option{
			opcua.SecurityModeString("None"),
		}
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load OPC-UA cert: %w", err)
		}
		pk, ok := cert.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("OPC-UA cert key must be RSA")
		}
		opts = append(opts,
			opcua.Certificate(cert.Certificate[0]),
			opcua.PrivateKey(pk),
		)
	}

	c, err := opcua.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("create OPC-UA client: %w", err)
	}

	return &OPCUAClient{
		client:   c,
		endpoint: cfg.Endpoint,
	}, nil
}

// Connect opens the OPC-UA session.
func (c *OPCUAClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.client.Connect(ctx); err != nil {
		return fmt.Errorf("opcua connect %s: %w", c.endpoint, err)
	}
	return nil
}

// Close closes the OPC-UA session.
func (c *OPCUAClient) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.Close(ctx)
}

// OPCUANode represents a single OPC-UA node reading.
type OPCUANode struct {
	NodeID string
	Name   string
	Value  any
}

// ReadNode reads a single OPC-UA node value by node ID string.
func (c *OPCUAClient) ReadNode(ctx context.Context, nodeID string) (*OPCUANode, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("parse node ID %q: %w", nodeID, err)
	}

	req := &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{
			{NodeID: id, AttributeID: ua.AttributeIDValue},
		},
	}

	resp, err := c.client.Read(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("read node %s: %w", nodeID, err)
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results for node %s", nodeID)
	}

	result := resp.Results[0]
	if result.Status != ua.StatusOK {
		return nil, fmt.Errorf("node %s status: %v", nodeID, result.Status)
	}

	return &OPCUANode{
		NodeID: nodeID,
		Value:  result.Value.Value(),
	}, nil
}

// BrowseChildren returns the child nodes of the given parent node.
func (c *OPCUAClient) BrowseChildren(ctx context.Context, parentNodeID string) ([]OPCUANode, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := ua.ParseNodeID(parentNodeID)
	if err != nil {
		return nil, fmt.Errorf("parse node ID %q: %w", parentNodeID, err)
	}

	req := &ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			{
				NodeID:          id,
				BrowseDirection: ua.BrowseDirectionForward,
				ReferenceTypeID: ua.NewNumericNodeID(0, 33), // HierarchicalReferences
				IncludeSubtypes: true,
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			},
		},
	}

	resp, err := c.client.Browse(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("browse node %s: %w", parentNodeID, err)
	}

	if len(resp.Results) == 0 {
		return nil, nil
	}

	nodes := make([]OPCUANode, 0, len(resp.Results[0].References))
	for _, ref := range resp.Results[0].References {
		nodes = append(nodes, OPCUANode{
			NodeID: ref.NodeID.NodeID.String(),
			Name:   ref.BrowseName.Name,
		})
	}

	return nodes, nil
}

// WriteNode writes a value to an OPC-UA node.
func (c *OPCUAClient) WriteNode(ctx context.Context, nodeID string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		return fmt.Errorf("parse node ID %q: %w", nodeID, err)
	}

	v, err := ua.NewVariant(value)
	if err != nil {
		return fmt.Errorf("create variant: %w", err)
	}

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					Value: v,
				},
			},
		},
	}

	resp, err := c.client.Write(ctx, req)
	if err != nil {
		return fmt.Errorf("write node %s: %w", nodeID, err)
	}

	if len(resp.Results) > 0 && resp.Results[0] != ua.StatusOK {
		return fmt.Errorf("write node %s status: %v", nodeID, resp.Results[0])
	}

	return nil
}

// Endpoint returns the OPC-UA endpoint this client is connected to.
func (c *OPCUAClient) Endpoint() string {
	return c.endpoint
}
