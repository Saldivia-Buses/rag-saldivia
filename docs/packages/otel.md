---
title: Package: pkg/otel
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ../architecture/observability.md
  - ./metrics.md
---

## Purpose

Single-call OpenTelemetry initialization for SDA services. `Setup` wires up
trace + metric exporters, a batch span processor, and the W3C TraceContext +
Baggage propagators, then returns a `Shutdown` callback for graceful
teardown. See `architecture/observability.md` for the full Tempo/Prometheus
pipeline. Most services don't import this directly — `pkg/server` calls it on
their behalf.

## Public API

Source: `pkg/otel/otel.go:4`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Config` | struct | `ServiceName`, `ServiceVersion`, `Endpoint` (default `localhost:4317`), `Insecure` |
| `Shutdown` | type | `func(ctx) error` returned by `Setup` |
| `Setup(ctx, cfg)` | func | Builds resource, OTLP gRPC exporters, batch processor, meter provider, propagator; returns `Shutdown` |

## Usage

```go
shutdown, err := otel.Setup(ctx, otel.Config{
    ServiceName:    "sda-auth",
    ServiceVersion: build.Version,
    Endpoint:       "otel-collector:4317",
    Insecure:       true,
})
if err != nil { /* traces disabled, log and continue */ }
defer shutdown(context.Background())
```

## Invariants

- If the metric exporter fails, the meter provider is left nil and metrics
  are skipped — never crash the service (`pkg/otel/otel.go:91`).
- Spans are batched: 5s batch timeout, 512 max batch size
  (`pkg/otel/otel.go:64`).
- Metrics are read on a 15s interval (`pkg/otel/otel.go:99`).
- `Shutdown` flushes meters first, then traces (`pkg/otel/otel.go:108`) so
  final metrics are exported before the trace exporter is closed.
- Propagators are `TraceContext` + `Baggage` (W3C standard) — every service
  uses the same composite, so trace IDs link end-to-end.

## Importers

`pkg/server/server.go:38` calls `Setup` automatically. Services using the
bootstrap helper (most of them) inherit it for free; only services that build
their HTTP server manually need to import `pkg/otel` directly.
