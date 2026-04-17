package migration

import (
	"sort"
	"sync"
)

// SkipReason is a categorical tag a migrator attaches to a dropped row so
// post-run analytics can tell "skipped because parent missing" apart from
// "skipped because amount was zero". Without this breakdown the operator
// only sees a total count and cannot tell whether a 20% skip rate is
// expected or a silent data-loss bug.
type SkipReason string

const (
	SkipLegacyIDZero    SkipReason = "legacy_id_zero"    // PK was 0 in legacy
	SkipParentNotFound  SkipReason = "parent_not_found"  // FK parent not in mapping
	SkipArticleNotFound SkipReason = "article_not_found" // article FK missing
	SkipEntityNotFound  SkipReason = "entity_not_found"  // entity FK missing
	SkipZeroAmount      SkipReason = "zero_amount"       // quantity/amount == 0
	SkipEmptyCode       SkipReason = "empty_code"        // required code is blank
	SkipUnresolvable    SkipReason = "unresolvable_fk"   // other FK resolution failure
	SkipDeprecatedType  SkipReason = "deprecated_type"   // legacy type we chose not to keep
	SkipInvalidData     SkipReason = "invalid_data"      // data fails target constraint
)

// SkipTracker aggregates per-reason skip counts per legacy table. It is
// process-wide and goroutine-safe so the same tracker can be shared across
// all migrators in a run.
type SkipTracker struct {
	mu sync.Mutex
	// table → reason → count
	counts map[string]map[SkipReason]int
}

// NewSkipTracker returns an empty tracker.
func NewSkipTracker() *SkipTracker {
	return &SkipTracker{counts: make(map[string]map[SkipReason]int)}
}

// Note records one skip for (table, reason). Safe from multiple goroutines.
func (s *SkipTracker) Note(table string, reason SkipReason) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inner, ok := s.counts[table]
	if !ok {
		inner = make(map[SkipReason]int)
		s.counts[table] = inner
	}
	inner[reason]++
}

// Summary returns a stable, sorted snapshot suitable for JSON serialisation
// or printing. Sorted by table then by reason to keep reports diff-friendly.
func (s *SkipTracker) Summary() []SkipSummaryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]SkipSummaryEntry, 0, len(s.counts))
	for table, reasons := range s.counts {
		entry := SkipSummaryEntry{Table: table, Total: 0, ByReason: make([]SkipCount, 0, len(reasons))}
		for r, n := range reasons {
			entry.Total += n
			entry.ByReason = append(entry.ByReason, SkipCount{Reason: r, Count: n})
		}
		sort.Slice(entry.ByReason, func(i, j int) bool { return entry.ByReason[i].Reason < entry.ByReason[j].Reason })
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Table < out[j].Table })
	return out
}

// SkipSummaryEntry is one row in the skip-reason breakdown.
type SkipSummaryEntry struct {
	Table    string      `json:"table"`
	Total    int         `json:"total"`
	ByReason []SkipCount `json:"by_reason"`
}

// SkipCount is a single (reason, count) pair inside a table summary.
type SkipCount struct {
	Reason SkipReason `json:"reason"`
	Count  int        `json:"count"`
}
