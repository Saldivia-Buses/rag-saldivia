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
	{Name: "agent", Port: "8004"},
	{Name: "notification", Port: "8005"},
	{Name: "platform", Port: "8006"},
	{Name: "ingest", Port: "8007"},
	{Name: "feedback", Port: "8008"},
	{Name: "traces", Port: "8009"},
	{Name: "search", Port: "8010"},
	{Name: "astro", Port: "8011"},
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

// DBQuery executes a read-only SELECT query against a tenant's database.
// Only SELECT statements are allowed — any other statement type is rejected.
func DBQuery(platformDBURL, tenantSlug, query string) ([]map[string]any, error) {
	query = strings.TrimSpace(query)
	upper := strings.ToUpper(query)
	if !strings.HasPrefix(upper, "SELECT") {
		return nil, fmt.Errorf("only SELECT queries are allowed")
	}
	// Reject dangerous patterns
	for _, blocked := range []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE", "CREATE", "GRANT", "REVOKE"} {
		if strings.Contains(upper, blocked) {
			return nil, fmt.Errorf("query contains blocked keyword: %s", blocked)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Resolve tenant DB URL from platform
	pConn, err := pgx.Connect(ctx, platformDBURL)
	if err != nil {
		return nil, fmt.Errorf("connect to platform db: %w", err)
	}
	defer pConn.Close(ctx)

	var tenantDBURL string
	err = pConn.QueryRow(ctx,
		`SELECT postgres_url FROM tenants WHERE slug = $1 AND enabled = true`, tenantSlug,
	).Scan(&tenantDBURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant %q not found or disabled", tenantSlug)
		}
		return nil, fmt.Errorf("resolve tenant: %w", err)
	}

	// Connect to tenant DB and execute
	tConn, err := pgx.Connect(ctx, tenantDBURL)
	if err != nil {
		return nil, fmt.Errorf("connect to tenant db: %w", err)
	}
	defer tConn.Close(ctx)

	rows, err := tConn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	cols := rows.FieldDescriptions()
	var results []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[string(col.Name)] = vals[i]
		}
		results = append(results, row)
		if len(results) >= 100 {
			break // safety limit
		}
	}
	return results, rows.Err()
}

// Deploy restarts a service via docker compose.
func Deploy(serviceName string) (string, error) {
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

	//nolint:gosec // serviceName is validated above
	cmd := exec.Command("docker", "compose", "-f", "deploy/docker-compose.prod.yml",
		"up", "-d", "--force-recreate", "--no-deps", serviceName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("deploy failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
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
	if lines > 10000 {
		lines = 10000
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
