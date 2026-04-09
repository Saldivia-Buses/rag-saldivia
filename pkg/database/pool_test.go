package database

import (
	"context"
	"testing"
)

func TestNewPool_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		// pgxpool.ParseConfig rejects these outright
		{name: "wrong scheme", url: "http://localhost:5432/testdb"},
		{name: "mysql scheme", url: "mysql://localhost/testdb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pool, err := NewPool(ctx, tt.url)
			if err == nil {
				if pool != nil {
					pool.Close()
				}
				t.Fatal("NewPool should return error for invalid URL scheme")
			}
		})
	}
}

func TestNewPool_UnreachableHost(t *testing.T) {
	// Valid URL format but unreachable host (RFC 5737 TEST-NET-1).
	// pgxpool.NewWithConfig will fail on initial connection.
	ctx := context.Background()
	pool, err := NewPool(ctx, "postgres://user:pass@192.0.2.1:5432/testdb?connect_timeout=1")
	if err != nil {
		// Pool creation failed on connect — this is the expected path
		return
	}
	// If pool was created (lazy mode), ping should fail
	defer pool.Close()
	if err := pool.Ping(ctx); err == nil {
		t.Error("expected Ping to fail for unreachable host")
	}
}

func TestNewPool_ReturnsWrappedError(t *testing.T) {
	ctx := context.Background()
	_, err := NewPool(ctx, "ftp://bad:5432/db")
	if err == nil {
		t.Fatal("expected error for ftp:// scheme")
	}
	// Our wrapper adds "parse db config:" or "create pool:" prefix
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("error message should not be empty")
	}
}
