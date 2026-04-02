// Package handler implements HTTP handlers for the RAG service.
package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Camionerou/rag-saldivia/services/rag/internal/service"
)

// RAGService defines the operations the handler needs from the service layer.
type RAGService interface {
	GenerateStream(ctx context.Context, tenantSlug string, req service.GenerateRequest) (io.ReadCloser, string, error)
	ListCollections(ctx context.Context, tenantSlug string) ([]string, error)
	Health(ctx context.Context) error
}

// RAG handles HTTP requests for RAG operations.
type RAG struct {
	ragSvc RAGService
}

// NewRAG creates RAG HTTP handlers.
func NewRAG(ragSvc RAGService) *RAG {
	return &RAG{ragSvc: ragSvc}
}

// Routes returns a chi router with all RAG routes.
func (h *RAG) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/generate", h.Generate)
	r.Get("/collections", h.ListCollections)
	return r
}

// Generate handles POST /v1/rag/generate — proxies streaming to the Blueprint.
func (h *RAG) Generate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
	tenantSlug := r.Header.Get("X-Tenant-Slug")

	var req service.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "messages are required"})
		return
	}

	body, contentType, err := h.ragSvc.GenerateStream(r.Context(), tenantSlug, req)
	if err != nil {
		reqID := middleware.GetReqID(r.Context())
		slog.Error("rag generate failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "rag server unavailable"})
		return
	}
	defer body.Close()

	// Stream SSE directly to the client
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		io.Copy(w, body)
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
}

// ListCollections handles GET /v1/rag/collections
func (h *RAG) ListCollections(w http.ResponseWriter, r *http.Request) {
	tenantSlug := r.Header.Get("X-Tenant-Slug")

	collections, err := h.ragSvc.ListCollections(r.Context(), tenantSlug)
	if err != nil {
		reqID := middleware.GetReqID(r.Context())
		slog.Error("list collections failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "rag server unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"collections": collections})
}

// Health checks if the RAG Blueprint is reachable.
func (h *RAG) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.ragSvc.Health(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy", "service": "rag", "error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "rag"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
