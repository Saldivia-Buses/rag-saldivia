package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Event represents an incoming NATS event that should generate a notification.
type Event struct {
	UserID  string          `json:"user_id"` // target user (who gets notified)
	Type    string          `json:"type"`     // e.g., "chat.new_message"
	Title   string          `json:"title"`
	Body    string          `json:"body"`
	Data    json.RawMessage `json:"data,omitempty"`
	Channel string          `json:"channel"` // "in_app", "email", "both"
}

// Consumer listens to NATS events and creates notifications + sends emails.
type Consumer struct {
	nc     *nats.Conn
	svc    *NotificationService
	mailer Mailer
	cons   jetstream.Consumer
	ctx    context.Context
	cancel context.CancelFunc
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

const (
	streamName   = "NOTIFICATIONS"
	durableName  = "notification-service"
	subjectFilter = "tenant.*.notify.>"
)

// Start creates a JetStream durable consumer for guaranteed delivery.
// Events are acked only after successful processing — if the service crashes,
// NATS redelivers unacked messages on restart.
func (c *Consumer) Start(ctx context.Context) error {
	js, err := jetstream.New(c.nc)
	if err != nil {
		return fmt.Errorf("create jetstream context: %w", err)
	}

	// Create or update the stream (idempotent)
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectFilter},
		Storage:  jetstream.FileStorage,
		MaxAge:   7 * 24 * 60 * 60 * 1e9, // 7 days retention
	})
	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}

	// Create durable consumer (survives restarts)
	cons, err := js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       durableName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: subjectFilter,
	})
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}
	c.cons = cons

	c.ctx, c.cancel = context.WithCancel(ctx)

	// Consume messages in a goroutine
	go c.consumeLoop()

	slog.Info("notification consumer started (JetStream durable)", "stream", streamName, "consumer", durableName)
	return nil
}

// Stop cancels the consumer loop.
func (c *Consumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Consumer) consumeLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		batch, err := c.cons.Fetch(10, jetstream.FetchMaxWait(5e9)) // 5s max wait
		if err != nil {
			if c.ctx.Err() != nil {
				return // shutting down
			}
			slog.Warn("jetstream fetch error", "error", err)
			continue
		}

		for msg := range batch.Messages() {
			c.handleEvent(msg)
		}
	}
}

func (c *Consumer) handleEvent(msg jetstream.Msg) {
	// Extract tenant slug from the trusted NATS subject, not from JSON body
	tenantSlug := tenantFromSubject(msg.Subject())
	if tenantSlug == "" {
		slog.Warn("could not extract tenant from subject", "subject", msg.Subject())
		msg.Nak()
		return
	}

	var evt Event
	if err := json.Unmarshal(msg.Data(), &evt); err != nil {
		slog.Warn("invalid notification event", "error", err, "subject", msg.Subject())
		msg.Term() // don't redeliver malformed messages
		return
	}

	if evt.UserID == "" || evt.Type == "" || evt.Title == "" {
		slog.Warn("notification event missing required fields", "subject", msg.Subject())
		msg.Term()
		return
	}

	ctx := context.Background()

	// Check user preferences
	prefs, err := c.svc.GetPreferences(ctx, evt.UserID)
	if err != nil {
		slog.Error("failed to get preferences", "error", err, "user_id", evt.UserID)
		prefs = &Preferences{EmailEnabled: true, InAppEnabled: true, MutedTypes: []string{}}
	}

	// Skip if user muted this type
	if slices.Contains(prefs.MutedTypes, evt.Type) {
		slog.Debug("notification muted by user", "type", evt.Type, "user_id", evt.UserID)
		msg.Ack()
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
			msg.Nak() // retry later
			return
		}
		c.publishToWS(tenantSlug, notif)
	}

	// Send email
	if prefs.EmailEnabled && (channel == "email" || channel == "both") && c.mailer != nil {
		email := extractEmail(evt.Data)
		if email != "" {
			if err := c.mailer.Send(ctx, email, evt.Title, evt.Body); err != nil {
				slog.Error("failed to send email", "error", err, "user_id", evt.UserID)
				// Don't nak for email failure — notification was persisted
			}
		}
	}

	msg.Ack()
}

// publishToWS publishes a notification event to the WS Hub via NATS.
// Uses the trusted tenant slug extracted from the inbound NATS subject.
func (c *Consumer) publishToWS(tenantSlug string, notif *Notification) {
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

// tenantFromSubject extracts the tenant slug from a NATS subject.
// Subject format: tenant.{slug}.notify.{type}
func tenantFromSubject(subject string) string {
	parts := strings.SplitN(subject, ".", 4)
	if len(parts) < 3 || parts[0] != "tenant" {
		return ""
	}
	return parts[1]
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
