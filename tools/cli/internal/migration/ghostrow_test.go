package migration

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TestWriteBatchReportsDedupedRows reproduces the Phase 0 ghost-row bug and
// demonstrates the fix's accounting.
//
// The bug: pipelines use INSERT ... ON CONFLICT DO NOTHING for idempotency.
// tag.RowsAffected() only counts newly-inserted rows — conflict-dedup'd rows
// are invisible to the writer. stats.RowsWritten += written then loses them,
// and `rows_read = rows_written + rows_skipped` stops holding. Prod saldivia
// had 214K such rows across 13 migrators (FACREMIT alone: 119K).
//
// The fix: pipeline now computes `duplicate = len(rows) - written` and tracks
// it as stats.RowsDuplicate so `rows_read = written + skipped + duplicate`
// holds. This test verifies the writer's `written` return is the actually-
// inserted count (not the attempted count), which is what the pipeline relies
// on to compute the duplicate delta.
func TestWriteBatchReportsDedupedRows(t *testing.T) {
	ctx, dsn := mustHavePG(t)
	pool := openBenchPool(t, dsn)
	tenantID := fmt.Sprintf("test-ghost-%s", uuid.NewString()[:8])
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM erp_articles WHERE tenant_id = $1`, tenantID)
	})

	w := NewBatchWriter(pool, tenantID)

	// Three rows, two with same code → one conflict on (tenant_id, code).
	id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()
	rows := [][]any{
		{id1, tenantID, "GHOST-001", "first",
			"material", decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, true},
		{id2, tenantID, "GHOST-002", "second",
			"material", decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, true},
		// third uses same code as first → INSERT ... ON CONFLICT drops it
		{id3, tenantID, "GHOST-001", "duplicate",
			"material", decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, true},
	}
	cols := []string{"id", "tenant_id", "code", "name", "article_type",
		"min_stock", "max_stock", "reorder_point", "last_cost", "avg_cost", "active"}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		t.Fatalf("set tenant: %v", err)
	}

	written, err := w.WriteBatch(ctx, tx, "erp_articles", cols, "", rows)
	if err != nil {
		t.Fatalf("write batch: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if written != 2 {
		t.Errorf("written = %d, want 2 (the third row collides on (tenant_id, code) and should not count)", written)
	}
	if dup := len(rows) - written; dup != 1 {
		t.Errorf("duplicate = %d (len(rows)=%d - written=%d), want 1", dup, len(rows), written)
	}

	// Confirm Phase 0 invariant by construction: read = written + skipped + duplicate.
	// Here skipped=0, read=3, written=2, duplicate=1. 3 = 2 + 0 + 1. ✓
	// Without this counter the pipeline would have reported (read=3, written=2,
	// skipped=0) — a ghost row, exactly the bug this test guards against.
}
