// Package storage provides a pluggable file storage interface.
// The default implementation uses S3-compatible APIs (MinIO locally, AWS S3 in production).
// All services that need to store or retrieve files use this interface.
package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound is returned when a requested key does not exist.
var ErrNotFound = errors.New("storage: not found")

// PutOptions configures a Put operation.
type PutOptions struct {
	ContentType string // e.g. "application/pdf", "image/png". Empty = "application/octet-stream".
}

// Store is the interface for file storage operations.
// Keys are slash-separated paths: "{tenant}/{doc_id}/original.pdf".
type Store interface {
	// Put stores a file at the given key, reading from r.
	// opts may be nil for defaults.
	Put(ctx context.Context, key string, r io.Reader, opts *PutOptions) error

	// Get returns a reader for the file at the given key.
	// Returns ErrNotFound if the key does not exist.
	// The caller must close the returned ReadCloser.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the file at the given key.
	// Returns nil if the file does not exist.
	Delete(ctx context.Context, key string) error

	// Exists checks whether a file exists at the given key.
	Exists(ctx context.Context, key string) (bool, error)
}
