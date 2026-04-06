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
	tracesStreamName = "TRACES"
	tracesSubject    = "tenant.*.traces.>"
)

// Consumer listens to NATS trace events and persists them via the Traces service.
type Consumer struct {
	nc     *nats.Conn
	svc    *Traces
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConsumer creates a traces NATS consumer.
func NewConsumer(nc *nats.Conn, svc *Traces) *Consumer {
	return &Consumer{nc: nc, svc: svc}
}

// Start creates a JetStream durable consumer for trace events.
func (c *Consumer) Start(ctx context.Context) error {
	js, err := jetstream.New(c.nc)
	if err != nil {
		return fmt.Errorf("create jetstream context: %w", err)
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     tracesStreamName,
		Subjects: []string{tracesSubject},
		Storage:  jetstream.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("create traces stream: %w", err)
	}

	cons, err := js.CreateOrUpdateConsumer(ctx, tracesStreamName, jetstream.ConsumerConfig{
		Durable:       "traces-consumer",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: tracesSubject,
		MaxDeliver:    5,
	})
	if err != nil {
		return fmt.Errorf("create traces consumer: %w", err)
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	go c.consumeLoop(cons)

	slog.Info("traces consumer started (JetStream durable)", "stream", tracesStreamName)
	return nil
}

// Stop cancels the consumer loop.
func (c *Consumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Consumer) consumeLoop(cons jetstream.Consumer) {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		batch, err := cons.Fetch(10, jetstream.FetchMaxWait(5e9))
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			slog.Warn("traces jetstream fetch error", "error", err)
			continue
		}

		for msg := range batch.Messages() {
			c.handleEvent(msg)
		}
	}
}

func (c *Consumer) handleEvent(msg jetstream.Msg) {
	subject := msg.Subject()

	// Extract action from subject: tenant.{slug}.traces.{action}
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		slog.Warn("invalid traces subject", "subject", subject)
		msg.Term()
		return
	}
	subjectTenant := parts[1]
	action := parts[3]

	ctx := c.ctx

	switch action {
	case "start":
		var evt TraceStartEvent
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			slog.Error("invalid trace start event", "error", err)
			msg.Term()
			return
		}
		if subjectTenant != "" && evt.TenantID != subjectTenant {
			slog.Error("tenant mismatch", "subject", subjectTenant, "payload", evt.TenantID)
			msg.Term()
			return
		}
		if err := c.svc.RecordTraceStart(ctx, evt); err != nil {
			slog.Error("record trace start failed", "error", err, "trace_id", evt.TraceID)
			msg.Nak()
			return
		}

	case "end":
		var evt TraceEndEvent
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			slog.Error("invalid trace end event", "error", err)
			msg.Term()
			return
		}
		if subjectTenant != "" && evt.TenantID != "" && evt.TenantID != subjectTenant {
			slog.Error("tenant mismatch in trace end", "subject", subjectTenant, "payload", evt.TenantID)
			msg.Term()
			return
		}
		if err := c.svc.RecordTraceEnd(ctx, evt); err != nil {
			slog.Error("record trace end failed", "error", err, "trace_id", evt.TraceID)
			msg.Nak()
			return
		}

	case "event":
		var evt TraceEvent
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			slog.Error("invalid trace event", "error", err)
			msg.Term()
			return
		}
		if subjectTenant != "" && evt.TenantID != "" && evt.TenantID != subjectTenant {
			slog.Error("tenant mismatch in trace event", "subject", subjectTenant, "payload", evt.TenantID)
			msg.Term()
			return
		}
		if err := c.svc.RecordEvent(ctx, evt); err != nil {
			slog.Error("record event failed", "error", err, "trace_id", evt.TraceID)
			msg.Nak()
			return
		}

	default:
		slog.Warn("unknown traces action", "action", action, "subject", subject)
		msg.Term()
		return
	}

	msg.Ack()
}
