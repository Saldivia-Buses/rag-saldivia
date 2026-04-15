package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rewriteTransport redirects all HTTP requests to a test server,
// keeping the original path but replacing the host.
type rewriteTransport struct {
	targetURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, _ := url.Parse(t.targetURL)
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func newTestServiceCollector(serverURL string) *Service {
	return &Service{
		httpClient: &http.Client{
			Transport: &rewriteTransport{targetURL: serverURL},
		},
	}
}

func TestService_CheckServices_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/health") {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/v1/info") {
			_ = json.NewEncoder(w).Encode(map[string]string{"version": "0.1.0"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := newTestServiceCollector(server.URL)
	check := c.checkService(context.Background(), "auth", "8001")
	assert.Equal(t, "auth", check.Name)
	assert.Equal(t, "healthy", check.Status)
	assert.Equal(t, "0.1.0", check.Version)
}

func TestService_CheckServices_Offline(t *testing.T) {
	c := NewService()
	// Use a port nothing is listening on
	check := c.checkService(context.Background(), "fake-service", "19999")
	assert.Equal(t, "fake-service", check.Name)
	assert.Equal(t, "offline", check.Status)
}

func TestService_CheckServices_Degraded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer server.Close()

	c := newTestServiceCollector(server.URL)
	check := c.checkService(context.Background(), "auth", "8001")
	assert.Equal(t, "degraded", check.Status)
}

func TestService_CheckServices_Parallel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/health") {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/v1/info") {
			_ = json.NewEncoder(w).Encode(map[string]string{"version": "0.1.0"})
			return
		}
	}))
	defer server.Close()

	c := newTestServiceCollector(server.URL)
	checks, err := c.CheckServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, checks, len(KnownServices))

	// All services should be healthy via the rewrite transport
	names := make(map[string]bool)
	for _, ch := range checks {
		names[ch.Name] = true
		assert.Equal(t, "healthy", ch.Status, fmt.Sprintf("service %s should be healthy", ch.Name))
	}
	for _, svc := range KnownServices {
		assert.True(t, names[svc], fmt.Sprintf("missing check for %s", svc))
	}
}
