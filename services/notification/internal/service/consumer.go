package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"

	"github.com/nats-io/nats.go"
)

// Event represents an incoming NATS event that should generate a notification.
type Event struct {
	TenantSlug string          `json:"tenant_slug"`
	UserID     string          `json:"user_id"`      // target user (who gets notified)
	Type       string          `json:"type"`          // e.g., "chat.new_message"
	Title      string          `json:"title"`
	Body       string          `json:"body"`
	Data       json.RawMessage `json:"data,omitempty"`
	Channel    string          `json:"channel"`       // "in_app", "email", "both"
}

// Consumer listens to NATS events and creates notifications + sends emails.
type Consumer struct {
	nc       *nats.Conn
	svc      *NotificationService
	mailer   Mailer
	subs     []*nats.Subscription
}

// Mailer sends email notifications.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// NewConsumer creates a NATS event consumer.
func NewConsumer(nc *nats.Conn, svc *NotificationService, mailer Mailer) *Consumer {
	return &Consumer{
		nc:     nc,
		svc:    svc,
		mailer: mailer,
	}
}

// Start subscribes to notification events from all services.
// Pattern: tenant.*.notify.> — services publish here when they want to trigger a notification.
func (c *Consumer) Start() error {
	sub, err := c.nc.Subscribe("tenant.*.notify.>", c.handleEvent)
	if err != nil {
		return err
	}
	c.subs = append(c.subs, sub)
	slog.Info("notification consumer started", "subject", "tenant.*.notify.>")
	return nil
}

// Stop unsubscribes from all NATS subjects.
func (c *Consumer) Stop() {
	for _, sub := range c.subs {
		sub.Unsubscribe()
	}
}

func (c *Consumer) handleEvent(msg *nats.Msg) {
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		slog.Warn("invalid notification event", "error", err, "subject", msg.Subject)
		return
	}

	if evt.UserID == "" || evt.Type == "" || evt.Title == "" {
		slog.Warn("notification event missing required fields", "subject", msg.Subject)
		return
	}

	ctx := context.Background()

	// Check user preferences
	prefs, err := c.svc.GetPreferences(ctx, evt.UserID)
	if err != nil {
		slog.Error("failed to get preferences", "error", err, "user_id", evt.UserID)
		// Continue with defaults
		prefs = &Preferences{EmailEnabled: true, InAppEnabled: true, MutedTypes: []string{}}
	}

	// Skip if user muted this type
	if slices.Contains(prefs.MutedTypes, evt.Type) {
		slog.Debug("notification muted by user", "type", evt.Type, "user_id", evt.UserID)
		return
	}

	channel := evt.Channel
	if channel == "" {
		channel = "in_app"
	}

	// Persist in-app notification
	if prefs.InAppEnabled && (channel == "in_app" || channel == "both") {
		notif, err := c.svc.Create(ctx, evt.UserID, evt.Type, evt.Title, evt.Body, evt.Data, channel)
		if err != nil {
			slog.Error("failed to create notification", "error", err, "user_id", evt.UserID)
		} else {
			// Publish to WS Hub for real-time push
			c.publishToWS(evt.TenantSlug, notif)
		}
	}

	// Send email
	if prefs.EmailEnabled && (channel == "email" || channel == "both") && c.mailer != nil {
		// Look up user email — for now we include it in the event data
		email := extractEmail(evt.Data)
		if email != "" {
			if err := c.mailer.Send(ctx, email, evt.Title, evt.Body); err != nil {
				slog.Error("failed to send email", "error", err, "user_id", evt.UserID)
			}
		}
	}
}

// publishToWS publishes a notification event to the WS Hub via NATS.
// The WS Hub subscribes to tenant.*.> and forwards to WebSocket clients.
func (c *Consumer) publishToWS(tenantSlug string, notif *Notification) {
	if tenantSlug == "" {
		return
	}

	payload, err := json.Marshal(map[string]any{
		"type":    "event",
		"channel": "notifications",
		"data":    notif,
	})
	if err != nil {
		slog.Error("failed to marshal ws notification", "error", err)
		return
	}

	subject := "tenant." + tenantSlug + ".notifications"
	if err := c.nc.Publish(subject, payload); err != nil {
		slog.Error("failed to publish to WS", "error", err, "subject", subject)
	}
}

func extractEmail(data json.RawMessage) string {
	if data == nil {
		return ""
	}
	var d map[string]any
	if err := json.Unmarshal(data, &d); err != nil {
		return ""
	}
	if email, ok := d["email"].(string); ok {
		return email
	}
	return ""
}
