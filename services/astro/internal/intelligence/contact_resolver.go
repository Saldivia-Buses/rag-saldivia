package intelligence

import "strings"

// ContactMatch holds a fuzzy match result for contact name resolution.
type ContactMatch struct {
	Name       string  `json:"name"`
	Score      float64 `json:"score"` // 0.0-1.0
	MatchType  string  `json:"match_type"` // "exact", "prefix", "contains", "fuzzy"
}

// ResolveContact finds the best matching contact name from a list.
// Uses progressive matching: exact → prefix → contains → fuzzy.
func ResolveContact(query string, contactNames []string) *ContactMatch {
	queryLower := strings.ToLower(strings.TrimSpace(query))
	if queryLower == "" {
		return nil
	}

	// Exact match
	for _, name := range contactNames {
		if strings.ToLower(name) == queryLower {
			return &ContactMatch{Name: name, Score: 1.0, MatchType: "exact"}
		}
	}

	// Prefix match (first name)
	var prefixMatches []ContactMatch
	for _, name := range contactNames {
		nameLower := strings.ToLower(name)
		if strings.HasPrefix(nameLower, queryLower) {
			prefixMatches = append(prefixMatches, ContactMatch{
				Name: name, Score: 0.9, MatchType: "prefix",
			})
		}
		// Also check first name match
		parts := strings.Fields(nameLower)
		if len(parts) > 0 && parts[0] == queryLower {
			return &ContactMatch{Name: name, Score: 0.95, MatchType: "prefix"}
		}
	}
	if len(prefixMatches) == 1 {
		return &prefixMatches[0]
	}

	// Contains match
	for _, name := range contactNames {
		if strings.Contains(strings.ToLower(name), queryLower) {
			return &ContactMatch{Name: name, Score: 0.7, MatchType: "contains"}
		}
	}

	// Fuzzy: check if query words appear in name
	queryWords := strings.Fields(queryLower)
	bestScore := 0.0
	bestName := ""
	for _, name := range contactNames {
		nameLower := strings.ToLower(name)
		matches := 0
		for _, w := range queryWords {
			if len(w) >= 3 && strings.Contains(nameLower, w) {
				matches++
			}
		}
		score := float64(matches) / float64(len(queryWords))
		if score > bestScore {
			bestScore = score
			bestName = name
		}
	}
	if bestScore >= 0.5 {
		return &ContactMatch{Name: bestName, Score: bestScore * 0.6, MatchType: "fuzzy"}
	}

	return nil
}
