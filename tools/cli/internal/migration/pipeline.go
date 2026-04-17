package migration

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Writer abstracts how a migrator's output rows reach PostgreSQL so we can
// swap the legacy multi-value INSERT path for COPY or ParallelCopy without
// touching orchestrator/migrator code.
type Writer interface {
	// WriteBatch inserts rows into `table` idempotently. Returns the number
	// of rows actually persisted (after ON CONFLICT dedup).
	// `q` is a pgx.Tx when the caller owns the transaction (sequential mode),
	// or nil when the writer manages its own per-chunk txs (parallel mode).
	WriteBatch(ctx context.Context, q pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error)
}

// batchWriterAdapter makes the legacy BatchWriter implement Writer without
// rewriting its callsites.
type batchWriterAdapter struct{ *BatchWriter }

func (a batchWriterAdapter) WriteBatch(ctx context.Context, q pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	return a.BatchWriter.WriteBatch(ctx, q, table, columns, conflictColumn, rows)
}

// AsWriter wraps a *BatchWriter as Writer. Kept here so callers don't have to
// know about the adapter type.
func AsWriter(w *BatchWriter) Writer { return batchWriterAdapter{BatchWriter: w} }

// copyWriterAdapter makes CopyWriter implement Writer.
type copyWriterAdapter struct{ *CopyWriter }

func (a copyWriterAdapter) WriteBatch(ctx context.Context, q pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	return a.CopyWriter.WriteBatch(ctx, q, table, columns, conflictColumn, rows)
}

// AsCopyWriter wraps a *CopyWriter as Writer.
func AsCopyWriter(w *CopyWriter) Writer { return copyWriterAdapter{CopyWriter: w} }

// parallelWriterAdapter adapts ParallelCopyWriter. Its WriteBatch ignores the
// passed-in tx and owns its own worker-tx set (one per chunk).
type parallelWriterAdapter struct{ *ParallelCopyWriter }

func (a parallelWriterAdapter) WriteBatch(ctx context.Context, _ pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	return a.ParallelCopyWriter.WriteBatch(ctx, table, columns, conflictColumn, rows)
}

// AsParallelWriter wraps a *ParallelCopyWriter as Writer.
func AsParallelWriter(w *ParallelCopyWriter) Writer { return parallelWriterAdapter{ParallelCopyWriter: w} }

// pipelineBatch carries a single transformed batch from the reader goroutine
// to the writer goroutine.
type pipelineBatch struct {
	rows       [][]any          // transformed, ready to INSERT/COPY
	readN      int              // number of legacy rows consumed to produce this batch
	skipped    int              // skipped during transform
	skippedRow []map[string]any // raw rows the transform rejected, kept when archiveSkipped
	nextKey    string           // resume checkpoint after this batch
}

// pipelineResult summarises everything a migrator produced.
type pipelineResult struct {
	stats    TableStats
	lastKey  string
	duration time.Duration
}

