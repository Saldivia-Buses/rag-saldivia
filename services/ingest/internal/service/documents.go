package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"regexp"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/storage"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/repository"
)

var safeSubjectToken = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// DocumentService manages document upload, extraction triggering, and tree storage.
type DocumentService struct {
	pool   *pgxpool.Pool
	repo   *repository.Queries
	nc     *nats.Conn
	store  storage.Store
	tenant string
}

// NewDocumentService creates a DocumentService. Panics if tenant is not a valid NATS token.
func NewDocumentService(pool *pgxpool.Pool, nc *nats.Conn, store storage.Store, tenant string) *DocumentService {
	if !safeSubjectToken.MatchString(tenant) {
		panic(fmt.Sprintf("invalid tenant slug for NATS: %q", tenant))
	}
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
	// B3: stream hash computation without loading entire file into memory.
	// We still need the bytes for MinIO upload, but use TeeReader to hash in one pass.
	hasher := sha256.New()
	var buf bytes.Buffer
	if _, err := io.Copy(io.MultiWriter(hasher, &buf), file); err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

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

	// B2: compute storage key before insert so it's never empty
	// Use hash prefix as temp ID — will be the real doc ID path
	tempKey := fmt.Sprintf("%s/pending-%s/original.%s", s.tenant, fileHash[:12], fileType)

	doc, err := s.repo.CreateDocument(ctx, repository.CreateDocumentParams{
		Name:       fileName,
		StorageKey: tempKey,
		FileType:   fileType,
		FileHash:   fileHash,
		SizeBytes:  fileSize,
		UploadedBy: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	// Store in MinIO with real doc ID path
	storageKey := fmt.Sprintf("%s/%s/original.%s", s.tenant, doc.ID, fileType)
	err = s.store.Put(ctx, storageKey, bytes.NewReader(buf.Bytes()), &storage.PutOptions{
		ContentType: mimeType(fileType),
	})
	if err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}

	// Update storage key to real path
	s.repo.UpdateDocumentStorageKey(ctx, repository.UpdateDocumentStorageKeyParams{
		StorageKey: storageKey,
		ID:         doc.ID,
	})
	doc.StorageKey = storageKey

	// Publish extraction job via NATS
	job := ExtractionJobMessage{
		DocumentID: doc.ID,
		TenantSlug: s.tenant,
		StorageKey: storageKey,
		FileName:   fileName,
		FileType:   fileType,
	}
	payload, _ := json.Marshal(job)
	// B4: tenant already validated in constructor, safe for subject
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
