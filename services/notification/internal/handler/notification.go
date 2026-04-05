// Package handler implements HTTP handlers for the notification service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Camionerou/rag-saldivia/services/notification/internal/service"
)

// NotificationService defines the operations the handler needs from the service layer.
type NotificationService interface {
	List(ctx context.Context, userID string, unreadOnly bool, limit int) ([]service.Notification, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, notifID, userID string) error
	MarkAllRead(ctx context.Context, userID string) (int64, error)
	GetPreferences(ctx context.Context, userID string) (*service.Preferences, error)
	UpdatePreferences(ctx context.Context, userID string, emailEnabled, inAppEnabled bool, quietStart, quietEnd *string, mutedTypes []string) (*service.Preferences, error)
}

// Notification handles HTTP requests for notification operations.
type Notification struct {
	svc NotificationService
}

// NewNotification creates notification HTTP handlers.
func NewNotification(svc NotificationService) *Notification {
	return &Notification{svc: svc}
}

// Routes returns a chi router with all notification routes.
func (h *Notification) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(requireUserID)

	r.Get("/", h.List)
	r.Get("/count", h.UnreadCount)
	r.Post("/read-all", h.MarkAllRead)
	r.Patch("/{notificationID}/read", h.MarkRead)

	r.Route("/preferences", func(r chi.Router) {
		r.Get("/", h.GetPreferences)
		r.Put("/", h.UpdatePreferences)
	})

	return r
}

// List handles GET /v1/notifications
func (h *Notification) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	unreadOnly, _ := strconv.ParseBool(r.URL.Query().Get("unread"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	notifications, err := h.svc.List(r.Context(), userID, unreadOnly, limit)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, notifications)
}

// UnreadCount handles GET /v1/notifications/count
func (h *Notification) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	count, err := h.svc.UnreadCount(r.Context(), userID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

// MarkRead handles PATCH /v1/notifications/{notificationID}/read
func (h *Notification) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	notifID := chi.URLParam(r, "notificationID")

	if err := h.svc.MarkRead(r.Context(), notifID, userID); err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "notification not found"})
			return
		}
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MarkAllRead handles POST /v1/notifications/read-all
func (h *Notification) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	count, err := h.svc.MarkAllRead(r.Context(), userID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"marked": count})
}

// GetPreferences handles GET /v1/notifications/preferences
func (h *Notification) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	prefs, err := h.svc.GetPreferences(r.Context(), userID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

type updatePreferencesRequest struct {
	EmailEnabled bool     `json:"email_enabled"`
	InAppEnabled bool     `json:"in_app_enabled"`
	QuietStart   *string  `json:"quiet_start,omitempty"`
	QuietEnd     *string  `json:"quiet_end,omitempty"`
	MutedTypes   []string `json:"muted_types"`
}

// UpdatePreferences handles PUT /v1/notifications/preferences
func (h *Notification) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	userID := r.Header.Get("X-User-ID")

	var req updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	prefs, err := h.svc.UpdatePreferences(r.Context(), userID, req.EmailEnabled, req.InAppEnabled, req.QuietStart, req.QuietEnd, req.MutedTypes)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, prefs)
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
