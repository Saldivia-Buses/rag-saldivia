package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupHook is a function called before migration starts, after run record creation.
type SetupHook func(ctx context.Context, mapper *Mapper) error

// Orchestrator coordinates the migration of legacy MySQL data to PostgreSQL.
type Orchestrator struct {
	mysql      *sql.DB
	pg         *pgxpool.Pool
	tenantID   string
	mapper     *Mapper
	writer     *BatchWriter
	readers    []TableMigrator
	setupHooks []SetupHook
}

// TableMigrator defines how to migrate a single legacy table.
type TableMigrator interface {
	// LegacyTable returns the MySQL table name.
	LegacyTable() string
	// SDATable returns the PostgreSQL table name.
	SDATable() string
	// Domain returns the migration domain.
	Domain() string
	// Columns returns the PostgreSQL column names for INSERT.
	Columns() []string
	// ConflictColumn returns the column name for ON CONFLICT.
	ConflictColumn() string
	// Transform converts a legacy row to PostgreSQL values.
	// Returns the row values, or nil to skip the row.
	Transform(ctx context.Context, row map[string]any, mapper *Mapper) ([]any, error)
	// Reader returns the legacy reader for this table.
	Reader() interface {
		ReadBatch(ctx context.Context, resumeKey string, limit int) ([]map[string]any, string, error)
	}
}

// NewOrchestrator creates a new migration orchestrator.
func NewOrchestrator(mysql *sql.DB, pg *pgxpool.Pool, tenantID string) *Orchestrator {
	return &Orchestrator{
		mysql:    mysql,
		pg:       pg,
		tenantID: tenantID,
		mapper:   NewMapper(pg, tenantID),
		writer:   NewBatchWriter(pg, tenantID),
	}
}

// RegisterMigrators adds table migrators in dependency order.
func (o *Orchestrator) RegisterMigrators(migrators ...TableMigrator) {
	o.readers = append(o.readers, migrators...)
}

// AddSetupHook adds a function to run before migration starts (for building indexes, caches).
func (o *Orchestrator) AddSetupHook(fn SetupHook) {
	o.setupHooks = append(o.setupHooks, fn)
}

// GetMapper returns the mapper (for external setup hooks).
func (o *Orchestrator) GetMapper() *Mapper {
	return o.mapper
}

// Run executes the full migration for specified domains (empty = all).
func (o *Orchestrator) Run(ctx context.Context, domains []string, dryRun bool) error {
	return o.RunWithID(ctx, uuid.Nil, domains, dryRun)
}

