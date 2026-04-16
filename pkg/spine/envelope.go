// Package spine implements the event envelope, subject builder, and consumer
// framework for the SDA spine event bus.
//
// The spine wraps NATS with typed envelopes, schema versioning, and consumer
// guarantees (panic recovery, idempotency, DLQ). Callers publish and consume
// Envelope[T] values whose type T is generated from CUE specs in
// pkg/events/spec. See docs/plans/2.0.x-plan26-spine.md.
package spine

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Envelope wraps a typed payload with canonical metadata for the spine bus.
//
// The payload T is typically generated from a CUE spec (pkg/events/spec/*.cue).
// All envelopes share the same outer shape regardless of T, so a consumer can
// PeekHeader to route by Type before committing to a typed handler.
type Envelope[T any] struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Type          string     `json:"type"`
	SchemaVersion uint8      `json:"schema_version"`
	OccurredAt    time.Time  `json:"occurred_at"`
	RecordedAt    time.Time  `json:"recorded_at"`
	TraceID       string     `json:"trace_id,omitempty"`
	SpanID        string     `json:"span_id,omitempty"`
	CorrelationID string     `json:"correlation_id,omitempty"`
	CausationID   *uuid.UUID `json:"causation_id,omitempty"`
	ActorUserID   *string    `json:"actor_user_id,omitempty"`
	Payload       T          `json:"payload"`
}

// Header is the subset of envelope metadata decodable without knowing the
// payload type. Consumers use this to dispatch by Type before invoking the
// typed handler.
type Header struct {
	ID            uuid.UUID `json:"id"`
	TenantID      string    `json:"tenant_id"`
	Type          string    `json:"type"`
	SchemaVersion uint8     `json:"schema_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	RecordedAt    time.Time `json:"recorded_at"`
	TraceID       string    `json:"trace_id,omitempty"`
	SpanID        string    `json:"span_id,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

// ErrSchemaVersion is returned when an envelope's SchemaVersion does not match
// the consumer's expectation. Consumers should log and route to DLQ rather
// than attempt to decode a mismatched version.
type ErrSchemaVersion struct {
	Type     string
	Expected uint8
	Got      uint8
}

func (e *ErrSchemaVersion) Error() string {
	return fmt.Sprintf("spine: schema version mismatch for %s: expected %d, got %d", e.Type, e.Expected, e.Got)
}

// New builds an envelope with server-generated ID (UUIDv7) and timestamps
// (OccurredAt == RecordedAt == now). Callers set OccurredAt after the fact if
// the event represents a past occurrence.
func New[T any](tenantID, eventType string, schemaVersion uint8, payload T) (Envelope[T], error) {
	if tenantID == "" {
		return Envelope[T]{}, errors.New("spine: tenant_id required")
	}
	if eventType == "" {
		return Envelope[T]{}, errors.New("spine: type required")
	}
	if schemaVersion == 0 {
		return Envelope[T]{}, errors.New("spine: schema_version must be >= 1")
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Envelope[T]{}, fmt.Errorf("spine: generate uuid v7: %w", err)
	}
	now := time.Now().UTC()
	return Envelope[T]{
		ID:            id,
		TenantID:      tenantID,
		Type:          eventType,
		SchemaVersion: schemaVersion,
		OccurredAt:    now,
		RecordedAt:    now,
		Payload:       payload,
	}, nil
}

// Encode serializes an envelope to JSON for publishing on NATS.
func Encode[T any](env Envelope[T]) ([]byte, error) {
	return json.Marshal(env)
}

// Decode deserializes a raw NATS message body into a typed envelope.
//
// Returns an error if the envelope is missing required metadata (id, tenant_id,
// type, schema_version). Callers that do not yet know the payload type should
// use PeekHeader first.
func Decode[T any](data []byte) (Envelope[T], error) {
	var env Envelope[T]
	if err := json.Unmarshal(data, &env); err != nil {
		return env, fmt.Errorf("spine: unmarshal envelope: %w", err)
	}
	if env.ID == uuid.Nil {
		return env, errors.New("spine: envelope missing id")
	}
	if env.TenantID == "" {
		return env, errors.New("spine: envelope missing tenant_id")
	}
	if env.Type == "" {
		return env, errors.New("spine: envelope missing type")
	}
	if env.SchemaVersion == 0 {
		return env, errors.New("spine: envelope missing schema_version")
	}
	return env, nil
}

// PeekHeader parses only the envelope metadata, ignoring the payload. Used by
// the WS bridge and consumer framework to route by Type without committing to
// a payload type.
func PeekHeader(data []byte) (Header, error) {
	var h Header
	if err := json.Unmarshal(data, &h); err != nil {
		return h, fmt.Errorf("spine: peek header: %w", err)
	}
	if h.ID == uuid.Nil || h.Type == "" || h.SchemaVersion == 0 {
		return h, errors.New("spine: envelope header incomplete")
	}
	return h, nil
}

// CheckSchemaVersion returns *ErrSchemaVersion if env.SchemaVersion != expected.
// Consumers call this right after Decode to fail fast on schema mismatch.
func CheckSchemaVersion[T any](env Envelope[T], expected uint8) error {
	if env.SchemaVersion != expected {
		return &ErrSchemaVersion{Type: env.Type, Expected: expected, Got: env.SchemaVersion}
	}
	return nil
}