// runTablePipeline replaces the old sequential read → transform → write loop
// with a two-stage pipeline: a reader goroutine fills a small channel with
// pre-transformed batches while a writer goroutine drains it. When the read
// side is I/O-bound (MySQL over the LAN) and the write side is CPU/network-
// bound (PG COPY), the two stages overlap almost perfectly.
//
// Checkpointing: the writer owns the checkpoint update. The reader advances
// its cursor eagerly to keep the channel full, but the `last_legacy_key` we
// persist in erp_migration_table_progress only moves after the batch is
// committed on the writer side. Crash + resume therefore cannot lose rows.
//
// Error fan-in: the first error from either side shuts the pipeline down and
// is returned to the caller. We use a context with cancellation so a write
// failure stops the reader mid-batch.
//
// Buffered channel size: 2. One batch in flight is being written, one is
// transformed and waiting. More than that pins memory for no throughput
// gain (PG COPY + fsync is the bottleneck, not the read).
func (o *Orchestrator) runTablePipeline(ctx context.Context, runID, progressID uuid.UUID, m TableMigrator, resumeKey string, readSoFar, writtenSoFar, skippedSoFar int, writer Writer) (TableStats, error) {
	slog.Info("migrating table (pipeline)", "legacy", m.LegacyTable(), "sda", m.SDATable(), "resume_key", resumeKey)
	start := time.Now()

	// Mark in_progress outside the pipeline so we do not race with the writer.
	if _, err := o.pg.Exec(ctx,
		`UPDATE erp_migration_table_progress SET status = 'in_progress', started_at = COALESCE(started_at, now()) WHERE id = $1`,
		progressID,
	); err != nil {
		slog.Warn("mark in_progress failed", "progress_id", progressID, "err", err)
	}

	stats := TableStats{RowsRead: readSoFar, RowsWritten: writtenSoFar, RowsSkipped: skippedSoFar}
	effectiveBatch := batchSizeFor(o.batchSize, len(m.Columns()))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Buffered: 8 slots so the MySQL reader can stay ahead of the PG writer
	// when they run at similar speeds (the 2-slot default was leaving 95%
	// CPU idle on the workstation). Memory cost is bounded by batch size.
	batches := make(chan pipelineBatch, 8)

	// Reader goroutine: pulls batches from legacy, runs Transform, publishes
	// to the channel. Owns the resumeKey cursor but never persists it.
	var readErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(batches)

		reader := m.Reader()
		cursor := resumeKey

		for {
			if ctx.Err() != nil {
				readErr = ctx.Err()
				return
			}

			rows, nextKey, err := reader.ReadBatch(ctx, cursor, effectiveBatch)
			if err != nil {
				readErr = fmt.Errorf("read %s: %w", m.LegacyTable(), err)
				return
			}
			if len(rows) == 0 {
				return // EOF
			}

			// Fork-join Transform: the read side of the pipeline was the
			// bottleneck — single-goroutine Transform left 95% CPU idle on
			// the workstation with 24 writer workers. Now we fan Transform
			// out across `o.transformWorkers` cores, preserving per-row
			// order via fixed-index slices so checkpoints stay correct.
			transformedByIdx := make([][]any, len(rows))
			skippedFlag := make([]bool, len(rows))
			var firstTransformErr error
			var transformErrMu sync.Mutex

			tw := o.transformWorkers
			if tw <= 1 {
				tw = 1
			}
			if tw > len(rows) {
				tw = len(rows)
			}
			if tw <= 1 {
				// Small batch — stay sequential, no point paying goroutine cost.
				for i, row := range rows {
					vals, err := m.Transform(ctx, row, o.mapper)
					if err != nil {
						firstTransformErr = fmt.Errorf("transform %s: %w", m.LegacyTable(), err)
						break
					}
					if vals == nil {
						skippedFlag[i] = true
					} else {
						transformedByIdx[i] = vals
					}
				}
			} else {
				var twg sync.WaitGroup
				chunk := (len(rows) + tw - 1) / tw
				for w := 0; w < tw; w++ {
					start := w * chunk
					end := start + chunk
					if end > len(rows) {
						end = len(rows)
					}
					if start >= end {
						continue
					}
					twg.Add(1)
					go func(start, end int) {
						defer twg.Done()
						for i := start; i < end; i++ {
							vals, err := m.Transform(ctx, rows[i], o.mapper)
							if err != nil {
								transformErrMu.Lock()
								if firstTransformErr == nil {
									firstTransformErr = fmt.Errorf("transform %s: %w", m.LegacyTable(), err)
								}
								transformErrMu.Unlock()
								return
							}
							if vals == nil {
								skippedFlag[i] = true
							} else {
								transformedByIdx[i] = vals
							}
						}
					}(start, end)
				}
				twg.Wait()
			}
			if firstTransformErr != nil {
				readErr = firstTransformErr
				return
			}

			skipped := 0
			transformed := make([][]any, 0, len(rows))
			var skippedRows []map[string]any
			if o.archiveSkipped {
				skippedRows = make([]map[string]any, 0, len(rows)/10)
			}
			for i := range rows {
				if skippedFlag[i] {
					skipped++
					if o.archiveSkipped {
						skippedRows = append(skippedRows, rows[i])
					}
					continue
				}
				if transformedByIdx[i] != nil {
					transformed = append(transformed, transformedByIdx[i])
				}
			}

			select {
			case batches <- pipelineBatch{
				rows:       transformed,
				readN:      len(rows),
				skipped:    skipped,
				skippedRow: skippedRows,
				nextKey:    nextKey,
			}:
			case <-ctx.Done():
				readErr = ctx.Err()
				return
			}
			cursor = nextKey
		}
	}()

	// Writer loop (this goroutine): drain the channel, one batch per tx.
	for b := range batches {
		stats.RowsRead += b.readN
		stats.RowsSkipped += b.skipped

		if len(b.rows) == 0 {
			// Nothing to write, but we still advance the checkpoint.
			if err := updateCheckpoint(ctx, o.pg, progressID, b.nextKey, stats); err != nil {
				cancel()
				<-batchesDrain(batches)
				wg.Wait()
				return stats, err
			}
			continue
		}

		tx, err := o.pg.Begin(ctx)
		if err != nil {
			cancel()
			wg.Wait()
			return stats, fmt.Errorf("begin tx: %w", err)
		}

		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", o.tenantID); err != nil {
			_ = tx.Rollback(ctx)
			cancel()
			wg.Wait()
			return stats, fmt.Errorf("set tenant: %w", err)
		}
		if _, err := tx.Exec(ctx, "SET LOCAL session_replication_role = 'replica'"); err != nil {
			slog.Warn("could not disable FK checks in tx", "err", err)
		}

		if err := o.mapper.FlushPending(ctx, tx); err != nil {
			_ = tx.Rollback(ctx)
			cancel()
			wg.Wait()
			return stats, err
		}

		written, err := writer.WriteBatch(ctx, tx, m.SDATable(), m.Columns(), m.ConflictColumn(), b.rows)
		if err != nil {
			_ = tx.Rollback(ctx)
			cancel()
			wg.Wait()
			markTableFailed(ctx, o.pg, progressID, err.Error())
			return stats, err
		}
		stats.RowsWritten += written

		// Archive anything the transform decided to drop so no row is ever
		// silently lost. Runs in the same tx as the batch — if the write
		// succeeds, the archive lands with it; if it fails, the rollback
		// covers both paths.
		if o.archiveSkipped && len(b.skippedRow) > 0 {
			if err := archiveSkippedRows(ctx, tx, o.tenantID, m.LegacyTable(), b.skippedRow); err != nil {
				_ = tx.Rollback(ctx)
				cancel()
				wg.Wait()
				markTableFailed(ctx, o.pg, progressID, err.Error())
				return stats, fmt.Errorf("archive skipped: %w", err)
			}
		}

		if _, err := tx.Exec(ctx,
			`UPDATE erp_migration_table_progress
			 SET last_legacy_key = $1, rows_read = $2, rows_written = $3, rows_skipped = $4
			 WHERE id = $5`,
			b.nextKey, stats.RowsRead, stats.RowsWritten, stats.RowsSkipped, progressID,
		); err != nil {
			slog.Error("failed to update progress checkpoint", "err", err)
		}

		if err := tx.Commit(ctx); err != nil {
			cancel()
			wg.Wait()
			markTableFailed(ctx, o.pg, progressID, err.Error())
			return stats, fmt.Errorf("commit batch: %w", err)
		}

		elapsed := time.Since(start)
		rowsPerSec := float64(stats.RowsRead) / elapsed.Seconds()
		slog.Info("batch progress (pipeline)",
			"table", m.LegacyTable(),
			"read", stats.RowsRead,
			"written", stats.RowsWritten,
			"skipped", stats.RowsSkipped,
			"rows/sec", int(rowsPerSec),
			"elapsed", elapsed.Round(time.Millisecond),
		)
	}

	wg.Wait()
	if readErr != nil {
		markTableFailed(ctx, o.pg, progressID, readErr.Error())
		return stats, readErr
	}

	if _, err := o.pg.Exec(ctx,
		`UPDATE erp_migration_table_progress
		 SET status = 'completed', completed_at = now(),
		     rows_read = $1, rows_written = $2, rows_skipped = $3
		 WHERE id = $4`,
		stats.RowsRead, stats.RowsWritten, stats.RowsSkipped, progressID,
	); err != nil {
		slog.Warn("mark completed failed", "progress_id", progressID, "err", err)
	}

	slog.Info("table completed (pipeline)",
		"table", m.LegacyTable(),
		"read", stats.RowsRead,
		"written", stats.RowsWritten,
		"skipped", stats.RowsSkipped,
		"duration", time.Since(start).Round(time.Millisecond),
	)
	return stats, nil
}

// updateCheckpoint persists the pipeline cursor when a batch yielded no
// transformable rows (all skipped). Uses the pool directly — cheap and does
// not race with the next tx because the writer loop is single-threaded.
func updateCheckpoint(ctx context.Context, pg *pgxpool.Pool, progressID uuid.UUID, nextKey string, s TableStats) error {
	_, err := pg.Exec(ctx,
		`UPDATE erp_migration_table_progress
		 SET last_legacy_key = $1, rows_read = $2, rows_written = $3, rows_skipped = $4
		 WHERE id = $5`,
		nextKey, s.RowsRead, s.RowsWritten, s.RowsSkipped, progressID,
	)
	return err
}

// batchesDrain is a helper to drain a channel the reader is still writing
// to, so the reader goroutine can exit cleanly after a writer error.
func batchesDrain(ch <-chan pipelineBatch) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range ch {
		}
	}()
	return done
}
