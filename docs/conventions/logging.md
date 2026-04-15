---
title: Convention: Structured Logging
audience: ai
last_reviewed: 2026-04-15
related:
  - ./error-handling.md
  - ./security.md
  - ../architecture/observability.md
  - ../packages/audit.md
---

All Go services log via Go's stdlib `log/slog`. Output is JSON in production (`slog.NewJSONHandler`) and text-friendly in dev. Logs ship to Loki via the OTel collector.

## Use slog, never fmt

DO use `slog.Info`, `slog.Warn`, `slog.Error`, `slog.Debug` with structured fields:
- `slog.Info("user logged in", "user_id", userID, "tenant", slug)`
- See `pkg/httperr/httperr.go:99-101` for the canonical error/warn split.

DON'T use `log.Printf`, `fmt.Println`, or print-style logging. They produce unparseable output and miss the trace context.

DON'T concatenate values into the message string. Keys go in fields, message stays static.

## Field naming

DO use snake_case keys: `user_id`, `tenant_slug`, `request_id`, `error`, `path`, `method`, `status_code`, `duration_ms`.

DO put the dynamic value as the field, never as part of the message:
- `slog.Info("login attempt", "email", req.Email)` — searchable
- `slog.Info("login attempt for " + req.Email)` — opaque

DO include `error` as a field (not in the message): `slog.Error("publish event", "error", err, "subject", subject)`.

## Severity levels

| Level | When |
|---|---|
| `Debug` | Verbose, off in prod by default. Use for development tracing. |
| `Info` | Successful operations worth recording (login, write, deploy step). |
| `Warn` | Client errors, retried operations, expected-but-notable events. 4xx HTTP. |
| `Error` | Server errors, integration failures, unexpected conditions. 5xx HTTP. |

DO emit `Warn` for 4xx (client mistake) and `Error` for 5xx (server mistake) — `pkg/httperr` does this automatically when you use `httperr.WriteError`.

DON'T log every error at `Error`. A 401 from bad credentials is `Warn`, not `Error`.

## Request-scoped logger

DO pass the logger through `context.Context` when middleware enriches it with request-scoped fields (`request_id`, `tenant_slug`, `user_id`).

DO retrieve via `slog.Default()` for non-request code (boot, scheduled tasks, NATS consumers). Per-request enrichment lives in the chi middleware chain.

## OpenTelemetry trace IDs

DO let the OTel slog handler attach `trace_id` and `span_id` automatically. The wiring is in `pkg/otel/`. Don't manually pull and pass them.

DO start a span for every external call (DB, HTTP, NATS) so logs and traces correlate in Grafana Tempo + Loki.

## PII and secrets

This is non-negotiable. Logs ship to disk, to Loki, and to backups.

DON'T log: passwords (raw or hashed), tokens (JWT, refresh, API keys, MFA codes), full credit card numbers, complete personal documents (DNI, passport).

DO log: user IDs (UUIDs), tenant slugs, email addresses (treated as identifier here, not as content), request paths, durations, error categories.

DO redact when in doubt: `slog.Info("issued token", "user_id", uid, "expires_in_s", 900)` — never log the token itself.

DON'T dump entire request bodies. Log specific fields you need for debugging.

## Audit log is separate

DO use `pkg/audit` for compliance-grade write records (who did what, when, against what resource). Audit entries go to an immutable table in the tenant DB, not to slog.

`slog` is operational telemetry; `pkg/audit` is the legal record.

## Performance

DO use `slog.Group` or pre-built `slog.Attr` slices when emitting many fields in a hot loop. Avoid `slog.Any` for typed values where a typed helper exists (`slog.String`, `slog.Int64`, `slog.Duration`).

DON'T log inside tight inner loops. If a request publishes 10k events, log a summary, not each one.
