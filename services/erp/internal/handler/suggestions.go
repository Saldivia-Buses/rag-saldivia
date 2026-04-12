// Package handler provides HTTP handlers for ERP modules.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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

// tenantSlug reads the tenant slug from the request headers (injected by auth middleware).
func tenantSlug(r *http.Request) string {
	return r.Header.Get("X-Tenant-Slug")
}

// writeSafeErr writes a JSON error response without leaking internal details.
// Known business errors pass through; unknown errors become "internal error".
func writeSafeErr(w http.ResponseWriter, err error, status int) {
	msg := err.Error()
	safe := "internal error"
	for _, prefix := range []string{
		"fiscal year", "draft entries", "result_account",
		"solo se pueden", "factura con CAE", "receipt",
		"not found", "already", "balance", "don't balance",
		"invalid transition", "warehouse_id required",
		"allocation", "not open", "not confirmed",
	} {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(prefix)) {
			safe = msg
			break
		}
	}
	http.Error(w, `{"error":"`+safe+`"}`, status)
}

// parseUUID parses a string into pgtype.UUID.
func parseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

// Routes returns the chi router for suggestion endpoints.
func (h *Suggestions) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.suggestions.read"))
		r.Get("/", h.List)
		r.Get("/unread", h.CountUnread)
		r.Get("/{id}", h.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.suggestions.write"))
		r.Post("/", h.Create)
		r.Post("/{id}/respond", h.Respond)
		r.Patch("/{id}/read", h.MarkRead)
	})

	return r
}

// List returns paginated suggestions.
func (h *Suggestions) List(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)

	suggestions, err := h.svc.List(r.Context(), slug, p.Limit(), p.Offset())
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
	slug := tenantSlug(r)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	suggestion, responses, err := h.svc.Get(r.Context(), id, slug)
	if err != nil {
		slog.Error("get suggestion failed", "error", err, "id", id)
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
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		Origin string `json:"origin"`
		Body   string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"missing user identity"}`, http.StatusUnauthorized)
		return
	}

	suggestion, err := h.svc.Create(r.Context(), service.CreateRequest{
		TenantID: slug,
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
	slug := tenantSlug(r)

	suggestionID, err := parseUUID(chi.URLParam(r, "id"))
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

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"missing user identity"}`, http.StatusUnauthorized)
		return
	}

	response, err := h.svc.Respond(r.Context(), service.RespondRequest{
		TenantID:     slug,
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
	slug := tenantSlug(r)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkRead(r.Context(), id, slug); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CountUnread returns the number of unread suggestions.
func (h *Suggestions) CountUnread(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	count, err := h.svc.CountUnread(r.Context(), slug)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"unread": count})
}
