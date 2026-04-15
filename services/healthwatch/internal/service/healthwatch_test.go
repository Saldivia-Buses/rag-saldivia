package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Camionerou/rag-saldivia/services/healthwatch/internal/collector"
)

// newTestCollectors creates collectors pointing at mock servers.
func newTestCollectors(t *testing.T) (*collector.Prometheus, *collector.Docker, *collector.Service) {
	t.Helper()

	promServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/alerts" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"alerts": []map[string]any{
						{
							"labels":   map[string]string{"alertname": "HighErrorRate", "service_name": "agent", "severity": "warning"},
							"state":    "firing",
							"activeAt": "2026-04-14T21:45:00Z",
						},
					},
				},
			})
			return
		}
		// metrics query
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"resultType": "vector",
				"result": []map[string]any{
					{"value": []any{1234567890.0, "0.003"}},
				},
			},
		})
	}))
	t.Cleanup(promServer.Close)

	dockerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"Names": []string{"/sda-auth"}, "Image": "sda-auth:0.1.0", "State": "running", "Status": "Up (healthy)"},
			{"Names": []string{"/sda-agent"}, "Image": "sda-agent:0.1.0", "State": "running", "Status": "Up (unhealthy)"},
			{"Names": []string{"/sda-postgres"}, "Image": "postgres:16", "State": "running", "Status": "Up"},
		})
	}))
	t.Cleanup(dockerServer.Close)

	svcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/health") {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/v1/info") {
			_ = json.NewEncoder(w).Encode(map[string]string{"version": "0.1.0"})
			return
		}
	}))
	t.Cleanup(svcServer.Close)

	prom := collector.NewPrometheus(promServer.URL)
	docker := collector.NewDocker(dockerServer.URL)

	// Use rewrite transport for service collector
	svcCollector := &collector.Service{}
	setTestTransport(svcCollector, svcServer.URL)

	return prom, docker, svcCollector
}

// setTestTransport sets a custom transport that rewrites URLs to the test server.
func setTestTransport(c *collector.Service, targetURL string) {
	// Access unexported httpClient field via the exported constructor pattern
	// We'll use a different approach: create a Service with the test client
	*c = *collector.NewServiceWithClient(&http.Client{
		Transport: &rewriteTransport{targetURL: targetURL},
	})
}

type rewriteTransport struct {
	targetURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	origHost := req.URL.Host
	_ = origHost
	req.URL.Scheme = "http"
	parts := strings.SplitN(t.targetURL, "://", 2)
	if len(parts) == 2 {
		req.URL.Host = parts[1]
	}
	return http.DefaultTransport.RoundTrip(req)
}

func TestSummary_Schema(t *testing.T) {
	prom, docker, svc := newTestCollectors(t)
	s := &Service{prom: prom, docker: docker, services: svc}

	summary, err := s.Summary(context.Background())
	require.NoError(t, err)
	require.NotNil(t, summary)

	// Validate top-level fields
	assert.NotZero(t, summary.Timestamp)
	assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, summary.OverallStatus)
	assert.NotNil(t, summary.Services)
	assert.NotNil(t, summary.ActiveAlerts)
	assert.NotNil(t, summary.Infrastructure)

	// Validate service status fields
	for _, svc := range summary.Services {
		assert.NotEmpty(t, svc.Name)
		assert.Contains(t, []string{"healthy", "degraded", "unhealthy", "offline"}, svc.Status)
	}

	// Validate infrastructure fields
	assert.GreaterOrEqual(t, summary.Infrastructure.ContainersTotal, 0)
	assert.NotNil(t, summary.Infrastructure.ContainersUnhealthy)
}

func TestSummary_ScrubbedOutput(t *testing.T) {
	prom, docker, svc := newTestCollectors(t)
	s := &Service{prom: prom, docker: docker, services: svc}

	summary, err := s.Summary(context.Background())
	require.NoError(t, err)

	// Serialize to JSON to check for leaks
	data, err := json.Marshal(summary)
	require.NoError(t, err)
	jsonStr := string(data)

	// M3: No raw errors, IPs, credentials, or stack traces in output
	ipPattern := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
	stackPattern := regexp.MustCompile(`goroutine \d+`)
	credPattern := regexp.MustCompile(`(?i)(password|secret|token|credential|api[_-]?key)`)
	connStrPattern := regexp.MustCompile(`(?i)(postgres://|redis://|nats://|mongodb://)`)

	assert.False(t, ipPattern.MatchString(jsonStr),
		"summary should not contain raw IP addresses: %s", jsonStr)
	assert.False(t, stackPattern.MatchString(jsonStr),
		"summary should not contain stack traces: %s", jsonStr)
	assert.False(t, credPattern.MatchString(jsonStr),
		"summary should not contain credential references: %s", jsonStr)
	assert.False(t, connStrPattern.MatchString(jsonStr),
		"summary should not contain connection strings: %s", jsonStr)

	// Should not contain raw error messages
	assert.NotContains(t, jsonStr, "panic")
	assert.NotContains(t, jsonStr, "sql:")
	assert.NotContains(t, jsonStr, "runtime error")
}

func TestSummary_OverallStatus_Degraded(t *testing.T) {
	// Test that overall status reflects the worst service status
	prom, docker, _ := newTestCollectors(t)

	// Create a service collector where one service returns degraded
	degradedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/health") {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/v1/info") {
			_ = json.NewEncoder(w).Encode(map[string]string{"version": "0.1.0"})
			return
		}
	}))
	defer degradedServer.Close()

	svcCollector := collector.NewServiceWithClient(&http.Client{
		Transport: &rewriteTransport{targetURL: degradedServer.URL},
	})

	s := &Service{prom: prom, docker: docker, services: svcCollector}

	summary, err := s.Summary(context.Background())
	require.NoError(t, err)

	// At least one service is degraded, so overall should be degraded or unhealthy
	assert.Contains(t, []string{"degraded", "unhealthy"}, summary.OverallStatus)
}
