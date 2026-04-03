package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventPublisher can publish notification events via NATS.
type EventPublisher interface {
	Notify(tenantSlug string, evt any) error
}

// Alerter checks health metrics against thresholds and creates alerts.
type Alerter struct {
	platformDB *pgxpool.Pool
	publisher  EventPublisher // nil = no notifications
}

// NewAlerter creates an alerter.
func NewAlerter(platformDB *pgxpool.Pool, publisher EventPublisher) *Alerter {
	return &Alerter{platformDB: platformDB, publisher: publisher}
}

// AlertCheck represents a threshold check to perform.
type AlertCheck struct {
	AlertType   string
	Severity    string
	Module      string
	Condition   func(scores HealthScores) bool
	Title       func(tenantSlug string) string
	Description func(tenantSlug string, scores HealthScores) string
	Threshold   string
	CurrentVal  func(scores HealthScores) string
}

// HealthScores holds the computed scores for threshold checking.
type HealthScores struct {
	Overall     float64
	AIQuality   float64
	ErrorRate   float64
	Performance float64
	Security    float64
	Usage       float64
}

// Define all threshold checks
var alertChecks = []AlertCheck{
	{
		AlertType: "quality_critical",
		Severity:  "critical",
		Condition: func(s HealthScores) bool { return s.AIQuality < 50 },
		Title:     func(slug string) string { return fmt.Sprintf("Calidad critica en tenant '%s'", slug) },
		Description: func(slug string, s HealthScores) string {
			return fmt.Sprintf("Calidad de IA en %.0f%% — mas del 50%% de respuestas negativas", s.AIQuality)
		},
		Threshold:  "ai_quality < 50",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f%%", s.AIQuality) },
	},
	{
		AlertType: "quality_drop",
		Severity:  "warning",
		Condition: func(s HealthScores) bool { return s.AIQuality < 70 && s.AIQuality >= 50 },
		Title:     func(slug string) string { return fmt.Sprintf("Calidad de IA baja en tenant '%s'", slug) },
		Description: func(slug string, s HealthScores) string {
			return fmt.Sprintf("Calidad de IA en %.0f%% — por debajo del umbral recomendado", s.AIQuality)
		},
		Threshold:  "ai_quality < 70",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f%%", s.AIQuality) },
	},
	{
		AlertType: "error_spike",
		Severity:  "warning",
		Condition: func(s HealthScores) bool { return s.ErrorRate < 50 },
		Title:     func(slug string) string { return fmt.Sprintf("Spike de errores en tenant '%s'", slug) },
		Description: func(slug string, s HealthScores) string {
			return fmt.Sprintf("Error rate score en %.0f — multiples errores en la ultima hora", s.ErrorRate)
		},
		Threshold:  "error_rate_score < 50",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f", s.ErrorRate) },
	},
	{
		AlertType: "error_critical",
		Severity:  "critical",
		Condition: func(s HealthScores) bool { return s.ErrorRate <= 30 },
		Title:     func(slug string) string { return fmt.Sprintf("Errores criticos en tenant '%s'", slug) },
		Description: func(slug string, s HealthScores) string {
			return "Errores criticos detectados en la ultima hora"
		},
		Threshold:  "error_rate_score <= 30",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f", s.ErrorRate) },
	},
	{
		AlertType: "latency_spike",
		Severity:  "warning",
		Condition: func(s HealthScores) bool { return s.Performance < 50 },
		Title:     func(slug string) string { return fmt.Sprintf("Latencia alta en tenant '%s'", slug) },
		Description: func(slug string, s HealthScores) string {
			return fmt.Sprintf("Performance score en %.0f — p95 por encima de 2s", s.Performance)
		},
		Threshold:  "performance_score < 50",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f", s.Performance) },
	},
	{
		AlertType: "security_anomaly",
		Severity:  "warning",
		Condition: func(s HealthScores) bool { return s.Security < 70 },
		Title:     func(slug string) string { return fmt.Sprintf("Actividad de seguridad inusual en '%s'", slug) },
		Description: func(slug string, s HealthScores) string {
			return "Multiples eventos de seguridad detectados en la ultima hora"
		},
		Threshold:  "security_score < 70",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f", s.Security) },
	},
	{
		AlertType: "security_critical",
		Severity:  "critical",
		Condition: func(s HealthScores) bool { return s.Security <= 30 },
		Title:     func(slug string) string { return fmt.Sprintf("Alerta de seguridad critica en '%s'", slug) },
		Description: func(slug string, s HealthScores) string {
			return "Eventos de seguridad criticos detectados"
		},
		Threshold:  "security_score <= 30",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f", s.Security) },
	},
	{
		AlertType: "inactive_tenant",
		Severity:  "info",
		Condition: func(s HealthScores) bool { return s.Usage < 25 },
		Title:     func(slug string) string { return fmt.Sprintf("Tenant '%s' con baja actividad", slug) },
		Description: func(slug string, s HealthScores) string {
			return "Actividad muy baja o nula en las ultimas horas"
		},
		Threshold:  "usage_score < 25",
		CurrentVal: func(s HealthScores) string { return fmt.Sprintf("%.0f", s.Usage) },
	},
}

