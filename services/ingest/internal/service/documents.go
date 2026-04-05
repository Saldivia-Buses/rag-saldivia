package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/storage"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/repository"
)

// DocumentService manages document upload, extraction triggering, and tree storage.
type DocumentService struct {
	pool    *pgxpool.Pool
	repo    *repository.Queries
	nc      *nats.Conn
	store   storage.Store
	tenant  string // tenant slug for NATS subjects and storage keys
}

// NewDocumentService creates a DocumentService.
func NewDocumentService(pool *pgxpool.Pool, nc *nats.Conn, store storage.Store, tenant string) *DocumentService {
	return &DocumentService{
		pool:   pool,
		repo:   repository.New(pool),
		nc:     nc,
		store:  store,
		tenant: tenant,
	}
}

// UploadDocument stores a file in MinIO, creates a DB record, and triggers extraction.
func (s *DocumentService) UploadDocument(ctx context.Context, userID, fileName string, fileSize int64, file multipart.File) (*repository.Document, error) {
	// Read file to compute hash
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	hash := sha256.Sum256(data)
	fileHash := hex.EncodeToString(hash[:])

	// Dedup check
	existing, err := s.repo.GetDocumentByHash(ctx, fileHash)
	if err == nil {
		slog.Info("duplicate document", "hash", fileHash, "existing_id", existing.ID)
		return &existing, nil
	}

	fileType := filepath.Ext(fileName)
	if len(fileType) > 0 {
		fileType = fileType[1:] // remove dot
	}

	// Generate ID for storage key
	doc, err := s.repo.CreateDocument(ctx, repository.CreateDocumentParams{
		Name:       fileName,
		StorageKey: "", // set after we know the ID
		FileType:   fileType,
		FileHash:   fileHash,
		SizeBytes:  fileSize,
		UploadedBy: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	// Store in MinIO
	storageKey := fmt.Sprintf("%s/%s/original.%s", s.tenant, doc.ID, fileType)
	if err := s.store.Put(ctx, storageKey, io.NopCloser(io.Reader(nil)), &storage.PutOptions{ContentType: mimeType(fileType)}); err != nil {
		// Actually put the real data
	}
	// Put actual data
	err = s.store.Put(ctx, storageKey, readCloserFromBytes(data), &storage.PutOptions{
		ContentType: mimeType(fileType),
	})
	if err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}

	// Update storage key in DB
	// (We need to add this query — for now use the key from creation)

	// Publish extraction job via NATS
	job := ExtractionJobMessage{
		DocumentID: doc.ID,
		TenantSlug: s.tenant,
		StorageKey: storageKey,
		FileName:   fileName,
		FileType:   fileType,
	}
	payload, _ := json.Marshal(job)
	subject := fmt.Sprintf("tenant.%s.extractor.job", s.tenant)
	if err := s.nc.Publish(subject, payload); err != nil {
		slog.Error("failed to publish extraction job", "error", err, "doc_id", doc.ID)
		s.repo.UpdateDocumentStatusWithError(ctx, repository.UpdateDocumentStatusWithErrorParams{
			Status: "error",
			Error:  pgtype.Text{String: "failed to trigger extraction", Valid: true},
			ID:     doc.ID,
		})
		return nil, fmt.Errorf("publish extraction job: %w", err)
	}

	// Update status to extracting
	s.repo.UpdateDocumentStatus(ctx, repository.UpdateDocumentStatusParams{
		Status: "extracting",
		ID:     doc.ID,
	})

	return &doc, nil
}

// ExtractionJobMessage is published to NATS for the Extractor.
type ExtractionJobMessage struct {
	DocumentID string `json:"document_id"`
	TenantSlug string `json:"tenant_slug"`
	StorageKey string `json:"storage_key"`
	FileName   string `json:"file_name"`
	FileType   string `json:"file_type"`
}

func mimeType(ext string) string {
	switch ext {
	case "pdf":
		return "application/pdf"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

type bytesReadCloser struct {
	*io.SectionReader
}

func (b *bytesReadCloser) Close() error { return nil }

func readCloserFromBytes(data []byte) io.Reader {
	return io.NopCloser(io.NewSectionReader(readerAtBytes(data), 0, int64(len(data))))
}

type readerAtBytes []byte

func (r readerAtBytes) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(r)) {
		return 0, io.EOF
	}
	n := copy(p, r[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}
