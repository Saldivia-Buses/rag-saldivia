// Package handler provides HTTP handlers for the Search Service.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/search/internal/service"
)

// Handler wraps the Search service for HTTP.
type Handler struct {
	svc   *service.Search
	audit *audit.Writer
}

// New creates a search Handler.
func New(svc *service.Search, auditWriter *audit.Writer) *Handler {
	return &Handler{svc: svc, audit: auditWriter}
}

// Routes returns the search router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	// D2: require chat.read permission for search
	r.With(sdamw.RequirePermission("chat.read")).Post("/query", h.SearchDocuments)
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
		httperr.WriteError(w, r, httperr.Unauthorized("tenant context missing"))
		return
	}

	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Query == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("query is required"))
		return
	}

	result, err := h.svc.SearchDocuments(r.Context(), req.Query, req.CollectionID, req.MaxNodes)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	// Audit search query
	if h.audit != nil {
		h.audit.Write(r.Context(), audit.Entry{
			Action:   "search.query",
			Resource: req.Query,
			Details:  map[string]any{"results": len(result.Selections), "duration_ms": result.DurationMS},
			IP:       r.RemoteAddr,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
