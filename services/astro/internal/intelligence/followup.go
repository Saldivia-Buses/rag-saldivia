package intelligence

import (
	"strings"
	"unicode"
)

// FollowUpContext holds inherited context from a previous exchange.
type FollowUpContext struct {
	IsFollowUp  bool   // true if the message is a follow-up
	ContactID   string // inherited from previous message
	DomainID    string // inherited from previous analysis
	PrevBrief   string // brief from previous exchange (cached)
	PrevResponse string // last assistant response (truncated)
}

// DetectFollowUp determines if a message is a short follow-up to a previous
// exchange rather than a new consultation. Uses multiple signals:
// 1. Message is short (< 25 words)
// 2. Does NOT contain birth data (date, time, place patterns)
// 3. Contains continuation indicators (pronouns, month references, conjunctions)
// 4. There is prior history to continue from
//
// Returns nil if not a follow-up.
func DetectFollowUp(msg string, hasHistory bool, prevDomain string, prevContactID string) *FollowUpContext {
	if !hasHistory || msg == "" {
		return nil
	}

	words := strings.Fields(msg)
	if len(words) > 25 {
		return nil // too long for a follow-up
	}

	lower := strings.ToLower(msg)

	// Reject if contains birth data patterns
	if containsBirthData(lower) {
		return nil
	}

	// Check for continuation indicators
	if !hasContinuationSignals(lower, words) {
		return nil
	}

	return &FollowUpContext{
		IsFollowUp: true,
		ContactID:  prevContactID,
		DomainID:   prevDomain,
	}
}

// containsBirthData checks for birth date/time patterns that indicate
// a new consultation, not a follow-up.
func containsBirthData(lower string) bool {
	// Date patterns: "15 de marzo", "1990", "naci el"
	birthIndicators := []string{
		"naci ", "nací ", "nacido", "nacida",
		"hora de nacimiento", "lugar de nacimiento",
		"de 19", "de 20", // decades
	}
	for _, ind := range birthIndicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}

	// Check for year-like numbers (1940-2025)
	for i := 0; i < len(lower)-3; i++ {
		if lower[i] >= '1' && lower[i] <= '2' &&
			lower[i+1] >= '0' && lower[i+1] <= '9' &&
			lower[i+2] >= '0' && lower[i+2] <= '9' &&
			lower[i+3] >= '0' && lower[i+3] <= '9' {
			year := (int(lower[i]-'0') * 1000) + (int(lower[i+1]-'0') * 100) +
				(int(lower[i+2]-'0') * 10) + int(lower[i+3]-'0')
			if year >= 1940 && year <= 2010 {
				return true // likely a birth year
			}
		}
	}

	return false
}

// hasContinuationSignals checks for patterns that indicate the message
// continues a previous exchange.
func hasContinuationSignals(lower string, words []string) bool {
	// Starts with conjunction or connector
	if len(words) > 0 {
		firstWord := strings.TrimLeftFunc(words[0], func(r rune) bool {
			return !unicode.IsLetter(r)
		})
		firstWord = strings.ToLower(firstWord)
		continuationStarts := map[string]bool{
			"y": true, "pero": true, "tambien": true, "también": true,
			"ademas": true, "además": true, "entonces": true, "o": true,
			"osea": true, "eso": true, "ese": true, "esa": true,
		}
		if continuationStarts[firstWord] {
			return true
		}
	}

	// Contains month/planet references without full context
	monthRefs := []string{
		"en enero", "en febrero", "en marzo", "en abril", "en mayo",
		"en junio", "en julio", "en agosto", "en septiembre", "en octubre",
		"en noviembre", "en diciembre",
		"enero?", "febrero?", "marzo?", "abril?", "mayo?",
		"junio?", "julio?", "agosto?", "septiembre?", "octubre?",
		"noviembre?", "diciembre?",
	}
	for _, ref := range monthRefs {
		if strings.Contains(lower, ref) {
			return true
		}
	}

	// Question about something already discussed
	followUpPatterns := []string{
		"que pasa con", "y que hay de", "como queda",
		"me explicas", "podes ampliar", "podés ampliar",
		"contame mas", "contame más", "dame mas detalle",
		"y eso que significa", "y eso qué significa",
		"en que mes", "en qué mes",
		"que significa eso", "qué significa eso",
	}
	for _, pat := range followUpPatterns {
		if strings.Contains(lower, pat) {
			return true
		}
	}

	// Very short question (1-4 words) ending with ? — likely a follow-up
	// BUT not if it contains a year (e.g., "Jupiter 2027?" is a new query)
	if len(words) <= 4 && strings.HasSuffix(strings.TrimSpace(lower), "?") {
		hasYear := false
		for _, w := range words {
			cleaned := strings.TrimRight(w, "?!.,;:")
			if len(cleaned) == 4 && cleaned[0] >= '1' && cleaned[0] <= '2' &&
				cleaned[1] >= '0' && cleaned[1] <= '9' &&
				cleaned[2] >= '0' && cleaned[2] <= '9' &&
				cleaned[3] >= '0' && cleaned[3] <= '9' {
				hasYear = true
				break
			}
		}
		if !hasYear {
			return true
		}
	}

	return false
}
