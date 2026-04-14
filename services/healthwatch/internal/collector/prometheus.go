package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Prometheus queries the Prometheus HTTP API for service metrics.
type Prometheus struct {
	baseURL    string
	httpClient *http.Client
}

// NewPrometheus creates a Prometheus collector.
func NewPrometheus(baseURL string) *Prometheus {
	return &Prometheus{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// QueryMetrics returns error rate and p99 latency for a service.
// Service name is validated against KnownServices whitelist — NEVER user input ([M3]).
func (c *Prometheus) QueryMetrics(ctx context.Context, svc string) (*PrometheusMetrics, error) {
	if !IsKnownService(svc) {
		return nil, fmt.Errorf("unknown service: %s", svc)
	}

	metrics := &PrometheusMetrics{}

	// Query error rate
	errorRate, err := c.queryScalar(ctx,
		fmt.Sprintf(`rate(http_server_request_duration_seconds_count{service_name="%s",http_status_code=~"5.."}[5m])`, svc))
	if err == nil {
		metrics.ErrorRate5m = errorRate
	}

	// Query p99 latency
	p99, err := c.queryScalar(ctx,
		fmt.Sprintf(`histogram_quantile(0.99, rate(http_server_request_duration_seconds_bucket{service_name="%s"}[5m]))`, svc))
	if err == nil {
		metrics.P99LatencyMs = p99 * 1000 // seconds to ms
	}

	return metrics, nil
}

// QueryAlerts returns currently firing Prometheus alerts.
func (c *Prometheus) QueryAlerts(ctx context.Context) ([]PrometheusAlert, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/alerts", nil)
	if err != nil {
		return nil, fmt.Errorf("create alerts request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus alerts returned %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Alerts []struct {
				Labels   map[string]string `json:"labels"`
				State    string            `json:"state"`
				ActiveAt string            `json:"activeAt"`
			} `json:"alerts"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode alerts: %w", err)
	}

	alerts := make([]PrometheusAlert, 0)
	for _, a := range result.Data.Alerts {
		if a.State != "firing" {
			continue
		}
		activeAt, _ := time.Parse(time.RFC3339, a.ActiveAt)
		alerts = append(alerts, PrometheusAlert{
			Name:     a.Labels["alertname"],
			Service:  a.Labels["service_name"],
			Severity: a.Labels["severity"],
			ActiveAt: activeAt,
		})
	}
	return alerts, nil
}

// queryScalar runs a PromQL instant query and returns the scalar result.
func (c *Prometheus) queryScalar(ctx context.Context, query string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/query", nil)
	if err != nil {
		return 0, fmt.Errorf("create query request: %w", err)
	}
	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("query prometheus: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("prometheus returned %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Result []struct {
				Value []json.RawMessage `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode query result: %w", err)
	}

	if len(result.Data.Result) == 0 || len(result.Data.Result[0].Value) < 2 {
		return 0, nil
	}

	var val float64
	if err := json.Unmarshal(result.Data.Result[0].Value[1], &val); err != nil {
		// Prometheus returns string values like "0.001"
		var strVal string
		if err := json.Unmarshal(result.Data.Result[0].Value[1], &strVal); err != nil {
			return 0, fmt.Errorf("parse value: %w", err)
		}
		_, _ = fmt.Sscanf(strVal, "%f", &val)
	}

	return val, nil
}
