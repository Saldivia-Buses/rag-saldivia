// Package remote provides SSH and WinRM clients for remote command execution
// and file transfer on network devices.
//
// WARNING: This package provides raw transport clients. For production use,
// always go through a service that enforces command allowlists, audit logging,
// credential management, and rate limiting (e.g., BigBrother).
package remote

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHClient wraps an SSH connection to a single host.
type SSHClient struct {
	client *ssh.Client
	addr   string
}

// SSHConfig holds connection parameters for an SSH client.
type SSHConfig struct {
	Host     string
	Port     int           // default 22
	User     string
	Password string        // password auth
	KeyBytes []byte        // PEM-encoded private key (alternative to password)
	Timeout  time.Duration // default 10s
}

// NewSSHClient connects to a remote host via SSH.
func NewSSHClient(cfg SSHConfig) (*SSHClient, error) {
	if cfg.Port == 0 {
		cfg.Port = 22
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	var authMethods []ssh.AuthMethod
	if len(cfg.KeyBytes) > 0 {
		signer, err := ssh.ParsePrivateKey(cfg.KeyBytes)
		if err != nil {
			return nil, fmt.Errorf("parse SSH key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("SSH: no auth method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		Timeout:         cfg.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: TOFU or known_hosts
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH connect %s: %w", addr, err)
	}

	return &SSHClient{client: client, addr: addr}, nil
}

// ExecResult holds the output of a remote command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Exec runs a command on the remote host and returns the output.
// The command string is controlled by the caller — for production use,
// always use an enum-based allowlist.
func (c *SSHClient) Exec(ctx context.Context, command string) (*ExecResult, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return nil, ctx.Err()
	case err := <-done:
		result := &ExecResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
			} else {
				return nil, fmt.Errorf("SSH exec: %w", err)
			}
		}
		return result, nil
	}
}

// ReadFile reads a file from the remote host via SFTP.
// The caller MUST validate the path against an allowlist before calling.
// This method checks for symlinks and rejects them (defense against
// compromised targets using symlinks to exfiltrate data).
func (c *SSHClient) ReadFile(ctx context.Context, path string, maxSize int64) ([]byte, error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return nil, fmt.Errorf("create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Symlink protection: reject symlinks to prevent path traversal
	info, err := sftpClient.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("SFTP lstat %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("SFTP: rejected symlink at %s (security)", path)
	}

	f, err := sftpClient.Open(path)
	if err != nil {
		return nil, fmt.Errorf("SFTP open %s: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxSize))
	if err != nil {
		return nil, fmt.Errorf("SFTP read %s: %w", path, err)
	}

	return data, nil
}

// Close closes the SSH connection.
func (c *SSHClient) Close() error {
	return c.client.Close()
}

// Addr returns the remote address this client is connected to.
func (c *SSHClient) Addr() string {
	return c.addr
}

// IsReachable checks if an SSH port is reachable without authenticating.
func IsReachable(host string, port int, timeout time.Duration) bool {
	if port == 0 {
		port = 22
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
