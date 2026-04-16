package spine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TenantPool resolves a tenant-scoped pgx pool from a tenant slug. Implemented
// by *pkg/tenant.Resolver. The interface keeps pkg/spine independent of
// pkg/tenant.
type TenantPool interface {
	PostgresPool(ctx context.Context, slug string) (*pgxpool.Pool, error)
}

// SubjectSlug returns the tenant slug embedded in a tenant-scoped subject of
// the form `tenant.{slug}.<...>`. Returns the empty string for non-tenant
// subjects (e.g. `platform.lifecycle.tenant_created`).
func SubjectSlug(subject string) string {
	parts := strings.SplitN(subject, ".", 3)
	if len(parts) < 3 || parts[0] != "tenant" {
		return ""
	}
	return parts[1]
}

// ValidateTenantMatch ensures the slug embedded in the subject matches the
// envelope's TenantID. Returns nil for non-tenant subjects (caller decides
// whether to allow). Mismatches are unrecoverable — the consumer should Term
// the message and log.
var ErrTenantMismatch = errors.New("spine: subject tenant slug != envelope tenant_id")

func ValidateTenantMatch(subject string, env Header) error {
	subjectSlug := SubjectSlug(subject)
	if subjectSlug == "" {
		return nil // platform-wide event; tenant_id may still be set but no enforcement
	}
	if env.TenantID != subjectSlug {
		return fmt.Errorf("%w: subject=%q envelope=%q", ErrTenantMismatch, subjectSlug, env.TenantID)
	}
	return nil
}

// ExtractTraceContext rebuilds an OTel context from envelope trace fields if
// present, otherwise from NATS headers (compat with non-spine producers).
// The returned context can be passed to handlers so spans nest correctly.
//
// A valid OTel SpanContext requires both a TraceID and a SpanID, so we only
// build one from envelope fields when both parse cleanly. If only TraceID is
// present we fall back to header extraction (which may still yield nothing).
func ExtractTraceContext(ctx context.Context, env Header, headers nats.Header) context.Context {
	if env.TraceID != "" && env.SpanID != "" {
		tid, terr := trace.TraceIDFromHex(env.TraceID)
		sid, serr := trace.SpanIDFromHex(env.SpanID)
		if terr == nil && serr == nil {
			sc := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    tid,
				SpanID:     sid,
				TraceFlags: trace.FlagsSampled,
				Remote:     true,
			})
			return trace.ContextWithRemoteSpanContext(ctx, sc)
		}
	}
	carrier := propagation.HeaderCarrier(headers)
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// EnsureFirstDelivery atomically marks an event_id as processed by a consumer.
// Returns (true, nil) if this is the first time we see the (event_id, consumer)
// pair (the caller should run the handler), (false, nil) if it was already
// processed (the caller should Ack and skip), or (false, err) on DB error.
//
// MUST be called inside the same tx as the handler so the dedup record and the
// handler's side-effects commit together.
func EnsureFirstDelivery(ctx context.Context, tx pgx.Tx, eventID, consumerName string) (firstTime bool, err error) {
	var inserted string
	row := tx.QueryRow(ctx,
		`INSERT INTO processed_events (event_id, consumer_name)
		 VALUES ($1, $2)
		 ON CONFLICT (event_id, consumer_name) DO NOTHING
		 RETURNING event_id`,
		eventID, consumerName,
	)
	err = row.Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("spine: insert processed_events: %w", err)
	}
	return true, nil
}

// Backoff computes the Nak delay for the n-th delivery attempt (1-indexed).
// Doubles base each attempt, capped at max. Returns base for attempt <=0.
func Backoff(attempt int, base, max time.Duration) time.Duration {
	if attempt <= 1 {
		return base
	}
	d := base
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= max {
			return max
		}
	}
	return d
}