// RunWithID executes the migration with an optional pre-created run ID.
// If runID is Nil, a new one is created.
func (o *Orchestrator) RunWithID(ctx context.Context, externalRunID uuid.UUID, domains []string, dryRun bool) error {
	runStart := time.Now()
	mode := "prod"
	if dryRun {
		mode = "dry_run"
	}

	// Use external run ID or create new one
	runID := externalRunID
	if runID == uuid.Nil {
		runID = uuid.New()
	}
	_, err := o.pg.Exec(ctx,
		`INSERT INTO erp_migration_runs (id, tenant_id, mode) VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO UPDATE SET mode = $3, status = 'running', started_at = now()`,
		runID, o.tenantID, mode,
	)
	if err != nil {
		return fmt.Errorf("create run: %w", err)
	}

	slog.Info("migration started",
		"run_id", runID,
		"tenant", o.tenantID,
		"mode", mode,
		"migrators", len(o.readers),
	)

	// Create progress records for each table
	for _, m := range o.readers {
		if len(domains) > 0 && !contains(domains, m.Domain()) {
			continue
		}
		_, err := o.pg.Exec(ctx,
			`INSERT INTO erp_migration_table_progress
			 (tenant_id, run_id, domain, legacy_table, sda_table)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (tenant_id, run_id, legacy_table) DO NOTHING`,
			o.tenantID, runID, m.Domain(), m.LegacyTable(), m.SDATable(),
		)
		if err != nil {
			return fmt.Errorf("create progress %s: %w", m.LegacyTable(), err)
		}
	}

	// Tenant context is set per-batch transaction via set_config (see runTable)

	// Run each migrator
	stats := make(map[string]TableStats)

	if dryRun {
		// Dry-run: parallel — no writes, no FK dependencies
		// Skip PG lookups for mapper (no data in mapping table during dry-run)
		o.mapper.SetDryRun(true)
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, 8) // max 8 concurrent readers
		var firstErr error

		for _, m := range o.readers {
			if len(domains) > 0 && !contains(domains, m.Domain()) {
				continue
			}
			wg.Add(1)
			go func(m TableMigrator) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				mu.Lock()
				if firstErr != nil {
					mu.Unlock()
					return
				}
				mu.Unlock()

				s, err := o.dryRunTable(ctx, runID, m)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					if firstErr == nil {
						firstErr = fmt.Errorf("dry-run %s: %w", m.LegacyTable(), err)
					}
					return
				}
				stats[m.LegacyTable()] = s
			}(m)
		}
		wg.Wait()
		if firstErr != nil {
			o.failRun(ctx, runID, firstErr)
			return firstErr
		}
	} else {
		// Prod: sequential with FK ordering + setup hooks
		// Disable FK checks during bulk load (re-enabled after completion)
		if _, err := o.pg.Exec(ctx, "SET session_replication_role = 'replica'"); err != nil {
			slog.Warn("could not disable FK checks", "err", err)
		} else {
			defer func() { _, _ = o.pg.Exec(ctx, "SET session_replication_role = 'origin'") }()
			slog.Info("FK checks disabled for bulk load")
		}
		hooksRan := false
		for _, m := range o.readers {
			if len(domains) > 0 && !contains(domains, m.Domain()) {
				continue
			}

			if !hooksRan && needsSetupHooks(m) {
				for _, hook := range o.setupHooks {
					if err := hook(ctx, o.mapper); err != nil {
						return fmt.Errorf("setup hook: %w", err)
					}
				}
				hooksRan = true
			}

			if _, err := o.pg.Exec(ctx,
				`UPDATE erp_migration_runs SET current_domain = $1, current_table = $2 WHERE id = $3`,
				m.Domain(), m.LegacyTable(), runID,
			); err != nil {
				slog.Error("failed to update run position", "err", err)
			}

			s, err := o.runTable(ctx, runID, m)
			if err != nil {
				o.failRun(ctx, runID, err)
				return fmt.Errorf("migrate %s: %w", m.LegacyTable(), err)
			}
			stats[m.LegacyTable()] = s
		}
	}

	// Update run stats and mark completed
	statsJSON, _ := json.Marshal(stats)
	status := "completed"
	if dryRun {
		status = "dry_run_ok"
	}
	_, err = o.pg.Exec(ctx,
		`UPDATE erp_migration_runs
		 SET status = $1, completed_at = now(), stats = $2, current_domain = NULL, current_table = NULL
		 WHERE id = $3`,
		status, statsJSON, runID,
	)
	if err != nil {
		return fmt.Errorf("complete run: %w", err)
	}

	slog.Info("migration completed", "run_id", runID, "status", status)
	printReport(stats, dryRun, time.Since(runStart))
	return nil
}

