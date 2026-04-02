// Package natspub provides typed NATS event publishing for SDA services.
// Services use this to publish notification events that the Notification
// Service consumes via JetStream on tenant.*.notify.>.
package natspub

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

// Event is the payload published to NATS for the notification service.
// The notification consumer expects this exact structure.
type Event struct {
	UserID  string          `json:"user_id"`           // target user (who gets notified)
	Type    string          `json:"type"`              // e.g., "chat.new_message"
	Title   string          `json:"title"`
	Body    string          `json:"body"`
	Data    json.RawMessage `json:"data,omitempty"`    // action-specific payload
	Channel string          `json:"channel,omitempty"` // "in_app" (default), "email", "both"
}

// Publisher wraps a NATS connection for publishing typed events.
type Publisher struct {
	nc *nats.Conn
}

// New creates a publisher. The connection should be established by the caller.
func New(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

// Notify publishes an event that the Notification Service will consume.
// Subject format: tenant.{slug}.notify.{eventType}
// evt can be an Event struct or any map/struct with a "type" field.
func (p *Publisher) Notify(tenantSlug string, evt any) error {
	if !isValidSubjectToken(tenantSlug) {
		return fmt.Errorf("invalid tenant slug for NATS subject: %q", tenantSlug)
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// Extract event type for the NATS subject
	var parsed struct{ Type string `json:"type"` }
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("extract event type: %w", err)
	}
	if parsed.Type == "" {
		return fmt.Errorf("event type is required")
	}

	subject := "tenant." + tenantSlug + ".notify." + parsed.Type
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish %s: %w", subject, err)
	}
	return nil
}

// Broadcast publishes a raw event directly to a WS Hub channel (bypassing
// the notification service). Use for real-time UI updates that don't need
// to be persisted as notifications (e.g., session list updates, typing indicators).
// Subject format: tenant.{slug}.{channel}
func (p *Publisher) Broadcast(tenantSlug, channel string, data any) error {
	if !isValidSubjectToken(tenantSlug) {
		return fmt.Errorf("invalid tenant slug for NATS subject: %q", tenantSlug)
	}
	if !isValidSubjectToken(channel) {
		return fmt.Errorf("invalid channel for NATS subject: %q", channel)
	}

	payload, err := json.Marshal(map[string]any{
		"type":    "event",
		"channel": channel,
		"data":    data,
	})
	if err != nil {
		return fmt.Errorf("marshal broadcast: %w", err)
	}

	subject := "tenant." + tenantSlug + "." + channel
	if err := p.nc.Publish(subject, payload); err != nil {
		return fmt.Errorf("broadcast %s: %w", subject, err)
	}
	return nil
}

// isValidSubjectToken checks that a string is safe to use as a NATS subject token.
// Rejects empty strings and strings containing NATS special characters.
func isValidSubjectToken(s string) bool {
	return s != "" && !strings.ContainsAny(s, ".*> \t\r\n")
}
