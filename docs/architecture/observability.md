---
title: Observability Stack
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/otel.md
  - ../packages/metrics.md
  - ../services/traces.md
  - ../services/healthwatch.md
  - ../operations/monitoring.md
---

This document describes how SDA exports traces, metrics, and logs. Read it
before adding a new metric, changing the OTel pipeline, wiring a new
exporter, or modifying any health check — observability is how operators
detect production regressions before users do.

## Three signals, one collector

Every Go service initialises OpenTelemetry via `otel.Setup(ctx, cfg)`
(`pkg/otel/otel.go:35`) which wires:

- An OTLP **gRPC** trace exporter to the OTel Collector
  (`OTEL_EXPORTER_OTLP_ENDPOINT`, default `localhost:4317`).
- A periodic OTLP gRPC metric exporter (15 s interval).
- The W3C TraceContext + Baggage propagators as global defaults.

Every service calls `Setup` from its `pkg/server` bootstrap and registers
the returned `Shutdown` on graceful shutdown. Trace and metric pipelines
fail soft — if the collector is unreachable the service still serves
requests; signals are dropped silently.

## The pipeline

```
service ──OTLP/gRPC──> otel-collector ──┬─> Tempo (traces)
                                        ├─> Prometheus (metrics, remote-write)
                                        `─> (passthrough)
                          docker logs ──> Promtail ──> Loki
```

The full stack is defined in
`deploy/observability/docker-compose.observability.yml`:

| Component   | Image / version             | Port  | Role                |
|-------------|------------------------------|-------|---------------------|
| OTel Collector | `otel/opentelemetry-collector-contrib:0.115.0` | 4317/4318 | Receive + fan-out |
| Tempo       | `grafana/tempo:2.7.2`        | 3200  | Trace storage       |
| Prometheus  | `prom/prometheus:v3.4.0`     | 9090  | Metric storage (30d)|
| Loki        | `grafana/loki:3.5.0`         | 3100  | Log storage         |
| Promtail    | `grafana/promtail:3.5.0`     | -     | Ship Docker logs    |
| Grafana     | `grafana/grafana:11.6.0`     | 3001  | Dashboards / explore|
| Alertmanager| `prom/alertmanager:v0.27.0`  | 9093  | Alert routing       |

Alertmanager forwards alerts to the Notification service's webhook so
operators get in-app + email notifications using the same paths as user
events.

## Trace propagation

- HTTP: `otelhttp.NewTransport` is wrapped around the LLM client
  (`pkg/llm/client.go:33`) and is the recommended transport for any new
  outbound HTTP call.
- NATS: `natspub.NotifyCtx` / `BroadcastCtx` inject the active trace
  context into NATS message headers so consumers continue the parent trace
  (`pkg/nats/publisher.go:142`).
- Database: `pkg/database/pool.go:28` notes the planned hook for
  `otelpgx`. Today queries appear as child spans of HTTP handlers but do
  not produce dedicated `pgx.query` spans until that import lands.

## Application metrics

`pkg/metrics/business.go` defines the canonical business-level instruments
used across services (`metrics.QueriesTotal`, `LLMTokensTotal`,
`DocumentsIngestedTotal`, `ToolCallsTotal`, `AuthLoginsTotal`,
`NATSMessagesTotal`, `WSConnectionsActive`, `LLMRequestDuration`). Add
service-specific labels via `metric.WithAttributes`. Always include
`tenant_slug` so dashboards can drill down per tenant.

## Execution traces (different signal)

The `traces` service consumes JetStream subjects `tenant.*.traces.>` and
persists agent + astro execution traces — token counts, tool calls, costs —
into the platform DB (`services/traces/cmd/main.go:49`). These are the
**business**-level traces shown in the UI; OpenTelemetry traces are the
**system**-level ones used by operators in Tempo. Both are tenant-scoped.

## Health checks

Every service mounts `/health` via `pkg/health.New(name)` and registers
checks for its dependencies — postgres pool ping, NATS connection state,
Redis blacklist ping — plus optional extras (e.g. WS client count). The
HealthWatch service polls Prometheus and the Docker socket for system-wide
state and runs daily AI triage (`services/healthwatch/cmd/main.go:47`).
