package intelligence

import (
	"fmt"
	"strings"

	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// BuildIntelligenceBrief produces a domain-aware, weighted brief.
// Ghost techniques are omitted. Cross-references are promoted to the top.
// This replaces the default brief when a query is provided.
func BuildIntelligenceBrief(fullCtx *astrocontext.FullContext, domain *ResolvedDomain, gate *GateResult, crossRefs []CrossReference) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# BRIEF DE INTELIGENCIA — %s — %d\n", fullCtx.ContactName, fullCtx.Year))
	b.WriteString(fmt.Sprintf("Dominio: %s\n\n", domain.Name))

	// Section 1: Cross-references (highest priority)
	if len(crossRefs) > 0 {
		b.WriteString("## CONVERGENCIAS CLAVE\n\n")
		for _, cr := range crossRefs {
			icon := "·"
			switch cr.Type {
			case CrossRefRuler:
				icon = "♛"
			case CrossRefPoint:
				icon = "⊕"
			case CrossRefTemporal:
				icon = "⏱"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", icon, cr.Description))
		}
		b.WriteString("\n")
	}

	// Section 2: One-line verdicts per validated technique (ordered by domain weight)
	if gate != nil && len(gate.Validated) > 0 {
		b.WriteString("## TÉCNICAS VALIDADAS\n\n")
		// Order by domain brief weight
		weightMap := make(map[string]float64)
		for _, tw := range domain.TechniquesBrief {
			weightMap[tw.ID] = tw.Weight
		}
		type weightedTech struct {
			id     string
			weight float64
			count  int
		}
		var sorted []weightedTech
		for _, v := range gate.Validated {
			w := weightMap[v.TechniqueID]
			if w == 0 {
				w = 0.1
			}
			sorted = append(sorted, weightedTech{v.TechniqueID, w, v.EntryCount})
		}
		// Sort by weight descending
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].weight > sorted[i].weight {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		for _, t := range sorted {
			stars := "★"
			if t.weight >= 0.8 {
				stars = "★★★"
			} else if t.weight >= 0.6 {
				stars = "★★"
			}
			b.WriteString(fmt.Sprintf("  %s %s (%d resultados)\n", stars, t.id, t.count))
		}
		b.WriteString("\n")
	}

	// Section 3: The actual computed brief (from context builder)
	// This is the full brief with all technique sections
	if fullCtx.Brief != "" {
		b.WriteString(fullCtx.Brief)
	}

	// Section 4: Precautions
	if len(domain.Precautions) > 0 {
		b.WriteString("\n## PRECAUCIONES\n\n")
		for _, p := range domain.Precautions {
			b.WriteString(fmt.Sprintf("⚠ %s\n", p))
		}
	}

	return b.String()
}
