// Package handler implements HTTP handlers for the notification service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
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
	Send(ctx context.Context, req service.SendRequest) error
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

	// /send requires admin role — prevents arbitrary email relay.
	// Called by triage workflow with platform admin service account JWT.
	r.Group(func(r chi.Router) {
		r.Use(requireAdmin)
		r.Post("/send", h.Send)
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
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, notifications)
}

// UnreadCount handles GET /v1/notifications/count
func (h *Notification) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	count, err := h.svc.UnreadCount(r.Context(), userID)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
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
			httperr.WriteError(w, r, httperr.NotFound("notification"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MarkAllRead handles POST /v1/notifications/read-all
func (h *Notification) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	count, err := h.svc.MarkAllRead(r.Context(), userID)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"marked": count})
}

// GetPreferences handles GET /v1/notifications/preferences
func (h *Notification) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	prefs, err := h.svc.GetPreferences(r.Context(), userID)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
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
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	prefs, err := h.svc.UpdatePreferences(r.Context(), userID, req.EmailEnabled, req.InAppEnabled, req.QuietStart, req.QuietEnd, req.MutedTypes)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

type sendRequest struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Send handles POST /v1/notifications/send
func (h *Notification) Send(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Type == "" || req.To == "" || req.Subject == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("type, to, and subject are required"))
		return
	}

	if req.Type != "email" && req.Type != "in_app" {
		httperr.WriteError(w, r, httperr.InvalidInput("type must be email or in_app"))
		return
	}

	if err := h.svc.Send(r.Context(), service.SendRequest{
		Type:    req.Type,
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
	}); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
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

// requireAdmin checks that the authenticated user has admin role.
// Used for privileged operations like programmatic notification sending.
func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-Role") != "admin" {
			httperr.WriteError(w, r, httperr.Forbidden("admin access required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
