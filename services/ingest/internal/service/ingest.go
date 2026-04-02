// Package service implements the Ingest service business logic.
// The Ingest Service handles document upload and async processing via NATS
// JetStream. Documents are staged to disk, a job is created in the tenant DB,
// and a NATS message triggers the worker to forward the file to the NVIDIA
// RAG Blueprint for vectorization.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

var ErrJobNotFound = errors.New("job not found")

// EventPublisher can publish notification events. Optional.
type EventPublisher interface {
	Notify(tenantSlug string, evt any) error
}

// Config holds ingest service configuration.
type Config struct {
	BlueprintURL string        // http://localhost:8081
	StagingDir   string        // /tmp/ingest-staging
	Timeout      time.Duration // request timeout for Blueprint uploads
}

// IngestMessage is published to NATS for the worker to process.
// Shared between Submit (producer) and Worker (consumer).
type IngestMessage struct {
	JobID      string `json:"job_id"`
	TenantSlug string `json:"tenant_slug"`
	UserID     string `json:"user_id"`
	Collection string `json:"collection"`
	FileName   string `json:"file_name"`
	StagedPath string `json:"staged_path"`
}

// Ingest manages document upload and job tracking.
type Ingest struct {
	pool      *pgxpool.Pool
	nc        *nats.Conn
	publisher EventPublisher
	cfg       Config
}

// New creates an Ingest service.
func New(pool *pgxpool.Pool, nc *nats.Conn, publisher EventPublisher, cfg Config) *Ingest {
	if cfg.StagingDir == "" {
		cfg.StagingDir = "/tmp/ingest-staging"
	}
	if err := os.MkdirAll(cfg.StagingDir, 0750); err != nil {
		slog.Error("failed to create staging directory", "error", err, "path", cfg.StagingDir)
	}

	return &Ingest{
		pool:      pool,
		nc:        nc,
		publisher: publisher,
		cfg:       cfg,
	}
}

// Job represents an ingest job.
type Job struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Collection string    `json:"collection"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	Status     string    `json:"status"`
	Error      *string   `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Submit stages a document to disk, creates a pending job, and publishes
// a NATS message for async processing. Returns immediately with 202.
func (s *Ingest) Submit(ctx context.Context, tenantSlug, userID, collection, fileName string, fileSize int64, file multipart.File) (*Job, error) {
	// Stage file to disk
	stagePath := filepath.Join(s.cfg.StagingDir, tenantSlug)
	if err := os.MkdirAll(stagePath, 0750); err != nil {
		return nil, fmt.Errorf("create staging dir: %w", err)
	}

	tmpFile, err := os.CreateTemp(stagePath, "ingest-*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	stagedPath := tmpFile.Name()

	written, err := io.Copy(tmpFile, file)
	tmpFile.Close()
	if err != nil {
		os.Remove(stagedPath)
		return nil, fmt.Errorf("stage file: %w", err)
	}
	if written == 0 {
		os.Remove(stagedPath)
		return nil, fmt.Errorf("empty file")
	}

	// Create job in DB
	var job Job
	err = s.pool.QueryRow(ctx,
		`INSERT INTO ingest_jobs (user_id, collection, file_name, file_size, status)
		 VALUES ($1, $2, $3, $4, 'pending')
		 RETURNING id, user_id, collection, file_name, file_size, status, error, created_at, updated_at`,
		userID, collection, fileName, fileSize,
	).Scan(&job.ID, &job.UserID, &job.Collection, &job.FileName, &job.FileSize,
		&job.Status, &job.Error, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		os.Remove(stagedPath)
		return nil, fmt.Errorf("create ingest job: %w", err)
	}

	// Publish to NATS for async processing (B3 fix: json.Marshal instead of Sprintf)
	subject := "tenant." + tenantSlug + ".ingest.process"
	payload, err := json.Marshal(IngestMessage{
		JobID:      job.ID,
		TenantSlug: tenantSlug,
		UserID:     userID,
		Collection: collection,
		FileName:   fileName,
		StagedPath: stagedPath,
	})
	if err != nil {
		os.Remove(stagedPath)
		return nil, fmt.Errorf("marshal ingest message: %w", err)
	}

	// S7 fix: fail the upload if NATS publish fails — prevents orphaned pending jobs
	if err := s.nc.Publish(subject, payload); err != nil {
		os.Remove(stagedPath)
		s.pool.Exec(ctx, `DELETE FROM ingest_jobs WHERE id = $1`, job.ID)
		return nil, fmt.Errorf("publish ingest job: %w", err)
	}

	return &job, nil
}

// ListJobs returns all jobs for a user, newest first.
func (s *Ingest) ListJobs(ctx context.Context, userID string, limit int) ([]Job, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, collection, file_name, file_size, status, error, created_at, updated_at
		 FROM ingest_jobs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list ingest jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.UserID, &j.Collection, &j.FileName, &j.FileSize,
			&j.Status, &j.Error, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan ingest job: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ingest jobs: %w", err)
	}
	if jobs == nil {
		jobs = []Job{}
	}
	return jobs, nil
}

// GetJob returns a single job by ID, verifying ownership.
func (s *Ingest) GetJob(ctx context.Context, jobID, userID string) (*Job, error) {
	var j Job
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, collection, file_name, file_size, status, error, created_at, updated_at
		 FROM ingest_jobs WHERE id = $1`,
		jobID,
	).Scan(&j.ID, &j.UserID, &j.Collection, &j.FileName, &j.FileSize,
		&j.Status, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get ingest job: %w", err)
	}
	if j.UserID != userID {
		return nil, ErrJobNotFound
	}
	return &j, nil
}

// DeleteJob removes a job record, verifying ownership.
func (s *Ingest) DeleteJob(ctx context.Context, jobID, userID string) error {
	result, err := s.pool.Exec(ctx,
		`DELETE FROM ingest_jobs WHERE id = $1 AND user_id = $2`, jobID, userID)
	if err != nil {
		return fmt.Errorf("delete ingest job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

// UpdateJobStatus updates a job's status. Used by the worker.
func (s *Ingest) UpdateJobStatus(ctx context.Context, jobID, status string, errMsg *string) error {
	if errMsg != nil {
		_, err := s.pool.Exec(ctx,
			`UPDATE ingest_jobs SET status = $1, error = $2, updated_at = now() WHERE id = $3`,
			status, *errMsg, jobID)
		return err
	}
	_, err := s.pool.Exec(ctx,
		`UPDATE ingest_jobs SET status = $1, updated_at = now() WHERE id = $2`,
		status, jobID)
	return err
}
