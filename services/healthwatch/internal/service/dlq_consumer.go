package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	dlqStreamName  = "DLQ"
	dlqSubject     = "dlq.>"
	dlqDurable     = "healthwatch-dlq"
	dlqMaxDeliver  = 5
	dlqRetention   = 30 * 24 * time.Hour // 30 days
)

// DLQConsumer reads from the DLQ JetStream stream and persists dead events
// to the platform DB. Operators interact with dead events via the admin
// handler (list, replay, drop).
type DLQConsumer struct {
	pool   *pgxpool.Pool
	nc     *nats.Conn
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDLQConsumer creates a DLQ consumer. Call Start to begin consuming.
func NewDLQConsumer(pool *pgxpool.Pool, nc *nats.Conn) *DLQConsumer {
	return &DLQConsumer{pool: pool, nc: nc}
}

// Start creates the DLQ JetStream stream + durable consumer and begins
// processing. Returns an error if setup fails; otherwise runs a goroutine
// until ctx is cancelled.
func (c *DLQConsumer) Start(ctx context.Context) error {
	js, err := jetstream.New(c.nc)
	if err != nil {
		return fmt.Errorf("dlq: create jetstream: %w", err)
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      dlqStreamName,
		Subjects:  []string{dlqSubject},
		Storage:   jetstream.FileStorage,
		MaxAge:    dlqRetention,
		Retention: jetstream.LimitsPolicy,
	})
	if err != nil {
		return fmt.Errorf("dlq: create stream: %w", err)
	}

	cons, err := js.CreateOrUpdateConsumer(ctx, dlqStreamName, jetstream.ConsumerConfig{
		Durable:       dlqDurable,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: dlqSubject,
		MaxDeliver:    dlqMaxDeliver,
		DeliverGroup:  dlqDurable, // queue subscription for HA (2+ replicas)
	})
	if err != nil {
		return fmt.Errorf("dlq: create consumer: %w", err)
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	go c.consumeLoop(cons)

	slog.Info("DLQ consumer started", "stream", dlqStreamName, "durable", dlqDurable)
	return nil
}

// Stop cancels the consume loop.
func (c *DLQConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *DLQConsumer) consumeLoop(cons jetstream.Consumer) {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		batch, err := cons.Fetch(10, jetstream.FetchMaxWait(5*time.Second))
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			slog.Warn("dlq: fetch error", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for msg := range batch.Messages() {
			c.handleDLQMessage(msg)
		}
	}
}

// dlqPayload mirrors spine.DLQEntry for deserialization without importing
// pkg/spine (keeps healthwatch's dep graph light).
type dlqPayload struct {
	OriginalSubject string              `json:"original_subject"`
	OriginalStream  string              `json:"original_stream"`
	ConsumerName    string              `json:"consumer_name"`
	DeliveryCount   int                 `json:"delivery_count"`
	LastError       string              `json:"last_error"`
	DeadAt          time.Time           `json:"dead_at"`
	Envelope        json.RawMessage     `json:"envelope"`
	Headers         map[string][]string `json:"headers,omitempty"`
}

// envelopeProbe extracts key fields from the envelope for indexing without
// fully deserializing the typed payload.
type envelopeProbe struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Type     string `json:"type"`
}

func (c *DLQConsumer) handleDLQMessage(msg jetstream.Msg) {
	var entry dlqPayload
	if err := json.Unmarshal(msg.Data(), &entry); err != nil {
		slog.Error("dlq: unmarshal failed", "error", err)
		_ = msg.Term()
		return
	}

	// Extract tenant + type from the embedded envelope.
	var probe envelopeProbe
	_ = json.Unmarshal(entry.Envelope, &probe)

	headersJSON, _ := json.Marshal(entry.Headers)

	// Use envelope event_id as dead_events.id for idempotent persistence.
	// If JetStream redelivers the same DLQ message, the ON CONFLICT skips
	// the duplicate instead of creating a second row.
	var deadID *string
	if probe.ID != "" {
		deadID = &probe.ID
	}

	_, err := c.pool.Exec(c.ctx,
		`INSERT INTO dead_events
			(id, original_subject, original_stream, consumer_name, tenant_id,
			 event_type, delivery_count, last_error, dead_at, envelope, headers)
		 VALUES (COALESCE($1::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (id) DO NOTHING`,
		deadID,
		entry.OriginalSubject,
		entry.OriginalStream,
		entry.ConsumerName,
		nilIfEmpty(probe.TenantID),
		nilIfEmpty(probe.Type),
		entry.DeliveryCount,
		entry.LastError,
		entry.DeadAt,
		entry.Envelope,
		headersJSON,
	)
	if err != nil {
		slog.Error("dlq: persist dead event failed", "error", err,
			"subject", entry.OriginalSubject, "consumer", entry.ConsumerName)
		_ = msg.Nak()
		return
	}

	_ = msg.Ack()
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
