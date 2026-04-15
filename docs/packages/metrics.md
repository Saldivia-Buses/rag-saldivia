---
title: Package: pkg/metrics
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./otel.md
---

## Purpose

Pre-declared OpenTelemetry instruments for SDA-wide business metrics
(queries, LLM tokens, documents, tool calls, logins, NATS messages, WebSocket
connections, LLM duration). All metrics use the global `otel.Meter`; they
flow through the OTel Collector to Prometheus → Grafana. Import this so
services emit the same metric names with the same labels — dashboards
expect this consistency.

## Public API

Source: `pkg/metrics/business.go:10`

Exported instruments (all lazily initialized in `init()`):

| Name | Type | Labels |
|------|------|--------|
| `QueriesTotal` | `Int64Counter` | service, tenant_slug, domain |
| `LLMTokensTotal` | `Int64Counter` | model, direction (`input`/`output`), tenant_slug |
| `DocumentsIngestedTotal` | `Int64Counter` | tenant_slug, status |
| `ToolCallsTotal` | `Int64Counter` | tool_name, status, tenant_slug |
| `AuthLoginsTotal` | `Int64Counter` | tenant_slug, result |
| `NATSMessagesTotal` | `Int64Counter` | subject_prefix, tenant_slug |
| `WSConnectionsActive` | `Int64UpDownCounter` | tenant_slug |
| `LLMRequestDuration` | `Float64Histogram` | model, tenant_slug |

The meter name is `sda-business`.

## Usage

```go
import (
    "github.com/Camionerou/rag-saldivia/pkg/metrics"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/metric"
)

metrics.QueriesTotal.Add(ctx, 1, metric.WithAttributes(
    attribute.String("service", "astro"),
    attribute.String("tenant_slug", slug),
    attribute.String("domain", "general"),
))
```

## Invariants

- Metric names are the contract. The list at `pkg/metrics/business.go:23`
  must match Grafana dashboards. Renaming requires updating dashboards.
- `init()` (`pkg/metrics/business.go:57`) registers a fallback `*_fallback`
  counter if registration fails, so services never panic on metric init.
- Always include `tenant_slug` for tenant-aware filtering in Grafana.
- This package is import-only; constructing your own meter instances bypasses
  the SDA naming convention.

## Importers

None in production code yet. Adopting services should import this package
and remove ad-hoc meter instances.
