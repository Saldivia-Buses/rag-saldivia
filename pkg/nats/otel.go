package natspub

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

// natsHeaderCarrier adapts nats.Header to OpenTelemetry's TextMapCarrier.
// This enables trace context propagation through NATS messages.
type natsHeaderCarrier struct {
	msg *nats.Msg
}

func (c natsHeaderCarrier) Get(key string) string {
	if c.msg.Header == nil {
		return ""
	}
	return c.msg.Header.Get(key)
}

func (c natsHeaderCarrier) Set(key, val string) {
	if c.msg.Header == nil {
		c.msg.Header = nats.Header{}
	}
	c.msg.Header.Set(key, val)
}

func (c natsHeaderCarrier) Keys() []string {
	if c.msg.Header == nil {
		return nil
	}
	keys := make([]string, 0, len(c.msg.Header))
	for k := range c.msg.Header {
		keys = append(keys, k)
	}
	return keys
}

// InjectTraceContext injects the current span's trace context into a NATS message.
// Call before publishing to propagate traces across service boundaries.
func InjectTraceContext(ctx context.Context, msg *nats.Msg) {
	if msg.Header == nil {
		msg.Header = nats.Header{}
	}
	otel.GetTextMapPropagator().Inject(ctx, natsHeaderCarrier{msg})
}

// ExtractTraceContext extracts trace context from a NATS message.
// Call in consumers to link received messages to the originating trace.
func ExtractTraceContext(ctx context.Context, msg *nats.Msg) context.Context {
	if msg.Header == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, natsHeaderCarrier{msg})
}
