package quality

import (
	"fmt"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/intelligence"
)

// AuditResult holds the deterministic quality audit of a response.
// No LLM calls — all checks are rule-based and fast (<10ms).
type AuditResult struct {
	ScoreTechnical    int            `json:"score_technical"`    // 0-100 data completeness
	ScoreCommunication int           `json:"score_communication"` // 0-100 narrative quality
	ScoreTotal        int            `json:"score_total"`        // weighted average
	Issues            []AuditIssue   `json:"issues"`
	TechniquesUsed    int            `json:"techniques_used"`
	TechniquesExpected int           `json:"techniques_expected"`
	PrecautionsMet    int            `json:"precautions_met"`
	PrecautionsTotal  int            `json:"precautions_total"`
}

// AuditIssue is a single quality problem found.
type AuditIssue struct {
	Severity    string `json:"severity"` // "critical", "warning", "info"
	Category    string `json:"category"` // "missing_technique", "hallucination", "precaution"
	Description string `json:"description"`
}

// RunAudit performs a deterministic quality audit of an LLM response.
// Checks:
// 1. Domain-required techniques mentioned in response
// 2. No ghost techniques cited (techniques with no computed data)
// 3. Precautions addressed
// 4. Response length and structure
func RunAudit(
	response string,
	domain *intelligence.ResolvedDomain,
	gate *intelligence.GateResult,
) *AuditResult {
	result := &AuditResult{}
	lower := strings.ToLower(response)

	// --- Technical score (data completeness) ---

	// Check how many validated techniques are mentioned in the response
	techMentioned := 0
	techExpected := 0
	for _, v := range gate.Validated {
		techExpected++
		// Check if the technique ID or a Spanish equivalent appears
		if containsTechReference(lower, v.TechniqueID) {
			techMentioned++
		}
	}
	result.TechniquesUsed = techMentioned
	result.TechniquesExpected = techExpected

	if techExpected > 0 {
		result.ScoreTechnical = (techMentioned * 100) / techExpected
	} else {
		result.ScoreTechnical = 50 // no techniques expected = neutral
	}

	// Penalize ghost technique mentions (citing data that doesn't exist)
	for _, g := range gate.Ghosts {
		if containsTechReference(lower, g.TechniqueID) {
			result.Issues = append(result.Issues, AuditIssue{
				Severity:    "critical",
				Category:    "hallucination",
				Description: fmt.Sprintf("Cita técnica '%s' que no tiene datos calculados", g.TechniqueID),
			})
			result.ScoreTechnical -= 15
		}
	}

	// Warn for high-weight techniques not mentioned
	for _, tw := range domain.TechniquesBrief {
		if tw.Weight >= 0.8 && !containsTechReference(lower, tw.ID) {
			// Check if it was validated (has data)
			for _, v := range gate.Validated {
				if v.TechniqueID == tw.ID {
					result.Issues = append(result.Issues, AuditIssue{
						Severity:    "warning",
						Category:    "missing_technique",
						Description: fmt.Sprintf("Técnica prioritaria '%s' (peso %.1f) no mencionada", tw.ID, tw.Weight),
					})
					result.ScoreTechnical -= 5
					break
				}
			}
		}
	}

	// --- Communication score (narrative quality) ---
	result.ScoreCommunication = 70 // baseline

	// Length check
	wordCount := len(strings.Fields(response))
	if wordCount < 50 {
		result.ScoreCommunication -= 20
		result.Issues = append(result.Issues, AuditIssue{
			Severity: "warning", Category: "length",
			Description: fmt.Sprintf("Respuesta muy corta (%d palabras)", wordCount),
		})
	} else if wordCount > 2000 {
		result.ScoreCommunication -= 10
		result.Issues = append(result.Issues, AuditIssue{
			Severity: "info", Category: "length",
			Description: fmt.Sprintf("Respuesta muy extensa (%d palabras)", wordCount),
		})
	} else if wordCount >= 200 {
		result.ScoreCommunication += 10 // good length
	}

	// Structure check: has headers/sections
	if strings.Contains(response, "##") || strings.Contains(response, "**") {
		result.ScoreCommunication += 10
	}

	// Actionable recommendation check
	actionWords := []string{"recomend", "suger", "consejo", "acción", "aprovech", "evitar", "cuidado"}
	hasAction := false
	for _, w := range actionWords {
		if strings.Contains(lower, w) {
			hasAction = true
			break
		}
	}
	if hasAction {
		result.ScoreCommunication += 5
	} else {
		result.ScoreCommunication -= 5
		result.Issues = append(result.Issues, AuditIssue{
			Severity: "info", Category: "actionable",
			Description: "Sin recomendación accionable detectada",
		})
	}

	// Narrative arc check: opening (direct answer), development (techniques), closing (wisdom/advice)
	sentences := strings.Split(response, ".")
	hasDirectOpening := false
	if len(sentences) >= 3 {
		// First 2 sentences should address the question directly (not preamble)
		opening := strings.ToLower(strings.Join(sentences[:2], "."))
		directPatterns := []string{"este año", "este período", "el año", "tu ", "vas ", "hay ", "se viene", "durante"}
		for _, dp := range directPatterns {
			if strings.Contains(opening, dp) {
				hasDirectOpening = true
				break
			}
		}
	}
	if hasDirectOpening {
		result.ScoreCommunication += 5
	} else if len(sentences) >= 3 {
		result.Issues = append(result.Issues, AuditIssue{
			Severity: "info", Category: "narrative",
			Description: "La respuesta no empieza con respuesta directa al consultante",
		})
	}

	// Timing precision: mentions specific months/weeks (not just vague "este año")
	monthNames := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio",
		"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
	monthCount := 0
	for _, m := range monthNames {
		if strings.Contains(lower, m) {
			monthCount++
		}
	}
	if monthCount >= 3 {
		result.ScoreCommunication += 5 // good timing precision
	} else if monthCount == 0 {
		result.ScoreCommunication -= 5
		result.Issues = append(result.Issues, AuditIssue{
			Severity: "warning", Category: "timing",
			Description: "Sin meses específicos mencionados — timing vago",
		})
	}

	// Quincena/week precision bonus
	weekPatterns := []string{"primera quincena", "segunda quincena", "principios de", "mediados de", "fines de",
		"primera semana", "segunda semana", "tercera semana", "cuarta semana"}
	hasWeekPrecision := false
	for _, wp := range weekPatterns {
		if strings.Contains(lower, wp) {
			hasWeekPrecision = true
			break
		}
	}
	if hasWeekPrecision {
		result.ScoreCommunication += 5
	}

	// Anti-jargon: check ratio of technical terms vs accessible language
	jargonTerms := []string{"orbe", "progresado", "mundano", "eclíptica", "cúspide",
		"dispositor", "almutén", "hyleg", "decenial", "topocéntrico", "regiomontanus"}
	jargonCount := 0
	for _, jt := range jargonTerms {
		if strings.Contains(lower, jt) {
			jargonCount++
		}
	}
	if jargonCount > 5 {
		result.ScoreCommunication -= 5
		result.Issues = append(result.Issues, AuditIssue{
			Severity: "info", Category: "jargon",
			Description: fmt.Sprintf("Exceso de jerga técnica (%d términos técnicos detectados)", jargonCount),
		})
	}

	// Predictive richness: response mentions specific events/outcomes, not just vague tendencies
	richPatterns := []string{"va a ", "se va a ", "puede ", "es probable que", "hay chances de",
		"momento ideal para", "cuidado con", "aprovechar", "evitar"}
	richCount := 0
	for _, rp := range richPatterns {
		if strings.Contains(lower, rp) {
			richCount++
		}
	}
	if richCount >= 3 {
		result.ScoreCommunication += 5
	}

	// Precautions check
	result.PrecautionsTotal = len(domain.Precautions)
	for _, prec := range domain.Precautions {
		precLower := strings.ToLower(prec)
		// Extract key words from precaution
		words := strings.Fields(precLower)
		matched := false
		for _, w := range words {
			if len(w) > 4 && strings.Contains(lower, w) {
				matched = true
				break
			}
		}
		if matched {
			result.PrecautionsMet++
		} else {
			result.Issues = append(result.Issues, AuditIssue{
				Severity: "info", Category: "precaution",
				Description: fmt.Sprintf("Precaución no abordada: %s", prec),
			})
		}
	}

	// Clamp scores
	if result.ScoreTechnical < 0 { result.ScoreTechnical = 0 }
	if result.ScoreTechnical > 100 { result.ScoreTechnical = 100 }
	if result.ScoreCommunication < 0 { result.ScoreCommunication = 0 }
	if result.ScoreCommunication > 100 { result.ScoreCommunication = 100 }

	// Total: 60% technical + 40% communication
	result.ScoreTotal = (result.ScoreTechnical*60 + result.ScoreCommunication*40) / 100

	return result
}

