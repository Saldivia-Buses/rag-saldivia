// Package handler implements HTTP handlers for the feedback service.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/services/feedback/internal/repository"
)

// Feedback handles tenant-scoped feedback endpoints (read-only).
type Feedback struct {
	repo *repository.Queries
}

// NewFeedback creates feedback HTTP handlers.
func NewFeedback(repo *repository.Queries) *Feedback {
	return &Feedback{repo: repo}
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

	hours := int32(parsePeriod(r.URL.Query().Get("period")))
	ctx := r.Context()

	// AI quality
	aiRow, err := h.repo.GetSummaryAIQuality(ctx, hours)
	if err != nil {
		slog.Error("summary: ai quality query failed", "error", err)
	}

	// Errors
	errRow, err := h.repo.GetSummaryErrors(ctx, hours)
	if err != nil {
		slog.Error("summary: errors query failed", "error", err)
	}

	// Feature requests
	featRow, err := h.repo.GetSummaryFeatures(ctx, hours)
	if err != nil {
		slog.Error("summary: features query failed", "error", err)
	}

	// NPS (30 day rolling)
	npsRow, err := h.repo.GetSummaryNPS(ctx)
	if err != nil {
		slog.Error("summary: nps query failed", "error", err)
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
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
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
			reqID := middleware.GetReqID(ctx)
			slog.Error("quality query failed", "error", err, "request_id", reqID)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
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
		reqID := middleware.GetReqID(ctx)
		slog.Error("quality query failed", "error", err, "request_id", reqID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
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
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
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
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	hours := int32(parsePeriod(r.URL.Query().Get("period")))
	ctx := r.Context()

	rows, err := h.repo.GetUsageByModule(ctx, hours)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
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
	json.NewEncoder(w).Encode(v)
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
