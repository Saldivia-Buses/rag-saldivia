package migration

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// To run: BENCH_PG_DSN=postgres://sda:sda@localhost:5432/sda_migration_bench?sslmode=disable go test -v -run TestBench -timeout 10m
//
// The benchmark targets `erp_bom_history` because it's the largest legacy
// table (3.3M rows in prod) and has a representative column mix: uuid FKs,
// numeric(14,4), date, bigint, timestamptz default. Results scale with real
// rag-saldivia workloads.

const benchTenant = "bench_tenant"

func requireBenchDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("BENCH_PG_DSN")
	if dsn == "" {
		t.Skip("BENCH_PG_DSN not set — skipping benchmark")
	}
	return dsn
}

func openBenchPool(t *testing.T, dsn string) *pgxpool.Pool {
	t.Helper()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	cfg.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func seedArticles(ctx context.Context, t *testing.T, pool *pgxpool.Pool, n int) []uuid.UUID {
	t.Helper()
	// Idempotent seed: wipe prior articles so repeated runs don't fight the
	// unique (tenant_id, code) constraint.
	if _, err := pool.Exec(ctx, "TRUNCATE erp_articles CASCADE"); err != nil {
		t.Fatalf("truncate articles: %v", err)
	}
	ids := make([]uuid.UUID, n)
	rows := make([][]any, n)
	for i := range ids {
		ids[i] = uuid.New()
		rows[i] = []any{
			ids[i], benchTenant,
			fmt.Sprintf("ART-%06d", i),
			fmt.Sprintf("Article %d", i),
			"material",
			decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero,
			true,
		}
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin seed: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", benchTenant); err != nil {
		t.Fatalf("set tenant: %v", err)
	}
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"erp_articles"},
		[]string{"id", "tenant_id", "code", "name", "article_type", "min_stock", "max_stock", "reorder_point", "last_cost", "avg_cost", "active"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		t.Fatalf("seed articles: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit seed: %v", err)
	}
	return ids
}

// genBomRows generates n synthetic erp_bom_history rows using the pool of
// article UUIDs created by seedArticles. The row shape is deterministic but
// the data values vary, so every run exercises the same write path.
func genBomRows(n int, articles []uuid.UUID, seed int64) [][]any {
	r := rand.New(rand.NewSource(seed))
	cols := 10 // id, tenant_id, parent_id, child_id, quantity, unit_id, version, effective_date, replaced_date, legacy_id
	_ = cols
	out := make([][]any, n)
	for i := 0; i < n; i++ {
		parent := articles[r.Intn(len(articles))]
		child := articles[r.Intn(len(articles))]
		for child == parent {
			child = articles[r.Intn(len(articles))]
		}
		qty := decimal.NewFromFloat(float64(r.Intn(10000)) / 100.0)
		eff := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, r.Intn(2000))
		out[i] = []any{
			uuid.New(),     // id
			benchTenant,    // tenant_id
			parent,         // parent_id
			child,          // child_id
			qty,            // quantity
			(*uuid.UUID)(nil), // unit_id (nullable)
			r.Intn(5) + 1,  // version
			&eff,           // effective_date
			(*time.Time)(nil), // replaced_date
			int64(i + 1),   // legacy_id
		}
	}
	return out
}

var bomColumns = []string{
	"id", "tenant_id", "parent_id", "child_id", "quantity",
	"unit_id", "version", "effective_date", "replaced_date", "legacy_id",
}

func truncateBom(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(ctx, "TRUNCATE erp_bom_history"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// writeAllInsert invokes the production BatchWriter in per-batch txs,
// matching how orchestrator.runTableResume actually operates.
func writeAllInsert(ctx context.Context, t *testing.T, pool *pgxpool.Pool, w *BatchWriter, rows [][]any, batchSize int) int {
	t.Helper()
	total := 0
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", benchTenant); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("set tenant: %v", err)
		}
		n, err := w.WriteBatch(ctx, tx, "erp_bom_history", bomColumns, "", rows[i:end])
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("write batch: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit: %v", err)
		}
		total += n
	}
	return total
}

func writeAllCopy(ctx context.Context, t *testing.T, pool *pgxpool.Pool, w *CopyWriter, rows [][]any, batchSize int) int {
	t.Helper()
	total := 0
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", benchTenant); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("set tenant: %v", err)
		}
		n, err := w.WriteBatch(ctx, tx, "erp_bom_history", bomColumns, "", rows[i:end])
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("copy batch: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit: %v", err)
		}
		total += n
	}
	return total
}

func writeAllParallel(ctx context.Context, t *testing.T, w *ParallelCopyWriter, rows [][]any, batchSize int) int {
	t.Helper()
	total := 0
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		n, err := w.WriteBatch(ctx, "erp_bom_history", bomColumns, "", rows[i:end])
		if err != nil {
			t.Fatalf("parallel write: %v", err)
		}
		total += n
	}
	return total
}

// writeAllCopyDirect bypasses the staging table — uses COPY straight into the
// target. Only correct when there are no pre-existing rows that might conflict.
// Measures the ceiling: what's the fastest path PG can ingest rows?
func writeAllCopyDirect(ctx context.Context, t *testing.T, pool *pgxpool.Pool, rows [][]any, batchSize int) int {
	t.Helper()
	total := 0
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", benchTenant); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("set tenant: %v", err)
		}
		n, err := tx.CopyFrom(ctx,
			pgx.Identifier{"erp_bom_history"},
			bomColumns,
			pgx.CopyFromRows(rows[i:end]),
		)
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("copy direct: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit: %v", err)
		}
		total += int(n)
	}
	return total
}