// CheckAndAlert evaluates all thresholds and creates alerts as needed.
func (a *Alerter) CheckAndAlert(ctx context.Context, tenantID, tenantSlug string, scores HealthScores) {
	for _, check := range alertChecks {
		if !check.Condition(scores) {
			// Condition not met — auto-resolve if there's an active alert
			a.autoResolve(ctx, tenantID, check.AlertType)
			continue
		}

		// Check if there's already an active alert of this type
		var existingID string
		err := a.platformDB.QueryRow(ctx,
			`SELECT id FROM feedback_alerts
			 WHERE tenant_id = $1 AND alert_type = $2 AND status = 'active'
			 LIMIT 1`,
			tenantID, check.AlertType,
		).Scan(&existingID)

		if err == nil {
			// Active alert already exists — don't spam
			continue
		}
		if err != pgx.ErrNoRows {
			slog.Error("failed to check existing alert", "error", err)
			continue
		}

		// Create new alert
		title := check.Title(tenantSlug)
		desc := check.Description(tenantSlug, scores)
		currentVal := check.CurrentVal(scores)

		_, err = a.platformDB.Exec(ctx,
			`INSERT INTO feedback_alerts (tenant_id, alert_type, severity, module, title, description, threshold, current_value)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			tenantID, check.AlertType, check.Severity, check.Module,
			title, desc, check.Threshold, currentVal,
		)
		if err != nil {
			slog.Error("failed to create alert", "error", err, "type", check.AlertType)
			continue
		}

		slog.Warn("feedback alert created",
			"type", check.AlertType,
			"severity", check.Severity,
			"tenant", tenantSlug,
			"value", currentVal,
		)

		// Notify via NATS
		if a.publisher != nil {
			a.publisher.Notify(tenantSlug, map[string]any{
				"user_id": "platform-admin-broadcast",
				"type":    "feedback.alert",
				"title":   title,
				"body":    desc,
				"channel": "both",
				"data": map[string]string{
					"alert_type":    check.AlertType,
					"severity":      check.Severity,
					"tenant_id":     tenantID,
					"current_value": currentVal,
					"threshold":     check.Threshold,
				},
			})
		}
	}
}

func (a *Alerter) autoResolve(ctx context.Context, tenantID, alertType string) {
	result, err := a.platformDB.Exec(ctx,
		`UPDATE feedback_alerts
		 SET status = 'auto_resolved', resolved_at = now()
		 WHERE tenant_id = $1 AND alert_type = $2 AND status = 'active'`,
		tenantID, alertType,
	)
	if err != nil {
		slog.Error("failed to auto-resolve alert", "error", err)
		return
	}
	if result.RowsAffected() > 0 {
		slog.Info("feedback alert auto-resolved", "type", alertType, "tenant", tenantID)
	}
}
