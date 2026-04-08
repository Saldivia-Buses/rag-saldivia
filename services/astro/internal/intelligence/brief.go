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

	// Section 3: Technique interpretations (wired from interpretations_full.go)
	if len(fullCtx.Transits) > 0 || len(fullCtx.SolarArc) > 0 {
		b.WriteString("## INTERPRETACIONES\n\n")
		for _, tr := range fullCtx.Transits {
			if tr.Orb < 2.0 { // only tight transits get interpretation
				b.WriteString("- " + InterpretTransit(tr.Transit, tr.Aspect, tr.Natal).Detail + "\n")
			}
		}
		for _, sa := range fullCtx.SolarArc {
			if sa.Orb < 1.0 {
				b.WriteString("- " + InterpretSA(sa.SAplanet, sa.Aspect, sa.NatPlanet) + "\n")
			}
		}
		if fullCtx.Profection != nil {
			b.WriteString("- " + InterpretProfection(fullCtx.Profection.ActiveHouse, fullCtx.Profection.Lord, fullCtx.Profection.Theme) + "\n")
		}
		if fullCtx.Firdaria != nil {
			b.WriteString("- " + InterpretFirdaria(fullCtx.Firdaria.MajorLord, fullCtx.Firdaria.SubLord) + "\n")
		}
		b.WriteString("\n")
	}

	// Section 4: The actual computed brief (from context builder)
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
