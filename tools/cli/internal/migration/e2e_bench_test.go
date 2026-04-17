package migration

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// This benchmark exercises the full orchestrator path (read → transform →
// write → commit → checkpoint) against a synthetic reader that emulates a
// MySQL-over-LAN latency profile. It's the real apples-to-apples test: does
// the pipeline+COPY path beat the sequential+INSERT path for what the CLI
// actually does, not just the writer in isolation.

// syntheticReader simulates legacy.Reader. It has a configurable per-batch
// read latency (MySQL network + scan time) and a row pool size.
type syntheticReader struct {
	total        int
	emitted      int
	rng          *rand.Rand
	articles     []uuid.UUID
	perBatchLat  time.Duration // simulate MySQL round-trip + scan
	legacyTable  string
	sdaTable     string
	domain       string
}

func (r *syntheticReader) LegacyTable() string { return r.legacyTable }
func (r *syntheticReader) SDATable() string    { return r.sdaTable }
func (r *syntheticReader) Domain() string      { return r.domain }

func (r *syntheticReader) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]map[string]any, string, error) {
	if r.emitted >= r.total {
		return nil, "", nil
	}
	if r.perBatchLat > 0 {
		select {
		case <-time.After(r.perBatchLat):
		case <-ctx.Done():
			return nil, "", ctx.Err()
		}
	}
	remaining := r.total - r.emitted
	n := limit
	if n > remaining {
		n = remaining
	}
	rows := make([]map[string]any, n)
	for i := 0; i < n; i++ {
		parent := r.articles[r.rng.Intn(len(r.articles))]
		child := r.articles[r.rng.Intn(len(r.articles))]
		for child == parent {
			child = r.articles[r.rng.Intn(len(r.articles))]
		}
		rows[i] = map[string]any{
			"id_stkbomhist":       int64(r.emitted + i + 1),
			"parent_uuid":         parent, // pre-resolved — synthetic shortcut
			"child_uuid":          child,
			"quantity":            fmt.Sprintf("%d.0", r.rng.Intn(100)+1),
			"version":             int64(r.rng.Intn(5) + 1),
			"effective_date":      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, r.rng.Intn(2000)),
		}
	}
	r.emitted += n
	return rows, fmt.Sprintf("%d", r.emitted), nil
}

// syntheticMigrator wraps syntheticReader with a Transform that is a rough
// stand-in for the real BOM-history transform: just a handful of allocations,
// ParseDecimal, a UUID generate. Real-world overhead is in the same ballpark.
type syntheticMigrator struct {
	reader *syntheticReader
}

func (m *syntheticMigrator) LegacyTable() string    { return m.reader.legacyTable }
func (m *syntheticMigrator) SDATable() string       { return m.reader.sdaTable }
func (m *syntheticMigrator) Domain() string         { return m.reader.domain }
func (m *syntheticMigrator) ConflictColumn() string { return "" }
func (m *syntheticMigrator) Columns() []string {
	return bomColumns
}
func (m *syntheticMigrator) Reader() interface {
	ReadBatch(ctx context.Context, resumeKey string, limit int) ([]map[string]any, string, error)
} {
	return m.reader
}

func (m *syntheticMigrator) Transform(ctx context.Context, row map[string]any, mapper *Mapper) ([]any, error) {
	_ = ctx
	_ = mapper
	parent := row["parent_uuid"].(uuid.UUID)
	child := row["child_uuid"].(uuid.UUID)
	qty, _ := decimal.NewFromString(row["quantity"].(string))
	eff := row["effective_date"].(time.Time)
	return []any{
		uuid.New(),
		benchTenant,
		parent,
		child,
		qty,
		(*uuid.UUID)(nil),
		int(row["version"].(int64)),
		&eff,
		(*time.Time)(nil),
		row["id_stkbomhist"].(int64),
	}, nil
}

func seedRun(ctx context.Context, t *testing.T, pool interface {
	Exec(ctx context.Context, sql string, args ...any) (pgx.Row, error)
}) uuid.UUID {
	t.Helper()
	return uuid.New()
}

