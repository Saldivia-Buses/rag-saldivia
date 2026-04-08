package quality

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// ValidationIssue records a factual error in the response.
type ValidationIssue struct {
	Severity    string `json:"severity"` // "error", "warning"
	Type        string `json:"type"`     // "wrong_month", "wrong_planet", "invented_date"
	Description string `json:"description"`
}

// monthPattern matches month references like "en marzo", "abril 2026", "mes 3".
var monthPattern = regexp.MustCompile(`(?i)(?:en |para |durante )?(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)`)
var monthNumPattern = regexp.MustCompile(`(?i)mes\s+(\d{1,2})`)

var monthNames = map[string]int{
	"enero": 1, "febrero": 2, "marzo": 3, "abril": 4,
	"mayo": 5, "junio": 6, "julio": 7, "agosto": 8,
	"septiembre": 9, "octubre": 10, "noviembre": 11, "diciembre": 12,
}

// ValidateResponse checks the LLM response against computed data for factual accuracy.
// Returns issues found. Empty slice = response is factually consistent.
func ValidateResponse(response string, fullCtx *astrocontext.FullContext) []ValidationIssue {
	var issues []ValidationIssue
	lower := strings.ToLower(response)

	// Collect months that actually have astrological activity
	activeMonths := make(map[int]bool)

	for _, tr := range fullCtx.Transits {
		for _, ep := range tr.EpDetails {
			for m := ep.MonthStart; m <= ep.MonthEnd; m++ {
				if m >= 1 && m <= 12 {
					activeMonths[m] = true
				}
			}
		}
	}
	for _, ecl := range fullCtx.Eclipses {
		if ecl.Eclipse.Month >= 1 && ecl.Eclipse.Month <= 12 {
			activeMonths[ecl.Eclipse.Month] = true
		}
	}
	for _, st := range fullCtx.Stations {
		if st.Month >= 1 && st.Month <= 12 {
			activeMonths[st.Month] = true
		}
	}

	// Check: does the response mention months that have zero activity?
	// This could indicate the LLM is inventing timing data
	mentionedMonths := extractMonths(lower)
	for _, m := range mentionedMonths {
		if !activeMonths[m] && len(activeMonths) > 3 {
			// Only flag if we have enough active months to be confident
			issues = append(issues, ValidationIssue{
				Severity:    "warning",
				Type:        "suspicious_month",
				Description: fmt.Sprintf("Mes %d mencionado en respuesta pero sin actividad astrológica calculada", m),
			})
		}
	}

	// Check: wrong year references
	yearStr := strconv.Itoa(fullCtx.Year)
	wrongYears := []string{
		strconv.Itoa(fullCtx.Year - 2),
		strconv.Itoa(fullCtx.Year + 2),
		strconv.Itoa(fullCtx.Year - 3),
		strconv.Itoa(fullCtx.Year + 3),
	}
	for _, wy := range wrongYears {
		if strings.Contains(response, wy) && !strings.Contains(response, yearStr) {
			issues = append(issues, ValidationIssue{
				Severity:    "error",
				Type:        "wrong_year",
				Description: fmt.Sprintf("Menciona año %s pero el análisis es para %d", wy, fullCtx.Year),
			})
		}
	}

	return issues
}

// extractMonths finds month references in text.
func extractMonths(text string) []int {
	var months []int
	seen := make(map[int]bool)

	// Named months
	matches := monthPattern.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			if num, ok := monthNames[strings.ToLower(m[1])]; ok && !seen[num] {
				months = append(months, num)
				seen[num] = true
			}
		}
	}

	// Numeric months ("mes 3", "mes 11")
	numMatches := monthNumPattern.FindAllStringSubmatch(text, -1)
	for _, m := range numMatches {
		if len(m) >= 2 {
			if num, err := strconv.Atoi(m[1]); err == nil && num >= 1 && num <= 12 && !seen[num] {
				months = append(months, num)
				seen[num] = true
			}
		}
	}

	return months
}
