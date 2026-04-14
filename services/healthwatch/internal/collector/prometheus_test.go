package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheus_QueryMetrics_ValidService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/query", r.URL.Path)

		query := r.URL.Query().Get("query")
		assert.Contains(t, query, `service_name="auth"`)

		json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"resultType": "vector",
				"result": []map[string]any{
					{"value": []any{1234567890.0, "0.005"}},
				},
			},
		})
	}))
	defer server.Close()

	c := NewPrometheus(server.URL)
	metrics, err := c.QueryMetrics(context.Background(), "auth")
	require.NoError(t, err)
	assert.NotNil(t, metrics)
}

func TestPrometheus_QueryMetrics_RejectsUnknownService(t *testing.T) {
	// Server should never be called
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server for unknown service")
	}))
	defer server.Close()

	c := NewPrometheus(server.URL)

	tests := []struct {
		name    string
		service string
	}{
		{"unknown name", "evil-service"},
		{"sql injection", "auth'; DROP TABLE users;--"},
		{"promql injection", `auth"}[5m]) or vector(1)`},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.QueryMetrics(context.Background(), tt.service)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unknown service")
		})
	}
}

func TestPrometheus_QueryMetrics_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"resultType": "vector",
				"result":     []any{},
			},
		})
	}))
	defer server.Close()

	c := NewPrometheus(server.URL)
	metrics, err := c.QueryMetrics(context.Background(), "auth")
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, float64(0), metrics.ErrorRate5m)
}

func TestPrometheus_QueryAlerts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/alerts", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"alerts": []map[string]any{
					{
						"labels": map[string]string{
							"alertname":    "HighErrorRate",
							"service_name": "agent",
							"severity":     "warning",
						},
						"state":    "firing",
						"activeAt": "2026-04-14T21:45:00Z",
					},
					{
						"labels": map[string]string{
							"alertname":    "LowDisk",
							"service_name": "platform",
							"severity":     "info",
						},
						"state":    "pending",
						"activeAt": "2026-04-14T21:00:00Z",
					},
				},
			},
		})
	}))
	defer server.Close()

	c := NewPrometheus(server.URL)
	alerts, err := c.QueryAlerts(context.Background())
	require.NoError(t, err)

	// Only firing alerts returned
	assert.Len(t, alerts, 1)
	assert.Equal(t, "HighErrorRate", alerts[0].Name)
	assert.Equal(t, "agent", alerts[0].Service)
	assert.Equal(t, "warning", alerts[0].Severity)
}

func TestPrometheus_QueryAlerts_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewPrometheus(server.URL)
	_, err := c.QueryAlerts(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
