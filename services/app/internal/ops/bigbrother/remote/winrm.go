package remote

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/masterzen/winrm"
)

// WinRMClient wraps a WinRM connection to a single Windows host.
// HTTPS only (port 5986) — HTTP (port 5985) is never used.
type WinRMClient struct {
	client *winrm.Client
	addr   string
}

// WinRMConfig holds connection parameters for a WinRM client.
type WinRMConfig struct {
	Host     string
	Port     int    // default 5986 (HTTPS only)
	User     string
	Password string
	Timeout  time.Duration // default 30s

	// CACertFile is the path to a CA certificate for server cert validation.
	// If empty, the system cert pool is used. Set Insecure=true to skip
	// validation entirely (NOT recommended for production).
	CACertFile string

	// Insecure skips TLS certificate validation. Use ONLY for testing.
	// In production, always validate certificates via CACertFile or system pool.
	Insecure bool
}

// NewWinRMClient creates a WinRM client. Always uses HTTPS (port 5986).
func NewWinRMClient(cfg WinRMConfig) (*WinRMClient, error) {
	if cfg.Port == 0 {
		cfg.Port = 5986
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	var caCert []byte
	if cfg.CACertFile != "" {
		var err error
		caCert, err = os.ReadFile(cfg.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("read CA cert: %w", err)
		}
	}

	var certPool *x509.CertPool
	if caCert != nil {
		certPool = x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("invalid CA certificate in %s", cfg.CACertFile)
		}
	}

	// Always HTTPS (useHTTPS=true), insecure controlled by config
	endpoint := winrm.NewEndpoint(
		cfg.Host,
		cfg.Port,
		true,          // useHTTPS — always true, never HTTP
		cfg.Insecure,  // insecure — false in production (validates certs)
		caCert,        // CA cert bytes
		nil,           // client cert
		nil,           // client key
		cfg.Timeout,
	)

	client, err := winrm.NewClient(endpoint, cfg.User, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("create WinRM client: %w", err)
	}

	return &WinRMClient{
		client: client,
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}, nil
}

// Exec runs a command on the remote Windows host and returns the output.
// The command string is controlled by the caller — for production use,
// always use an enum-based allowlist.
func (c *WinRMClient) Exec(ctx context.Context, command string) (*ExecResult, error) {
	var stdout, stderr bytes.Buffer

	exitCode, err := c.client.RunWithContext(ctx, command, &stdout, &stderr)
	if err != nil {
		return nil, fmt.Errorf("WinRM exec on %s: %w", c.addr, err)
	}

	return &ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// Addr returns the remote address this client is connected to.
func (c *WinRMClient) Addr() string {
	return c.addr
}
