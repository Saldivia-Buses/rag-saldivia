package otel

import (
	"context"
	"testing"
	"time"
)

func TestSetup_UnreachableEndpoint_NoPanic(t *testing.T) {
	// Setup should succeed even if the collector is unreachable.
	// The OTLP exporter connects lazily; traces are dropped silently.
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "0.0.1",
		Endpoint:       "localhost:19999", // nothing listening here
	}

	shutdown, err := Setup(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Setup with unreachable endpoint should not error, got: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown function should not be nil")
	}

	// Shutdown must also complete without error
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestSetup_ShutdownCallable(t *testing.T) {
	cfg := Config{
		ServiceName:    "shutdown-test",
		ServiceVersion: "1.0.0",
		Endpoint:       "localhost:19998",
	}

	shutdown, err := Setup(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	// Calling shutdown multiple times should not panic.
	// The first call shuts down the tracer provider; subsequent calls
	// may return an error but must not crash.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	if err := shutdown(ctx); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}

	// Second call -- should not panic (may or may not error)
	_ = shutdown(ctx)
}

func TestSetup_DefaultEndpoint(t *testing.T) {
	// When Endpoint is empty, Setup should default to localhost:4317
	// and not fail. We verify indirectly: if it doesn't error or panic,
	// the default was applied (the real endpoint is unreachable but
	// OTLP connects lazily).
	cfg := Config{
		ServiceName:    "default-ep",
		ServiceVersion: "0.1.0",
		Endpoint:       "", // should default to localhost:4317
	}

	shutdown, err := Setup(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Setup with empty endpoint: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestSetup_EmptyServiceName(t *testing.T) {
	// An empty service name is allowed by OTel (falls back to "unknown_service").
	// Verify it does not panic or error.
	cfg := Config{
		ServiceName:    "",
		ServiceVersion: "",
		Endpoint:       "localhost:19997",
	}

	shutdown, err := Setup(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Setup with empty service name: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestSetup_CancelledContext(t *testing.T) {
	// Calling Setup with an already-cancelled context should return an error,
	// not panic.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	cfg := Config{
		ServiceName:    "cancelled",
		ServiceVersion: "0.0.1",
		Endpoint:       "localhost:19996",
	}

	shutdown, err := Setup(ctx, cfg)
	// Either an error or a valid shutdown is acceptable -- the key invariant
	// is no panic.
	if err != nil {
		// Expected: context cancelled during resource or exporter creation.
		return
	}
	if shutdown != nil {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutCancel()
		_ = shutdown(shutCtx)
	}
}
