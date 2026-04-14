// Package service implements the healthwatch business logic.
// Aggregates data from collectors and persists health snapshots to platform DB.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/services/healthwatch/internal/collector"
)

// ErrCheckCooldown is returned when a manual check is triggered too soon.
var ErrCheckCooldown = errors.New("manual check on cooldown, try again later")

// HealthSummary is the scrubbed executive summary response ([M3]).
// No raw errors, IPs, credentials, or stack traces.
type HealthSummary struct {
	Timestamp      time.Time       `json:"timestamp"`
	OverallStatus  string          `json:"overall_status"`
	Services       []ServiceStatus `json:"services"`
	ActiveAlerts   []Alert         `json:"active_alerts"`
	Infrastructure *InfraStatus    `json:"infrastructure"`
}

// ServiceStatus represents health state of a single service.
type ServiceStatus struct {
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	Version        string  `json:"version,omitempty"`
	UptimeHours    float64 `json:"uptime_hours,omitempty"`
	ErrorRate5m    float64 `json:"error_rate_5m,omitempty"`
	P99LatencyMs   float64 `json:"p99_latency_ms,omitempty"`
	ActiveRequests int     `json:"active_requests,omitempty"`
}

// Alert represents an active Prometheus alert (scrubbed).
type Alert struct {
	Name     string    `json:"name"`
	Service  string    `json:"service,omitempty"`
	Severity string    `json:"severity"`
	Since    time.Time `json:"since"`
}

// InfraStatus aggregates container and host stats.
type InfraStatus struct {
	ContainersTotal     int      `json:"containers_total"`
	ContainersHealthy   int      `json:"containers_healthy"`
	ContainersUnhealthy []string `json:"containers_unhealthy"`
	DiskFreeGB          float64  `json:"disk_free_gb,omitempty"`
	MemoryUsedPct       float64  `json:"memory_used_pct,omitempty"`
}

