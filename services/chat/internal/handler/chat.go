// Package handler implements HTTP handlers for the chat service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Camionerou/rag-saldivia/services/chat/internal/service"
)

// ChatService defines the operations the handler needs from the service layer.
type ChatService interface {
	CreateSession(ctx context.Context, userID, title string, collection *string) (*service.Session, error)
	GetSession(ctx context.Context, sessionID, userID string) (*service.Session, error)
	ListSessions(ctx context.Context, userID string) ([]service.Session, error)
	DeleteSession(ctx context.Context, sessionID, userID string) error
	RenameSession(ctx context.Context, sessionID, userID, title string) error
	AddMessage(ctx context.Context, sessionID, userID, role, content string, thinking *string, sources, metadata []byte) (*service.Message, error)
	GetMessages(ctx context.Context, sessionID string) ([]service.Message, error)
}

// Chat handles HTTP requests for chat operations.
type Chat struct {
	chatSvc ChatService
}

// NewChat creates chat HTTP handlers.
func NewChat(chatSvc ChatService) *Chat {
	return &Chat{chatSvc: chatSvc}
}

// Routes returns a chi router with all chat routes.
func (h *Chat) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(requireUserID)

	r.Get("/", h.ListSessions)
	r.Post("/", h.CreateSession)
	r.Get("/{sessionID}", h.GetSession)
	r.Delete("/{sessionID}", h.DeleteSession)
	r.Patch("/{sessionID}", h.RenameSession)
	r.Get("/{sessionID}/messages", h.GetMessages)
	r.Post("/{sessionID}/messages", h.AddMessage)

	return r
}

type createSessionRequest struct {
	Title      string  `json:"title"`
	Collection *string `json:"collection,omitempty"`
}

type renameRequest struct {
	Title string `json:"title"`
}

type addMessageRequest struct {
	Role     string          `json:"role"`
	Content  string          `json:"content"`
	Thinking *string         `json:"thinking,omitempty"`
	Sources  json.RawMessage `json:"sources,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// ListSessions handles GET /v1/chat/sessions
func (h *Chat) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	sessions, err := h.chatSvc.ListSessions(r.Context(), userID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

// CreateSession handles POST /v1/chat/sessions
func (h *Chat) CreateSession(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	userID := r.Header.Get("X-User-ID")

	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	title := req.Title
	if title == "" {
		title = "Nueva conversacion"
	}

	session, err := h.chatSvc.CreateSession(r.Context(), userID, title, req.Collection)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// GetSession handles GET /v1/chat/sessions/{sessionID}
func (h *Chat) GetSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	sessionID := chi.URLParam(r, "sessionID")

	session, err := h.chatSvc.GetSession(r.Context(), sessionID, userID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) || errors.Is(err, service.ErrNotOwner) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// DeleteSession handles DELETE /v1/chat/sessions/{sessionID}
func (h *Chat) DeleteSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	sessionID := chi.URLParam(r, "sessionID")

	if err := h.chatSvc.DeleteSession(r.Context(), sessionID, userID); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RenameSession handles PATCH /v1/chat/sessions/{sessionID}
func (h *Chat) RenameSession(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	userID := r.Header.Get("X-User-ID")
	sessionID := chi.URLParam(r, "sessionID")

	var req renameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	if err := h.chatSvc.RenameSession(r.Context(), sessionID, userID, req.Title); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetMessages handles GET /v1/chat/sessions/{sessionID}/messages
func (h *Chat) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	sessionID := chi.URLParam(r, "sessionID")

	// Verify ownership before returning messages
	if _, err := h.chatSvc.GetSession(r.Context(), sessionID, userID); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) || errors.Is(err, service.ErrNotOwner) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		serverError(w, r, err)
		return
	}

	messages, err := h.chatSvc.GetMessages(r.Context(), sessionID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

var validRoles = map[string]bool{"user": true, "assistant": true, "system": true}

// AddMessage handles POST /v1/chat/sessions/{sessionID}/messages
func (h *Chat) AddMessage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	userID := r.Header.Get("X-User-ID")
	sessionID := chi.URLParam(r, "sessionID")

	var req addMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Role == "" || req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role and content are required"})
		return
	}
	if !validRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role must be user, assistant, or system"})
		return
	}

	// Verify ownership before adding message
	if _, err := h.chatSvc.GetSession(r.Context(), sessionID, userID); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) || errors.Is(err, service.ErrNotOwner) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		serverError(w, r, err)
		return
	}

	msg, err := h.chatSvc.AddMessage(r.Context(), sessionID, userID, req.Role, req.Content, req.Thinking, req.Sources, req.Metadata)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func requireUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-ID") == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user identity"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetReqID(r.Context())
	slog.Error("internal error", "error", err, "request_id", reqID)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}
