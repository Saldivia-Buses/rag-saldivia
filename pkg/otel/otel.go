// Package otel provides shared OpenTelemetry setup for SDA services.
// Each service calls Setup() on startup and Shutdown() on teardown.
// Traces, metrics, and logs are exported to the OTel Collector via OTLP gRPC.
package otel

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config holds the OTel configuration for a service.
type Config struct {
	ServiceName    string // e.g. "sda-auth"
	ServiceVersion string // e.g. "1.0.0"
	Endpoint       string // OTel Collector gRPC endpoint, e.g. "localhost:4317"
	Insecure       bool   // true = no TLS (default for local collector), false = require TLS
}

// Shutdown is returned by Setup and should be called on service shutdown.
type Shutdown func(ctx context.Context) error

// Setup initializes OpenTelemetry tracing and returns a shutdown function.
// If the collector is unreachable, traces are dropped silently (no crash).
func Setup(ctx context.Context, cfg Config) (Shutdown, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "localhost:4317"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	// OTLP gRPC exporter — connects to OTel Collector
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithTimeout(5 * time.Second),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otel exporter: %w", err)
	}

	// Batch span processor — buffers and sends in batches
	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithBatchTimeout(5*time.Second),
		sdktrace.WithMaxExportBatchSize(512),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("otel initialized", "service", cfg.ServiceName, "endpoint", cfg.Endpoint)

	shutdown := func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}

	return shutdown, nil
}
