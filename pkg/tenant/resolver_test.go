package tenant

import (
	"context"
	"testing"
)

func TestNewResolver(t *testing.T) {
	r := NewResolver(nil) // nil platformDB is fine for construction test
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
	if r.PoolMaxConns != 4 {
		t.Fatalf("expected PoolMaxConns=4, got %d", r.PoolMaxConns)
	}
	if r.closed {
		t.Fatal("expected resolver not closed")
	}
}

func TestResolver_Close(t *testing.T) {
	r := NewResolver(nil)
	r.Close()

	if !r.closed {
		t.Fatal("expected resolver closed after Close()")
	}

	// After close, methods should return error
	_, err := r.PostgresPool(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error after Close, got nil")
	}

	_, err = r.RedisClient(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error after Close, got nil")
	}
}

func TestResolver_Close_Idempotent(t *testing.T) {
	r := NewResolver(nil)
	r.Close()
	r.Close() // should not panic
}

func TestResolver_PostgresPool_NoDatabase(t *testing.T) {
	// Without a real platform DB, resolveConnInfo will fail
	r := NewResolver(nil)
	defer r.Close()

	_, err := r.PostgresPool(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nil platformDB, got nil")
	}
}

func TestResolver_RedisClient_NoDatabase(t *testing.T) {
	r := NewResolver(nil)
	defer r.Close()

	_, err := r.RedisClient(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nil platformDB, got nil")
	}
}
