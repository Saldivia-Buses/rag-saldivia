// Package handler provides HTTP handlers for ERP modules.
package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// SuggestionsService is the interface the Suggestions handler depends on.
type SuggestionsService interface {
	List(ctx context.Context, tenantID string, limit, offset int) ([]repository.ListSuggestionsRow, error)
	Get(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpSuggestion, []repository.ErpSuggestionResponse, error)
	Create(ctx context.Context, req service.CreateRequest) (repository.ErpSuggestion, error)
	Respond(ctx context.Context, req service.RespondRequest) (repository.ErpSuggestionResponse, error)
	MarkRead(ctx context.Context, id pgtype.UUID, tenantID string) error
	CountUnread(ctx context.Context, tenantID string) (int32, error)
}

// Suggestions handles suggestion endpoints.
type Suggestions struct {
	svc SuggestionsService
}

// NewSuggestions creates a suggestion handler.
func NewSuggestions(svc SuggestionsService) *Suggestions {
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encErr := json.NewEncoder(w).Encode(map[string]string{"error": safe}); encErr != nil {
		slog.Warn("write error response failed", "error", encErr)
	}
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
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
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}

	suggestion, responses, err := h.svc.Get(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("suggestion"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
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
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		erperrors.WriteError(w, r, erperrors.Wrap(nil, erperrors.CodeUnauthorized, "missing user identity", http.StatusUnauthorized))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(suggestion)
}

// Respond adds a response to a suggestion.
func (h *Suggestions) Respond(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	suggestionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		erperrors.WriteError(w, r, erperrors.Wrap(nil, erperrors.CodeUnauthorized, "missing user identity", http.StatusUnauthorized))
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
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// MarkRead marks a suggestion as read.
func (h *Suggestions) MarkRead(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}

	if err := h.svc.MarkRead(r.Context(), id, slug); err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CountUnread returns the number of unread suggestions.
func (h *Suggestions) CountUnread(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	count, err := h.svc.CountUnread(r.Context(), slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"unread": count})
}
