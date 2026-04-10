// Package handler provides HTTP handlers for ERP modules.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// Suggestions handles suggestion endpoints.
type Suggestions struct {
	svc *service.Suggestions
}

// NewSuggestions creates a suggestion handler.
func NewSuggestions(svc *service.Suggestions) *Suggestions {
	return &Suggestions{svc: svc}
}

// Routes returns the chi router for suggestion endpoints.
func (h *Suggestions) Routes() chi.Router {
	r := chi.NewRouter()

	// All users can read + create suggestions
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.read"))
		r.Get("/", h.List)
		r.Get("/unread", h.CountUnread)
		r.Get("/{id}", h.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.write"))
		r.Post("/", h.Create)
		r.Post("/{id}/respond", h.Respond)
		r.Patch("/{id}/read", h.MarkRead)
	})

	return r
}

// List returns paginated suggestions.
func (h *Suggestions) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(r)
	if !ok {
		http.Error(w, `{"error":"missing tenant context"}`, http.StatusBadRequest)
		return
	}

	p := pagination.Parse(r)

	suggestions, err := h.svc.List(r.Context(), tenantID, p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list suggestions failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"suggestions": suggestions,
		"page":        p.Page,
		"page_size":   p.PageSize,
	})
}

// Get returns a suggestion with its response thread.
func (h *Suggestions) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(r)
	if !ok {
		http.Error(w, `{"error":"missing tenant context"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	suggestion, responses, err := h.svc.Get(r.Context(), id, tenantID)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"suggestion": suggestion,
		"responses":  responses,
	})
}

// Create creates a new suggestion.
func (h *Suggestions) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(r)
	if !ok {
		http.Error(w, `{"error":"missing tenant context"}`, http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16<<10) // 16KB max

	var body struct {
		Origin string `json:"origin"`
		Body   string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	userID, _ := parseUserID(r)

	suggestion, err := h.svc.Create(r.Context(), service.CreateRequest{
		TenantID: tenantID,
		UserID:   userID,
		Origin:   body.Origin,
		Body:     body.Body,
		IP:       r.RemoteAddr,
	})
	if err != nil {
		slog.Error("create suggestion failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(suggestion)
}

// Respond adds a response to a suggestion.
func (h *Suggestions) Respond(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(r)
	if !ok {
		http.Error(w, `{"error":"missing tenant context"}`, http.StatusBadRequest)
		return
	}

	suggestionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	userID, _ := parseUserID(r)

	response, err := h.svc.Respond(r.Context(), service.RespondRequest{
		TenantID:     tenantID,
		SuggestionID: suggestionID,
		UserID:       userID,
		Body:         body.Body,
		IP:           r.RemoteAddr,
	})
	if err != nil {
		slog.Error("respond to suggestion failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// MarkRead marks a suggestion as read.
func (h *Suggestions) MarkRead(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(r)
	if !ok {
		http.Error(w, `{"error":"missing tenant context"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkRead(r.Context(), id, tenantID); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CountUnread returns the number of unread suggestions.
func (h *Suggestions) CountUnread(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(r)
	if !ok {
		http.Error(w, `{"error":"missing tenant context"}`, http.StatusBadRequest)
		return
	}

	count, err := h.svc.CountUnread(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"unread": count})
}

// parseTenantID extracts tenant ID from JWT claims (set by auth middleware).
func parseTenantID(r *http.Request) (uuid.UUID, bool) {
	tid := r.Header.Get("X-Tenant-ID")
	if tid == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(tid)
	return id, err == nil
}

// parseUserID extracts user ID from JWT claims.
func parseUserID(r *http.Request) (uuid.UUID, bool) {
	uid := r.Header.Get("X-User-ID")
	if uid == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(uid)
	return id, err == nil
}
