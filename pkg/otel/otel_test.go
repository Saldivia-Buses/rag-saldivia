package otel

import (
	"context"
	"testing"
	"time"
)

// These tests verify that Setup and Shutdown do not panic when the OTel
// collector is unreachable. Shutdown may return an error (the gRPC exporter
// drops data silently when the endpoint is down) — that is expected behavior
// and is NOT asserted here. The tests only verify no panic occurs.

func TestSetup_UnreachableEndpoint_NoPanic(t *testing.T) {
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "0.0.1",
		Endpoint:       "localhost:19999",
	}
	shutdown, err := Setup(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Setup with unreachable endpoint should not error, got: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown function should not be nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx) // error expected — endpoint unreachable
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
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx) // first call — error expected, must not panic
	_ = shutdown(ctx) // second call — must not panic either
}

func TestSetup_DefaultEndpoint(t *testing.T) {
	cfg := Config{
		ServiceName:    "default-ep",
		ServiceVersion: "0.1.0",
		Endpoint:       "",
	}
	shutdown, err := Setup(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Setup with empty endpoint: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx)
}

func TestSetup_EmptyServiceName(t *testing.T) {
	cfg := Config{
		ServiceName:    "",
		ServiceVersion: "",
		Endpoint:       "localhost:19997",
	}
	shutdown, err := Setup(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Setup with empty service name: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx)
}

func TestSetup_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	cfg := Config{
		ServiceName:    "cancelled",
		ServiceVersion: "0.0.1",
		Endpoint:       "localhost:19996",
	}
	shutdown, err := Setup(ctx, cfg)
	if err != nil {
		return // expected: context cancelled during setup
	}
	if shutdown != nil {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer shutCancel()
		_ = shutdown(shutCtx)
	}
}
