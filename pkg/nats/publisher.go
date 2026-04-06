// Package natspub provides typed NATS event publishing for SDA services.
// Services use this to publish notification events that the Notification
// Service consumes via JetStream on tenant.*.notify.>.
package natspub

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/nats-io/nats.go"
)

// safeTokenRegex is the canonical allowlist for NATS subject tokens.
// Used by IsValidSubjectToken and should match the Python extractor's _SAFE_SUBJECT_RE.
var safeTokenRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Connect creates a NATS connection with the standard project options.
// MaxReconnects(-1), RetryOnFailedConnect, disconnect/reconnect logging.
// All services should use this instead of nats.Connect() directly.
func Connect(url string) (*nats.Conn, error) {
	return nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				slog.Warn("nats disconnected", "error", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.Info("nats reconnected", "url", nc.ConnectedUrl())
		}),
	)
}

// IsValidSubjectToken checks that a string is safe to use as a single NATS subject token.
// Uses allowlist regex: only alphanumeric, underscore, and hyphen.
// This is the canonical validation — Go services use this function,
// Python extractor mirrors the same regex (^[a-zA-Z0-9_-]+$).
func IsValidSubjectToken(s string) bool {
	return s != "" && safeTokenRegex.MatchString(s)
}

// eventTypeRegex allows dots for hierarchical event types like "chat.new_message"
// but rejects wildcards, spaces, and control characters.
var eventTypeRegex = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$`)

// IsValidEventType checks that an event type is safe for NATS subject interpolation.
// Allows dots (for chat.new_message style) unlike IsValidSubjectToken.
func IsValidEventType(s string) bool {
	return s != "" && eventTypeRegex.MatchString(s)
}

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
	if !IsValidSubjectToken(tenantSlug) {
		return fmt.Errorf("invalid tenant slug for NATS subject: %q", tenantSlug)
	}

	// Extract event type without double-serialization
	var eventType string
	switch e := evt.(type) {
	case Event:
		eventType = e.Type
	case *Event:
		eventType = e.Type
	default:
		// Fallback: marshal once, extract type from JSON
		if m, ok := evt.(map[string]any); ok {
			if t, ok := m["type"].(string); ok {
				eventType = t
			}
		}
	}
	if eventType == "" {
		return fmt.Errorf("event type is required")
	}
	if !IsValidEventType(eventType) {
		return fmt.Errorf("invalid event type for NATS subject: %q", eventType)
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := "tenant." + tenantSlug + ".notify." + eventType
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
	if !IsValidSubjectToken(tenantSlug) {
		return fmt.Errorf("invalid tenant slug for NATS subject: %q", tenantSlug)
	}
	if !IsValidSubjectToken(channel) {
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

