// Package handler implements HTTP handlers for the feedback service.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	"github.com/Camionerou/rag-saldivia/services/app/internal/core/feedback/repository"
)

// Feedback handles tenant-scoped feedback endpoints (read-only).
type Feedback struct {
	repo       *repository.Queries
	platformDB *pgxpool.Pool // for health scores (lives in platform DB)
}

// NewFeedback creates feedback HTTP handlers.
func NewFeedback(repo *repository.Queries, platformDB *pgxpool.Pool) *Feedback {
	return &Feedback{repo: repo, platformDB: platformDB}
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
		httperr.WriteError(w, r, httperr.Unauthorized("not authenticated"))
		return
	}

	hours := int32(parsePeriod(r.URL.Query().Get("period")))
	ctx := r.Context()

	// AI quality
	aiRow, err := h.repo.GetSummaryAIQuality(ctx, hours)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	// Errors
	errRow, err := h.repo.GetSummaryErrors(ctx, hours)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	// Feature requests
	featRow, err := h.repo.GetSummaryFeatures(ctx, hours)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	// NPS (30 day rolling)
	npsRow, err := h.repo.GetSummaryNPS(ctx)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"period": r.URL.Query().Get("period"),
		"ai_quality": map[string]any{
			"total_feedback": aiRow.TotalFeedback,
			"positive_rate":  aiRow.PositiveRate,
			"avg_score":      aiRow.AvgScore,
		},
		"errors": map[string]any{
			"total":    errRow.Total,
			"open":     errRow.Open,
			"critical": errRow.Critical,
		},
		"feature_requests": map[string]any{
			"total": featRow.Total,
			"open":  featRow.Open,
		},
		"nps": map[string]any{
			"score":     npsRow.Score,
			"responses": npsRow.Responses,
		},
	})
}

// Quality handles GET /v1/feedback/quality
func (h *Feedback) Quality(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("not authenticated"))
		return
	}

	hours := int32(parsePeriod(r.URL.Query().Get("period")))
	module := r.URL.Query().Get("module")
	limit := int32(parseIntParam(r.URL.Query().Get("limit"), 50, 200))
	ctx := r.Context()

	if module != "" {
		rows, err := h.repo.ListQualityEventsByModule(ctx, repository.ListQualityEventsByModuleParams{
			Hours:  hours,
			Module: module,
			Limit:  limit,
		})
		if err != nil {
			httperr.WriteError(w, r, httperr.Internal(err))
			return
		}

		items := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			items = append(items, qualityEventToMap(row.ID, row.Category, row.Module, row.Score, row.Thumbs, row.Comment, row.CreatedAt))
		}
		writeJSON(w, http.StatusOK, items)
		return
	}

	rows, err := h.repo.ListQualityEvents(ctx, repository.ListQualityEventsParams{
		Hours: hours,
		Limit: limit,
	})
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, qualityEventToMap(row.ID, row.Category, row.Module, row.Score, row.Thumbs, row.Comment, row.CreatedAt))
	}
	writeJSON(w, http.StatusOK, items)
}

// Errors handles GET /v1/feedback/errors
func (h *Feedback) Errors(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("not authenticated"))
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "open"
	}
	limit := int32(parseIntParam(r.URL.Query().Get("limit"), 50, 200))
	ctx := r.Context()

	rows, err := h.repo.ListErrorEvents(ctx, repository.ListErrorEventsParams{
		Status: status,
		Limit:  limit,
	})
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		var comment *string
		if row.Comment.Valid {
			comment = &row.Comment.String
		}
		var severity *string
		if row.Severity.Valid {
			severity = &row.Severity.String
		}
		items = append(items, map[string]any{
			"id": row.ID, "module": row.Module,
			"severity": severity, "status": row.Status,
			"context": json.RawMessage(row.Context), "comment": comment,
			"created_at": row.CreatedAt.Time,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

// Usage handles GET /v1/feedback/usage
func (h *Feedback) Usage(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("not authenticated"))
		return
	}

	hours := int32(parsePeriod(r.URL.Query().Get("period")))
	ctx := r.Context()

	rows, err := h.repo.GetUsageByModule(ctx, hours)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]any{
			"module": row.Module, "actions": row.Actions, "unique_users": row.UniqueUsers,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

// HealthScore handles GET /v1/feedback/health-score
// Queries tenant_health_scores from platform DB for the current tenant.
func (h *Feedback) HealthScore(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("not authenticated"))
		return
	}

	if h.platformDB == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal, "platform DB not configured", http.StatusServiceUnavailable))
		return
	}

	var score struct {
		OverallScore     float64 `json:"overall_score"`
		AIQualityScore   float64 `json:"ai_quality_score"`
		ErrorRateScore   float64 `json:"error_rate_score"`
		PerformanceScore float64 `json:"performance_score"`
		SecurityScore    float64 `json:"security_score"`
		UsageScore       float64 `json:"usage_score"`
		Period           string  `json:"period"`
	}

	err := h.platformDB.QueryRow(r.Context(),
		`SELECT overall_score, ai_quality_score, error_rate_score,
		        performance_score, security_score, usage_score, period::text
		 FROM tenant_health_scores
		 WHERE tenant_id = $1
		 ORDER BY period DESC LIMIT 1`,
		tenantID,
	).Scan(&score.OverallScore, &score.AIQualityScore, &score.ErrorRateScore,
		&score.PerformanceScore, &score.SecurityScore, &score.UsageScore, &score.Period)

	if err != nil {
		// No scores yet — return zeros (new tenant)
		slog.Debug("no health scores for tenant", "tenant_id", tenantID, "error", err)
		writeJSON(w, http.StatusOK, score)
		return
	}

	writeJSON(w, http.StatusOK, score)
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

func parseIntParam(s string, fallback, max int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return fallback
	}
	if v > max {
		return max
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// qualityEventToMap builds a response map from typed quality event fields.
func qualityEventToMap(id, category, module string, score pgtype.Int4, thumbs, comment pgtype.Text, createdAt pgtype.Timestamptz) map[string]any {
	m := map[string]any{
		"id":         id,
		"category":   category,
		"module":     module,
		"created_at": createdAt.Time,
	}
	if score.Valid {
		m["score"] = score.Int32
	}
	if thumbs.Valid {
		m["thumbs"] = thumbs.String
	}
	if comment.Valid {
		m["comment"] = comment.String
	}
	return m
}
