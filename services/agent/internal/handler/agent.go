// Package handler provides HTTP handlers for the Agent Runtime.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/agent/internal/llm"
	"github.com/Camionerou/rag-saldivia/services/agent/internal/service"
)

// Handler wraps the Agent service for HTTP.
type Handler struct {
	svc *service.Agent
}

// New creates an agent Handler.
func New(svc *service.Agent) *Handler {
	return &Handler{svc: svc}
}

// Routes returns the agent router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/query", h.Query)
	return r
}

type queryRequest struct {
	Message string        `json:"message"`
	History []llm.Message `json:"history,omitempty"`
}

// Query handles POST /v1/agent/query
func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024) // 256KB max

	// Extract JWT from auth header for tool passthrough
	jwt := r.Header.Get("Authorization")
	if len(jwt) > 7 && jwt[:7] == "Bearer " {
		jwt = jwt[7:]
	}

	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, `{"error":"message is required"}`, http.StatusBadRequest)
		return
	}

	result, err := h.svc.Query(r.Context(), jwt, req.Message, req.History)
	if err != nil {
		slog.Error("agent query failed", "error", err)
		http.Error(w, `{"error":"query failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
