package spine

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"

	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
)

// Publisher publishes spine envelopes to NATS with validated subjects and
// OpenTelemetry trace context propagation.
type Publisher struct {
	nc *nats.Conn
}

// NewPublisher wraps a NATS connection for envelope publishing. The connection
// must already be established (see natspub.Connect for the project-standard
// options).
func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

// Publish serializes an envelope to JSON and publishes it to the given
// subject with OTel trace context propagated via NATS headers.
//
// The envelope's Type and SchemaVersion are authoritative — the subject is a
// separate concern (see BuildSubject). Callers should produce the subject
// from the generated constant (e.g. notify.ChatNewMessageSubject) via
// BuildSubject with the tenant slug substituted.
func Publish[T any](ctx context.Context, p *Publisher, subject string, env Envelope[T]) error {
	if p == nil {
		return fmt.Errorf("spine: nil publisher")
	}
	if err := ValidateSubject(subject); err != nil {
		return err
	}
	data, err := Encode(env)
	if err != nil {
		return fmt.Errorf("spine: encode envelope: %w", err)
	}
	msg := &nats.Msg{Subject: subject, Data: data, Header: nats.Header{}}
	natspub.InjectTraceContext(ctx, msg)
	if err := p.nc.PublishMsg(msg); err != nil {
		return fmt.Errorf("spine: publish to %s: %w", subject, err)
	}
	return nil
}
