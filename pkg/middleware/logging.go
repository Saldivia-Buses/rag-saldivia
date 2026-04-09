package middleware

import (
	"context"
	"log/slog"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type logContextKey string

const loggerKey logContextKey = "slog-logger"

// EnrichLogger is an HTTP middleware that injects tenant_id, user_id, request_id,
// and trace_id into the slog context. All downstream log calls using LoggerFromCtx
// will include these fields automatically.
//
// Usage in service main.go:
//
//	r.Use(middleware.EnrichLogger)
//
// Then in handlers:
//
//	logger := middleware.LoggerFromCtx(r.Context())
//	logger.Error("something failed", "detail", err)
//	// Output includes: tenant_id, user_id, request_id, trace_id
func EnrichLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract identifiers from headers (set by Traefik + auth middleware)
		tenantID := r.Header.Get("X-Tenant-ID")
		userID := r.Header.Get("X-User-ID")
		requestID := chimw.GetReqID(r.Context())

		// Extract trace ID from OpenTelemetry span context
		traceID := ""
		if span := trace.SpanFromContext(r.Context()); span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}

		// Build enriched logger
		logger := slog.Default().With(
			"tenant_id", tenantID,
			"user_id", userID,
			"request_id", requestID,
			"trace_id", traceID,
		)

		// Store in context
		ctx := context.WithValue(r.Context(), loggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerFromCtx retrieves the enriched logger from context.
// Returns slog.Default() if no enriched logger was set.
func LoggerFromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
