package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitConfig defines a rate limiting policy.
type RateLimitConfig struct {
	// Requests per window. E.g., 5 requests per minute.
	Requests int
	Window   time.Duration
	// KeyFunc extracts the rate-limit key from the request.
	// Common: ByIP (per source IP) or ByUser (per authenticated user).
	// If nil, defaults to ByIP.
	KeyFunc func(r *http.Request) string
}

// ByIP returns the client IP as the rate-limit key.
func ByIP(r *http.Request) string {
	// chi's RealIP middleware rewrites RemoteAddr
	return r.RemoteAddr
}

// ByUser returns the authenticated user ID as the rate-limit key.
// Falls back to IP if no user is authenticated.
func ByUser(r *http.Request) string {
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return "user:" + uid
	}
	return r.RemoteAddr
}

// RateLimit returns a chi middleware that enforces per-key rate limiting.
// Uses an in-memory token bucket (golang.org/x/time/rate). For multi-node
// production deployments, migrate to a Redis-backed sliding window.
//
// The token bucket allows a burst of up to Requests tokens, then refills
// at a steady rate of Requests/Window. This means short bursts up to the
// limit are allowed, but sustained traffic is capped at the configured rate.
//
// Stale entries are cleaned up every 10 minutes.
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = ByIP
	}

	lim := &limiterMap{
		limiters: make(map[string]*limiterEntry),
		rate:     rate.Limit(float64(cfg.Requests) / cfg.Window.Seconds()),
		burst:    cfg.Requests,
	}

	// Background cleanup of stale entries
	go lim.cleanup(10 * time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.KeyFunc(r)
			l := lim.get(key)

			if !l.Allow() {
				retryAfter := int(cfg.Window.Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type limiterMap struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	rate     rate.Limit
	burst    int
}

func (m *limiterMap) get(key string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.limiters[key]
	if !ok {
		entry = &limiterEntry{
			limiter: rate.NewLimiter(m.rate, m.burst),
		}
		m.limiters[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (m *limiterMap) cleanup(interval time.Duration) {
	for {
		time.Sleep(interval)
		m.mu.Lock()
		for key, entry := range m.limiters {
			if time.Since(entry.lastSeen) > interval {
				delete(m.limiters, key)
			}
		}
		m.mu.Unlock()
	}
}
