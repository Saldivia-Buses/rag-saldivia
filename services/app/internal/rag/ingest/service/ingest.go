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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/services/app/internal/rag/ingest/repository"
)

var ErrJobNotFound = errors.New("job not found")

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
	pool    *pgxpool.Pool
	repo    *repository.Queries
	nc      *nats.Conn
	auditor *audit.Writer
	cfg     Config
}

// New creates an Ingest service. Event publishing for job completion is handled
// via the transactional outbox in the Worker (see completeJob).
func New(pool *pgxpool.Pool, nc *nats.Conn, cfg Config) *Ingest {
	if cfg.StagingDir == "" {
		cfg.StagingDir = "/tmp/ingest-staging"
	}
	if err := os.MkdirAll(cfg.StagingDir, 0750); err != nil {
		slog.Error("failed to create staging directory", "error", err, "path", cfg.StagingDir)
	}

	return &Ingest{
		pool:    pool,
		repo:    repository.New(pool),
		nc:      nc,
		auditor: audit.NewWriter(pool),
		cfg:     cfg,
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

// toJob converts a sqlc-generated IngestJob to the service-layer Job type.
func toJob(r repository.IngestJob) Job {
	j := Job{
		ID:         r.ID,
		UserID:     r.UserID,
		Collection: r.Collection,
		FileName:   r.FileName,
		FileSize:   r.FileSize,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt.Time,
		UpdatedAt:  r.UpdatedAt.Time,
	}
	if r.Error.Valid {
		j.Error = &r.Error.String
	}
	return j
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
	_ = tmpFile.Close()
	if err != nil {
		_ = os.Remove(stagedPath)
		return nil, fmt.Errorf("stage file: %w", err)
	}
	if written == 0 {
		_ = os.Remove(stagedPath)
		return nil, fmt.Errorf("empty file")
	}

	// Create job in DB
	row, err := s.repo.CreateJob(ctx, repository.CreateJobParams{
		UserID:     userID,
		Collection: collection,
		FileName:   fileName,
		FileSize:   fileSize,
	})
	if err != nil {
		_ = os.Remove(stagedPath)
		return nil, fmt.Errorf("create ingest job: %w", err)
	}
	job := toJob(row)

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
		_ = os.Remove(stagedPath)
		return nil, fmt.Errorf("marshal ingest message: %w", err)
	}

	// S7 fix: fail the upload if NATS publish fails — prevents orphaned pending jobs
	//nolint:forbidigo // Plan 26 Fase 3 migrates ingest publishes to outbox.PublishTx.
	if err := s.nc.Publish(subject, payload); err != nil {
		_ = os.Remove(stagedPath)
		_ = s.repo.DeleteJobByID(ctx, job.ID)
		return nil, fmt.Errorf("publish ingest job: %w", err)
	}

	s.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "ingest.upload", Resource: job.ID,
		Details: map[string]any{"file": fileName, "collection": collection, "size": fileSize},
	})

	return &job, nil
}

// ListJobs returns all jobs for a user, newest first.
func (s *Ingest) ListJobs(ctx context.Context, userID string, limit int) ([]Job, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.repo.ListJobsByUser(ctx, repository.ListJobsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list ingest jobs: %w", err)
	}

	jobs := make([]Job, len(rows))
	for i, r := range rows {
		jobs[i] = toJob(r)
	}
	return jobs, nil
}

// GetJob returns a single job by ID, verifying ownership at the query level.
func (s *Ingest) GetJob(ctx context.Context, jobID, userID string) (*Job, error) {
	row, err := s.repo.GetJob(ctx, repository.GetJobParams{
		ID:     jobID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get ingest job: %w", err)
	}
	job := toJob(row)
	return &job, nil
}

// DeleteJob removes a job record, verifying ownership.
func (s *Ingest) DeleteJob(ctx context.Context, jobID, userID string) error {
	n, err := s.repo.DeleteJob(ctx, repository.DeleteJobParams{
		ID:     jobID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete ingest job: %w", err)
	}
	if n == 0 {
		return ErrJobNotFound
	}

	s.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "ingest.delete_job", Resource: jobID,
	})

	return nil
}

// UpdateJobStatus updates a job's status. Used by the worker.
func (s *Ingest) UpdateJobStatus(ctx context.Context, jobID, status string, errMsg *string) error {
	if errMsg != nil {
		return s.repo.UpdateJobStatusWithError(ctx, repository.UpdateJobStatusWithErrorParams{
			Status: status,
			Error:  pgtype.Text{String: *errMsg, Valid: true},
			ID:     jobID,
		})
	}
	return s.repo.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
		Status: status,
		ID:     jobID,
	})
}

// ListCollections returns all collections.
func (s *Ingest) ListCollections(ctx context.Context) ([]repository.Collection, error) {
	return s.repo.ListCollections(ctx)
}

// CreateCollection creates a new collection.
func (s *Ingest) CreateCollection(ctx context.Context, name, description string) (repository.Collection, error) {
	return s.repo.CreateCollection(ctx, repository.CreateCollectionParams{
		Name:        name,
		Description: pgtype.Text{String: description, Valid: description != ""},
	})
}
