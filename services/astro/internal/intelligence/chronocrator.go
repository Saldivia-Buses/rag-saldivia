package intelligence

import (
	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// ChronocratorGate filters technique results to only show those relevant
// to the active chronocrator (time lord). When the profection lord is Saturn,
// Saturn-related activations get boosted and non-Saturn activations get demoted.
type ChronocratorFilter struct {
	Lord          string   `json:"lord"`           // active year chronocrator
	MonthLord     string   `json:"month_lord"`     // active month chronocrator
	BoostPlanets  []string `json:"boost_planets"`  // planets to emphasize (2.0-2.5x)
	DemotePlanets []string `json:"demote_planets"` // planets to de-emphasize (0.3x)
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

	// Build boost and demote lists (Valens tradition)
	// Boost: the lord itself + its sub-lords from other time-lord systems
	boost := []string{lord}

	// Add firdaria sub-lord if different from main lord
	if fullCtx.Firdaria != nil && fullCtx.Firdaria.SubLord != "" && fullCtx.Firdaria.SubLord != lord {
		boost = append(boost, fullCtx.Firdaria.SubLord)
	}

	// Month lord: from firdaria sub-lord as proxy (cascade not in FullContext yet)
	monthLord := ""
	if fullCtx.Firdaria != nil && fullCtx.Firdaria.SubLord != "" {
		monthLord = fullCtx.Firdaria.SubLord
	}

	// All classical planets NOT in boost list get demoted
	allPlanets := []string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno"}
	boostSet := make(map[string]bool)
	for _, b := range boost {
		boostSet[b] = true
	}
	var demote []string
	for _, p := range allPlanets {
		if !boostSet[p] && p != monthLord {
			demote = append(demote, p)
		}
	}

	return &ChronocratorFilter{
		Lord:          lord,
		MonthLord:     monthLord,
		BoostPlanets:  boost,
		DemotePlanets: demote,
	}
}

// RelevanceMultiplier returns a multiplier for activation relevance based on
// the active time lord. Valens tradition: year lord involvement = 2.5x,
// target involving year lord = 2.0x, month lord = 1.5x, no involvement = 0.3x.
// Cap: 3.0x. Ported from Python astro-v2 chronocrator_gate.py.
func RelevanceMultiplier(activationPlanet, targetPlanet string, filter *ChronocratorFilter) float64 {
	if filter == nil {
		return 1.0
	}

	mult := 0.3 // base: no involvement

	// Check activation planet
	for _, bp := range filter.BoostPlanets {
		if bp == activationPlanet {
			mult = 2.5 // activation involves year lord
			break
		}
	}

	// Check target planet
	if mult < 2.0 {
		for _, bp := range filter.BoostPlanets {
			if bp == targetPlanet {
				mult = 2.0 // target involves year lord
				break
			}
		}
	}

	// Month lord bonus
	if filter.MonthLord != "" {
		if activationPlanet == filter.MonthLord {
			if mult < 2.0 {
				mult = 2.0
			}
		} else if targetPlanet == filter.MonthLord {
			if mult < 1.5 {
				mult = 1.5
			}
		}
	}

	// Cap at 3.0
	if mult > 3.0 {
		mult = 3.0
	}

	return mult
}
