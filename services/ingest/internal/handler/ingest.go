// Package handler implements HTTP handlers for the Ingest service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/service"
)

// MaxUploadSize is the maximum document upload size (100MB).
const MaxUploadSize = 100 << 20

var allowedExts = map[string]bool{
	".pdf": true, ".docx": true, ".doc": true, ".txt": true,
	".md": true, ".csv": true, ".xlsx": true, ".pptx": true,
	".html": true, ".json": true, ".xml": true,
}

// IngestService defines the operations the handler needs from the service layer.
type IngestService interface {
	Submit(ctx context.Context, tenantSlug, userID, collection, fileName string, fileSize int64, file multipart.File) (*service.Job, error)
	ListJobs(ctx context.Context, userID string, limit int) ([]service.Job, error)
	GetJob(ctx context.Context, jobID, userID string) (*service.Job, error)
	DeleteJob(ctx context.Context, jobID, userID string) error
	ListCollections(ctx context.Context) ([]repository.Collection, error)
	CreateCollection(ctx context.Context, name, description string) (repository.Collection, error)
}

// Ingest handles HTTP requests for document ingestion.
type Ingest struct {
	svc IngestService
}

// NewIngest creates Ingest HTTP handlers.
func NewIngest(svc IngestService) *Ingest {
	return &Ingest{svc: svc}
}

// Routes returns a chi router with all ingest routes.
func (h *Ingest) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(sdamw.RequirePermission("ingest.write")).Post("/upload", h.Upload)
	r.With(sdamw.RequirePermission("ingest.write")).Get("/jobs", h.ListJobs)
	r.With(sdamw.RequirePermission("ingest.write")).Get("/jobs/{jobID}", h.GetJob)
	r.With(sdamw.RequirePermission("ingest.write")).Delete("/jobs/{jobID}", h.DeleteJob)
	r.With(sdamw.RequirePermission("collections.read")).Get("/collections", h.ListCollections)
	r.With(sdamw.RequirePermission("collections.write")).Post("/collections", h.CreateCollection)
	return r
}

// requireIdentity extracts and validates identity headers set by auth middleware.
func requireIdentity(r *http.Request) (userID, tenantSlug string, ok bool) {
	userID = r.Header.Get("X-User-ID")
	tenantSlug = r.Header.Get("X-Tenant-Slug")
	return userID, tenantSlug, userID != "" && tenantSlug != ""
}

// Upload handles POST /v1/ingest/upload — multipart document upload.
func (h *Ingest) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	userID, tenantSlug, ok := requireIdentity(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB in memory, rest to disk
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	// S6: validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExts[ext] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported file type: " + ext})
		return
	}
	// P1: sanitize filename — strip path components, reject dangerous patterns
	safeName := filepath.Base(header.Filename)
	if safeName == "." || safeName == ".." || strings.ContainsAny(safeName, "/\\") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid file name"})
		return
	}
	header.Filename = safeName

	collection := r.FormValue("collection")
	if collection == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "collection is required"})
		return
	}

	job, err := h.svc.Submit(r.Context(), tenantSlug, userID, collection, header.Filename, header.Size, file)
	if err != nil {
		reqID := middleware.GetReqID(r.Context())
		slog.Error("ingest submit failed", "error", err, "request_id", reqID, "file", header.Filename)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to queue document"})
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

// ListJobs handles GET /v1/ingest/jobs
func (h *Ingest) ListJobs(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireIdentity(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	jobs, err := h.svc.ListJobs(r.Context(), userID, limit)
	if err != nil {
		reqID := middleware.GetReqID(r.Context())
		slog.Error("list ingest jobs failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
}

// GetJob handles GET /v1/ingest/jobs/{jobID}
func (h *Ingest) GetJob(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireIdentity(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
		return
	}

	jobID := chi.URLParam(r, "jobID")
	job, err := h.svc.GetJob(r.Context(), jobID, userID)
	if err != nil {
		if errors.Is(err, service.ErrJobNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		reqID := middleware.GetReqID(r.Context())
		slog.Error("get ingest job failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// DeleteJob handles DELETE /v1/ingest/jobs/{jobID}
func (h *Ingest) DeleteJob(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireIdentity(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
		return
	}

	jobID := chi.URLParam(r, "jobID")
	if err := h.svc.DeleteJob(r.Context(), jobID, userID); err != nil {
		if errors.Is(err, service.ErrJobNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		reqID := middleware.GetReqID(r.Context())
		slog.Error("delete ingest job failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// ListCollections handles GET /v1/ingest/collections.
func (h *Ingest) ListCollections(w http.ResponseWriter, r *http.Request) {
	collections, err := h.svc.ListCollections(r.Context())
	if err != nil {
		slog.Error("list collections failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list collections"})
		return
	}
	writeJSON(w, http.StatusOK, collections)
}

// CreateCollection handles POST /v1/ingest/collections.
func (h *Ingest) CreateCollection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	col, err := h.svc.CreateCollection(r.Context(), strings.TrimSpace(req.Name), strings.TrimSpace(req.Description))
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "collection already exists"})
			return
		}
		slog.Error("create collection failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create collection"})
		return
	}
	writeJSON(w, http.StatusCreated, col)
}
