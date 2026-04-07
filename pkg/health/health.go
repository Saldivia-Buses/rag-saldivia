// Package health provides a shared health check handler for all SDA services.
// Each service registers its dependencies (DB, Redis, NATS) and the handler
// reports their real status with latency. Returns HTTP 200 if all healthy,
// HTTP 503 if any dependency is down.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// CheckFunc is a function that checks a dependency's health.
// Returns nil if healthy, error if not.
type CheckFunc func(ctx context.Context) error

// ExtraFunc returns a key-value pair to include in the health response.
type ExtraFunc func() (string, any)

// Checker holds registered health checks for a service.
type Checker struct {
	service string
	checks  map[string]CheckFunc
	extras  []ExtraFunc
}

// New creates a health checker for the given service name.
func New(service string) *Checker {
	return &Checker{
		service: service,
		checks:  make(map[string]CheckFunc),
	}
}

// Add registers a named dependency check.
func (c *Checker) Add(name string, fn CheckFunc) {
	c.checks[name] = fn
}

// AddExtra registers a function that provides extra key-value data
// in the health response (e.g., connected client count).
func (c *Checker) AddExtra(fn ExtraFunc) {
	c.extras = append(c.extras, fn)
}

// DependencyStatus is the health status of a single dependency.
type DependencyStatus struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

// Response is the JSON response from the health endpoint.
type Response struct {
	Status       string                      `json:"status"`
	Service      string                      `json:"service"`
	Dependencies map[string]DependencyStatus `json:"dependencies,omitempty"`
	Extra        map[string]any              `json:"extra,omitempty"`
}

// Handler returns an http.HandlerFunc that runs all registered checks.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		resp := Response{
			Status:       "ok",
			Service:      c.service,
			Dependencies: make(map[string]DependencyStatus, len(c.checks)),
		}

		var mu sync.Mutex
		var wg sync.WaitGroup
		allHealthy := true

		for name, fn := range c.checks {
			wg.Add(1)
			go func(name string, fn CheckFunc) {
				defer wg.Done()
				start := time.Now()
				err := fn(ctx)
				latency := time.Since(start).Milliseconds()

				dep := DependencyStatus{
					Status:    "up",
					LatencyMs: latency,
				}
				if err != nil {
					dep.Status = "down"
					dep.Error = err.Error()
				}

				mu.Lock()
				resp.Dependencies[name] = dep
				if err != nil {
					allHealthy = false
				}
				mu.Unlock()
			}(name, fn)
		}
		wg.Wait()

		if !allHealthy {
			resp.Status = "degraded"
		}

		if len(c.extras) > 0 {
			resp.Extra = make(map[string]any, len(c.extras))
			for _, fn := range c.extras {
				k, v := fn()
				resp.Extra[k] = v
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if allHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(resp)
	}
}
