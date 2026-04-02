// Package handler implements HTTP handlers for the Ingest service.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Camionerou/rag-saldivia/services/ingest/internal/service"
)

// MaxUploadSize is the maximum document upload size (100MB).
const MaxUploadSize = 100 << 20

// Ingest handles HTTP requests for document ingestion.
type Ingest struct {
	svc *service.Ingest
}

// NewIngest creates Ingest HTTP handlers.
func NewIngest(svc *service.Ingest) *Ingest {
	return &Ingest{svc: svc}
}

// Routes returns a chi router with all ingest routes.
func (h *Ingest) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/upload", h.Upload)
	r.Get("/jobs", h.ListJobs)
	r.Get("/jobs/{jobID}", h.GetJob)
	r.Delete("/jobs/{jobID}", h.DeleteJob)
	return r
}

// Upload handles POST /v1/ingest/upload — multipart document upload.
func (h *Ingest) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	userID := r.Header.Get("X-User-ID")
	tenantSlug := r.Header.Get("X-Tenant-Slug")
	if userID == "" || tenantSlug == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
		return
	}

	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

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
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
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
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
		return
	}

	jobID := chi.URLParam(r, "jobID")
	job, err := h.svc.GetJob(r.Context(), jobID, userID)
	if err != nil {
		if err.Error() == "job not found" {
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
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
		return
	}

	jobID := chi.URLParam(r, "jobID")
	if err := h.svc.DeleteJob(r.Context(), jobID, userID); err != nil {
		if err.Error() == "job not found" {
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
