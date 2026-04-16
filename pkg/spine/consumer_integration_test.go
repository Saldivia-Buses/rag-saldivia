//go:build integration

// Integration tests for the spine consumer framework — require Docker
// (testcontainers spins up postgres). Run: go test -tags=integration ./spine/

package spine_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Camionerou/rag-saldivia/pkg/spine"
)

func setupPostgresWithProcessedEvents(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("spine_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, `
		CREATE TABLE processed_events (
			event_id      uuid NOT NULL,
			consumer_name text NOT NULL,
			processed_at  timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (event_id, consumer_name)
		);
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	return pool
}

func TestEnsureFirstDelivery_FirstTime(t *testing.T) {
	pool := setupPostgresWithProcessedEvents(t)
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	first, err := spine.EnsureFirstDelivery(ctx, tx, "0194f010-1234-7abc-8def-000000000001", "test-consumer")
	if err != nil {
		t.Fatalf("EnsureFirstDelivery: %v", err)
	}
	if !first {
		t.Error("expected firstTime=true on initial call")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func TestEnsureFirstDelivery_DuplicateInSameConsumer(t *testing.T) {
	pool := setupPostgresWithProcessedEvents(t)
	ctx := context.Background()
	eventID := "0194f010-1234-7abc-8def-000000000002"

	// First delivery: insert + commit.
	insertOnce(t, pool, eventID, "test-consumer", true)

	// Second delivery: should return firstTime=false.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	first, err := spine.EnsureFirstDelivery(ctx, tx, eventID, "test-consumer")
	if err != nil {
		t.Fatalf("EnsureFirstDelivery: %v", err)
	}
	if first {
		t.Error("expected firstTime=false on duplicate")
	}
}

func TestEnsureFirstDelivery_SameEventDifferentConsumers(t *testing.T) {
	pool := setupPostgresWithProcessedEvents(t)
	eventID := "0194f010-1234-7abc-8def-000000000003"

	insertOnce(t, pool, eventID, "consumer-a", true)
	insertOnce(t, pool, eventID, "consumer-b", true) // independent — should also be firstTime=true
}

func TestEnsureFirstDelivery_RollbackKeepsTableEmpty(t *testing.T) {
	pool := setupPostgresWithProcessedEvents(t)
	ctx := context.Background()
	eventID := "0194f010-1234-7abc-8def-000000000004"

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	first, err := spine.EnsureFirstDelivery(ctx, tx, eventID, "test-consumer")
	if err != nil {
		t.Fatalf("EnsureFirstDelivery: %v", err)
	}
	if !first {
		t.Fatal("expected firstTime=true")
	}
	// Simulate handler error → rollback.
	_ = tx.Rollback(ctx)

	// Retry in a new tx — should still be firstTime=true because rollback wiped the row.
	tx2, _ := pool.Begin(ctx)
	defer func() { _ = tx2.Rollback(ctx) }()
	first, err = spine.EnsureFirstDelivery(ctx, tx2, eventID, "test-consumer")
	if err != nil {
		t.Fatalf("EnsureFirstDelivery (retry): %v", err)
	}
	if !first {
		t.Error("expected firstTime=true after rollback — handler retry should re-execute")
	}
}

func insertOnce(t *testing.T, pool *pgxpool.Pool, eventID, consumer string, expectFirst bool) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() {
		if t.Failed() {
			_ = tx.Rollback(ctx)
		}
	}()
	first, err := spine.EnsureFirstDelivery(ctx, tx, eventID, consumer)
	if err != nil {
		t.Fatalf("EnsureFirstDelivery: %v", err)
	}
	if first != expectFirst {
		t.Errorf("got firstTime=%v, want %v (consumer=%s)", first, expectFirst, consumer)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

// TestBackoff_ProductionDefaults sanity-checks the 1s..60s defaults.
func TestBackoff_ProductionDefaults(t *testing.T) {
	if d := spine.Backoff(7, 1*time.Second, 60*time.Second); d != 60*time.Second {
		t.Errorf("attempt 7 should cap at 60s, got %v", d)
	}
}
