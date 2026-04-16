package spine

import (
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// metrics holds the Prometheus-exported counters/histograms emitted by the
// spine framework. Initialized once via initMetrics; safe to call repeatedly
// (idempotent).
var (
	spineMeter = otel.Meter("sda-spine")

	// PublishTotal counts envelope publishes.
	// Labels: service, subject, result ("success", "error").
	PublishTotal metric.Int64Counter

	// ConsumeTotal counts envelopes processed by a consumer.
	// Labels: consumer, subject, result ("success", "error", "panic", "skipped_dup", "term").
	ConsumeTotal metric.Int64Counter

	// ConsumeDuration measures handler execution time.
	// Labels: consumer, subject.
	ConsumeDuration metric.Float64Histogram

	// DLQDepth is incremented every time PushDLQ is called.
	// Labels: stream, consumer.
	DLQTotal metric.Int64Counter

	// OutboxUnpublished is the count of unpublished rows seen by the drainer
	// on its last poll. Updated as a gauge by Fase 3.
	OutboxUnpublished metric.Int64UpDownCounter
)

func init() {
	var err error
	PublishTotal, err = spineMeter.Int64Counter("spine_publish_total",
		metric.WithDescription("Spine envelope publishes by service, subject, result"),
	)
	if err != nil {
		slog.Error("spine: init PublishTotal", "error", err)
	}

	ConsumeTotal, err = spineMeter.Int64Counter("spine_consume_total",
		metric.WithDescription("Spine envelopes processed by consumer, subject, result"),
	)
	if err != nil {
		slog.Error("spine: init ConsumeTotal", "error", err)
	}

	ConsumeDuration, err = spineMeter.Float64Histogram("spine_consume_duration_seconds",
		metric.WithDescription("Spine handler execution time"),
		metric.WithUnit("s"),
	)
	if err != nil {
		slog.Error("spine: init ConsumeDuration", "error", err)
	}

	DLQTotal, err = spineMeter.Int64Counter("spine_dlq_total",
		metric.WithDescription("Spine DLQ pushes by stream, consumer"),
	)
	if err != nil {
		slog.Error("spine: init DLQTotal", "error", err)
	}

	OutboxUnpublished, err = spineMeter.Int64UpDownCounter("spine_outbox_unpublished",
		metric.WithDescription("Spine outbox rows pending publish (per tenant)"),
	)
	if err != nil {
		slog.Error("spine: init OutboxUnpublished", "error", err)
	}
}