// Resume continues a failed migration from where it left off.
func (o *Orchestrator) Resume(ctx context.Context, resumeRunID string) error {
	runID, err := uuid.Parse(resumeRunID)
	if err != nil {
		return fmt.Errorf("invalid run ID: %w", err)
	}

	// Load run
	var status, mode string
	err = o.pg.QueryRow(ctx,
		`SELECT status, mode FROM erp_migration_runs WHERE id = $1 AND tenant_id = $2`,
		runID, o.tenantID,
	).Scan(&status, &mode)
	if err != nil {
		return fmt.Errorf("load run %s: %w", runID, err)
	}
	if status != "failed" {
		return fmt.Errorf("run %s has status %q, only failed runs can be resumed", runID, status)
	}

	// Mark as running again
	if _, err := o.pg.Exec(ctx,
		`UPDATE erp_migration_runs SET status = 'running', error_message = NULL WHERE id = $1`,
		runID,
	); err != nil {
		slog.Warn("could not mark run as running", "err", err)
	}

	// Tenant context is set per-batch transaction via set_config (see runTable)

	// Load pending/failed tables
	rows, err := o.pg.Query(ctx,
		`SELECT id, domain, legacy_table, sda_table, last_legacy_key, rows_read, rows_written, rows_skipped
		 FROM erp_migration_table_progress
		 WHERE run_id = $1 AND status IN ('pending', 'in_progress', 'failed')
		 ORDER BY id`,
		runID,
	)
	if err != nil {
		return fmt.Errorf("list pending: %w", err)
	}
	defer rows.Close()

	type pendingTable struct {
		ID           uuid.UUID
		Domain       string
		LegacyTable  string
		SDATable     string
		LastKey      string
		RowsRead     int
		RowsWritten  int
		RowsSkipped  int
	}
	var pending []pendingTable
	for rows.Next() {
		var t pendingTable
		if err := rows.Scan(&t.ID, &t.Domain, &t.LegacyTable, &t.SDATable, &t.LastKey, &t.RowsRead, &t.RowsWritten, &t.RowsSkipped); err != nil {
			return fmt.Errorf("scan pending: %w", err)
		}
		pending = append(pending, t)
	}

	slog.Info("resuming migration", "run_id", runID, "pending_tables", len(pending))

	stats := make(map[string]TableStats)
	for _, t := range pending {
		// Find the matching migrator
		var migrator TableMigrator
		for _, m := range o.readers {
			if m.LegacyTable() == t.LegacyTable {
				migrator = m
				break
			}
		}
		if migrator == nil {
			slog.Warn("no migrator found for table, skipping", "table", t.LegacyTable)
			continue
		}

		// Preload dependent domains for FK resolution
		if err := o.mapper.PreloadDomain(ctx, t.Domain); err != nil {
			slog.Warn("could not preload domain", "domain", t.Domain, "err", err)
		}

		s, err := o.runTableResume(ctx, runID, t.ID, migrator, t.LastKey, t.RowsRead, t.RowsWritten, t.RowsSkipped)
		if err != nil {
			o.failRun(ctx, runID, err)
			return fmt.Errorf("resume %s: %w", t.LegacyTable, err)
		}
		stats[t.LegacyTable] = s
	}

	// Mark completed
	statsJSON, _ := json.Marshal(stats)
	if _, err := o.pg.Exec(ctx,
		`UPDATE erp_migration_runs
		 SET status = 'completed', completed_at = now(), stats = $1
		 WHERE id = $2`,
		statsJSON, runID,
	); err != nil {
		slog.Warn("could not mark resume run as completed", "err", err)
	}

	slog.Info("resume completed", "run_id", runID)
	printReport(stats, false, 0)
	return nil
}

// TableStats tracks per-table migration statistics.
type TableStats struct {
	RowsRead    int `json:"rows_read"`
	RowsWritten int `json:"rows_written"`
	RowsSkipped int `json:"rows_skipped"`
}

const batchSize = 5000

func (o *Orchestrator) runTable(ctx context.Context, runID uuid.UUID, m TableMigrator) (TableStats, error) {
	return o.runTableResume(ctx, runID, uuid.Nil, m, "", 0, 0, 0)
}

