package config_test

import (
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
