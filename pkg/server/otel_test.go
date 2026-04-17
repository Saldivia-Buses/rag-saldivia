package server

import (
	"context"
	"testing"
	"time"
)

// These tests verify that setupOTel and its shutdown do not panic when the
// OTel collector is unreachable. Shutdown may return an error — the gRPC
// exporter drops data silently when the endpoint is down. That is expected
// behavior and is NOT asserted here. The tests only verify no panic occurs.

func TestSetupOTel_UnreachableEndpoint_NoPanic(t *testing.T) {
	cfg := otelConfig{
		ServiceName:    "test-service",
		ServiceVersion: "0.0.1",
		Endpoint:       "localhost:19999",
	}
	shutdown, err := setupOTel(t.Context(), cfg)
	if err != nil {
		t.Fatalf("setupOTel with unreachable endpoint should not error, got: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown function should not be nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx)
}

func TestSetupOTel_ShutdownCallable(t *testing.T) {
	cfg := otelConfig{
		ServiceName:    "shutdown-test",
		ServiceVersion: "1.0.0",
		Endpoint:       "localhost:19998",
	}
	shutdown, err := setupOTel(t.Context(), cfg)
	if err != nil {
		t.Fatalf("setupOTel: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx)
	_ = shutdown(ctx)
}

func TestSetupOTel_DefaultEndpoint(t *testing.T) {
	cfg := otelConfig{
		ServiceName:    "default-ep",
		ServiceVersion: "0.1.0",
		Endpoint:       "",
	}
	shutdown, err := setupOTel(t.Context(), cfg)
	if err != nil {
		t.Fatalf("setupOTel with empty endpoint: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx)
}

func TestSetupOTel_EmptyServiceName(t *testing.T) {
	cfg := otelConfig{
		ServiceName:    "",
		ServiceVersion: "",
		Endpoint:       "localhost:19997",
	}
	shutdown, err := setupOTel(t.Context(), cfg)
	if err != nil {
		t.Fatalf("setupOTel with empty service name: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx)
}

func TestSetupOTel_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	cfg := otelConfig{
		ServiceName:    "cancelled",
		ServiceVersion: "0.0.1",
		Endpoint:       "localhost:19996",
	}
	shutdown, err := setupOTel(ctx, cfg)
	if err != nil {
		return
	}
	if shutdown != nil {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer shutCancel()
		_ = shutdown(shutCtx)
	}
}
