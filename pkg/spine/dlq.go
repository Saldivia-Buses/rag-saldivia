package spine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// DLQEntry is the structured payload published to dlq.{stream}.{subject} when
// a consumer exhausts its MaxDeliver attempts. Healthwatch's DLQ supervisor
// (Fase 4) consumes these and persists to dead_events.
//
// The original envelope is included verbatim as a json.RawMessage so the
// supervisor can deserialize against the canonical shape without an
// additional schema layer.
type DLQEntry struct {
	OriginalSubject string              `json:"original_subject"`
	OriginalStream  string              `json:"original_stream"`
	ConsumerName    string              `json:"consumer_name"`
	DeliveryCount   uint64              `json:"delivery_count"`
	LastError       string              `json:"last_error"`
	DeadAt          time.Time           `json:"dead_at"`
	Envelope        json.RawMessage     `json:"envelope"`
	Headers         map[string][]string `json:"headers,omitempty"`
}

// PushDLQ publishes a DLQ entry to dlq.{stream}.{originalSubject} via the
// given NATS connection. Always logs via slog so the dead message is visible
// even if the publish fails (e.g. NATS down). Returns nil on logged-and-sent;
// returns the publish error if NATS is connected but rejects.
//
// Designed to be called from the consumer middleware when MaxDeliver is hit.
// nc may be nil — in that case only the slog record is emitted (test mode).
func PushDLQ(ctx context.Context, nc *nats.Conn, entry DLQEntry) error {
	slog.Error("spine: message sent to DLQ",
		"original_subject", entry.OriginalSubject,
		"stream", entry.OriginalStream,
		"consumer", entry.ConsumerName,
		"delivery_count", entry.DeliveryCount,
		"error", entry.LastError,
	)

	if nc == nil {
		return nil
	}

	subject := fmt.Sprintf("dlq.%s.%s", entry.OriginalStream, entry.OriginalSubject)
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("spine: marshal dlq entry: %w", err)
	}
	msg := &nats.Msg{Subject: subject, Data: data, Header: nats.Header{}}
	if err := nc.PublishMsg(msg); err != nil {
		return fmt.Errorf("spine: publish dlq to %s: %w", subject, err)
	}
	return nil
}
