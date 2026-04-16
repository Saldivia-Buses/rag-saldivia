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

func TestDocker_ListContainers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/containers/json", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "filters=")

		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"Names":  []string{"/sda-auth"},
				"Image":  "ghcr.io/saldivia-buses/sda-auth:0.1.0",
				"State":  "running",
				"Status": "Up 2 hours (healthy)",
				"Labels": map[string]string{"com.docker.compose.project": "sda"},
			},
			{
				"Names":  []string{"/sda-agent"},
				"Image":  "ghcr.io/saldivia-buses/sda-agent:0.1.0",
				"State":  "running",
				"Status": "Up 2 hours (unhealthy)",
				"Labels": map[string]string{"com.docker.compose.project": "sda"},
			},
			{
				"Names":  []string{"/sda-postgres"},
				"Image":  "postgres:16-alpine",
				"State":  "running",
				"Status": "Up 2 hours",
				"Labels": map[string]string{"com.docker.compose.project": "sda"},
			},
		})
	}))
	defer server.Close()

	c := NewDocker(server.URL)
	containers, err := c.ListContainers(context.Background())
	require.NoError(t, err)
	assert.Len(t, containers, 3)

	// Auth should be healthy
	assert.Equal(t, "sda-auth", containers[0].Name)
	assert.True(t, containers[0].Healthy)

	// Agent should be unhealthy
	assert.Equal(t, "sda-agent", containers[1].Name)
	assert.False(t, containers[1].Healthy)

	// Postgres should be healthy (running, no unhealthy in status)
	assert.Equal(t, "sda-postgres", containers[2].Name)
	assert.True(t, containers[2].Healthy)
}

func TestDocker_ListContainers_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]any{})
	}))
	defer server.Close()

	c := NewDocker(server.URL)
	containers, err := c.ListContainers(context.Background())
	require.NoError(t, err)
	assert.Empty(t, containers)
}

func TestDocker_ListContainers_ProxyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := NewDocker(server.URL)
	_, err := c.ListContainers(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
