// Package outbox implements the transactional outbox pattern for the spine
// event bus. Services insert envelopes into event_outbox within the same
// database tx as the business write. A background DrainerWorker polls for
// unpublished rows, publishes to NATS, and marks them done.
//
// The pattern guarantees that every committed business write eventually
// produces a NATS event (at-least-once), and the spine consumer framework's
// idempotency (EnsureFirstDelivery) deduplicates redeliveries — achieving
// effectively-once semantics.
//
// See docs/plans/2.0.x-plan26-spine.md § Outbox drainer.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Camionerou/rag-saldivia/services/app/internal/spine"
)

// PublishTx inserts an envelope into event_outbox within the caller's tx.
// The row is NOT published to NATS yet — the DrainerWorker does that
// asynchronously after the tx commits.
//
// This MUST be called inside the same tx as the business write:
//
//	tx, _ := pool.Begin(ctx)
//	repo.InsertMessage(ctx, tx, msg)
//	outbox.PublishTx(ctx, tx, subject, env)
//	tx.Commit(ctx)
//
// If the tx rolls back (handler error, constraint violation), the outbox
// row disappears with it — no orphaned events.
func PublishTx[T any](ctx context.Context, tx pgx.Tx, subject string, env spine.Envelope[T]) error {
	if err := spine.ValidateSubject(subject); err != nil {
		return err
	}

	payload, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("outbox: marshal envelope: %w", err)
	}

	headers, err := json.Marshal(map[string]any{
		"trace_id": env.TraceID,
		"span_id":  env.SpanID,
	})
	if err != nil {
		return fmt.Errorf("outbox: marshal headers: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO event_outbox (id, subject, payload, headers)
		 VALUES ($1, $2, $3, $4)`,
		env.ID, subject, payload, headers,
	)
	if err != nil {
		return fmt.Errorf("outbox: insert into event_outbox: %w", err)
	}
	return nil
}

// Row represents a single unpublished outbox entry fetched by the drainer.
type Row struct {
	ID       uuid.UUID
	Subject  string
	Payload  json.RawMessage
	Headers  json.RawMessage
	Attempts int
}
