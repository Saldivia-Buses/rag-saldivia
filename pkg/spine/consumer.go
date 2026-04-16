package spine

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// TypedHandler is the user-provided event handler. It runs inside a tx that
// already holds the processed_events insert, so callers MUST use this tx for
// any side-effect that should be atomic with the dedup record.
type TypedHandler[T any] func(ctx context.Context, tx pgx.Tx, env Envelope[T]) error

// ConsumerConfig configures a spine consumer.
type ConsumerConfig struct {
	// ConsumerName is the durable name AND the value stored in
	// processed_events.consumer_name. Must be unique per (stream, version).
	ConsumerName string

	// Stream is the JetStream stream name the consumer reads from.
	Stream string

	// FilterSubject narrows the stream to a subset (e.g. tenant.*.notify.>).
	FilterSubject string

	// MaxDeliver caps redelivery attempts before the message is sent to DLQ
	// and Term'd. Defaults to 5.
	MaxDeliver int

	// AckWait is the JetStream ack timeout per delivery. Defaults to 30s.
	AckWait time.Duration

	// BackoffBase is the Nak delay for the first failed attempt; doubles
	// each subsequent attempt up to BackoffMax. Defaults to 1s/60s.
	BackoffBase time.Duration
	BackoffMax  time.Duration

	// FetchBatch is how many messages each pull fetches. Defaults to 10.
	FetchBatch int

	// FetchWait is the max time the consumer blocks waiting for a batch.
	// Defaults to 5s.
	FetchWait time.Duration
}

func (c *ConsumerConfig) defaults() {
	if c.MaxDeliver == 0 {
		c.MaxDeliver = 5
	}
	if c.AckWait == 0 {
		c.AckWait = 30 * time.Second
	}
	if c.BackoffBase == 0 {
		c.BackoffBase = 1 * time.Second
	}
	if c.BackoffMax == 0 {
		c.BackoffMax = 60 * time.Second
	}
	if c.FetchBatch == 0 {
		c.FetchBatch = 10
	}
	if c.FetchWait == 0 {
		c.FetchWait = 5 * time.Second
	}
}

// Consume creates (or updates) a JetStream durable consumer and runs a loop
// that processes envelopes of type T. Returns an error if the consumer cannot
// be created; otherwise spawns a goroutine that runs until ctx is cancelled.
//
// Built-ins per delivery:
//  1. Panic recovery — handler panics are logged with stack and Naked.
//  2. Tenant validation — subject slug must match envelope.TenantID.
//  3. Idempotency — handler runs inside the same tx as the processed_events
//     INSERT. Duplicates are Ack'd without re-running the handler.
//  4. OTel trace context extracted from envelope (or NATS headers).
//  5. MaxDeliver exhaustion → DLQ push + Term.
//  6. Exponential backoff on Nak (BackoffBase * 2^attempt, capped at BackoffMax).
//  7. Prometheus counters: spine_consume_total{result}, spine_consume_duration_seconds.
func Consume[T any](
	ctx context.Context,
	nc *nats.Conn,
	js jetstream.JetStream,
	pool TenantPool,
	cfg ConsumerConfig,
	handler TypedHandler[T],
) error {
	cfg.defaults()

	cons, err := js.CreateOrUpdateConsumer(ctx, cfg.Stream, jetstream.ConsumerConfig{
		Durable:       cfg.ConsumerName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: cfg.FilterSubject,
		MaxDeliver:    cfg.MaxDeliver,
		AckWait:       cfg.AckWait,
	})
	if err != nil {
		return fmt.Errorf("spine: create consumer %s: %w", cfg.ConsumerName, err)
	}

	go consumeLoop(ctx, nc, cons, pool, cfg, handler)
	slog.Info("spine consumer started",
		"consumer", cfg.ConsumerName,
		"stream", cfg.Stream,
		"subject", cfg.FilterSubject,
	)
	return nil
}

func consumeLoop[T any](
	ctx context.Context,
	nc *nats.Conn,
	cons jetstream.Consumer,
	pool TenantPool,
	cfg ConsumerConfig,
	handler TypedHandler[T],
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		batch, err := cons.Fetch(cfg.FetchBatch, jetstream.FetchMaxWait(cfg.FetchWait))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Warn("spine: jetstream fetch error", "consumer", cfg.ConsumerName, "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for msg := range batch.Messages() {
			processOne(ctx, nc, pool, cfg, handler, msg)
		}
	}
}

