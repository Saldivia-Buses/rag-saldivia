package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	fbStreamName    = "FEEDBACK"
	fbDurableName   = "feedback-service"
	fbSubjectFilter = "tenant.*.feedback.>"
)

// validCategories maps the last segment of the NATS subject to the feedback category.
var validCategories = map[string]bool{
	"response_quality": true,
	"agent_quality":    true,
	"extraction":       true,
	"detection":        true,
	"error_report":     true,
	"feature_request":  true,
	"nps":              true,
	"usage":            true,
	"performance":      true,
	"security":         true,
}

// Consumer listens to NATS feedback events and persists them.
type Consumer struct {
	nc     *nats.Conn
	svc    *Feedback
	cons   jetstream.Consumer
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConsumer creates a NATS feedback consumer.
func NewConsumer(nc *nats.Conn, svc *Feedback) *Consumer {
	return &Consumer{nc: nc, svc: svc}
}

// Start creates a JetStream durable consumer for feedback events.
func (c *Consumer) Start(ctx context.Context) error {
	js, err := jetstream.New(c.nc)
	if err != nil {
		return fmt.Errorf("create jetstream context: %w", err)
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     fbStreamName,
		Subjects: []string{fbSubjectFilter},
		Storage:  jetstream.FileStorage,
		MaxAge:   7 * 24 * 60 * 60 * 1e9, // 7 days
	})
	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}

	cons, err := js.CreateOrUpdateConsumer(ctx, fbStreamName, jetstream.ConsumerConfig{
		Durable:       fbDurableName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: fbSubjectFilter,
		MaxDeliver:    5,
	})
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}
	c.cons = cons

	c.ctx, c.cancel = context.WithCancel(ctx)
	go c.consumeLoop()

	slog.Info("feedback consumer started (JetStream durable)",
		"stream", fbStreamName,
		"consumer", fbDurableName,
	)
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

		batch, err := c.cons.Fetch(10, jetstream.FetchMaxWait(5e9))
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			slog.Warn("feedback jetstream fetch error", "error", err)
			continue
		}

		for msg := range batch.Messages() {
			c.handleEvent(msg)
		}
	}
}

func (c *Consumer) handleEvent(msg jetstream.Msg) {
	// Extract tenant + category from subject: tenant.{slug}.feedback.{category}
	tenantSlug, category := parseSubject(msg.Subject())
	if tenantSlug == "" || category == "" {
		slog.Warn("invalid feedback subject", "subject", msg.Subject())
		msg.Term()
		return
	}

	if !validCategories[category] {
		slog.Warn("unknown feedback category", "category", category, "subject", msg.Subject())
		msg.Term()
		return
	}

	// Parse payload
	var payload map[string]any
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		slog.Warn("invalid feedback payload", "error", err, "subject", msg.Subject())
		msg.Term()
		return
	}

	// Build the feedback event
	evt := FeedbackEvent{
		Category: category,
		Module:   stringFromMap(payload, "module"),
		UserID:   stringFromMap(payload, "user_id"),
		Thumbs:   stringFromMap(payload, "thumbs"),
		Severity: stringFromMap(payload, "severity"),
		Comment:  stringFromMap(payload, "comment"),
		Context:  msg.Data(), // store the full original payload as context
	}

	// Infer module from category if not provided
	if evt.Module == "" {
		evt.Module = moduleFromCategory(category)
	}

	// Extract score if present
	if scoreVal, ok := payload["score"]; ok {
		if scoreFloat, ok := scoreVal.(float64); ok {
			score := int(scoreFloat)
			evt.Score = &score
		}
	}

	// Infer severity for error_report if not provided
	if category == "error_report" && evt.Severity == "" {
		evt.Severity = "error"
	}

	ctx := c.ctx
	if err := c.svc.RecordEvent(ctx, evt); err != nil {
		slog.Error("failed to record feedback event",
			"error", err,
			"category", category,
			"tenant", tenantSlug,
		)
		msg.Nak()
		return
	}

	slog.Debug("feedback event processed",
		"category", category,
		"tenant", tenantSlug,
		"module", evt.Module,
	)
	msg.Ack()
}

// parseSubject extracts tenant slug and category from: tenant.{slug}.feedback.{category}
func parseSubject(subject string) (tenantSlug, category string) {
	parts := strings.Split(subject, ".")
	if len(parts) < 4 || parts[0] != "tenant" || parts[2] != "feedback" {
		return "", ""
	}
	return parts[1], parts[3]
}

func stringFromMap(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func moduleFromCategory(category string) string {
	switch category {
	case "response_quality":
		return "chat"
	case "agent_quality":
		return "agent"
	case "extraction":
		return "docai"
	case "detection":
		return "vision"
	case "security":
		return "auth"
	case "performance":
		return "system"
	case "nps", "usage", "error_report", "feature_request":
		return "platform"
	default:
		return "unknown"
	}
}
