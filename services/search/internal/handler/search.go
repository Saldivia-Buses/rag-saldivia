// Package handler provides HTTP handlers for the Search Service.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/search/internal/service"
)

// Handler wraps the Search service for HTTP.
type Handler struct {
	svc *service.Search
}

// New creates a search Handler.
func New(svc *service.Search) *Handler {
	return &Handler{svc: svc}
}

// Routes returns the search router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/query", h.SearchDocuments)
	return r
}

type searchRequest struct {
	Query        string `json:"query"`
	CollectionID string `json:"collection_id,omitempty"`
	MaxNodes     int    `json:"max_nodes,omitempty"`
}

// SearchDocuments handles POST /v1/search/query
func (h *Handler) SearchDocuments(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

	// C1: tenant isolation — read from JWT context
	ti, err := tenant.FromContext(r.Context())
	if err != nil || ti.ID == "" {
		http.Error(w, `{"error":"tenant context missing"}`, http.StatusUnauthorized)
		return
	}

	reqID := middleware.GetReqID(r.Context())

	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, `{"error":"query is required"}`, http.StatusBadRequest)
		return
	}

	result, err := h.svc.SearchDocuments(r.Context(), req.Query, req.CollectionID, req.MaxNodes)
	if err != nil {
		slog.Error("search failed", "error", err, "query", req.Query,
			"request_id", reqID, "tenant_id", ti.ID)
		http.Error(w, `{"error":"search failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
