package intelligence

import (
	"regexp"
	"strings"
)

// Intent is the parsed query intent.
type Intent struct {
	PrimaryDomain   string   // domain ID (e.g., "carrera")
	SecondaryDomains []string // additional domains detected
	Confidence      float64  // 0.0-1.0
	MatchedKeywords []string // which keywords triggered
	FocusPoints     []string // planets/houses mentioned
}

// intentRule maps keywords to a domain with priority.
type intentRule struct {
	domain  string
	weight  float64
	keywords []string
	patterns []*regexp.Regexp
}

// focusPatterns detect planet/house mentions in queries.
var focusPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(sol|luna|mercurio|venus|marte|j[úu]piter|saturno|urano|neptuno|plut[oó]n|quir[oó]n)\b`),
	regexp.MustCompile(`(?i)\bcasa\s+(\d{1,2})\b`),
	regexp.MustCompile(`(?i)\b(asc|ascendente|mc|medio cielo)\b`),
}

// ParseIntent analyzes a query and returns the detected intent.
// Uses keyword matching with Spanish astrological vocabulary.
func ParseIntent(query string, registry *DomainRegistry) *Intent {
	lower := strings.ToLower(query)
	intent := &Intent{
		Confidence: 0.5,
	}

	// Extract focus points (planets/houses mentioned)
	for _, pat := range focusPatterns {
		matches := pat.FindAllString(query, -1)
		for _, m := range matches {
			intent.FocusPoints = append(intent.FocusPoints, strings.ToLower(m))
		}
	}

	// Score each domain by keyword matches
	type domainScore struct {
		id       string
		score    float64
		keywords []string
	}
	var scores []domainScore

	for _, id := range registry.AllIDs() {
		d := registry.Get(id)
		if d == nil || len(d.Keywords) == 0 {
			continue
		}

		score := 0.0
		var matched []string
		for _, kw := range d.Keywords {
			if strings.Contains(lower, kw) {
				score += 1.0
				matched = append(matched, kw)
				// Bonus for longer keyword matches (more specific)
				if len(kw) > 6 {
					score += 0.5
				}
			}
		}

		if score > 0 {
			// Subdomains get a slight bonus (more specific)
			if d.Parent != "" {
				score += 0.3
			}
			scores = append(scores, domainScore{id, score, matched})
		}
	}

	if len(scores) == 0 {
		// No keyword match — default to predictivo
		intent.PrimaryDomain = "predictivo"
		intent.Confidence = 0.3
		return intent
	}

	// Sort by score descending
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	intent.PrimaryDomain = scores[0].id
	intent.MatchedKeywords = scores[0].keywords
	intent.Confidence = min(scores[0].score/3.0, 1.0) // normalize to 0-1

	// Secondary domains (score > 50% of primary)
	for _, s := range scores[1:] {
		if s.score >= scores[0].score*0.5 {
			intent.SecondaryDomains = append(intent.SecondaryDomains, s.id)
		}
	}

	return intent
}

// min64 removed — Go 1.21+ builtin min works with float64
