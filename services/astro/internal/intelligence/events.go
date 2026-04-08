package intelligence

import (
	"regexp"
	"strings"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// LifeEvent represents a past-tense event detected from user messages.
// Used for auto-rectification: events with known dates help narrow birth time.
type LifeEvent struct {
	Description string `json:"description"`
	Category    string `json:"category"` // contrato, cierre, perdida, ganancia, inicio, etc.
	DateText    string `json:"date_text"` // raw date reference (e.g., "mayo 2025")
	Month       int    `json:"month,omitempty"`
	Year        int    `json:"year,omitempty"`
}

// Past-tense verb patterns that indicate life events.
var eventVerbs = regexp.MustCompile(`(?i)\b(firm[eé]|contrat[eé]|perd[ií]|gan[eé]|empec[eé]|cerr[eé]|abr[ií]|vend[ií]|compr[eé]|mud[eé]|separ[eé]|divorci[eé]|cas[eé]|oper[eé]|enferme|renunci[eé]|ascend[ií]|cambi[eé]|viaj[eé]|naci[oó]|muri[oó]|jubil[eé])\b`)

// Month name patterns for date extraction.
var monthPattern = regexp.MustCompile(`(?i)(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)\s*(?:de\s+)?(\d{4})`)
var monthOnly = regexp.MustCompile(`(?i)(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)`)

var monthNames = map[string]int{
	"enero": 1, "febrero": 2, "marzo": 3, "abril": 4,
	"mayo": 5, "junio": 6, "julio": 7, "agosto": 8,
	"septiembre": 9, "octubre": 10, "noviembre": 11, "diciembre": 12,
}

// ExtractLifeEvents scans user messages for past-tense events with dates.
// Returns events suitable for auto-rectification.
func ExtractLifeEvents(messages []string) []LifeEvent {
	var events []LifeEvent

	for _, msg := range messages {
		if !eventVerbs.MatchString(msg) {
			continue
		}

		// Extract the verb match for category
		verb := eventVerbs.FindString(msg)
		category := categorizeVerb(strings.ToLower(verb))

		// Try to extract date
		event := LifeEvent{
			Description: truncate(msg, 200),
			Category:    category,
		}

		if m := monthPattern.FindStringSubmatch(msg); len(m) >= 3 {
			event.DateText = m[0]
			if mo, ok := monthNames[strings.ToLower(m[1])]; ok {
				event.Month = mo
			}
			// Parse year
			for _, c := range m[2] {
				event.Year = event.Year*10 + int(c-'0')
			}
		} else if m := monthOnly.FindString(msg); m != "" {
			event.DateText = m
			if mo, ok := monthNames[strings.ToLower(m)]; ok {
				event.Month = mo
			}
		}

		events = append(events, event)
	}

	return events
}

// ShouldAutoRectify returns true if enough events have been accumulated
// to trigger automatic birth time rectification.
func ShouldAutoRectify(events []LifeEvent) bool {
	// Need at least 3 events with dates for meaningful rectification
	dated := 0
	for _, e := range events {
		if e.Month > 0 && e.Year > 0 {
			dated++
		}
	}
	return dated >= 3
}

// ToRectificationEvents converts LifeEvents to technique.RectificationEvent
// for use with the rectification engine.
func ToRectificationEvents(events []LifeEvent) []technique.RectificationEvent {
	var result []technique.RectificationEvent
	for _, e := range events {
		if e.Month > 0 && e.Year > 0 {
			result = append(result, technique.RectificationEvent{
				Description: e.Description,
				Category:    e.Category,
				Date:        time.Date(e.Year, time.Month(e.Month), 15, 0, 0, 0, 0, time.UTC),
			})
		}
	}
	return result
}

func categorizeVerb(verb string) string {
	switch {
	case strings.HasPrefix(verb, "firm") || strings.HasPrefix(verb, "contrat"):
		return "contrato"
	case strings.HasPrefix(verb, "perd"):
		return "perdida"
	case strings.HasPrefix(verb, "gan"):
		return "ganancia"
	case strings.HasPrefix(verb, "empec"):
		return "inicio"
	case strings.HasPrefix(verb, "cerr"):
		return "cierre"
	case strings.HasPrefix(verb, "vend"):
		return "venta"
	case strings.HasPrefix(verb, "compr"):
		return "compra"
	case strings.HasPrefix(verb, "mud"):
		return "mudanza"
	case strings.HasPrefix(verb, "separ") || strings.HasPrefix(verb, "divorci"):
		return "separacion"
	case strings.HasPrefix(verb, "cas"):
		return "matrimonio"
	case strings.HasPrefix(verb, "oper") || strings.HasPrefix(verb, "enferm"):
		return "salud"
	case strings.HasPrefix(verb, "renunci"):
		return "renuncia"
	case strings.HasPrefix(verb, "ascend"):
		return "ascenso"
	case strings.HasPrefix(verb, "viaj"):
		return "viaje"
	default:
		return "evento"
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
