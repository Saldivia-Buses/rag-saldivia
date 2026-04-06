package natspub

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// LogDLQ logs a dead letter event and optionally publishes to a DLQ subject.
// Call this when a message has exhausted MaxDeliver attempts.
//
// The DLQ subject format is: dlq.{stream}.{originalSubject}
// If nc is nil or publish fails, the event is still logged via slog (independent
// of notification service — works even if notification is the failing consumer).
func LogDLQ(nc *nats.Conn, stream, subject string, data []byte, err error) {
	slog.Error("message sent to DLQ",
		"stream", stream,
		"subject", subject,
		"error", err,
		"data_size", len(data),
	)

	if nc == nil {
		return
	}

	dlqSubject := fmt.Sprintf("dlq.%s.%s", stream, subject)
	if pubErr := nc.Publish(dlqSubject, data); pubErr != nil {
		slog.Error("failed to publish to DLQ", "subject", dlqSubject, "error", pubErr)
	}
}
