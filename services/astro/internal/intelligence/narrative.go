package intelligence

import (
	"fmt"
	"strings"
)

// NarrativeArc structures the LLM response into a coherent narrative.
type NarrativeArc struct {
	Opening     string   `json:"opening"`      // hook based on dominant theme
	Convergences []string `json:"convergences"` // key cross-technique findings
	Sections    []string `json:"sections"`      // ordered technique sections
	Closing     string   `json:"closing"`       // actionable recommendation
}

// BuildNarrativeArc produces a narrative structure from cross-references and domain.
// The LLM uses this as a roadmap for its response.
func BuildNarrativeArc(crossRefs []CrossReference, domain *ResolvedDomain) *NarrativeArc {
	arc := &NarrativeArc{}

	// Opening: based on the strongest convergence
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

// FormatNarrativeGuide produces a text guide for the LLM to follow.
func FormatNarrativeGuide(arc *NarrativeArc) string {
	var b strings.Builder
	b.WriteString("## GUÍA NARRATIVA\n\n")
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
