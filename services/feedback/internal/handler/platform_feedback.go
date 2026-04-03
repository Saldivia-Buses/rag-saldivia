package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlatformFeedback handles platform-admin feedback endpoints (read-only, cross-tenant).
type PlatformFeedback struct {
	platformDB *pgxpool.Pool
}

// NewPlatformFeedback creates platform feedback HTTP handlers.
func NewPlatformFeedback(platformDB *pgxpool.Pool) *PlatformFeedback {
	return &PlatformFeedback{platformDB: platformDB}
}

// Routes returns the chi router for /v1/platform/feedback/*.
func (h *PlatformFeedback) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/tenants", h.Tenants)
	r.Get("/alerts", h.Alerts)
	r.Get("/quality", h.Quality)
	return r
}

// Tenants handles GET /v1/platform/feedback/tenants
// Returns health scores for all tenants.
func (h *PlatformFeedback) Tenants(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("X-User-Role")
	slug := r.Header.Get("X-Tenant-Slug")
	if role != "admin" || (slug != "" && slug != "platform") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "platform admin required"})
		return
	}

	ctx := r.Context()
	rows, err := h.platformDB.Query(ctx,
		`SELECT hs.tenant_id, t.name, hs.overall_score,
			hs.ai_quality_score, hs.error_rate_score, hs.performance_score,
			hs.security_score, hs.usage_score, hs.nps_score, hs.period,
			(SELECT COUNT(*) FROM feedback_alerts fa WHERE fa.tenant_id = hs.tenant_id AND fa.status = 'active')
		 FROM tenant_health_scores hs
		 JOIN tenants t ON t.id = hs.tenant_id
		 WHERE hs.period = (SELECT MAX(period) FROM tenant_health_scores WHERE tenant_id = hs.tenant_id)
		 ORDER BY hs.overall_score ASC`, // worst first
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	var tenants []map[string]any
	for rows.Next() {
		var tenantID, name string
		var overall, ai, errors, perf, security, usage float64
		var nps *float64
		var period string
		var activeAlerts int

		rows.Scan(&tenantID, &name, &overall, &ai, &errors, &perf, &security, &usage, &nps, &period, &activeAlerts)
		tenants = append(tenants, map[string]any{
			"tenant_id":         tenantID,
			"tenant_name":       name,
			"overall_score":     overall,
			"ai_quality_score":  ai,
			"error_rate_score":  errors,
			"performance_score": perf,
			"security_score":    security,
			"usage_score":       usage,
			"nps_score":         nps,
			"active_alerts":     activeAlerts,
			"last_updated":      period,
		})
	}

	if tenants == nil {
		tenants = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, tenants)
}

// Alerts handles GET /v1/platform/feedback/alerts
func (h *PlatformFeedback) Alerts(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("X-User-Role")
	slug := r.Header.Get("X-Tenant-Slug")
	if role != "admin" || (slug != "" && slug != "platform") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "platform admin required"})
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}
	limit := parseIntParam(r.URL.Query().Get("limit"), 50, 200)
	ctx := r.Context()

	rows, err := h.platformDB.Query(ctx,
		`SELECT a.id, a.tenant_id, t.name, a.alert_type, a.severity, a.module,
			a.title, a.description, a.threshold, a.current_value, a.status,
			a.created_at, a.resolved_at
		 FROM feedback_alerts a
		 JOIN tenants t ON t.id = a.tenant_id
		 WHERE a.status = $1
		 ORDER BY a.created_at DESC LIMIT $2`, status, limit,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	var alerts []map[string]any
	for rows.Next() {
		var id, tenantID, tenantName, alertType, severity, title, desc, st string
		var module, threshold, currentVal *string
		var createdAt string
		var resolvedAt *string

		rows.Scan(&id, &tenantID, &tenantName, &alertType, &severity, &module,
			&title, &desc, &threshold, &currentVal, &st, &createdAt, &resolvedAt)
		alerts = append(alerts, map[string]any{
			"id": id, "tenant_id": tenantID, "tenant_name": tenantName,
			"alert_type": alertType, "severity": severity, "module": module,
			"title": title, "description": desc, "threshold": threshold,
			"current_value": currentVal, "status": st,
			"created_at": createdAt, "resolved_at": resolvedAt,
		})
	}

	if alerts == nil {
		alerts = []map[string]any{}
	}

	// Also return summary counts
	var activeCount, warningCount, criticalCount int
	h.platformDB.QueryRow(ctx,
		`SELECT COUNT(*),
			COUNT(*) FILTER (WHERE severity = 'warning'),
			COUNT(*) FILTER (WHERE severity = 'critical')
		 FROM feedback_alerts WHERE status = 'active'`,
	).Scan(&activeCount, &warningCount, &criticalCount)

	writeJSON(w, http.StatusOK, map[string]any{
		"summary": map[string]int{
			"active":   activeCount,
			"warning":  warningCount,
			"critical": criticalCount,
		},
		"alerts": alerts,
	})
}

// Quality handles GET /v1/platform/feedback/quality
// Cross-tenant AI quality comparison.
func (h *PlatformFeedback) Quality(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("X-User-Role")
	slug := r.Header.Get("X-Tenant-Slug")
	if role != "admin" || (slug != "" && slug != "platform") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "platform admin required"})
		return
	}

	hours := parsePeriod(r.URL.Query().Get("period"))
	ctx := r.Context()

	rows, err := h.platformDB.Query(ctx,
		`SELECT fm.tenant_id, t.name, fm.module,
			SUM(fm.total_events), SUM(fm.positive), SUM(fm.negative),
			AVG(fm.avg_score)
		 FROM feedback_metrics fm
		 JOIN tenants t ON t.id = fm.tenant_id
		 WHERE fm.category IN ('response_quality','agent_quality','extraction','detection')
		   AND fm.period > now() - make_interval(hours => $1)
		 GROUP BY fm.tenant_id, t.name, fm.module
		 ORDER BY SUM(fm.negative) DESC`, hours,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	type qualityRow struct {
		TenantID   string  `json:"tenant_id"`
		TenantName string  `json:"tenant_name"`
		Module     string  `json:"module"`
		Total      int     `json:"total_events"`
		Positive   int     `json:"positive"`
		Negative   int     `json:"negative"`
		AvgScore   float64 `json:"avg_score"`
	}

	var items []qualityRow
	for rows.Next() {
		var qr qualityRow
		rows.Scan(&qr.TenantID, &qr.TenantName, &qr.Module, &qr.Total, &qr.Positive, &qr.Negative, &qr.AvgScore)
		items = append(items, qr)
	}

	if items == nil {
		items = []qualityRow{}
	}

	out, _ := json.Marshal(items)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}
