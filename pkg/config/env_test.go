package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Camionerou/rag-saldivia/pkg/config"
)

func TestEnv_WithValue(t *testing.T) {
	t.Setenv("TEST_ENV_KEY", "hello")

	got := config.Env("TEST_ENV_KEY", "default")
	if got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}

func TestEnv_Fallback(t *testing.T) {
	got := config.Env("NONEXISTENT_KEY_12345", "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestMustEnv_WithValue(t *testing.T) {
	t.Setenv("TEST_MUST_KEY", "value")

	got := config.MustEnv("TEST_MUST_KEY")
	if got != "value" {
		t.Fatalf("expected value, got %s", got)
	}
}

func TestMustEnv_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing env var")
		}
	}()
	config.MustEnv("DEFINITELY_NOT_SET_98765")
}

func TestEnvOrFile_EnvMode(t *testing.T) {
	t.Setenv("TEST_ENVORFILE", "direct")
	t.Setenv("TEST_ENVORFILE_FILE", "")
	if got := config.EnvOrFile("TEST_ENVORFILE"); got != "direct" {
		t.Fatalf("expected direct, got %q", got)
	}
}

func TestEnvOrFile_FileMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret")
	if err := os.WriteFile(path, []byte("fromfile\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("TEST_ENVORFILE", "")
	t.Setenv("TEST_ENVORFILE_FILE", path)
	if got := config.EnvOrFile("TEST_ENVORFILE"); got != "fromfile" {
		t.Fatalf("expected fromfile (trimmed), got %q", got)
	}
}

func TestEnvOrFile_EnvPreferredOverFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret")
	if err := os.WriteFile(path, []byte("fromfile"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("TEST_ENVORFILE", "direct")
	t.Setenv("TEST_ENVORFILE_FILE", path)
	if got := config.EnvOrFile("TEST_ENVORFILE"); got != "direct" {
		t.Fatalf("expected direct (env wins over file), got %q", got)
	}
}

func TestEnvOrFile_NeitherSet(t *testing.T) {
	t.Setenv("TEST_ENVORFILE", "")
	t.Setenv("TEST_ENVORFILE_FILE", "")
	if got := config.EnvOrFile("TEST_ENVORFILE"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestEnvOrFile_FilePathMissing(t *testing.T) {
	t.Setenv("TEST_ENVORFILE", "")
	t.Setenv("TEST_ENVORFILE_FILE", "/nonexistent/path/no-such-file")
	// Silent failure: returns "" rather than panicking. Callers must check.
	if got := config.EnvOrFile("TEST_ENVORFILE"); got != "" {
		t.Fatalf("expected empty on missing file, got %q", got)
	}
}
