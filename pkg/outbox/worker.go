package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/spine"
)

// DrainerWorker drains the event_outbox for a single tenant DB. It runs as a
// background goroutine, polling for unpublished rows, publishing to NATS, and
// marking them done. Multiple replicas of the same service can run concurrently
// against the same DB — FOR UPDATE SKIP LOCKED prevents double-publish.
type DrainerWorker struct {
	pool       *pgxpool.Pool
	nc         *nats.Conn
	tenantSlug string
	pollIdle   time.Duration // 2s default
	pollActive time.Duration // 200ms default
	logger     *slog.Logger
}

// DrainerOpt configures a DrainerWorker.
type DrainerOpt func(*DrainerWorker)

// WithPollIntervals overrides poll timing (default 200ms active, 2s idle).
func WithPollIntervals(active, idle time.Duration) DrainerOpt {
	return func(d *DrainerWorker) {
		d.pollActive = active
		d.pollIdle = idle
	}
}

// NewDrainer creates a drainer for a single tenant's event_outbox.
func NewDrainer(pool *pgxpool.Pool, nc *nats.Conn, tenantSlug string, opts ...DrainerOpt) *DrainerWorker {
	d := &DrainerWorker{
		pool:       pool,
		nc:         nc,
		tenantSlug: tenantSlug,
		pollIdle:   2 * time.Second,
		pollActive: 200 * time.Millisecond,
		logger:     slog.With("component", "outbox-drainer", "tenant", tenantSlug),
	}
	for _, o := range opts {
		o(d)
	}
	return d
}

// Run starts the drain loop. Blocks until ctx is cancelled.
func (d *DrainerWorker) Run(ctx context.Context) {
	d.logger.Info("outbox drainer started")
	defer d.logger.Info("outbox drainer stopped")

	// Try to LISTEN for immediate wakeup. Falls back to poll if LISTEN fails.
	notify := make(chan struct{}, 1)
	go d.listenNotify(ctx, notify)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := d.drainBatch(ctx)
		if err != nil {
			d.logger.Error("drain batch failed", "error", err)
			sleep(ctx, d.pollIdle)
			continue
		}

		// Update gauge.
		spine.OutboxUnpublished.Add(ctx, int64(-n))

		if n > 0 {
			// More work likely — poll fast.
			sleep(ctx, d.pollActive)
		} else {
			// Idle — wait for NOTIFY or poll timeout.
			select {
			case <-ctx.Done():
				return
			case <-notify:
				// Woken by NOTIFY — drain immediately.
			case <-time.After(d.pollIdle):
			}
		}
	}
}

// drainBatch fetches up to batchSize unpublished rows, publishes each to NATS,
// and marks them published. Returns the count of successfully published rows.
func (d *DrainerWorker) drainBatch(ctx context.Context) (int, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Dynamic batch size: clamp(unpublished/100, 1, 1000).
	// For normal operation this is 1-10; for backlog it scales up.
	var unpublished int
	_ = tx.QueryRow(ctx,
		`SELECT count(*) FROM event_outbox WHERE published_at IS NULL`,
	).Scan(&unpublished)
	batchSize := unpublished / 100
	if batchSize < 1 {
		batchSize = 1
	}
	if batchSize > 1000 {
		batchSize = 1000
	}

	rows, err := tx.Query(ctx,
		`SELECT id, subject, payload, headers, attempts
		 FROM event_outbox
		 WHERE published_at IS NULL
		   AND (next_attempt_at IS NULL OR next_attempt_at <= now())
		 ORDER BY created_at ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		batchSize,
	)
	if err != nil {
		return 0, fmt.Errorf("select unpublished: %w", err)
	}
	defer rows.Close()

	var entries []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.Subject, &r.Payload, &r.Headers, &r.Attempts); err != nil {
			return 0, fmt.Errorf("scan row: %w", err)
		}
		entries = append(entries, r)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("rows iteration: %w", err)
	}

	if len(entries) == 0 {
		return 0, tx.Commit(ctx) // release locks
	}

	published := 0
	for _, e := range entries {
		msg := &nats.Msg{
			Subject: e.Subject,
			Data:    e.Payload,
			Header:  natsHeadersFromJSON(e.Headers),
		}
		if err := d.nc.PublishMsg(msg); err != nil {
			d.markFailed(ctx, tx, e, err)
			continue
		}
		if _, err := tx.Exec(ctx,
			`UPDATE event_outbox SET published_at = now() WHERE id = $1`, e.ID,
		); err != nil {
			d.logger.Error("mark published failed", "id", e.ID, "error", err)
			continue
		}
		published++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit drain batch: %w", err)
	}

	if published > 0 {
		d.logger.Debug("drained batch", "published", published, "total", len(entries))
	}
	return published, nil
}

func (d *DrainerWorker) markFailed(ctx context.Context, tx pgx.Tx, e Row, publishErr error) {
	backoff := time.Duration(math.Min(
		float64(time.Duration(1<<uint(e.Attempts))*time.Second),
		float64(60*time.Second),
	))
	_, err := tx.Exec(ctx,
		`UPDATE event_outbox
		 SET attempts = attempts + 1,
		     last_error = $2,
		     next_attempt_at = now() + $3::interval
		 WHERE id = $1`,
		e.ID, publishErr.Error(), fmt.Sprintf("%d seconds", int(backoff.Seconds())),
	)
	if err != nil {
		d.logger.Error("mark failed error", "id", e.ID, "error", err)
	}
}

func (d *DrainerWorker) listenNotify(ctx context.Context, ch chan<- struct{}) {
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		d.logger.Warn("LISTEN acquire failed, falling back to poll", "error", err)
		return
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN spine_outbox_new")
	if err != nil {
		d.logger.Warn("LISTEN failed, falling back to poll", "error", err)
		return
	}

	d.logger.Debug("LISTEN spine_outbox_new active")
	for {
		_, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			d.logger.Warn("WaitForNotification error", "error", err)
			return
		}
		select {
		case ch <- struct{}{}:
		default: // non-blocking — drainer already waking
		}
	}
}

func natsHeadersFromJSON(raw json.RawMessage) nats.Header {
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return nats.Header{}
	}
	h := nats.Header{}
	for k, v := range m {
		if v != "" {
			h.Set(k, v)
		}
	}
	return h
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
