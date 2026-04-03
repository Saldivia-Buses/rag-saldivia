// Package admin provides shared logic for CLI and MCP server to manage
// the SDA Framework platform — service health, tenant queries, logs.
package admin

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ServiceStatus represents the health of a single microservice.
type ServiceStatus struct {
	Name    string        `json:"name"`
	Port    string        `json:"port"`
	Status  string        `json:"status"` // UP, DOWN, ERR(code)
	Latency time.Duration `json:"latency"`
}

// KnownServices is the canonical list of SDA services and their ports.
var KnownServices = []struct {
	Name string
	Port string
}{
	{Name: "auth", Port: "8001"},
	{Name: "ws", Port: "8002"},
	{Name: "chat", Port: "8003"},
	{Name: "rag", Port: "8004"},
	{Name: "notification", Port: "8005"},
	{Name: "platform", Port: "8006"},
	{Name: "ingest", Port: "8007"},
	{Name: "feedback", Port: "8008"},
}

// ServiceHealth checks the /health endpoint of all known services.
func ServiceHealth(baseHost string) []ServiceStatus {
	client := &http.Client{Timeout: 3 * time.Second}
	results := make([]ServiceStatus, 0, len(KnownServices))

	for _, svc := range KnownServices {
		url := fmt.Sprintf("http://%s:%s/health", baseHost, svc.Port)
		start := time.Now()
		resp, err := client.Get(url)
		latency := time.Since(start)

		st := ServiceStatus{
			Name:    svc.Name,
			Port:    svc.Port,
			Latency: latency,
		}

		if err != nil {
			st.Status = "DOWN"
			st.Latency = 0
		} else {
			if resp.StatusCode == http.StatusOK {
				st.Status = "UP"
			} else {
				st.Status = fmt.Sprintf("ERR(%d)", resp.StatusCode)
			}
			resp.Body.Close()
		}

		results = append(results, st)
	}

	return results
}

// TenantSummary represents a row from the tenants table.
type TenantSummary struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	PlanID    string    `json:"plan_id"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// TenantList queries the platform DB for all tenants.
func TenantList(platformDBURL string) ([]TenantSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, platformDBURL)
	if err != nil {
		return nil, fmt.Errorf("connect to platform db: %w", err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx,
		`SELECT id, slug, name, plan_id, enabled, created_at
		 FROM tenants
		 ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []TenantSummary
	for rows.Next() {
		var t TenantSummary
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.PlanID, &t.Enabled, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tenant row: %w", err)
		}
		tenants = append(tenants, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenants: %w", err)
	}

	return tenants, nil
}

// TenantDetail extends TenantSummary with additional runtime info.
type TenantDetail struct {
	TenantSummary
	ModuleCount int `json:"module_count"`
}

// TenantStatus gets detailed status of a specific tenant by slug.
func TenantStatus(platformDBURL, slug string) (*TenantDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, platformDBURL)
	if err != nil {
		return nil, fmt.Errorf("connect to platform db: %w", err)
	}
	defer conn.Close(ctx)

	var d TenantDetail
	err = conn.QueryRow(ctx,
		`SELECT t.id, t.slug, t.name, t.plan_id, t.enabled, t.created_at,
		        COALESCE((SELECT COUNT(*) FROM tenant_modules tm WHERE tm.tenant_id = t.id), 0)
		 FROM tenants t
		 WHERE t.slug = $1`, slug).Scan(
		&d.ID, &d.Slug, &d.Name, &d.PlanID, &d.Enabled, &d.CreatedAt, &d.ModuleCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant %q not found", slug)
		}
		return nil, fmt.Errorf("query tenant status: %w", err)
	}

	return &d, nil
}

// ServiceLogs reads the last N lines from Docker logs for a service.
// The service name is validated against KnownServices to prevent injection.
func ServiceLogs(serviceName string, lines int) (string, error) {
	// Validate service name against known services.
	valid := false
	for _, svc := range KnownServices {
		if svc.Name == serviceName {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("unknown service %q", serviceName)
	}

	if lines <= 0 {
		lines = 50
	}

	// Docker container names follow the pattern: sda-{service}-1
	containerName := fmt.Sprintf("sda-%s-1", serviceName)

	//nolint:gosec // serviceName is validated above
	cmd := exec.Command("docker", "logs", "--tail", fmt.Sprintf("%d", lines), containerName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: try service name directly (compose v1 naming)
		cmd = exec.Command("docker", "logs", "--tail", fmt.Sprintf("%d", lines), serviceName)
		out2, err2 := cmd.CombinedOutput()
		if err2 != nil {
			return "", fmt.Errorf("docker logs failed: %s (also tried %s: %s)",
				strings.TrimSpace(string(out)), containerName, strings.TrimSpace(string(out2)))
		}
		return string(out2), nil
	}

	return string(out), nil
}
