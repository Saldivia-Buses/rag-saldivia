package intelligence

import (
	_ "embed"
	"encoding/json"
	"math"
	"strings"
	"unicode"
)

//go:embed utterances.json
var utterancesJSON []byte

// UtteranceRouter implements IntentParser using TF-IDF-like scoring
// over 1,817 utterances across 55 routes from astro-v2.
// Falls back to KeywordParser when confidence is below threshold.
type UtteranceRouter struct {
	// inverted index: token → list of (routeID, utteranceIndex, tokenCount)
	index map[string][]indexEntry
	// route IDs
	routes []string
	// token counts per route (for IDF-like weighting)
	routeTokenCounts map[string]int
	// total utterances
	totalUtterances int
	// fallback
	keyword *KeywordParser
}

type indexEntry struct {
	route      string
	tokenCount int // total tokens in the utterance
}

// NewUtteranceRouter builds the inverted index from embedded utterances.
func NewUtteranceRouter(registry *DomainRegistry) *UtteranceRouter {
	r := &UtteranceRouter{
		index:            make(map[string][]indexEntry),
		routeTokenCounts: make(map[string]int),
		keyword:          NewKeywordParser(registry),
	}

	// Parse utterances JSON
	var data map[string][]string
	if err := json.Unmarshal(utterancesJSON, &data); err != nil {
		// If parsing fails, UtteranceRouter degrades to keyword-only
		return r
	}

	// Build inverted index
	routeSet := make(map[string]bool)
	for route, utterances := range data {
		routeSet[route] = true
		r.totalUtterances += len(utterances)
		for _, utt := range utterances {
			tokens := tokenize(utt)
			r.routeTokenCounts[route] += len(tokens)
			for _, tok := range tokens {
				r.index[tok] = append(r.index[tok], indexEntry{
					route:      route,
					tokenCount: len(tokens),
				})
			}
		}
	}
	r.routes = make([]string, 0, len(routeSet))
	for route := range routeSet {
		r.routes = append(r.routes, route)
	}

	return r
}

// Parse implements IntentParser. Scores query against all utterances
// using token overlap with IDF-like weighting.
func (r *UtteranceRouter) Parse(query string) *Intent {
	if len(r.index) == 0 {
		// No index built — fall back to keyword
		return r.keyword.Parse(query)
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return r.keyword.Parse(query)
	}

	// Score each route by token overlap
	routeScores := make(map[string]float64)
	routeHits := make(map[string]int)

	for _, tok := range queryTokens {
		entries, ok := r.index[tok]
		if !ok {
			continue
		}
		// IDF: tokens that appear in fewer routes are more discriminative
		routesWithToken := make(map[string]bool)
		for _, e := range entries {
			routesWithToken[e.route] = true
		}
		idf := math.Log(float64(len(r.routes)+1) / float64(len(routesWithToken)+1))
		if idf < 0.1 {
			idf = 0.1 // floor for very common tokens
		}

		for _, e := range entries {
			// Weight by: IDF * (1 / utterance length) — shorter utterances with matching tokens score higher
			weight := idf / math.Sqrt(float64(e.tokenCount))
			routeScores[e.route] += weight
			routeHits[e.route]++
		}
	}

	if len(routeScores) == 0 {
		// No utterance matches — fall back to keyword
		return r.keyword.Parse(query)
	}

	// Find top scoring route
	bestRoute := ""
	bestScore := 0.0
	for route, score := range routeScores {
		if score > bestScore {
			bestScore = score
			bestRoute = route
		}
	}

	// Normalize confidence: proportion of query tokens that matched
	confidence := float64(routeHits[bestRoute]) / float64(len(queryTokens))
	if confidence > 1.0 {
		confidence = 1.0
	}

	// If confidence is too low, fall back to keyword parser
	if confidence < 0.3 {
		kwIntent := r.keyword.Parse(query)
		// If keyword parser found something with higher confidence, use that
		if kwIntent.Confidence > confidence {
			return kwIntent
		}
	}

	// Build intent from utterance match
	intent := &Intent{
		PrimaryDomain:   bestRoute,
		Confidence:      confidence,
		MatchedKeywords: matchedTokens(queryTokens, r.index, bestRoute),
	}

	// Extract focus points (same as keyword parser)
	for _, pat := range focusPatterns {
		matches := pat.FindAllString(query, -1)
		for _, m := range matches {
			intent.FocusPoints = append(intent.FocusPoints, strings.ToLower(m))
		}
	}

	// Secondary domains (routes scoring > 50% of best)
	for route, score := range routeScores {
		if route != bestRoute && score >= bestScore*0.5 {
			intent.SecondaryDomains = append(intent.SecondaryDomains, route)
		}
	}

	return intent
}

// QuickDomain does a fast domain-only detection without full intent analysis.
// Used by handler for domain-aware lazy calc before full Analyze() pipeline.
func (r *UtteranceRouter) QuickDomain(query string) string {
	intent := r.Parse(query)
	return intent.PrimaryDomain
}

// tokenize splits text into lowercase tokens, removing stopwords and punctuation.
func tokenize(text string) []string {
	lower := strings.ToLower(text)
	// Replace accented characters
	lower = removeAccentsIntent(lower)

	words := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	// Filter stopwords
	var tokens []string
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if stopwords[w] {
			continue
		}
		tokens = append(tokens, w)
	}
	return tokens
}

// removeAccentsIntent replaces common accented characters.
func removeAccentsIntent(s string) string {
	r := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"ü", "u", "ñ", "n",
	)
	return r.Replace(s)
}

// matchedTokens returns query tokens that matched entries for a given route.
func matchedTokens(queryTokens []string, index map[string][]indexEntry, route string) []string {
	var matched []string
	for _, tok := range queryTokens {
		entries, ok := index[tok]
		if !ok {
			continue
		}
		for _, e := range entries {
			if e.route == route {
				matched = append(matched, tok)
				break
			}
		}
	}
	return matched
}

// Spanish stopwords — common words that don't carry domain-specific meaning.
var stopwords = map[string]bool{
	"de": true, "la": true, "el": true, "en": true, "que": true,
	"un": true, "una": true, "los": true, "las": true, "del": true,
	"al": true, "es": true, "por": true, "con": true, "para": true,
	"se": true, "su": true, "no": true, "lo": true, "como": true,
	"mas": true, "pero": true, "sus": true, "le": true, "ya": true,
	"este": true, "esta": true, "si": true, "hay": true, "me": true,
	"mi": true, "muy": true, "sin": true, "ser": true, "ha": true,
	"yo": true, "eso": true, "son": true, "todo": true, "fue": true,
	"nos": true, "tan": true, "ni": true, "te": true, "ti": true,
	"ese": true, "esa": true, "cual": true, "quien": true, "va": true,
	"puede": true, "sobre": true, "entre": true, "ano": true, "viene": true,
}