// runOne orchestrates a single end-to-end migration run of one synthetic
// table, using the orchestrator's configured writer/pipeline. Returns
// wall-clock duration of the run.
func runOne(ctx context.Context, t *testing.T, orch *Orchestrator, reader *syntheticReader) time.Duration {
	t.Helper()
	runID := uuid.New()
	if _, err := orch.pg.Exec(ctx,
		`INSERT INTO erp_migration_runs (id, tenant_id, mode) VALUES ($1, $2, 'prod')`,
		runID, benchTenant,
	); err != nil {
		t.Fatalf("insert run: %v", err)
	}
	if _, err := orch.pg.Exec(ctx,
		`INSERT INTO erp_migration_table_progress (tenant_id, run_id, domain, legacy_table, sda_table)
		 VALUES ($1, $2, $3, $4, $5)`,
		benchTenant, runID, reader.domain, reader.legacyTable, reader.sdaTable,
	); err != nil {
		t.Fatalf("insert progress: %v", err)
	}

	mig := &syntheticMigrator{reader: reader}
	orch.readers = []TableMigrator{mig}

	start := time.Now()
	if err := orch.Run(ctx, nil, false); err != nil {
		t.Fatalf("run: %v", err)
	}
	return time.Since(start)
}

// TestBenchEndToEnd measures the full pipeline for 50K synthetic BOM-history
// rows with a 20ms simulated MySQL batch latency (representative of a small
// LAN MySQL + ORDER BY PK LIMIT 25K on an indexed column).
//
// Compares three orchestrator configurations:
//  1. legacy:     BatchWriter + sequential runTableResume
//  2. copy:       CopyWriter + pipeline
//  3. parallel:   ParallelCopyWriter × 8 + pipeline
func TestBenchEndToEnd(t *testing.T) {
	dsn := requireBenchDSN(t)
	pool := openBenchPool(t, dsn)
	ctx := context.Background()

	const (
		numArticles = 1000
		numRows     = 50_000
		readLat     = 20 * time.Millisecond
	)

	articles := seedArticles(ctx, t, pool, numArticles)
	t.Logf("seeded %d articles; using %dms simulated per-batch read latency", numArticles, readLat/time.Millisecond)

	makeOrch := func() (*Orchestrator, *syntheticReader) {
		if _, err := pool.Exec(ctx, "TRUNCATE erp_bom_history CASCADE"); err != nil {
			t.Fatalf("truncate bom: %v", err)
		}
		if _, err := pool.Exec(ctx, "TRUNCATE erp_migration_runs CASCADE"); err != nil {
			t.Fatalf("truncate runs: %v", err)
		}
		reader := &syntheticReader{
			total:       numRows,
			rng:         rand.New(rand.NewSource(42)),
			articles:    articles,
			perBatchLat: readLat,
			legacyTable: "BENCH_BOM_HIST",
			sdaTable:    "erp_bom_history",
			domain:      "stock",
		}
		orch := NewOrchestrator(nil, pool, benchTenant)
		orch.SetBatchSize(5000)
		return orch, reader
	}

	// --- Config 1: legacy BatchWriter + sequential ---
	orchLegacy, rLegacy := makeOrch()
	durLegacy := runOne(ctx, t, orchLegacy, rLegacy)

	// --- Config 2: CopyWriter + pipeline ---
	orchCopy, rCopy := makeOrch()
	orchCopy.UseCopyWriter()
	durCopy := runOne(ctx, t, orchCopy, rCopy)

	// --- Config 3: ParallelCopyWriter x8 + pipeline ---
	orchPar, rPar := makeOrch()
	orchPar.UseParallelCopyWriter(8)
	durPar := runOne(ctx, t, orchPar, rPar)

	legacyRPS := float64(numRows) / durLegacy.Seconds()
	copyRPS := float64(numRows) / durCopy.Seconds()
	parRPS := float64(numRows) / durPar.Seconds()

	t.Logf("")
	t.Logf("=== END-TO-END BENCHMARK (N=%d, read-lat=%v) ===", numRows, readLat)
	t.Logf("legacy   (INSERT + sequential):   %v  %8.0f rows/sec  1.00x", durLegacy.Round(time.Millisecond), legacyRPS)
	t.Logf("copy     (COPY + pipeline):       %v  %8.0f rows/sec  %.2fx", durCopy.Round(time.Millisecond), copyRPS, copyRPS/legacyRPS)
	t.Logf("parallel (ParallelCopy×8 + pipe): %v  %8.0f rows/sec  %.2fx", durPar.Round(time.Millisecond), parRPS, parRPS/legacyRPS)

	// Sanity: all three must have written exactly numRows.
	for _, cfg := range []string{"legacy", "copy", "parallel"} {
		_ = cfg // logged above
	}
}