func (o *Orchestrator) runTableResume(ctx context.Context, runID, progressID uuid.UUID, m TableMigrator, resumeKey string, readSoFar, writtenSoFar, skippedSoFar int) (TableStats, error) {
	slog.Info("migrating table", "legacy", m.LegacyTable(), "sda", m.SDATable(), "resume_key", resumeKey)
	start := time.Now()

	// Get or create progress record
	if progressID == uuid.Nil {
		err := o.pg.QueryRow(ctx,
			`SELECT id FROM erp_migration_table_progress
			 WHERE run_id = $1 AND legacy_table = $2 AND tenant_id = $3`,
			runID, m.LegacyTable(), o.tenantID,
		).Scan(&progressID)
		if err != nil {
			return TableStats{}, fmt.Errorf("get progress id: %w", err)
		}
	}

	// Mark in_progress
	o.pg.Exec(ctx,
		`UPDATE erp_migration_table_progress SET status = 'in_progress', started_at = COALESCE(started_at, now()) WHERE id = $1`,
		progressID,
	)

	stats := TableStats{RowsRead: readSoFar, RowsWritten: writtenSoFar, RowsSkipped: skippedSoFar}
	reader := m.Reader()
	lastKey := resumeKey

	for {
		rows, nextKey, err := reader.ReadBatch(ctx, lastKey, batchSize)
		if err != nil {
			markTableFailed(ctx, o.pg, progressID, err.Error())
			return stats, err
		}
		if len(rows) == 0 {
			break
		}

		stats.RowsRead += len(rows)

		// Each batch runs in a transaction for atomicity (mapping + data + progress)
		tx, err := o.pg.Begin(ctx)
		if err != nil {
			markTableFailed(ctx, o.pg, progressID, err.Error())
			return stats, fmt.Errorf("begin tx: %w", err)
		}

		// Set tenant context for RLS on business tables (LOCAL = tx-scoped)
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", o.tenantID); err != nil {
			_ = tx.Rollback(ctx)
			slog.Warn("could not set app.tenant_id", "err", err)
		}
		// Disable FK checks within this TX for bulk load
		if _, err := tx.Exec(ctx, "SET LOCAL session_replication_role = 'replica'"); err != nil {
			slog.Warn("could not disable FK checks in tx", "err", err)
		}

		// Transform and collect batch
		var insertRows [][]any
		for _, row := range rows {
			vals, err := m.Transform(ctx, row, o.mapper)
			if err != nil {
				_ = tx.Rollback(ctx)
				markTableFailed(ctx, o.pg, progressID, err.Error())
				return stats, fmt.Errorf("transform %s row: %w", m.LegacyTable(), err)
			}
			if vals == nil {
				stats.RowsSkipped++
				continue
			}
			insertRows = append(insertRows, vals)
		}

		// Flush pending mapper inserts in batch (instead of per-row)
		if err := o.mapper.FlushPending(ctx, tx); err != nil {
			_ = tx.Rollback(ctx)
			markTableFailed(ctx, o.pg, progressID, err.Error())
			return stats, err
		}

		// Write batch within transaction (tx has FK checks disabled)
		if len(insertRows) > 0 {
			written, err := o.writer.WriteBatch(ctx, tx, m.SDATable(), m.Columns(), m.ConflictColumn(), insertRows)
			if err != nil {
				_ = tx.Rollback(ctx)
				markTableFailed(ctx, o.pg, progressID, err.Error())
				return stats, err
			}
			stats.RowsWritten += written
		}

		lastKey = nextKey

		// Update progress checkpoint within same transaction
		if _, err := tx.Exec(ctx,
			`UPDATE erp_migration_table_progress
			 SET last_legacy_key = $1, rows_read = $2, rows_written = $3, rows_skipped = $4
			 WHERE id = $5`,
			lastKey, stats.RowsRead, stats.RowsWritten, stats.RowsSkipped, progressID,
		); err != nil {
			slog.Error("failed to update progress checkpoint", "err", err)
		}

		if err := tx.Commit(ctx); err != nil {
			markTableFailed(ctx, o.pg, progressID, err.Error())
			return stats, fmt.Errorf("commit batch: %w", err)
		}

		// Log batch speed
		elapsed := time.Since(start)
		rowsPerSec := float64(stats.RowsRead) / elapsed.Seconds()
		slog.Info("batch progress",
			"table", m.LegacyTable(),
			"read", stats.RowsRead,
			"written", stats.RowsWritten,
			"skipped", stats.RowsSkipped,
			"rows/sec", int(rowsPerSec),
			"elapsed", elapsed.Round(time.Millisecond),
		)
	}

	// Mark completed
	o.pg.Exec(ctx,
		`UPDATE erp_migration_table_progress
		 SET status = 'completed', completed_at = now(),
		     rows_read = $1, rows_written = $2, rows_skipped = $3
		 WHERE id = $4`,
		stats.RowsRead, stats.RowsWritten, stats.RowsSkipped, progressID,
	)

	slog.Info("table completed",
		"table", m.LegacyTable(),
		"read", stats.RowsRead,
		"written", stats.RowsWritten,
		"skipped", stats.RowsSkipped,
		"duration", time.Since(start).Round(time.Millisecond),
	)
	return stats, nil
}

