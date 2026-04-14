package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Service checks /health and /v1/info endpoints of all known services.
type Service struct {
	httpClient *http.Client
}

// NewService creates a service collector.
func NewService() *Service {
	return &Service{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// NewServiceWithClient creates a service collector with a custom HTTP client.
// Used in tests to inject mock transports.
func NewServiceWithClient(client *http.Client) *Service {
	return &Service{httpClient: client}
}

// CheckServices queries /health for all known services in parallel.
func (c *Service) CheckServices(ctx context.Context) ([]ServiceCheck, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup

	checks := make([]ServiceCheck, 0, len(KnownServices))

	for _, svc := range KnownServices {
		port, ok := ServicePortMap[svc]
		if !ok {
			continue
		}

		wg.Add(1)
		go func(name, port string) {
			defer wg.Done()
			check := c.checkService(ctx, name, port)
			mu.Lock()
			checks = append(checks, check)
			mu.Unlock()
		}(svc, port)
	}

	wg.Wait()
	return checks, nil
}

func (c *Service) checkService(ctx context.Context, name, port string) ServiceCheck {
	check := ServiceCheck{
		Name:   name,
		Status: "offline",
	}

	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check /health
	healthURL := fmt.Sprintf("http://%s:%s/health", name, port)
	req, err := http.NewRequestWithContext(healthCtx, http.MethodGet, healthURL, nil)
	if err != nil {
		return check
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return check
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		check.Status = "healthy"
	} else if resp.StatusCode == http.StatusServiceUnavailable {
		check.Status = "degraded"
	} else {
		check.Status = "unhealthy"
	}

	// Try to parse health response for details
	var healthResp struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err == nil {
		if healthResp.Status == "degraded" {
			check.Status = "degraded"
		}
	}

	// Check /v1/info for version
	infoURL := fmt.Sprintf("http://%s:%s/v1/info", name, port)
	infoReq, err := http.NewRequestWithContext(healthCtx, http.MethodGet, infoURL, nil)
	if err != nil {
		return check
	}

	infoResp, err := c.httpClient.Do(infoReq)
	if err != nil {
		return check
	}
	defer infoResp.Body.Close()

	var info struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(infoResp.Body).Decode(&info); err == nil {
		check.Version = info.Version
	}

	return check
}
