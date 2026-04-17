package server

// OpenTelemetry setup — tracing + metrics via OTLP gRPC. Used by server.New().
// Kept internal to this package (was pkg/otel, inlined as single-importer).

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

type otelConfig struct {
	ServiceName    string
	ServiceVersion string
	Endpoint       string
	Insecure       bool
}

type otelShutdownFn func(ctx context.Context) error

// setupOTel initializes OpenTelemetry tracing + metrics and returns a shutdown function.
// If the collector is unreachable, traces are dropped silently (no crash).
func setupOTel(ctx context.Context, cfg otelConfig) (otelShutdownFn, error) {
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

	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithBatchTimeout(5*time.Second),
		sdktrace.WithMaxExportBatchSize(512),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

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
