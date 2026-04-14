package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	"github.com/Camionerou/rag-saldivia/services/notification/internal/service"
)

// Mailer sends email notifications (satisfied by service.SMTPMailer).
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// AlertStore persists infrastructure alerts to the platform database.
type AlertStore interface {
	SaveAlert(ctx context.Context, alert service.InfraAlert) error
}

// AlertWebhook handles incoming Alertmanager webhook payloads.
// Mounted on a separate internal router (no JWT middleware, no Traefik).
type AlertWebhook struct {
	token   string     // expected shared secret from Docker secret
	store   AlertStore // platform DB persistence
	mailer  Mailer     // for critical severity email
	alertTo string     // email recipient for critical alerts
}

// NewAlertWebhook creates an alert webhook handler.
func NewAlertWebhook(token string, store AlertStore, mailer Mailer, alertTo string) *AlertWebhook {
	return &AlertWebhook{
		token:   token,
		store:   store,
		mailer:  mailer,
		alertTo: alertTo,
	}
}

// HandleAlertWebhook handles POST /internal/webhook/alert from Alertmanager.
// Auth: Bearer token validated against Docker secret (DS3).
// Persists alerts in platform DB. Does NOT publish NATS events (infra alerts
// have no tenant — documented exception to invariant 4).
func (h *AlertWebhook) HandleAlertWebhook(w http.ResponseWriter, r *http.Request) {
	// Limit body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Validate Bearer token
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		httperr.WriteError(w, r, httperr.Unauthorized("missing authorization"))
		return
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if subtle.ConstantTimeCompare([]byte(token), []byte(h.token)) != 1 {
		httperr.WriteError(w, r, httperr.Unauthorized("invalid token"))
		return
	}

	// Decode Alertmanager payload
	var payload alertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if isMaxBytesError(err) {
			httperr.WriteError(w, r, httperr.Wrap(err, "payload_too_large", "payload too large", http.StatusRequestEntityTooLarge))
			return
		}
		httperr.WriteError(w, r, httperr.InvalidInput("invalid JSON payload"))
		return
	}

	ctx := r.Context()
	persisted := 0

	for _, a := range payload.Alerts {
		alert := toInfraAlert(a)

		if err := h.store.SaveAlert(ctx, alert); err != nil {
			slog.Error("failed to persist alert", "alertname", alert.AlertName, "error", err)
			httperr.WriteError(w, r, httperr.Internal(fmt.Errorf("persist alert: %w", err)))
			return
		}
		persisted++

		// Send email for critical severity firing alerts
		if alert.Severity == "critical" && alert.Status == "firing" && h.mailer != nil && h.alertTo != "" {
			subject := fmt.Sprintf("[CRITICAL] %s — %s", alert.AlertName, alert.Service)
			body := fmt.Sprintf("Alert: %s\nService: %s\nSeverity: %s\nStatus: %s\n\n%s\n\n%s",
				alert.AlertName, alert.Service, alert.Severity, alert.Status,
				alert.Summary, alert.Description)
			if err := h.mailer.Send(ctx, h.alertTo, subject, body); err != nil {
				slog.Error("failed to send critical alert email", "alertname", alert.AlertName, "error", err)
				// Don't fail the request for email errors — alert is already persisted
			}
		}
	}

	slog.Info("alert webhook processed", "persisted", persisted, "status", payload.Status)
	writeJSON(w, http.StatusOK, map[string]int{"persisted": persisted})
}

// --- Alertmanager payload types ---

type alertmanagerPayload struct {
	Version  string              `json:"version"`
	GroupKey string              `json:"groupKey"`
	Status   string              `json:"status"`
	Receiver string              `json:"receiver"`
	Alerts   []alertmanagerAlert `json:"alerts"`
}

type alertmanagerAlert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
	Fingerprint string            `json:"fingerprint"`
}

func toInfraAlert(a alertmanagerAlert) service.InfraAlert {
	labels, _ := json.Marshal(a.Labels)
	annotations, _ := json.Marshal(a.Annotations)

	alert := service.InfraAlert{
		Fingerprint: a.Fingerprint,
		Status:      a.Status,
		Severity:    a.Labels["severity"],
		AlertName:   a.Labels["alertname"],
		Service:     a.Labels["service_name"],
		Summary:     a.Annotations["summary"],
		Description: a.Annotations["description"],
		Labels:      labels,
		Annotations: annotations,
		StartsAt:    a.StartsAt,
	}

	// Alertmanager uses zero time for "not ended yet"
	if !a.EndsAt.IsZero() && a.EndsAt.Year() > 1 {
		alert.EndsAt = &a.EndsAt
	}

	return alert
}

// isMaxBytesError checks if the error is from http.MaxBytesReader.
func isMaxBytesError(err error) bool {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}
	return strings.Contains(err.Error(), "http: request body too large")
}
