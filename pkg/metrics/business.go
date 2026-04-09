// Package metrics provides business-level Prometheus metrics for SDA services.
// All metrics use the OTel meter API and are exported to Prometheus via OTel Collector.
//
// Usage in services:
//
//	metrics.QueriesTotal.Add(ctx, 1, metric.WithAttributes(
//	    attribute.String("service", "astro"),
//	    attribute.String("tenant_slug", tenantSlug),
//	))
package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("sda-business")

// ── Counters ────────────────────────────────────────────────────────────

// QueriesTotal counts queries across all services.
// Labels: service, tenant_slug, domain (astro domain or "general").
var QueriesTotal metric.Int64Counter

// LLMTokensTotal counts LLM tokens consumed.
// Labels: model, direction ("input" or "output"), tenant_slug.
var LLMTokensTotal metric.Int64Counter

// DocumentsIngestedTotal counts documents processed by the ingest service.
// Labels: tenant_slug, status ("success" or "error").
var DocumentsIngestedTotal metric.Int64Counter

// ToolCallsTotal counts tool executions in the agent service.
// Labels: tool_name, status ("success" or "error"), tenant_slug.
var ToolCallsTotal metric.Int64Counter

// AuthLoginsTotal counts login attempts.
// Labels: tenant_slug, result ("success", "failed", "locked").
var AuthLoginsTotal metric.Int64Counter

// NATSMessagesTotal counts NATS messages published.
// Labels: subject_prefix, tenant_slug.
var NATSMessagesTotal metric.Int64Counter

// ── Gauges ──────────────────────────────────────────────────────────────

// WSConnectionsActive tracks active WebSocket connections.
// Labels: tenant_slug.
var WSConnectionsActive metric.Int64UpDownCounter

// ── Histograms ──────────────────────────────────────────────────────────

// LLMRequestDuration measures LLM response time.
// Labels: model, tenant_slug.
var LLMRequestDuration metric.Float64Histogram

func init() {
	var err error

	QueriesTotal, err = meter.Int64Counter("sda_queries_total",
		metric.WithDescription("Total queries across SDA services"))
	if err != nil {
		QueriesTotal, _ = meter.Int64Counter("sda_queries_total_fallback")
	}

	LLMTokensTotal, err = meter.Int64Counter("sda_llm_tokens_total",
		metric.WithDescription("Total LLM tokens consumed"))
	if err != nil {
		LLMTokensTotal, _ = meter.Int64Counter("sda_llm_tokens_total_fallback")
	}

	DocumentsIngestedTotal, err = meter.Int64Counter("sda_documents_ingested_total",
		metric.WithDescription("Total documents processed"))
	if err != nil {
		DocumentsIngestedTotal, _ = meter.Int64Counter("sda_documents_ingested_total_fallback")
	}

	ToolCallsTotal, err = meter.Int64Counter("sda_tool_calls_total",
		metric.WithDescription("Total agent tool executions"))
	if err != nil {
		ToolCallsTotal, _ = meter.Int64Counter("sda_tool_calls_total_fallback")
	}

	AuthLoginsTotal, err = meter.Int64Counter("sda_auth_logins_total",
		metric.WithDescription("Total login attempts"))
	if err != nil {
		AuthLoginsTotal, _ = meter.Int64Counter("sda_auth_logins_total_fallback")
	}

	NATSMessagesTotal, err = meter.Int64Counter("sda_nats_messages_total",
		metric.WithDescription("Total NATS messages published"))
	if err != nil {
		NATSMessagesTotal, _ = meter.Int64Counter("sda_nats_messages_total_fallback")
	}

	WSConnectionsActive, err = meter.Int64UpDownCounter("sda_ws_connections_active",
		metric.WithDescription("Active WebSocket connections"))
	if err != nil {
		WSConnectionsActive, _ = meter.Int64UpDownCounter("sda_ws_connections_active_fallback")
	}

	LLMRequestDuration, err = meter.Float64Histogram("sda_llm_request_duration_seconds",
		metric.WithDescription("LLM response time in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		LLMRequestDuration, _ = meter.Float64Histogram("sda_llm_request_duration_seconds_fallback")
	}
}