// TestBenchWriters compares BatchWriter (multi-value INSERT) vs CopyWriter
// (binary COPY FROM + staging INSERT ... SELECT) writing the same N rows
// across the same batch size. Reports rows/sec for each path and the speedup.
func TestBenchWriters(t *testing.T) {
	dsn := requireBenchDSN(t)
	pool := openBenchPool(t, dsn)
	ctx := context.Background()

	const (
		numArticles = 1000
		numRows     = 100_000
		// The multi-value INSERT path is hard-capped by PG's 65535 bind-param
		// limit. For 10 columns the orchestrator clamps to ~6400. We use the
		// same clamped size for both paths so the comparison is apples-to-apples.
		batchSize = 2000
	)

	t.Logf("seeding %d articles…", numArticles)
	articles := seedArticles(ctx, t, pool, numArticles)

	t.Logf("generating %d synthetic bom_history rows…", numRows)
	rows := genBomRows(numRows, articles, 42)

	// --- Pass 1: current BatchWriter ---
	truncateBom(ctx, t, pool)
	insertW := NewBatchWriter(pool, benchTenant)
	startInsert := time.Now()
	insertedN := writeAllInsert(ctx, t, pool, insertW, rows, batchSize)
	durInsert := time.Since(startInsert)

	// --- Pass 2: CopyWriter ---
	truncateBom(ctx, t, pool)
	copyW := NewCopyWriter(pool, benchTenant)
	startCopy := time.Now()
	copiedN := writeAllCopy(ctx, t, pool, copyW, rows, batchSize)
	durCopy := time.Since(startCopy)

	// --- Pass 3: CopyWriter with larger batch (COPY has no 65535 limit) ---
	truncateBom(ctx, t, pool)
	startCopyLarge := time.Now()
	copiedLargeN := writeAllCopy(ctx, t, pool, copyW, rows, 25_000)
	durCopyLarge := time.Since(startCopyLarge)

	// --- Pass 4: COPY direct (no staging, fastest path when table is empty) ---
	truncateBom(ctx, t, pool)
	startDirect := time.Now()
	directN := writeAllCopyDirect(ctx, t, pool, rows, 25_000)
	durDirect := time.Since(startDirect)

	// --- Pass 5: ParallelCopyWriter 4 workers, big batch ---
	truncateBom(ctx, t, pool)
	pw4 := NewParallelCopyWriter(pool, benchTenant, 4)
	startP4 := time.Now()
	parallelN4 := writeAllParallel(ctx, t, pw4, rows, 25_000)
	durP4 := time.Since(startP4)

	// --- Pass 6: ParallelCopyWriter 8 workers, big batch ---
	truncateBom(ctx, t, pool)
	pw8 := NewParallelCopyWriter(pool, benchTenant, 8)
	startP8 := time.Now()
	parallelN8 := writeAllParallel(ctx, t, pw8, rows, 25_000)
	durP8 := time.Since(startP8)

	insertRPS := float64(insertedN) / durInsert.Seconds()
	copyRPS := float64(copiedN) / durCopy.Seconds()
	copyLargeRPS := float64(copiedLargeN) / durCopyLarge.Seconds()
	directRPS := float64(directN) / durDirect.Seconds()
	p4RPS := float64(parallelN4) / durP4.Seconds()
	p8RPS := float64(parallelN8) / durP8.Seconds()

	t.Logf("")
	t.Logf("=== WRITER BENCHMARK (N=%d) ===", numRows)
	t.Logf("BatchWriter    (INSERT, batch=2K):           inserted=%d  time=%v  %8.0f rows/sec  %.2fx",
		insertedN, durInsert.Round(time.Millisecond), insertRPS, 1.0)
	t.Logf("CopyWriter     (COPY+staging, batch=2K):     inserted=%d  time=%v  %8.0f rows/sec  %.2fx",
		copiedN, durCopy.Round(time.Millisecond), copyRPS, insertRPS/copyRPS)
	t.Logf("CopyWriter     (COPY+staging, batch=25K):    inserted=%d  time=%v  %8.0f rows/sec  %.2fx",
		copiedLargeN, durCopyLarge.Round(time.Millisecond), copyLargeRPS, copyLargeRPS/insertRPS)
	t.Logf("CopyDirect     (COPY no staging, batch=25K): inserted=%d  time=%v  %8.0f rows/sec  %.2fx",
		directN, durDirect.Round(time.Millisecond), directRPS, directRPS/insertRPS)
	t.Logf("ParallelCopy×4 (COPY+staging, batch=25K):    inserted=%d  time=%v  %8.0f rows/sec  %.2fx",
		parallelN4, durP4.Round(time.Millisecond), p4RPS, p4RPS/insertRPS)
	t.Logf("ParallelCopy×8 (COPY+staging, batch=25K):    inserted=%d  time=%v  %8.0f rows/sec  %.2fx",
		parallelN8, durP8.Round(time.Millisecond), p8RPS, p8RPS/insertRPS)

	if insertedN != numRows || copiedN != numRows || copiedLargeN != numRows ||
		directN != numRows || parallelN4 != numRows || parallelN8 != numRows {
		t.Errorf("row count mismatch: insert=%d copy=%d copyLarge=%d direct=%d p4=%d p8=%d (want %d)",
			insertedN, copiedN, copiedLargeN, directN, parallelN4, parallelN8, numRows)
	}
}
