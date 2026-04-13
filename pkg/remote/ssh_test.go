package remote

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// generateTestKey returns a fresh ed25519 SSH signer and its public key.
func generateTestKey(t *testing.T) (ssh.Signer, ssh.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}
	pubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("create public key: %v", err)
	}
	return signer, pubKey
}

// fakeAddr is a minimal net.Addr for testing.
type fakeAddr struct{ addr string }

func (f fakeAddr) Network() string { return "tcp" }
func (f fakeAddr) String() string  { return f.addr }

// TestHostKeyCallback_InsecureSentinel verifies the "-" sentinel disables checks.
func TestHostKeyCallback_InsecureSentinel(t *testing.T) {
	cb, err := hostKeyCallback("-")
	if err != nil {
		t.Fatalf("hostKeyCallback(\"-\"): %v", err)
	}
	_, pubKey := generateTestKey(t)
	if err := cb("somehost:22", fakeAddr{"somehost:22"}, pubKey); err != nil {
		t.Errorf("insecure callback should accept any key, got: %v", err)
	}
}

// TestHostKeyCallback_TOFU_FirstUse verifies that an unknown host is trusted on first use
// and its key is written to the known_hosts file.
func TestHostKeyCallback_TOFU_FirstUse(t *testing.T) {
	dir := t.TempDir()
	khPath := filepath.Join(dir, "known_hosts")

	_, pubKey := generateTestKey(t)

	cb, err := hostKeyCallback(khPath)
	if err != nil {
		t.Fatalf("hostKeyCallback: %v", err)
	}

	host := "192.0.2.1:22"
	remote := fakeAddr{host}

	// First connection — host unknown, should be trusted (TOFU).
	if err := cb(host, remote, pubKey); err != nil {
		t.Fatalf("TOFU first-use should succeed, got: %v", err)
	}

	// The key must now be written to the file.
	data, err := os.ReadFile(khPath)
	if err != nil {
		t.Fatalf("read known_hosts: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("known_hosts file is empty after TOFU trust")
	}
}

// TestHostKeyCallback_TOFU_SubsequentMatch verifies that a known host with the
// same key is accepted on subsequent connections.
func TestHostKeyCallback_TOFU_SubsequentMatch(t *testing.T) {
	dir := t.TempDir()
	khPath := filepath.Join(dir, "known_hosts")

	_, pubKey := generateTestKey(t)
	host := "192.0.2.2:22"
	remote := fakeAddr{host}

	// Seed the known_hosts file with the key.
	f, err := os.Create(khPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintln(f, knownhosts.Line([]string{host}, pubKey)); err != nil {
		t.Fatalf("write known host: %v", err)
	}
	f.Close()

	cb, err := hostKeyCallback(khPath)
	if err != nil {
		t.Fatalf("hostKeyCallback: %v", err)
	}

	// Same key — should be accepted.
	if err := cb(host, remote, pubKey); err != nil {
		t.Errorf("subsequent connection with same key should succeed, got: %v", err)
	}
}

// TestHostKeyCallback_TOFU_KeyMismatch verifies that a changed host key is rejected.
func TestHostKeyCallback_TOFU_KeyMismatch(t *testing.T) {
	dir := t.TempDir()
	khPath := filepath.Join(dir, "known_hosts")

	_, originalKey := generateTestKey(t)
	_, newKey := generateTestKey(t)
	host := "192.0.2.3:22"
	remote := fakeAddr{host}

	// Seed the known_hosts file with the original key.
	f, err := os.Create(khPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintln(f, knownhosts.Line([]string{host}, originalKey)); err != nil {
		t.Fatalf("write known host: %v", err)
	}
	f.Close()

	cb, err := hostKeyCallback(khPath)
	if err != nil {
		t.Fatalf("hostKeyCallback: %v", err)
	}

	// Present a different key — must be rejected as MITM.
	err = cb(host, remote, newKey)
	if err == nil {
		t.Fatal("expected error for key mismatch (MITM), got nil")
	}
	if !strings.Contains(err.Error(), "MITM") {
		t.Errorf("error should mention MITM, got: %v", err)
	}
}

// TestHostKeyCallback_CreatesParentDirs verifies that missing parent directories
// for the known_hosts file are created automatically.
func TestHostKeyCallback_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	khPath := filepath.Join(dir, "nested", "subdir", "known_hosts")

	_, err := hostKeyCallback(khPath)
	if err != nil {
		t.Fatalf("hostKeyCallback should create parent dirs, got: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(khPath)); os.IsNotExist(err) {
		t.Error("parent directory was not created")
	}
}

// TestFingerprint verifies the fingerprint function produces SHA256: prefixed output.
func TestFingerprint(t *testing.T) {
	_, pubKey := generateTestKey(t)
	fp := fingerprint(pubKey)
	if !strings.HasPrefix(fp, "SHA256:") {
		t.Errorf("fingerprint should start with SHA256:, got: %s", fp)
	}
}

// TestNewSSHClient_NoAuthMethod verifies that an empty config is rejected before
// any network dial is attempted.
func TestNewSSHClient_NoAuthMethod(t *testing.T) {
	_, err := NewSSHClient(SSHConfig{
		Host:           "127.0.0.1",
		Port:           22,
		User:           "test",
		KnownHostsFile: "-", // disable TOFU for this unit test
	})
	if err == nil {
		t.Fatal("expected error for missing auth method")
	}
	if !strings.Contains(err.Error(), "no auth method") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestIsReachable_Unreachable verifies the helper returns false for a closed port.
func TestIsReachable_Unreachable(t *testing.T) {
	// Pick a port that is almost certainly not listening.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot bind test listener")
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close() // close immediately so the port is not listening

	if IsReachable("127.0.0.1", port, 200e6) {
		t.Error("expected IsReachable to return false for closed port")
	}
}
