package intelligence

import (
	"fmt"
	"strings"
)

// NarrativeArc structures the LLM response into a coherent narrative.
type NarrativeArc struct {
	Archetype    *Archetype `json:"archetype,omitempty"` // dominant narrative archetype
	Opening      string     `json:"opening"`              // hook based on dominant theme
	Convergences []string   `json:"convergences"`         // key cross-technique findings
	Sections     []string   `json:"sections"`             // ordered technique sections
	Closing      string     `json:"closing"`              // actionable recommendation
}

// BuildNarrativeArc produces a narrative structure from cross-references and domain.
// The LLM uses this as a roadmap for its response.
func BuildNarrativeArc(crossRefs []CrossReference, domain *ResolvedDomain) *NarrativeArc {
	arc := &NarrativeArc{}

	// Detect archetype from cross-references
	arc.Archetype = DetectArchetype(crossRefs)

	// Opening: based on the strongest convergence + archetype
	if len(crossRefs) > 0 {
		cr := crossRefs[0]
		switch cr.Type {
		case CrossRefRuler:
			arc.Opening = fmt.Sprintf("El año está dominado por %s como señor del tiempo en múltiples técnicas.", cr.Planet)
		case CrossRefPoint:
			arc.Opening = fmt.Sprintf("Hay una convergencia significativa sobre %s — múltiples técnicas apuntan al mismo punto.", cr.NatalPoint)
		case CrossRefTemporal:
			arc.Opening = fmt.Sprintf("El mes %d concentra una actividad astrológica inusual — varias técnicas convergen.", cr.Month)
		default:
			arc.Opening = "Hay señales claras de activación astrológica este período."
		}
	} else {
		arc.Opening = "El panorama astrológico muestra un período de transición gradual."
	}

	// Convergences
	for _, cr := range crossRefs {
		arc.Convergences = append(arc.Convergences, cr.Description)
	}

	// Sections ordered by domain technique weights
	for _, tw := range domain.TechniquesBrief {
		if tw.Weight >= 0.5 {
			arc.Sections = append(arc.Sections, tw.ID)
		}
	}

	// Closing: domain-specific actionable advice
	switch domain.ID {
	case "carrera":
		arc.Closing = "Recomendación concreta para tu carrera en los próximos meses."
	case "salud":
		arc.Closing = "Períodos a cuidar y hábitos a reforzar."
	case "amor":
		arc.Closing = "Ventanas favorables para la conexión y el diálogo."
	case "dinero":
		arc.Closing = "Timing para decisiones financieras clave."
	case "empresa":
		arc.Closing = "Acciones concretas para el negocio este trimestre."
	default:
		arc.Closing = "Próximos pasos y fechas clave a tener en cuenta."
	}

	return arc
}

// Archetype represents a narrative archetype for the year.
type Archetype struct {
	Name        string `json:"name"`
	Planet      string `json:"planet"`
	Description string `json:"description"`
}

// narrativeArchetypes maps dominant planet/house to story archetype.
// Ported from Python astro-v2 narrative_arc.py — 8 archetypes.
var narrativeArchetypes = map[string]Archetype{
	"Saturno": {"Constructor", "Saturno", "Año de estructura, responsabilidad, y logros sólidos a largo plazo"},
	"Plutón":  {"Transformador", "Plutón", "Año de muerte y renacimiento — lo que muere abre paso a algo mayor"},
	"Júpiter": {"Expansor", "Júpiter", "Año de crecimiento, oportunidades, y ampliar horizontes"},
	"Venus":   {"Conector", "Venus", "Año de relaciones, valores, y armonía — lo vincular domina"},
	"Urano":   {"Liberador", "Urano", "Año de rupturas, independencia, y reinvención radical"},
	"Quirón":  {"Sanador", "Quirón", "Año de heridas que sanan — la vulnerabilidad es el camino"},
	"Marte":   {"Guerrero", "Marte", "Año de acción, iniciativa, y conquista — no es momento de esperar"},
	"Mercurio": {"Maestro", "Mercurio", "Año de aprendizaje, comunicación, y transmisión de conocimiento"},
}

// DetectArchetype determines the narrative archetype based on the dominant
// cross-reference planet. Uses ruler-type cross-refs as primary signal.
func DetectArchetype(crossRefs []CrossReference) *Archetype {
	// Count planet mentions across cross-references
	planetCount := make(map[string]int)
	for _, cr := range crossRefs {
		if cr.Planet != "" {
			planetCount[cr.Planet]++
		}
	}

	// Find dominant planet
	bestPlanet := ""
	bestCount := 0
	for p, c := range planetCount {
		if c > bestCount {
			bestCount = c
			bestPlanet = p
		}
	}

	if bestPlanet == "" {
		return nil
	}

	if arch, ok := narrativeArchetypes[bestPlanet]; ok {
		return &arch
	}

	// Default archetype for planets without specific archetype (Sol, Luna, Neptuno)
	return &Archetype{
		Name:        "Transitante",
		Planet:      bestPlanet,
		Description: fmt.Sprintf("Año marcado por la energía de %s — su tema central domina el período", bestPlanet),
	}
}

// FormatNarrativeGuide produces a text guide for the LLM to follow.
func FormatNarrativeGuide(arc *NarrativeArc) string {
	var b strings.Builder
	b.WriteString("## GUÍA NARRATIVA\n\n")
	if arc.Archetype != nil {
		b.WriteString(fmt.Sprintf("**Arquetipo del año: %s** — %s\n\n", arc.Archetype.Name, arc.Archetype.Description))
	}
	b.WriteString("Estructura tu respuesta así:\n\n")
	b.WriteString("1. APERTURA: " + arc.Opening + "\n")
	if len(arc.Convergences) > 0 {
		b.WriteString("2. CONVERGENCIAS:\n")
		for _, c := range arc.Convergences {
			b.WriteString("   - " + c + "\n")
		}
	}
	b.WriteString("3. CIERRE: " + arc.Closing + "\n")
	return b.String()
}
