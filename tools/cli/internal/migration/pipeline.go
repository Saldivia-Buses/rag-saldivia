package migration

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/tools/cli/internal/legacy"
)

// Writer abstracts how a migrator's output rows reach PostgreSQL so we can
// swap the legacy multi-value INSERT path for COPY or ParallelCopy without
// touching orchestrator/migrator code.
type Writer interface {
	// WriteBatch inserts rows into `table` idempotently. Returns the number
	// of rows actually persisted (after ON CONFLICT dedup).
	WriteBatch(ctx context.Context, q pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error)
}

// batchWriterAdapter makes the legacy BatchWriter implement Writer without
// rewriting its callsites.
type batchWriterAdapter struct{ *BatchWriter }

func (a batchWriterAdapter) WriteBatch(ctx context.Context, q pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	return a.BatchWriter.WriteBatch(ctx, q, table, columns, conflictColumn, rows)
}

// AsWriter wraps a *BatchWriter as Writer.
func AsWriter(w *BatchWriter) Writer { return batchWriterAdapter{BatchWriter: w} }

type copyWriterAdapter struct{ *CopyWriter }

func (a copyWriterAdapter) WriteBatch(ctx context.Context, q pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	return a.CopyWriter.WriteBatch(ctx, q, table, columns, conflictColumn, rows)
}

// AsCopyWriter wraps a *CopyWriter as Writer.
func AsCopyWriter(w *CopyWriter) Writer { return copyWriterAdapter{CopyWriter: w} }

type parallelWriterAdapter struct{ *ParallelCopyWriter }

func (a parallelWriterAdapter) WriteBatch(ctx context.Context, _ pgx.Tx, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	return a.ParallelCopyWriter.WriteBatch(ctx, table, columns, conflictColumn, rows)
}

// AsParallelWriter wraps a *ParallelCopyWriter as Writer.
func AsParallelWriter(w *ParallelCopyWriter) Writer { return parallelWriterAdapter{ParallelCopyWriter: w} }

// pipelineBatch carries a single transformed batch from reader goroutines to
// the writer goroutine.
type pipelineBatch struct {
	rows       [][]any
	readN      int
	skipped    int
	skippedRow []map[string]any
	nextKey    string // last-PK hint — only meaningful when single reader
	fragIdx    int    // 0-based fragment id, -1 when single reader
}

