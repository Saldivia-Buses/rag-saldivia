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
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
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

	// Meter provider — exports metrics via OTLP gRPC to OTel Collector
	metricOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		otlpmetricgrpc.WithTimeout(5 * time.Second),
	}
	if cfg.Insecure {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}
	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		slog.Warn("otel metric exporter failed, metrics disabled", "error", err)
	}

	var mp *sdkmetric.MeterProvider
	if metricExporter != nil {
		mp = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(15*time.Second),
			)),
		)
		otel.SetMeterProvider(mp)
	}

	slog.Info("otel initialized", "service", cfg.ServiceName, "endpoint", cfg.Endpoint, "metrics", metricExporter != nil)

	shutdown := func(ctx context.Context) error {
		var errs []error
		// Shutdown meters first so final metrics are captured before traces stop
		if mp != nil {
			errs = append(errs, mp.Shutdown(ctx))
		}
		errs = append(errs, tp.Shutdown(ctx))
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
		return nil
	}

	return shutdown, nil
}