// techAliases maps technique IDs to Spanish words that might appear in a response.
var techAliases = map[string][]string{
	intelligence.TechTransits:     {"tránsito", "transito", "transita"},
	intelligence.TechSolarArc:     {"arco solar", "arcos solares", "sa "},
	intelligence.TechPrimaryDir:   {"dirección primaria", "direcciones primarias", "dp "},
	intelligence.TechProgressions: {"progresión", "progresion", "progresada"},
	intelligence.TechSolarReturn:  {"revolución solar", "revolucion solar", "rs "},
	intelligence.TechProfections:  {"profección", "profeccion", "cronócrata", "cronocrata"},
	intelligence.TechFirdaria:     {"firdaria"},
	intelligence.TechZR:           {"zodiacal releasing", "zr ", "loosing"},
	intelligence.TechEclipses:     {"eclipse"},
	intelligence.TechFixedStars:   {"estrella fija", "estrellas fijas"},
	intelligence.TechStations:     {"estación", "estacion", "retrógrado", "retrogrado"},
	intelligence.TechDecennials:   {"decenial", "decennials"},
	intelligence.TechLots:         {"lote", "parte de fortuna", "parte de"},
	intelligence.TechMidpoints:    {"punto medio", "puntos medios", "ebertin"},
	intelligence.TechPlanetCycles: {"retorno de", "ciclo de", "return"},
	intelligence.TechAlmuten:      {"almutén", "almuten", "dignidad"},
	intelligence.TechLunations:    {"luna nueva", "luna llena", "lunación"},
	intelligence.TechFastTransits: {"tránsito rápido", "transito rapido"},
}

func containsTechReference(responseLower, techID string) bool {
	// Direct ID match
	if strings.Contains(responseLower, strings.ToLower(techID)) {
		return true
	}
	// Alias match
	aliases := techAliases[techID]
	for _, alias := range aliases {
		if strings.Contains(responseLower, alias) {
			return true
		}
	}
	return false
}
