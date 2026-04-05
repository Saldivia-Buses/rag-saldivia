// Package handler provides HTTP handlers for the Agent Runtime.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/llm"
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
	// H5: require chat.read permission for agent endpoints
	r.With(sdamw.RequirePermission("chat.read")).Post("/query", h.Query)
	r.With(sdamw.RequirePermission("chat.read")).Post("/confirm", h.Confirm)
	return r
}

type queryRequest struct {
	Message string        `json:"message"`
	History []llm.Message `json:"history,omitempty"`
}

// Query handles POST /v1/agent/query
func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	jwt := extractJWT(r)

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
		slog.Error("agent query failed", "error", err,
			"request_id", chimw.GetReqID(r.Context()))
		http.Error(w, `{"error":"query failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type confirmRequest struct {
	Tool   string          `json:"tool"`
	Params json.RawMessage `json:"params"`
}

// Confirm handles POST /v1/agent/confirm — executes a tool after user approval.
func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

	jwt := extractJWT(r)

	var req confirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Tool == "" {
		http.Error(w, `{"error":"tool is required"}`, http.StatusBadRequest)
		return
	}

	result, err := h.svc.ExecuteConfirmed(r.Context(), jwt, req.Tool, req.Params)
	if err != nil {
		slog.Error("confirm execution failed", "error", err, "tool", req.Tool)
		http.Error(w, `{"error":"execution failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func extractJWT(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
