package migration

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ParallelCopyWriter fan-outs a large row set to N workers, each COPYing
// its own chunk on its own connection. PG serialises disk writes per relation
// but the WAL path, binary decoding and index maintenance all parallelise
// cleanly. For rag-saldivia tables this lifts the per-batch ceiling that
// single-connection COPY hits once the network and WAL flush dominate.
//
// Each worker owns its tx, so ON CONFLICT dedup still works within the chunk.
// Across chunks, duplicates are resolved by the staging table's INSERT ...
// SELECT ... ON CONFLICT DO NOTHING — two workers inserting the same natural
// key just see ON CONFLICT on their own side.
type ParallelCopyWriter struct {
	pool     *pgxpool.Pool
	tenantID string
	workers  int
}

// NewParallelCopyWriter returns a writer that will fan-out to `workers`
// connections. workers<=0 defaults to 4. workers>pool.MaxConns() is clamped
// by the pool, so it's safe to request more than the pool can serve.
func NewParallelCopyWriter(pool *pgxpool.Pool, tenantID string, workers int) *ParallelCopyWriter {
	if workers <= 0 {
		workers = 4
	}
	return &ParallelCopyWriter{pool: pool, tenantID: tenantID, workers: workers}
}

// WriteBatch splits rows into `workers` near-equal chunks, runs COPY+staging
// on each, and returns the sum of rows persisted. Callers must pass rows that
// fit memory — this is not a streaming API.
//
// Semantics match CopyWriter.WriteBatch for idempotency. When a deadlock is
// detected (common when multiple workers touch the same unique index ranges),
// the worker falls back to a single-threaded retry in-process.
func (w *ParallelCopyWriter) WriteBatch(ctx context.Context, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	// Single-worker fallback: target tables with a hot unique constraint
	// (erp_entities, erp_articles, etc.) deadlock under parallel COPY because
	// each worker locks overlapping index ranges during INSERT ... SELECT.
	// If any chunk fails with SQLSTATE 40P01 we retry the whole batch on one
	// connection, which serialises the writes.
	workers := w.workers
	if len(rows) < workers {
		workers = len(rows)
	}

	chunkSize := (len(rows) + workers - 1) / workers
	chunks := make([][][]any, 0, workers)
	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunks = append(chunks, rows[i:end])
	}

	var (
		mu       sync.Mutex
		total    int
		firstErr error
		wg       sync.WaitGroup
	)
	for _, chunk := range chunks {
		wg.Add(1)
		go func(chunk [][]any) {
			defer wg.Done()

			tx, err := w.pool.Begin(ctx)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("worker begin: %w", err)
				}
				mu.Unlock()
				return
			}
			// Tenant context must live in every tx that touches an RLS table.
			if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", w.tenantID); err != nil {
				_ = tx.Rollback(ctx)
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("worker set tenant: %w", err)
				}
				mu.Unlock()
				return
			}

			n, err := copyChunkTx(ctx, tx, table, columns, conflictColumn, chunk)
			if err != nil {
				_ = tx.Rollback(ctx)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			if err := tx.Commit(ctx); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("worker commit: %w", err)
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			total += n
			mu.Unlock()
		}(chunk)
	}
	wg.Wait()

	if firstErr != nil {
		// Deadlock (SQLSTATE 40P01): fall back to single-threaded retry.
		// Conservative: check the error text; pgx doesn't expose a nicer
		// API from this surface and the cost of the check is trivial.
		if strings.Contains(firstErr.Error(), "deadlock detected") ||
			strings.Contains(firstErr.Error(), "40P01") {
			return w.singleThreadedFallback(ctx, table, columns, conflictColumn, rows)
		}
		return 0, firstErr
	}
	return total, nil
}

// singleThreadedFallback retries a whole batch through one COPY+staging tx,
// bypassing the fan-out entirely. Used only when parallel execution hit a
// deadlock on a hot unique index.
func (w *ParallelCopyWriter) singleThreadedFallback(ctx context.Context, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("fallback begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", w.tenantID); err != nil {
		return 0, fmt.Errorf("fallback tenant: %w", err)
	}
	n, err := copyChunkTx(ctx, tx, table, columns, conflictColumn, rows)
	if err != nil {
		return 0, fmt.Errorf("fallback copy: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("fallback commit: %w", err)
	}
	return n, nil
}

// copyChunkTx is the shared staging-then-insert path used by both the single
// and parallel writers. Keeping it here avoids drift between implementations.
func copyChunkTx(ctx context.Context, tx pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	stagingName := fmt.Sprintf("staging_%s", table)

	createStaging := fmt.Sprintf(
		"CREATE TEMP TABLE IF NOT EXISTS %s (LIKE %s INCLUDING DEFAULTS) ON COMMIT DROP",
		stagingName, table,
	)
	if _, err := tx.Exec(ctx, createStaging); err != nil {
		return 0, fmt.Errorf("create staging %s: %w", stagingName, err)
	}
	if _, err := tx.Exec(ctx, "TRUNCATE "+stagingName); err != nil {
		return 0, fmt.Errorf("truncate staging %s: %w", stagingName, err)
	}

	copied, err := tx.CopyFrom(ctx,
		pgx.Identifier{stagingName},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, fmt.Errorf("copy into %s: %w", stagingName, err)
	}
	if copied == 0 {
		return 0, nil
	}

	colList := joinCols(columns)
	var insertSQL string
	if conflictColumn != "" {
		insertSQL = fmt.Sprintf(
			"INSERT INTO %s (%s) SELECT %s FROM %s ON CONFLICT (%s) DO NOTHING",
			table, colList, colList, stagingName, conflictColumn,
		)
	} else {
		insertSQL = fmt.Sprintf(
			"INSERT INTO %s (%s) SELECT %s FROM %s ON CONFLICT DO NOTHING",
			table, colList, colList, stagingName,
		)
	}
	tag, err := tx.Exec(ctx, insertSQL)
	if err != nil {
		return 0, fmt.Errorf("insert from staging %s -> %s: %w", stagingName, table, err)
	}
	return int(tag.RowsAffected()), nil
}

func joinCols(cols []string) string {
	s := ""
	for i, c := range cols {
		if i > 0 {
			s += ","
		}
		s += c
	}
	return s
}
