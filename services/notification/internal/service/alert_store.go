package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InfraAlert represents a persisted infrastructure alert from Alertmanager.
type InfraAlert struct {
	Fingerprint string
	Status      string // "firing" or "resolved"
	Severity    string // "critical", "warning", "info"
	AlertName   string
	Service     string // service_name label
	Summary     string
	Description string
	Labels      json.RawMessage
	Annotations json.RawMessage
	StartsAt    time.Time
	EndsAt      *time.Time
}

// PgAlertStore persists infrastructure alerts in the platform database.
type PgAlertStore struct {
	pool *pgxpool.Pool
}

// NewPgAlertStore creates an alert store backed by the platform DB.
func NewPgAlertStore(pool *pgxpool.Pool) *PgAlertStore {
	return &PgAlertStore{pool: pool}
}

// SaveAlert persists an alert from Alertmanager into the infra_alerts table.
func (s *PgAlertStore) SaveAlert(ctx context.Context, alert InfraAlert) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO infra_alerts (fingerprint, status, severity, alertname, service, summary, description, labels, annotations, starts_at, ends_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		alert.Fingerprint, alert.Status, alert.Severity, alert.AlertName,
		alert.Service, alert.Summary, alert.Description,
		alert.Labels, alert.Annotations,
		alert.StartsAt, alert.EndsAt,
	)
	if err != nil {
		return fmt.Errorf("save alert: %w", err)
	}
	return nil
}