func (o *Orchestrator) dryRunTable(ctx context.Context, runID uuid.UUID, m TableMigrator) (TableStats, error) {
	slog.Info("dry-run table", "legacy", m.LegacyTable(), "sda", m.SDATable())

	var progressID uuid.UUID
	_ = o.pg.QueryRow(ctx,
		`SELECT id FROM erp_migration_table_progress
		 WHERE run_id = $1 AND legacy_table = $2 AND tenant_id = $3`,
		runID, m.LegacyTable(), o.tenantID,
	).Scan(&progressID)

	if _, err := o.pg.Exec(ctx,
		`UPDATE erp_migration_table_progress SET status = 'in_progress', started_at = now() WHERE id = $1`,
		progressID,
	); err != nil {
		slog.Warn("could not mark dry-run table as in_progress", "err", err)
	}

	stats := TableStats{}
	reader := m.Reader()
	lastKey := ""

	for {
		rows, nextKey, err := reader.ReadBatch(ctx, lastKey, batchSize)
		if err != nil {
			return stats, err
		}
		if len(rows) == 0 {
			break
		}
		stats.RowsRead += len(rows)

		for _, row := range rows {
			vals, err := m.Transform(ctx, row, o.mapper)
			if err != nil {
				return stats, fmt.Errorf("dry-run transform %s: %w", m.LegacyTable(), err)
			}
			if vals == nil {
				stats.RowsSkipped++
			} else {
				stats.RowsWritten++ // would-be-written count
			}
		}
		lastKey = nextKey
	}

	if _, err := o.pg.Exec(ctx,
		`UPDATE erp_migration_table_progress
		 SET status = 'completed', completed_at = now(),
		     rows_read = $1, rows_written = $2, rows_skipped = $3
		 WHERE id = $4`,
		stats.RowsRead, stats.RowsWritten, stats.RowsSkipped, progressID,
	); err != nil {
		slog.Warn("could not mark dry-run table as completed", "err", err)
	}

	slog.Info("dry-run table done", "table", m.LegacyTable(), "read", stats.RowsRead, "would_write", stats.RowsWritten, "skip", stats.RowsSkipped)
	return stats, nil
}

func (o *Orchestrator) failRun(ctx context.Context, runID uuid.UUID, err error) {
	status := "failed"
	o.pg.Exec(ctx,
		`UPDATE erp_migration_runs SET status = $1, error_message = $2, completed_at = now() WHERE id = $3`,
		status, err.Error(), runID,
	)
}

func markTableFailed(ctx context.Context, pg *pgxpool.Pool, progressID uuid.UUID, errMsg string) {
	pg.Exec(ctx,
		`UPDATE erp_migration_table_progress SET status = 'failed', error_message = $1, completed_at = now() WHERE id = $2`,
		errMsg, progressID,
	)
}

// FindLastDryRun finds the most recent dry_run_ok for a tenant.
func FindLastDryRun(ctx context.Context, pg *pgxpool.Pool, tenantID string) (string, error) {
	var status string
	err := pg.QueryRow(ctx,
		`SELECT status FROM erp_migration_runs
		 WHERE tenant_id = $1 AND mode = 'dry_run'
		 ORDER BY started_at DESC LIMIT 1`,
		tenantID,
	).Scan(&status)
	return status, err
}

func printReport(stats map[string]TableStats, dryRun bool, totalDuration time.Duration) {
	mode := "MIGRATION"
	if dryRun {
		mode = "DRY-RUN"
	}
	fmt.Printf("\n=== %s REPORT ===\n", mode)
	totalRead, totalWritten, totalSkipped := 0, 0, 0
	for table, s := range stats {
		label := "written"
		if dryRun {
			label = "would_write"
		}
		fmt.Printf("  %-30s read=%-8d %s=%-8d skipped=%-6d\n", table, s.RowsRead, label, s.RowsWritten, s.RowsSkipped)
		totalRead += s.RowsRead
		totalWritten += s.RowsWritten
		totalSkipped += s.RowsSkipped
	}
	label := "written"
	if dryRun {
		label = "would_write"
	}
	fmt.Printf("\n  TOTAL: read=%d %s=%d skipped=%d\n", totalRead, label, totalWritten, totalSkipped)
	fmt.Printf("  DURATION: %s\n", totalDuration.Round(time.Millisecond))
	if totalDuration.Seconds() > 0 {
		fmt.Printf("  THROUGHPUT: %.0f rows/sec read, %.0f rows/sec written\n",
			float64(totalRead)/totalDuration.Seconds(),
			float64(totalWritten)/totalDuration.Seconds())
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// needsSetupHooks returns true for tables that depend on code indexes or date caches.
func needsSetupHooks(m TableMigrator) bool {
	switch m.SDATable() {
	case "erp_journal_lines", "erp_stock_movements", "erp_purchase_order_lines", "erp_production_orders":
		return true
	}
	// FACDETAL (invoice lines) needs the regmovim index built after IVAVENTAS/IVACOMPRAS.
	if m.LegacyTable() == "FACDETAL" {
		return true
	}
	return false
}
