// Package handler provides HTTP handlers for the Agent Runtime.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/app/internal/httperr"
	"github.com/Camionerou/rag-saldivia/pkg/llm"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/app/internal/rag/agent/service"
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
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Message == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("message is required"))
		return
	}

	userID := r.Header.Get("X-User-ID")
	result, err := h.svc.Query(r.Context(), jwt, userID, req.Message, req.History)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
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
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Tool == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("tool is required"))
		return
	}

	result, err := h.svc.ExecuteConfirmed(r.Context(), jwt, req.Tool, req.Params)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func extractJWT(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
