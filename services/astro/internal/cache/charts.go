package cache

import (
	"sync"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// ChartRegistry is an in-memory LRU cache of computed natal charts.
// Thread-safe. Key format: "{tenant_id}:{contact_id}".
// Max 500 entries (~1MB total). Evicts least-recently-used on overflow.
type ChartRegistry struct {
	mu       sync.RWMutex
	entries  map[string]*chartEntry
	order    []string // LRU order: most recent at end
	maxSize  int
}

type chartEntry struct {
	chart *natal.Chart
}

const defaultMaxCharts = 500

// NewChartRegistry creates a chart cache.
func NewChartRegistry() *ChartRegistry {
	return &ChartRegistry{
		entries: make(map[string]*chartEntry, defaultMaxCharts),
		maxSize: defaultMaxCharts,
	}
}

// Get retrieves a cached chart. Returns nil if not found.
func (r *ChartRegistry) Get(tenantID, contactID string) *natal.Chart {
	key := tenantID + ":" + contactID
	r.mu.RLock()
	e, ok := r.entries[key]
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	// Move to end (most recent) — requires write lock
	r.mu.Lock()
	r.touch(key)
	r.mu.Unlock()
	return e.chart
}

// Put stores a chart in the cache.
func (r *ChartRegistry) Put(tenantID, contactID string, chart *natal.Chart) {
	key := tenantID + ":" + contactID
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[key]; exists {
		r.entries[key] = &chartEntry{chart: chart}
		r.touch(key)
		return
	}

	// Evict if at capacity
	for len(r.entries) >= r.maxSize && len(r.order) > 0 {
		oldest := r.order[0]
		r.order = r.order[1:]
		delete(r.entries, oldest)
	}

	r.entries[key] = &chartEntry{chart: chart}
	r.order = append(r.order, key)
}

// Invalidate removes a specific contact's chart from cache.
func (r *ChartRegistry) Invalidate(tenantID, contactID string) {
	key := tenantID + ":" + contactID
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, key)
	for i, k := range r.order {
		if k == key {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
}

// touch moves a key to the end of the LRU order. Caller must hold write lock.
func (r *ChartRegistry) touch(key string) {
	for i, k := range r.order {
		if k == key {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	r.order = append(r.order, key)
}

// Size returns the number of cached charts.
func (r *ChartRegistry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}
