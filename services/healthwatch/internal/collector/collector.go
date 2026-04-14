// Package collector provides data collectors for the healthwatch service.
// Each collector gathers health data from a specific source.
package collector

import (
	"slices"
	"time"
)

// KnownServices is the hardcoded service whitelist.
// NEVER accept service names from user input — prevents PromQL injection ([M3]).
var KnownServices = []string{
	"auth", "ws", "chat", "agent", "search", "astro",
	"traces", "notification", "platform", "ingest",
	"feedback", "bigbrother", "erp", "healthwatch",
}

// ServicePortMap maps service names to their HTTP ports.
var ServicePortMap = map[string]string{
	"auth":         "8001",
	"ws":           "8002",
	"chat":         "8003",
	"agent":        "8004",
	"notification": "8005",
	"platform":     "8006",
	"ingest":       "8007",
	"feedback":     "8008",
	"traces":       "8009",
	"search":       "8010",
	"astro":        "8011",
	"bigbrother":   "8012",
	"erp":          "8013",
	"healthwatch":  "8014",
}

// IsKnownService validates a service name against the whitelist.
func IsKnownService(name string) bool {
	return slices.Contains(KnownServices, name)
}

// ServiceCheck is the raw result from checking a service's /health endpoint.
type ServiceCheck struct {
	Name    string
	Status  string
	Version string
	Details map[string]any
}

// PrometheusMetrics holds scraped metrics for a single service.
type PrometheusMetrics struct {
	ErrorRate5m  float64
	P99LatencyMs float64
}

// PrometheusAlert represents a firing Prometheus alert.
type PrometheusAlert struct {
	Name     string
	Service  string
	Severity string
	ActiveAt time.Time
}

// ContainerInfo represents a Docker container's status.
type ContainerInfo struct {
	Name    string
	Image   string
	Status  string
	Healthy bool
}
