package intelligence

import (
	"fmt"
	"strings"
)

// WakeupContext is the compressed memory block injected at the start of every brief.
// Gives the agent memory between sessions (~200-400 tokens).
type WakeupContext struct {
	ContactName    string            `json:"contact_name"`
	TrackRecord    *TrackRecord      `json:"track_record,omitempty"`
	LastSession    *SessionSummary   `json:"last_session,omitempty"`
	PendingPreds   []string          `json:"pending_predictions,omitempty"`
	RecurringThemes map[string]int   `json:"recurring_themes,omitempty"` // domain → count
	RectStatus     string            `json:"rect_status,omitempty"`      // birth time status
}

// TrackRecord holds prediction accuracy stats.
type TrackRecord struct {
	Total     int `json:"total"`
	Confirmed int `json:"confirmed"`
	Partial   int `json:"partial"`
	Failed    int `json:"failed"`
	Pending   int `json:"pending"`
	Accuracy  int `json:"accuracy"` // 0-100
}

// SessionSummary is a compressed view of the last consultation.
type SessionSummary struct {
	Date    string `json:"date"`
	Domain  string `json:"domain"`
	Summary string `json:"summary"`
}

// BuildWakeupContext assembles the memory block for a contact.
// Inputs come from the DB (sessions, predictions, contacts).
func BuildWakeupContext(
	contactName string,
	predictions []PredictionRecord,
	sessions []SessionRecord,
	birthTimeKnown bool,
	rectSuggestionPending bool,
) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## MEMORIA — %s\n\n", contactName))

	// Track record
	if len(predictions) > 0 {
		confirmed, partial, failed, pending := 0, 0, 0, 0
		for _, p := range predictions {
			switch p.Outcome {
			case "correct": confirmed++
			case "partial": partial++
			case "incorrect": failed++
			default: pending++
			}
		}
		total := confirmed + partial + failed
		accuracy := 0
		if total > 0 {
			accuracy = int(float64(confirmed+partial/2) / float64(total) * 100)
		}
		b.WriteString(fmt.Sprintf("Track record: %d predicciones — %d%% acierto (%d confirmadas, %d parciales, %d fallidas, %d pendientes)\n",
			len(predictions), accuracy, confirmed, partial, failed, pending))

		// Pending predictions (max 3)
		pendingPreds := 0
		for _, p := range predictions {
			if p.Outcome == "pending" || p.Outcome == "" {
				if pendingPreds < 3 {
					summary := p.Description
					if len(summary) > 60 { summary = summary[:57] + "..." }
					b.WriteString(fmt.Sprintf("  Pendiente: %s\n", summary))
				}
				pendingPreds++
			}
		}
	}

	// Last session
	if len(sessions) > 0 {
		last := sessions[0]
		b.WriteString(fmt.Sprintf("Última consulta: %s — %s\n", last.Date, last.Domain))
		if last.Summary != "" {
			summary := last.Summary
			if len(summary) > 80 { summary = summary[:77] + "..." }
			b.WriteString(fmt.Sprintf("  %s\n", summary))
		}
	}

	// Recurring themes
	if len(sessions) >= 3 {
		themes := make(map[string]int)
		for _, s := range sessions {
			if s.Domain != "" { themes[s.Domain]++ }
		}
		if len(themes) > 0 {
			var parts []string
			for d, n := range themes {
				parts = append(parts, fmt.Sprintf("%s×%d", d, n))
			}
			b.WriteString(fmt.Sprintf("Temas recurrentes: %s\n", strings.Join(parts, ", ")))
		}
	}

	// Rectification status
	if !birthTimeKnown {
		if rectSuggestionPending {
			b.WriteString("⚠ Hora desconocida — sugerencia de rectificación pendiente\n")
		} else {
			b.WriteString("⚠ Hora desconocida — técnicas dependientes de casas tienen margen\n")
		}
	}

	return b.String()
}

// PredictionRecord is a simplified prediction for memory building.
type PredictionRecord struct {
	Description string
	Outcome     string // "correct", "incorrect", "partial", "pending", ""
	TargetDate  string
}

// SessionRecord is a simplified session for memory building.
type SessionRecord struct {
	Date    string
	Domain  string
	Summary string
}

// EventRecord represents a detected life event from chat.
type EventRecord struct {
	Contact     string `json:"contact"`
	Date        string `json:"date"`        // YYYY-MM-DD
	Type        string `json:"type"`        // "contrato", "pago", "pérdida", "cierre", "compra", "viaje", etc.
	Description string `json:"description"`
	Source      string `json:"source"`      // "chat", "followup"
}

// ExtractEventsFromText detects past-tense life events in user chat messages.
// Uses Spanish verb patterns to identify event mentions.
func ExtractEventsFromText(text string, contactNames []string) []EventRecord {
	lower := strings.ToLower(text)
	var events []EventRecord

	// Match contact
	matchedContact := ""
	for _, name := range contactNames {
		nameLower := strings.ToLower(name)
		firstName := strings.Fields(nameLower)[0]
		if strings.Contains(lower, nameLower) || strings.Contains(lower, firstName) {
			matchedContact = name
			break
		}
	}
	if matchedContact == "" { return nil }

	// Past-tense event verb detection (Spanish)
	eventVerbs := map[string]string{
		"firmamos": "contrato", "firmó": "contrato", "cerró": "cierre",
		"cerramos": "cierre", "pagó": "pago", "pagamos": "pago",
		"perdimos": "pérdida", "perdió": "pérdida", "ganamos": "ganancia",
		"compró": "compra", "compramos": "compra", "vendió": "venta",
		"vendimos": "venta", "viajó": "viaje", "se mudó": "mudanza",
		"se casó": "matrimonio", "nació": "nacimiento", "falleció": "fallecimiento",
		"operaron": "cirugía", "operó": "cirugía", "renunció": "renuncia",
		"lo ascendieron": "ascenso", "la ascendieron": "ascenso",
	}

	for verb, evType := range eventVerbs {
		if strings.Contains(lower, verb) {
			events = append(events, EventRecord{
				Contact:     matchedContact,
				Type:        evType,
				Description: text,
				Source:      "chat",
			})
			break // one event per message
		}
	}

	return events
}
