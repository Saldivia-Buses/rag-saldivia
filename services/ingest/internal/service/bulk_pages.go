package service

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PageData represents a single page for bulk insertion.
type PageData struct {
	DocumentID string
	PageNumber int32
	Text       string
	Tables     json.RawMessage
	Images     json.RawMessage
}

// BulkInsertPages inserts multiple pages in a single round-trip using pgx.CopyFrom.
// Deletes existing pages for the document first (atomic via transaction).
// Returns the number of pages inserted.
func BulkInsertPages(ctx context.Context, pool *pgxpool.Pool, docID string, pages []PageData) (int64, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Delete existing pages for this document (idempotent re-processing)
	_, err = tx.Exec(ctx, "DELETE FROM document_pages WHERE document_id = $1", docID)
	if err != nil {
		return 0, err
	}

	// Build rows for CopyFrom
	rows := make([][]any, len(pages))
	for i, p := range pages {
		tables := p.Tables
		if tables == nil {
			tables = json.RawMessage("[]")
		}
		images := p.Images
		if images == nil {
			images = json.RawMessage("[]")
		}
		rows[i] = []any{p.DocumentID, p.PageNumber, p.Text, []byte(tables), []byte(images)}
	}

	n, err := tx.CopyFrom(ctx,
		pgx.Identifier{"document_pages"},
		[]string{"document_id", "page_number", "text", "tables", "images"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, err
	}

	return n, tx.Commit(ctx)
}