// TriageRecord is an AI-generated triage analysis.
type TriageRecord struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Analysis    string    `json:"analysis"`
	Services    []string  `json:"services"`
	GitHubIssue *int      `json:"github_issue,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// Service handles health monitoring operations.
type Service struct {
	db        *pgxpool.Pool
	prom      *collector.Prometheus
	docker    *collector.Docker
	services  *collector.Service

	// Rate limiting for manual check trigger
	checkMu       sync.Mutex
	lastCheckTime time.Time
	checkCooldown time.Duration
}

// New creates a healthwatch service.
func New(db *pgxpool.Pool, prom *collector.Prometheus, docker *collector.Docker, services *collector.Service) *Service {
	return &Service{
		db:            db,
		prom:          prom,
		docker:        docker,
		services:      services,
		checkCooldown: 10 * time.Second,
	}
}

// Summary returns a scrubbed executive summary of system health.
func (s *Service) Summary(ctx context.Context) (*HealthSummary, error) {
	statuses, err := s.ServiceStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect service statuses: %w", err)
	}

	alerts, err := s.ActiveAlerts(ctx)
	if err != nil {
		// Alerts are best-effort — don't fail the whole summary
		slog.Warn("failed to collect alerts", "error", err)
		alerts = []Alert{}
	}

	infra, err := s.collectInfra(ctx)
	if err != nil {
		slog.Warn("failed to collect infra status", "error", err)
		infra = &InfraStatus{}
	}

	overall := "healthy"
	for _, svc := range statuses {
		if svc.Status == "unhealthy" || svc.Status == "offline" {
			overall = "unhealthy"
			break
		}
		if svc.Status == "degraded" {
			overall = "degraded"
		}
	}

	summary := &HealthSummary{
		Timestamp:      time.Now().UTC(),
		OverallStatus:  overall,
		Services:       statuses,
		ActiveAlerts:   alerts,
		Infrastructure: infra,
	}

	// Persist snapshots to DB — async to avoid blocking the response
	go s.persistSnapshots(context.WithoutCancel(ctx), statuses)

	return summary, nil
}

// ServiceStatuses returns per-service health from the service collector.
func (s *Service) ServiceStatuses(ctx context.Context) ([]ServiceStatus, error) {
	checks, err := s.services.CheckServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("check services: %w", err)
	}

	statuses := make([]ServiceStatus, 0, len(checks))
	for _, c := range checks {
		st := ServiceStatus{
			Name:    c.Name,
			Status:  c.Status,
			Version: c.Version,
		}
		// Enrich with Prometheus metrics if available
		metrics, err := s.prom.QueryMetrics(ctx, c.Name)
		if err == nil && metrics != nil {
			st.ErrorRate5m = metrics.ErrorRate5m
			st.P99LatencyMs = metrics.P99LatencyMs
		}
		statuses = append(statuses, st)
	}
	return statuses, nil
}

// ActiveAlerts returns active Prometheus alerts (scrubbed output).
func (s *Service) ActiveAlerts(ctx context.Context) ([]Alert, error) {
	promAlerts, err := s.prom.QueryAlerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}

	alerts := make([]Alert, 0, len(promAlerts))
	for _, a := range promAlerts {
		alerts = append(alerts, Alert{
			Name:     a.Name,
			Service:  a.Service,
			Severity: a.Severity,
			Since:    a.ActiveAt,
		})
	}
	return alerts, nil
}

// TriggerCheck runs a manual health check with cooldown rate limiting.
// Minimum 10s between manual triggers to prevent flooding internal services.
func (s *Service) TriggerCheck(ctx context.Context) (*HealthSummary, error) {
	s.checkMu.Lock()
	if time.Since(s.lastCheckTime) < s.checkCooldown {
		s.checkMu.Unlock()
		return nil, ErrCheckCooldown
	}
	s.lastCheckTime = time.Now()
	s.checkMu.Unlock()

	return s.Summary(ctx)
}

// ListTriageRecords returns recent triage records from the database.
// Uses raw SQL instead of sqlc — healthwatch queries the shared platform DB
// (health_snapshots, triage_records) which is not owned by this service.
// Setting up a full sqlc config for 2 simple queries is not justified.
func (s *Service) ListTriageRecords(ctx context.Context, limit int) ([]TriageRecord, error) {
	if s.db == nil {
		return []TriageRecord{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, severity, title, analysis, services, github_issue, status, created_at, resolved_at
		 FROM triage_records ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list triage records: %w", err)
	}
	defer rows.Close()

	records := make([]TriageRecord, 0)
	for rows.Next() {
		var r TriageRecord
		var githubIssue *int
		var resolvedAt *time.Time
		if err := rows.Scan(&r.ID, &r.Severity, &r.Title, &r.Analysis,
			&r.Services, &githubIssue, &r.Status, &r.CreatedAt, &resolvedAt); err != nil {
			return nil, fmt.Errorf("scan triage record: %w", err)
		}
		r.GitHubIssue = githubIssue
		r.ResolvedAt = resolvedAt
		records = append(records, r)
	}
	return records, nil
}

// collectInfra gathers infrastructure status from Docker.
func (s *Service) collectInfra(ctx context.Context) (*InfraStatus, error) {
	containers, err := s.docker.ListContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	infra := &InfraStatus{
		ContainersTotal:     len(containers),
		ContainersUnhealthy: make([]string, 0),
	}

	for _, c := range containers {
		if c.Healthy {
			infra.ContainersHealthy++
		} else {
			infra.ContainersUnhealthy = append(infra.ContainersUnhealthy, c.Name)
		}
	}

	return infra, nil
}

// persistSnapshots saves health check results to the database.
func (s *Service) persistSnapshots(ctx context.Context, statuses []ServiceStatus) {
	if s.db == nil {
		return
	}
	for _, st := range statuses {
		_, err := s.db.Exec(ctx,
			`INSERT INTO health_snapshots (service, status, version, checked_at)
			 VALUES ($1, $2, $3, now())`,
			st.Name, st.Status, st.Version)
		if err != nil {
			slog.Warn("failed to persist health snapshot", "service", st.Name, "error", err)
		}
	}
}

// StartRetentionCleanup runs a background goroutine that deletes snapshots older than 7 days.
// Runs once at startup and then every 24 hours.
func (s *Service) StartRetentionCleanup(ctx context.Context) {
	go func() {
		s.cleanupOldSnapshots(ctx)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.cleanupOldSnapshots(ctx)
			}
		}
	}()
}

func (s *Service) cleanupOldSnapshots(ctx context.Context) {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM health_snapshots WHERE checked_at < now() - interval '7 days'`)
	if err != nil {
		slog.Warn("retention cleanup failed", "error", err)
		return
	}
	if tag.RowsAffected() > 0 {
		slog.Info("retention cleanup completed", "deleted", tag.RowsAffected())
	}
}
