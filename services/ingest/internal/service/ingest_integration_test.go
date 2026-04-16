//go:build integration

// Integration tests for the ingest service (job tracking layer only).
// The Blueprint proxy is not tested here — that requires the Blueprint running.
// Run: go test -tags=integration -v ./internal/service/

package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("sda_test"),
		postgres.WithUsername("sda"),
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

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	// Apply base users schema + ingest schema
	migration := `
		CREATE TABLE users (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT NOT NULL DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		INSERT INTO users (id, email, name) VALUES ('u-1', 'alice@test.com', 'Alice');
		INSERT INTO users (id, email, name) VALUES ('u-2', 'bob@test.com', 'Bob');

		CREATE TABLE ingest_jobs (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			user_id TEXT NOT NULL REFERENCES users(id),
			collection TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_size BIGINT NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending'
				CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
			error TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE INDEX idx_ingest_jobs_user ON ingest_jobs (user_id, created_at DESC);
	`
	if _, err := pool.Exec(ctx, migration); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}
	return pool, cleanup
}

// seedJob inserts a job directly for testing read/delete paths.
func seedJob(t *testing.T, pool *pgxpool.Pool, userID, collection, fileName, status string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO ingest_jobs (user_id, collection, file_name, file_size, status)
		 VALUES ($1, $2, $3, 1024, $4) RETURNING id`,
		userID, collection, fileName, status,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	return id
}

func TestListJobs_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool, nil, Config{})
	ctx := context.Background()

	seedJob(t, pool, "u-1", "docs", "a.pdf", "completed")
	seedJob(t, pool, "u-1", "docs", "b.pdf", "pending")
	seedJob(t, pool, "u-2", "other", "c.pdf", "completed")

	jobs, err := svc.ListJobs(ctx, "u-1", 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs for u-1, got %d", len(jobs))
	}
}

func TestGetJob_Ownership_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool, nil, Config{})
	ctx := context.Background()

	jobID := seedJob(t, pool, "u-1", "docs", "private.pdf", "completed")

	// Owner can access
	job, err := svc.GetJob(ctx, jobID, "u-1")
	if err != nil {
		t.Fatalf("owner should access: %v", err)
	}
	if job.FileName != "private.pdf" {
		t.Errorf("expected file name 'private.pdf', got %q", job.FileName)
	}

	// Non-owner gets ErrJobNotFound
	_, err = svc.GetJob(ctx, jobID, "u-2")
	if err != ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound for non-owner, got: %v", err)
	}
}

func TestDeleteJob_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool, nil, Config{})
	ctx := context.Background()

	jobID := seedJob(t, pool, "u-1", "docs", "delete-me.pdf", "completed")

	if err := svc.DeleteJob(ctx, jobID, "u-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := svc.GetJob(ctx, jobID, "u-1")
	if err != ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound after delete, got: %v", err)
	}
}

func TestDeleteJob_NonOwner_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool, nil, Config{})
	ctx := context.Background()

	jobID := seedJob(t, pool, "u-1", "docs", "protected.pdf", "completed")

	err := svc.DeleteJob(ctx, jobID, "u-2")
	if err != ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound for non-owner delete, got: %v", err)
	}

	// Verify job still exists
	job, err := svc.GetJob(ctx, jobID, "u-1")
	if err != nil {
		t.Fatalf("job should still exist: %v", err)
	}
	if job.ID != jobID {
		t.Errorf("expected job %s, got %s", jobID, job.ID)
	}
}

func TestUpdateJobStatus_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool, nil, Config{})
	ctx := context.Background()

	jobID := seedJob(t, pool, "u-1", "docs", "status.pdf", "pending")

	// Update to processing
	if err := svc.UpdateJobStatus(ctx, jobID, "processing", nil); err != nil {
		t.Fatalf("update to processing: %v", err)
	}

	job, _ := svc.GetJob(ctx, jobID, "u-1")
	if job.Status != "processing" {
		t.Errorf("expected status 'processing', got %q", job.Status)
	}

	// Update to failed with error
	errMsg := "blueprint returned HTTP 500"
	if err := svc.UpdateJobStatus(ctx, jobID, "failed", &errMsg); err != nil {
		t.Fatalf("update to failed: %v", err)
	}

	job, _ = svc.GetJob(ctx, jobID, "u-1")
	if job.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", job.Status)
	}
	if job.Error == nil || *job.Error != errMsg {
		t.Errorf("expected error message %q, got %v", errMsg, job.Error)
	}
}

func TestListJobs_Limit_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool, nil, Config{})
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		seedJob(t, pool, "u-1", "docs", "file.pdf", "completed")
	}

	jobs, err := svc.ListJobs(ctx, "u-1", 3)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs (limit), got %d", len(jobs))
	}
}