// processOne runs the full middleware chain for a single message. Errors are
// not returned — they're handled by Nak/Term/DLQ.
func processOne[T any](
	ctx context.Context,
	nc *nats.Conn,
	pool TenantPool,
	cfg ConsumerConfig,
	handler TypedHandler[T],
	msg jetstream.Msg,
) {
	start := time.Now()
	subject := msg.Subject()

	meta, _ := msg.Metadata()
	delivery := uint64(1)
	if meta != nil {
		delivery = meta.NumDelivered
	}

	attrs := metric.WithAttributes(
		attribute.String("consumer", cfg.ConsumerName),
		attribute.String("subject", subject),
	)
	defer func() {
		ConsumeDuration.Record(ctx, time.Since(start).Seconds(), attrs)
	}()

	defer func() {
		if r := recover(); r != nil {
			ConsumeTotal.Add(ctx, 1, metric.WithAttributes(
				attribute.String("consumer", cfg.ConsumerName),
				attribute.String("subject", subject),
				attribute.String("result", "panic"),
			))
			slog.Error("spine: handler panic",
				"consumer", cfg.ConsumerName,
				"subject", subject,
				"panic", r,
				"stack", string(debug.Stack()),
				"delivery", delivery,
			)
			nakWithBackoff(msg, cfg, int(delivery))
		}
	}()

	// MaxDeliver hit — push to DLQ + Term so JetStream stops redelivering.
	if int(delivery) > cfg.MaxDeliver {
		pushDLQAndTerm(ctx, nc, cfg, msg, fmt.Sprintf("max_deliver exceeded (%d)", delivery))
		return
	}

	header, err := PeekHeader(msg.Data())
	if err != nil {
		recordResult(ctx, cfg, subject, "error")
		slog.Error("spine: peek header failed", "consumer", cfg.ConsumerName, "subject", subject, "error", err)
		_ = msg.Term()
		return
	}

	if err := ValidateTenantMatch(subject, header); err != nil {
		recordResult(ctx, cfg, subject, "tenant_mismatch")
		slog.Error("spine: tenant mismatch", "consumer", cfg.ConsumerName, "error", err)
		_ = msg.Term()
		return
	}

	env, err := Decode[T](msg.Data())
	if err != nil {
		recordResult(ctx, cfg, subject, "decode_error")
		slog.Error("spine: decode failed", "consumer", cfg.ConsumerName, "type", header.Type, "error", err)
		_ = msg.Term()
		return
	}

	handlerCtx := ExtractTraceContext(ctx, header, msg.Headers())

	tenantPool, err := pool.PostgresPool(handlerCtx, header.TenantID)
	if err != nil {
		recordResult(ctx, cfg, subject, "pool_error")
		slog.Error("spine: tenant pool resolve failed", "tenant", header.TenantID, "error", err)
		nakWithBackoff(msg, cfg, int(delivery))
		return
	}

	tx, err := tenantPool.Begin(handlerCtx)
	if err != nil {
		recordResult(ctx, cfg, subject, "tx_error")
		slog.Error("spine: begin tx failed", "tenant", header.TenantID, "error", err)
		nakWithBackoff(msg, cfg, int(delivery))
		return
	}
	defer func() { _ = tx.Rollback(handlerCtx) }()

	first, err := EnsureFirstDelivery(handlerCtx, tx, env.ID.String(), cfg.ConsumerName)
	if err != nil {
		recordResult(ctx, cfg, subject, "dedup_error")
		slog.Error("spine: ensure first delivery failed", "consumer", cfg.ConsumerName, "error", err)
		nakWithBackoff(msg, cfg, int(delivery))
		return
	}
	if !first {
		// Already processed — commit (no-op) and Ack.
		_ = tx.Commit(handlerCtx)
		recordResult(ctx, cfg, subject, "skipped_dup")
		_ = msg.Ack()
		return
	}

	if err := handler(handlerCtx, tx, env); err != nil {
		recordResult(ctx, cfg, subject, "handler_error")
		slog.Error("spine: handler error",
			"consumer", cfg.ConsumerName,
			"type", env.Type,
			"event_id", env.ID,
			"delivery", delivery,
			"error", err,
		)
		nakWithBackoff(msg, cfg, int(delivery))
		return
	}

	if err := tx.Commit(handlerCtx); err != nil {
		recordResult(ctx, cfg, subject, "commit_error")
		slog.Error("spine: commit failed", "consumer", cfg.ConsumerName, "error", err)
		nakWithBackoff(msg, cfg, int(delivery))
		return
	}

	recordResult(ctx, cfg, subject, "success")
	_ = msg.Ack()
}

func recordResult(ctx context.Context, cfg ConsumerConfig, subject, result string) {
	ConsumeTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("consumer", cfg.ConsumerName),
		attribute.String("subject", subject),
		attribute.String("result", result),
	))
}

func nakWithBackoff(msg jetstream.Msg, cfg ConsumerConfig, attempt int) {
	delay := Backoff(attempt, cfg.BackoffBase, cfg.BackoffMax)
	if err := msg.NakWithDelay(delay); err != nil {
		slog.Warn("spine: NakWithDelay failed", "error", err)
	}
}

func pushDLQAndTerm(ctx context.Context, nc *nats.Conn, cfg ConsumerConfig, msg jetstream.Msg, reason string) {
	entry := DLQEntry{
		OriginalSubject: msg.Subject(),
		OriginalStream:  cfg.Stream,
		ConsumerName:    cfg.ConsumerName,
		LastError:       reason,
		DeadAt:          time.Now().UTC(),
		Envelope:        msg.Data(),
		Headers:         msg.Headers(),
	}
	if meta, err := msg.Metadata(); err == nil && meta != nil {
		entry.DeliveryCount = meta.NumDelivered
	}
	if err := PushDLQ(ctx, nc, entry); err != nil {
		slog.Error("spine: push dlq failed", "error", err)
	}
	DLQTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("consumer", cfg.ConsumerName),
		attribute.String("stream", cfg.Stream),
	))
	recordResult(ctx, cfg, msg.Subject(), "term")
	if err := msg.Term(); err != nil {
		slog.Warn("spine: Term failed", "error", err)
	}
}

