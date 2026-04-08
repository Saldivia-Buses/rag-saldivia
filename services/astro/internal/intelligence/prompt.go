package intelligence

import (
	"fmt"
	"strings"
)

// BuildSystemPrompt constructs the LLM system prompt for a domain.
func BuildSystemPrompt(domain *ResolvedDomain, gate *GateResult, crossRefs []CrossReference) string {
	var b strings.Builder

	// Base persona
	b.WriteString(`Eres Astro, un agente de predicción astrológica de clase mundial. `)
	b.WriteString(`Operás con el Sistema Topocéntrico Polich-Page + Swiss Ephemeris + 55 técnicas simultáneas. `)
	b.WriteString(`Tu ventaja: la INTERDISCIPLINARIEDAD — encontrás las 2-3 convergencias más significativas y las tejés en la narrativa.`)
	b.WriteString("\n\n")

	// Domain-specific instructions
	b.WriteString(fmt.Sprintf("## DOMINIO: %s\n\n", domain.Name))

	// Technique hierarchy
	if len(domain.TechniquesBrief) > 0 {
		b.WriteString("Técnicas por orden de importancia para esta consulta:\n")
		for _, tw := range domain.TechniquesBrief {
			stars := "★"
			if tw.Weight >= 0.8 {
				stars = "★★★"
			} else if tw.Weight >= 0.6 {
				stars = "★★"
			}
			focus := ""
			if tw.Focus != "" {
				focus = fmt.Sprintf(" (foco: %s)", tw.Focus)
			}
			b.WriteString(fmt.Sprintf("  %s %s%s\n", stars, tw.ID, focus))
		}
		b.WriteString("\n")
	}

	// Cross-references to weave into narrative
	if len(crossRefs) > 0 {
		b.WriteString("## CONVERGENCIAS DETECTADAS (tejer en la narrativa):\n\n")
		for _, cr := range crossRefs {
			b.WriteString(fmt.Sprintf("- %s\n", cr.Description))
		}
		b.WriteString("\n")
	}

	// Ghost techniques — DO NOT mention these
	if gate != nil && len(gate.Ghosts) > 0 {
		b.WriteString("## TÉCNICAS SIN DATOS (NO mencionar):\n")
		for _, g := range gate.Ghosts {
			b.WriteString(fmt.Sprintf("  - %s\n", g.TechniqueID))
		}
		b.WriteString("\n")
	}

	// Precautions
	if len(domain.Precautions) > 0 {
		b.WriteString("## PRECAUCIONES:\n")
		for _, p := range domain.Precautions {
			b.WriteString(fmt.Sprintf("- %s\n", p))
		}
		b.WriteString("\n")
	}

	// Universal rules
	b.WriteString("## REGLAS UNIVERSALES:\n")
	b.WriteString("- NUNCA inventar datos — solo usar lo que está en el Brief\n")
	b.WriteString("- Citar mes/año cuando una técnica da timing preciso\n")
	b.WriteString("- Empezar con las convergencias (2-3 técnicas apuntando al mismo tema)\n")
	b.WriteString("- Terminar con recomendación concreta y accionable\n")
	b.WriteString("- Idioma: español rioplatense profesional\n")

	return b.String()
}
