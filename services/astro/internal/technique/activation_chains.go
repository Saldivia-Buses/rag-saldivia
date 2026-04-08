package technique

import (
	"fmt"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
)

// ActivationChain records when multiple techniques activate the same planet/point.
type ActivationChain struct {
	Planet      string   `json:"planet"`       // the focal planet
	Techniques  []string `json:"techniques"`   // which techniques activate it
	Count       int      `json:"count"`        // number of activations
	Significance string  `json:"significance"` // "alta", "media"
	Description string   `json:"description"`
}

// CalcActivationChains finds planets/points that are activated by 3+ techniques.
// This is a cross-technique analysis that runs AFTER individual techniques.
func CalcActivationChains(
	solarArcs []SolarArcResult,
	directions []PrimaryDirection,
	transits []TransitActivation,
	eclipses []EclipseActivation,
	stations []Station,
) []ActivationChain {
	// Count activations per natal point across all techniques
	type activation struct {
		technique string
		aspect    string
	}
	pointActivations := make(map[string][]activation)

	// Solar Arcs
	for _, sa := range solarArcs {
		pointActivations[sa.NatPlanet] = append(pointActivations[sa.NatPlanet],
			activation{"SA_" + sa.SAplanet, sa.Aspect})
	}

	// Primary Directions
	for _, pd := range directions {
		if pd.OrbDeg < 1.5 { // only tight directions
			pointActivations[pd.Significator] = append(pointActivations[pd.Significator],
				activation{"PD_" + pd.Promissor, pd.Aspect})
		}
	}

	// Transits
	for _, tr := range transits {
		pointActivations[tr.Natal] = append(pointActivations[tr.Natal],
			activation{"TR_" + tr.Transit, tr.Aspect})
	}

	// Eclipses
	for _, ecl := range eclipses {
		pointActivations[ecl.NatPoint] = append(pointActivations[ecl.NatPoint],
			activation{"ECL_" + ecl.Eclipse.Type, ecl.Aspect})
	}

	// Stations
	for _, st := range stations {
		if st.NatPoint != "" {
			pointActivations[st.NatPoint] = append(pointActivations[st.NatPoint],
				activation{"STAT_" + st.Planet, "conjunction"})
		}
	}

	// Find chains (3+ activations)
	var chains []ActivationChain
	for point, acts := range pointActivations {
		if len(acts) < 3 {
			continue
		}

		techs := make([]string, len(acts))
		for i, a := range acts {
			techs[i] = a.technique + "_" + a.aspect
		}

		significance := "media"
		if len(acts) >= 4 {
			significance = "alta"
		}

		keywords := astromath.PlanetKeywords[point]
		if keywords == "" {
			keywords = point
		}

		chains = append(chains, ActivationChain{
			Planet:       point,
			Techniques:   techs,
			Count:        len(acts),
			Significance: significance,
			Description: fmt.Sprintf("%s activado por %d técnicas — foco en %s",
				point, len(acts), keywords),
		})
	}

	// Sort by count descending
	for i := 0; i < len(chains); i++ {
		for j := i + 1; j < len(chains); j++ {
			if chains[j].Count > chains[i].Count {
				chains[i], chains[j] = chains[j], chains[i]
			}
		}
	}

	return chains
}
