// Package handler implements HTTP handlers for the chat service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/app/internal/guardrails"
	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/app/internal/realtime/chat/service"
)

// ChatService defines the operations the handler needs from the service layer.
type ChatService interface {
	CreateSession(ctx context.Context, userID, title string, collection *string) (*service.Session, error)
	GetSession(ctx context.Context, sessionID, userID string) (*service.Session, error)
	ListSessions(ctx context.Context, userID string, limit, offset int32) ([]service.Session, error)
	DeleteSession(ctx context.Context, sessionID, userID string) error
	RenameSession(ctx context.Context, sessionID, userID, title string) error
	AddMessage(ctx context.Context, sessionID, userID, role, content string, thinking *string, sources, metadata []byte) (*service.Message, error)
	GetMessages(ctx context.Context, sessionID string, limit int32) ([]service.Message, error)
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

	// Read operations — require chat.read
	r.With(sdamw.RequirePermission("chat.read")).Get("/", h.ListSessions)
	r.With(sdamw.RequirePermission("chat.read")).Get("/{sessionID}", h.GetSession)
	r.With(sdamw.RequirePermission("chat.read")).Get("/{sessionID}/messages", h.GetMessages)

	// Write operations — require chat.write
	r.With(sdamw.RequirePermission("chat.write")).Post("/", h.CreateSession)
	r.With(sdamw.RequirePermission("chat.write")).Post("/{sessionID}/messages", h.AddMessage)
	r.With(sdamw.RequirePermission("chat.write")).Delete("/{sessionID}", h.DeleteSession)
	r.With(sdamw.RequirePermission("chat.write")).Patch("/{sessionID}", h.RenameSession)

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

// ListSessions handles GET /v1/chat/sessions (paginated)
func (h *Chat) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	pg := pagination.Parse(r)
	sessions, err := h.chatSvc.ListSessions(r.Context(), userID, int32(pg.Limit()), int32(pg.Offset()))
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
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
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	title := req.Title
	if title == "" {
		title = "Nueva conversacion"
	}

	session, err := h.chatSvc.CreateSession(r.Context(), userID, title, req.Collection)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
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
			httperr.WriteError(w, r, httperr.NotFound("session"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
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
			httperr.WriteError(w, r, httperr.NotFound("session"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
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
		httperr.WriteError(w, r, httperr.InvalidInput("title is required"))
		return
	}

	if err := h.chatSvc.RenameSession(r.Context(), sessionID, userID, req.Title); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			httperr.WriteError(w, r, httperr.NotFound("session"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
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
			httperr.WriteError(w, r, httperr.NotFound("session"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	// Messages use limit-only (no offset) — they're loaded oldest-first sequentially
	limit := int32(pagination.DefaultPageSize)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= pagination.MaxPageSize {
			limit = int32(n)
		}
	}
	messages, err := h.chatSvc.GetMessages(r.Context(), sessionID, limit)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
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
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Role == "" || req.Content == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("role and content are required"))
		return
	}
	if !validRoles[req.Role] {
		httperr.WriteError(w, r, httperr.InvalidInput("role must be user, assistant, or system"))
		return
	}
	// C4: block system role from external clients — only internal services should set system messages
	if req.Role == "system" {
		httperr.WriteError(w, r, httperr.Forbidden("system messages cannot be added via API"))
		return
	}

	// P1: validate user message content through guardrails
	if req.Role == "user" {
		sanitized, err := guardrails.ValidateInput(r.Context(), req.Content, guardrails.DefaultInputConfig(50000), nil)
		if err != nil {
			httperr.WriteError(w, r, httperr.InvalidInput("message blocked by guardrails"))
			return
		}
		req.Content = sanitized
	}

	// Verify ownership before adding message
	if _, err := h.chatSvc.GetSession(r.Context(), sessionID, userID); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) || errors.Is(err, service.ErrNotOwner) {
			httperr.WriteError(w, r, httperr.NotFound("session"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	msg, err := h.chatSvc.AddMessage(r.Context(), sessionID, userID, req.Role, req.Content, req.Thinking, req.Sources, req.Metadata)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func requireUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-ID") == "" {
			httperr.WriteError(w, r, httperr.Unauthorized("missing user identity"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

