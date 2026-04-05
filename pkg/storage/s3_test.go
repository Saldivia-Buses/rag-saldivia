package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/Camionerou/rag-saldivia/pkg/storage"
)

func newTestStore(t *testing.T) *storage.S3Store {
	t.Helper()

	endpoint := os.Getenv("STORAGE_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:9000"
	}

	ctx := context.Background()
	store, err := storage.NewS3Store(ctx, storage.S3Config{
		Endpoint:  endpoint,
		Bucket:    "sda-test",
		AccessKey: envOr("STORAGE_ACCESS_KEY", "sda-admin"),
		SecretKey: envOr("STORAGE_SECRET_KEY", "sda-dev-secret"),
	})
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	if err := store.EnsureBucket(ctx); err != nil {
		t.Skipf("MinIO not available: %v", err)
	}

	return store
}

func TestPutGetDelete(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	key := "test-tenant/doc-1/original.pdf"
	content := []byte("fake pdf content for testing")

	// Put
	if err := store.Put(ctx, key, bytes.NewReader(content)); err != nil {
		t.Fatalf("put: %v", err)
	}

	// Exists
	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if !exists {
		t.Fatal("file should exist after put")
	}

	// Get
	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("content mismatch: got %q, want %q", got, content)
	}

	// Delete
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Exists after delete
	exists, err = store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("exists after delete: %v", err)
	}
	if exists {
		t.Fatal("file should not exist after delete")
	}
}

func TestExistsNotFound(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	exists, err := store.Exists(ctx, "nonexistent/key")
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if exists {
		t.Fatal("nonexistent key should return false")
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
