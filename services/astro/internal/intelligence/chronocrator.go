package intelligence

import (
	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// ChronocratorGate filters technique results to only show those relevant
// to the active chronocrator (time lord). When the profection lord is Saturn,
// Saturn-related activations get boosted and non-Saturn activations get demoted.
type ChronocratorFilter struct {
	Lord         string   `json:"lord"`          // active chronocrator
	BoostPlanets []string `json:"boost_planets"` // planets to emphasize
	DemotePlanets []string `json:"demote_planets"` // planets to de-emphasize
}

// DetermineChronocrator extracts the active chronocrator from the FullContext.
// Priority: profection lord > firdaria major lord > ZR L1 lord.
func DetermineChronocrator(fullCtx *astrocontext.FullContext) *ChronocratorFilter {
	lord := ""

	// Primary: profection chronocrator
	if fullCtx.Profection != nil && fullCtx.Profection.Lord != "" {
		lord = fullCtx.Profection.Lord
	}

	// Fallback: firdaria major lord
	if lord == "" && fullCtx.Firdaria != nil {
		lord = fullCtx.Firdaria.MajorLord
	}

	// Fallback: ZR Fortune L1
	if lord == "" && fullCtx.ZRFortune != nil && fullCtx.ZRFortune.Level1 != nil {
		lord = fullCtx.ZRFortune.Level1.Lord
	}

	if lord == "" {
		return nil
	}

	// Determine which planets to boost/demote
	// Boost: the lord itself + planets it rules by sign
	// Demote: nothing — just boost the lord
	return &ChronocratorFilter{
		Lord:         lord,
		BoostPlanets: []string{lord},
	}
}

// FilterByChronocrator adjusts activation relevance based on the active time lord.
// Returns a relevance multiplier (1.0 = normal, 1.5 = boosted, 0.7 = demoted).
func RelevanceMultiplier(planet string, filter *ChronocratorFilter) float64 {
	if filter == nil {
		return 1.0
	}
	for _, bp := range filter.BoostPlanets {
		if bp == planet {
			return 1.5
		}
	}
	return 1.0
}