// runTablePipeline streams legacy rows through the pipeline: N reader
// goroutines (one per PK fragment when readWorkers>1 and the migrator uses a
// GenericReader, else one) feed a channel of transformed batches; the writer
// loop drains the channel, one tx per batch.
//
// Checkpointing: when readers >1, last_legacy_key is meaningless (fragments
// advance independently) so we only persist rows_* counters mid-run. Single-
// reader runs keep the old resumeKey semantics.
func (o *Orchestrator) runTablePipeline(ctx context.Context, runID, progressID uuid.UUID, m TableMigrator, resumeKey string, readSoFar, writtenSoFar, skippedSoFar int, writer Writer) (TableStats, error) {
	slog.Info("migrating table (pipeline)", "legacy", m.LegacyTable(), "sda", m.SDATable(), "resume_key", resumeKey)
	start := time.Now()

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

	chanSize := 8
	if o.readWorkers > 1 {
		chanSize = 8 * o.readWorkers
	}
	batches := make(chan pipelineBatch, chanSize)

	var readErr error
	var readErrMu sync.Mutex
	setReadErr := func(err error) {
		readErrMu.Lock()
		if readErr == nil {
			readErr = err
		}
		readErrMu.Unlock()
	}

	// Detect multi-reader eligibility: readWorkers>1 + no resume + reader is
	// a GenericReader under the readerAdapter wrapper.
	multiReader := false
	var genericReader *legacy.GenericReader
	var pkFragments [][2]int64

	if o.readWorkers > 1 && resumeKey == "" {
		if adapter, ok := m.Reader().(*readerAdapter); ok {
			if gr, ok := adapter.r.(*legacy.GenericReader); ok {
				minPK, maxPK, err := gr.PKRange(ctx)
				if err != nil {
					slog.Warn("pk range scan failed, falling back to single reader", "table", m.LegacyTable(), "err", err)
				} else if maxPK > minPK && maxPK-minPK > int64(effectiveBatch)*int64(o.readWorkers) {
					genericReader = gr
					step := (maxPK - minPK) / int64(o.readWorkers)
					for i := 0; i < o.readWorkers; i++ {
						// startExclusive: include minPK on fragment 0, then
						// each subsequent fragment picks up where the last
						// finished (endInclusive of previous).
						start := minPK - 1 + step*int64(i)
						end := start + step
						if i == o.readWorkers-1 {
							end = maxPK
						}
						pkFragments = append(pkFragments, [2]int64{start, end})
					}
					multiReader = true
					slog.Info("pipeline multi-reader", "table", m.LegacyTable(),
						"fragments", len(pkFragments), "min", minPK, "max", maxPK)
				}
			}
		}
	}

	// sendTransformed fans Transform across o.transformWorkers goroutines,
	// preserving per-row order via fixed-index slices, then sends the batch.
	sendTransformed := func(rows []map[string]any, nextKey string, fragIdx int) error {
		transformedByIdx := make([][]any, len(rows))
		skippedFlag := make([]bool, len(rows))
		var firstTransformErr error
		var transformErrMu sync.Mutex

		tw := o.transformWorkers
		if tw < 1 {
			tw = 1
		}
		if tw > len(rows) {
			tw = len(rows)
		}
		if tw <= 1 {
			for i := range rows {
				vals, err := m.Transform(ctx, rows[i], o.mapper)
				if err != nil {
					return fmt.Errorf("transform %s: %w", m.LegacyTable(), err)
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
				startI := w * chunk
				endI := startI + chunk
				if endI > len(rows) {
					endI = len(rows)
				}
				if startI >= endI {
					continue
				}
				twg.Add(1)
				go func(s, e int) {
					defer twg.Done()
					for i := s; i < e; i++ {
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
				}(startI, endI)
			}
			twg.Wait()
			if firstTransformErr != nil {
				return firstTransformErr
			}
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
			fragIdx:    fragIdx,
		}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	var wg sync.WaitGroup

	// Single-reader loop: uses the existing TableMigrator.Reader adapter.
	singleReaderLoop := func(cursor string) {
		defer wg.Done()
		reader := m.Reader()
		for {
			if ctx.Err() != nil {
				setReadErr(ctx.Err())
				return
			}
			rows, nextKey, err := reader.ReadBatch(ctx, cursor, effectiveBatch)
			if err != nil {
				setReadErr(fmt.Errorf("read %s: %w", m.LegacyTable(), err))
				return
			}
			if len(rows) == 0 {
				return
			}
			if err := sendTransformed(rows, nextKey, -1); err != nil {
				setReadErr(err)
				return
			}
			cursor = nextKey
		}
	}

	// Range-reader loop for multi-reader mode. Each goroutine owns a disjoint
	// PK fragment. nextKey isn't persisted (fragments progress independently),
	// but we still emit it so the writer can log progress per fragment.
	rangeReaderLoop := func(startExcl, endIncl int64, fragIdx int) {
		defer wg.Done()
		cursor := startExcl
		for {
			if ctx.Err() != nil {
				setReadErr(ctx.Err())
				return
			}
			rows, nextKeyStr, err := genericReader.ReadBatchRange(ctx, cursor, endIncl, effectiveBatch)
			if err != nil {
				setReadErr(fmt.Errorf("read %s[frag %d]: %w", m.LegacyTable(), fragIdx, err))
				return
			}
			if len(rows) == 0 {
				return
			}
			out := make([]map[string]any, len(rows))
			for i, r := range rows {
				out[i] = map[string]any(r)
			}
			if err := sendTransformed(out, nextKeyStr, fragIdx); err != nil {
				setReadErr(err)
				return
			}
			last, _ := strconv.ParseInt(nextKeyStr, 10, 64)
			if last <= cursor {
				return
			}
			cursor = last
		}
	}

	if multiReader {
		for i, frag := range pkFragments {
			wg.Add(1)
			go rangeReaderLoop(frag[0], frag[1], i)
		}
	} else {
		wg.Add(1)
		go singleReaderLoop(resumeKey)
	}
	go func() { wg.Wait(); close(batches) }()

	// Writer loop: one tx per batch.
	for b := range batches {
		stats.RowsRead += b.readN
		stats.RowsSkipped += b.skipped

		if len(b.rows) == 0 && len(b.skippedRow) == 0 {
			// Nothing to persist. Checkpoint only when single-reader.
			if !multiReader && b.fragIdx == -1 {
				if err := updateCheckpoint(ctx, o.pg, progressID, b.nextKey, stats); err != nil {
					cancel()
					wg.Wait()
					return stats, err
				}
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

		if len(b.rows) > 0 {
			written, err := writer.WriteBatch(ctx, tx, m.SDATable(), m.Columns(), m.ConflictColumn(), b.rows)
			if err != nil {
				_ = tx.Rollback(ctx)
				cancel()
				wg.Wait()
				markTableFailed(ctx, o.pg, progressID, err.Error())
				return stats, err
			}
			stats.RowsWritten += written
		}

		if o.archiveSkipped && len(b.skippedRow) > 0 {
			if err := archiveSkippedRows(ctx, tx, o.tenantID, m.LegacyTable(), b.skippedRow); err != nil {
				_ = tx.Rollback(ctx)
				cancel()
				wg.Wait()
				markTableFailed(ctx, o.pg, progressID, err.Error())
				return stats, fmt.Errorf("archive skipped: %w", err)
			}
		}

		// Only update last_legacy_key when single-reader; multi-reader
		// fragments don't have a linear key.
		if !multiReader && b.fragIdx == -1 {
			if _, err := tx.Exec(ctx,
				`UPDATE erp_migration_table_progress
				 SET last_legacy_key = $1, rows_read = $2, rows_written = $3, rows_skipped = $4
				 WHERE id = $5`,
				b.nextKey, stats.RowsRead, stats.RowsWritten, stats.RowsSkipped, progressID,
			); err != nil {
				slog.Error("failed to update progress checkpoint", "err", err)
			}
		} else {
			if _, err := tx.Exec(ctx,
				`UPDATE erp_migration_table_progress
				 SET rows_read = $1, rows_written = $2, rows_skipped = $3
				 WHERE id = $4`,
				stats.RowsRead, stats.RowsWritten, stats.RowsSkipped, progressID,
			); err != nil {
				slog.Error("failed to update progress counters", "err", err)
			}
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
