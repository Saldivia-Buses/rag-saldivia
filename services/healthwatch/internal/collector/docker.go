package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Docker queries the docker-socket-proxy for container status.
// NEVER connects to the raw Docker socket — always via proxy ([DS5]).
type Docker struct {
	proxyURL   string
	httpClient *http.Client
}

// NewDocker creates a Docker collector.
func NewDocker(proxyURL string) *Docker {
	return &Docker{
		proxyURL: proxyURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ListContainers returns SDA containers from the Docker socket proxy.
func (c *Docker) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filters := url.QueryEscape(`{"label":["com.docker.compose.project=sda"]}`)
	reqURL := fmt.Sprintf("%s/containers/json?filters=%s", c.proxyURL, filters)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create docker request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query docker proxy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker proxy returned %d", resp.StatusCode)
	}

	var raw []struct {
		Names  []string          `json:"Names"`
		Image  string            `json:"Image"`
		State  string            `json:"State"`
		Status string            `json:"Status"`
		Labels map[string]string `json:"Labels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode containers: %w", err)
	}

	containers := make([]ContainerInfo, 0, len(raw))
	for _, c := range raw {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		healthy := c.State == "running" && !strings.Contains(c.Status, "unhealthy")
		containers = append(containers, ContainerInfo{
			Name:    name,
			Image:   c.Image,
			Status:  c.State,
			Healthy: healthy,
		})
	}
	return containers, nil
}
