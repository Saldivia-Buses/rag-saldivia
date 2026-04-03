// Package handler implements HTTP handlers for the feedback service.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Feedback handles tenant-scoped feedback endpoints (read-only).
type Feedback struct {
	tenantDB *pgxpool.Pool
}

// NewFeedback creates feedback HTTP handlers.
func NewFeedback(tenantDB *pgxpool.Pool) *Feedback {
	return &Feedback{tenantDB: tenantDB}
}

// Routes returns the chi router for /v1/feedback/*.
func (h *Feedback) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/summary", h.Summary)
	r.Get("/quality", h.Quality)
	r.Get("/errors", h.Errors)
	r.Get("/usage", h.Usage)
	r.Get("/health-score", h.HealthScore)
	return r
}

// Summary handles GET /v1/feedback/summary
func (h *Feedback) Summary(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	hours := parsePeriod(r.URL.Query().Get("period"))
	ctx := r.Context()

	// AI quality
	var positiveRate float64
	var totalFeedback int
	var avgScore float64
	h.tenantDB.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(AVG(score), 0),
			CASE WHEN COUNT(*) > 0 THEN COUNT(*) FILTER (WHERE thumbs = 'up')::float / COUNT(*)
			ELSE 0 END
		 FROM feedback_events
		 WHERE category IN ('response_quality','agent_quality','extraction','detection')
		   AND created_at > now() - make_interval(hours => $1)`, hours,
	).Scan(&totalFeedback, &avgScore, &positiveRate)

	// Errors
	var totalErrors, openErrors, criticalErrors int
	h.tenantDB.QueryRow(ctx,
		`SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'open'),
			COUNT(*) FILTER (WHERE severity = 'critical')
		 FROM feedback_events WHERE category = 'error_report'
		   AND created_at > now() - make_interval(hours => $1)`, hours,
	).Scan(&totalErrors, &openErrors, &criticalErrors)

	// Feature requests
	var totalFeatures, openFeatures int
	h.tenantDB.QueryRow(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'open')
		 FROM feedback_events WHERE category = 'feature_request'
		   AND created_at > now() - make_interval(hours => $1)`, hours,
	).Scan(&totalFeatures, &openFeatures)

	// NPS (30 day rolling)
	var npsScore float64
	var npsResponses int
	h.tenantDB.QueryRow(ctx,
		`SELECT COUNT(*),
			COALESCE(
				(COUNT(*) FILTER (WHERE score >= 9)::float - COUNT(*) FILTER (WHERE score < 7)::float)
				/ NULLIF(COUNT(*), 0) * 100, 0)
		 FROM feedback_events WHERE category = 'nps'
		   AND created_at > now() - interval '30 days'`,
	).Scan(&npsResponses, &npsScore)

	writeJSON(w, http.StatusOK, map[string]any{
		"period": r.URL.Query().Get("period"),
		"ai_quality": map[string]any{
			"total_feedback": totalFeedback,
			"positive_rate":  positiveRate,
			"avg_score":      avgScore,
		},
		"errors": map[string]any{
			"total":    totalErrors,
			"open":     openErrors,
			"critical": criticalErrors,
		},
		"feature_requests": map[string]any{
			"total": totalFeatures,
			"open":  openFeatures,
		},
		"nps": map[string]any{
			"score":     npsScore,
			"responses": npsResponses,
		},
	})
}

// Quality handles GET /v1/feedback/quality
func (h *Feedback) Quality(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	hours := parsePeriod(r.URL.Query().Get("period"))
	module := r.URL.Query().Get("module")
	limit := parseIntParam(r.URL.Query().Get("limit"), 50)
	ctx := r.Context()

	query := `SELECT id, category, module, score, thumbs, comment, created_at
		FROM feedback_events
		WHERE category IN ('response_quality','agent_quality','extraction','detection')
		  AND created_at > now() - make_interval(hours => $1)`
	args := []any{hours}

	if module != "" {
		query += ` AND module = $2`
		args = append(args, module)
	}
	query += ` ORDER BY created_at DESC LIMIT ` + strconv.Itoa(limit)

	rows, err := h.tenantDB.Query(ctx, query, args...)
	if err != nil {
		reqID := middleware.GetReqID(ctx)
		slog.Error("quality query failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	var items []map[string]any
	for rows.Next() {
		var id, category, mod string
		var score *int
		var thumbs, comment *string
		var createdAt string
		rows.Scan(&id, &category, &mod, &score, &thumbs, &comment, &createdAt)
		items = append(items, map[string]any{
			"id": id, "category": category, "module": mod,
			"score": score, "thumbs": thumbs, "comment": comment,
			"created_at": createdAt,
		})
	}

	if items == nil {
		items = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, items)
}

// Errors handles GET /v1/feedback/errors
func (h *Feedback) Errors(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "open"
	}
	limit := parseIntParam(r.URL.Query().Get("limit"), 50)
	ctx := r.Context()

	rows, err := h.tenantDB.Query(ctx,
		`SELECT id, module, severity, status, context, comment, created_at
		 FROM feedback_events WHERE category = 'error_report' AND status = $1
		 ORDER BY created_at DESC LIMIT $2`, status, limit,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	var items []map[string]any
	for rows.Next() {
		var id, mod, sev, st string
		var context json.RawMessage
		var comment *string
		var createdAt string
		rows.Scan(&id, &mod, &sev, &st, &context, &comment, &createdAt)
		items = append(items, map[string]any{
			"id": id, "module": mod, "severity": sev, "status": st,
			"context": context, "comment": comment, "created_at": createdAt,
		})
	}

	if items == nil {
		items = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, items)
}

// Usage handles GET /v1/feedback/usage
func (h *Feedback) Usage(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	hours := parsePeriod(r.URL.Query().Get("period"))
	ctx := r.Context()

	rows, err := h.tenantDB.Query(ctx,
		`SELECT module, COUNT(*), COUNT(DISTINCT user_id)
		 FROM feedback_events WHERE category = 'usage'
		   AND created_at > now() - make_interval(hours => $1)
		 GROUP BY module ORDER BY COUNT(*) DESC`, hours,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	var items []map[string]any
	for rows.Next() {
		var module string
		var actions, users int
		rows.Scan(&module, &actions, &users)
		items = append(items, map[string]any{
			"module": module, "actions": actions, "unique_users": users,
		})
	}

	if items == nil {
		items = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, items)
}

// HealthScore handles GET /v1/feedback/health-score
func (h *Feedback) HealthScore(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	// This queries the platform DB — but for now returns a stub
	// since the handler only has tenantDB. In production, this would
	// query tenant_health_scores from platform DB.
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "health score endpoint — requires platform DB integration",
	})
}

// --- Helpers ---

func parsePeriod(s string) int {
	switch s {
	case "7d":
		return 7 * 24
	case "90d":
		return 90 * 24
	default: // 30d
		return 30 * 24
	}
}

func parseIntParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
