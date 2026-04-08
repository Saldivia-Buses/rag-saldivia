package intelligence

import (
	"fmt"

	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// DataRichness classifies how much real data a technique produced.
type DataRichness int

const (
	RichnessGhost   DataRichness = 0
	RichnessMention DataRichness = 1
	RichnessFull    DataRichness = 2
)

func (r DataRichness) String() string {
	switch r {
	case RichnessFull:
		return "full"
	case RichnessMention:
		return "mention"
	default:
		return "ghost"
	}
}

// TechniqueStatus is the gate's assessment of one technique.
type TechniqueStatus struct {
	TechniqueID string
	Richness    DataRichness
	EntryCount  int
}

// GateResult is the full gate assessment.
type GateResult struct {
	Validated []TechniqueStatus
	Ghosts    []TechniqueStatus
	Warnings  []string
	Coverage  float64 // 0.0-1.0
}

// techniqueAccessor checks a FullContext field and returns (count, richness).
type techniqueAccessor func(*astrocontext.FullContext) (int, DataRichness)

// techniqueFieldMap maps technique IDs to struct field inspectors.
var techniqueFieldMap = map[string]techniqueAccessor{
	TechTransits: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.Transits)
		if n > 5 { return n, RichnessFull }
		if n > 0 { return n, RichnessMention }
		return 0, RichnessGhost
	},
	TechSolarArc: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.SolarArc)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechPrimaryDir: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.Directions)
		if n > 3 { return n, RichnessFull }
		if n > 0 { return n, RichnessMention }
		return 0, RichnessGhost
	},
	TechProgressions: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Progressions == nil { return 0, RichnessGhost }
		n := len(c.Progressions.Positions)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechSolarReturn: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.SolarReturn != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechLunarReturn: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.LunarReturns)
		if n > 10 { return n, RichnessFull }
		if n > 0 { return n, RichnessMention }
		return 0, RichnessGhost
	},
	TechProfections: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Profection != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechFirdaria: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Firdaria != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechEclipses: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.Eclipses)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechFixedStars: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.FixedStars)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechZR: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.ZRFortune != nil && c.ZRFortune.Level1 != nil { return 1, RichnessFull }
		if c.ZRFortune != nil { return 1, RichnessMention }
		return 0, RichnessGhost
	},
	TechStations: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.Stations)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechAlmuten: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Almuten != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechLots: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.Lots)
		if n > 10 { return n, RichnessFull }
		if n > 0 { return n, RichnessMention }
		return 0, RichnessGhost
	},
	TechDisposition: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Disposition != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechSect: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Sect != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechDecennials: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Decennials != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechTertiaryProg: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.TertiaryProg != nil && len(c.TertiaryProg.Positions) > 0 { return len(c.TertiaryProg.Positions), RichnessFull }
		return 0, RichnessGhost
	},
	TechFastTransits: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.FastTransits)
		if n > 5 { return n, RichnessFull }
		if n > 0 { return n, RichnessMention }
		return 0, RichnessGhost
	},
	TechLunations: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Lunations != nil && len(c.Lunations.Lunations) > 0 { return len(c.Lunations.Lunations), RichnessFull }
		return 0, RichnessGhost
	},
	TechPrenatalEcl: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.PrenatalEclipse != nil && (c.PrenatalEclipse.Solar != nil || c.PrenatalEclipse.Lunar != nil) { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechEclTriggers: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.EclipseTriggers)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechPlanetCycles: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.PlanetaryCycles)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechMidpoints: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Midpoints != nil && len(c.Midpoints.Activated) > 0 { return len(c.Midpoints.Activated), RichnessFull }
		return 0, RichnessGhost
	},
	TechDeclinations: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Declinations != nil && len(c.Declinations.Positions) > 0 { return len(c.Declinations.Positions), RichnessFull }
		return 0, RichnessGhost
	},
	TechActivChains: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.ActivationChains)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechTimingWindows: func(c *astrocontext.FullContext) (int, DataRichness) {
		n := len(c.TimingWindows)
		if n > 0 { return n, RichnessFull }
		return 0, RichnessGhost
	},
	TechTemperament: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Temperament != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechMelothesia: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Melothesia != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
	TechHyleg: func(c *astrocontext.FullContext) (int, DataRichness) {
		if c.Hyleg != nil { return 1, RichnessFull }
		return 0, RichnessGhost
	},
}

// ValidateTechniques scans the FullContext and determines data richness for each technique.
func ValidateTechniques(fullCtx *astrocontext.FullContext, domain *ResolvedDomain) *GateResult {
	result := &GateResult{}

	requiredSet := make(map[string]bool)
	for _, t := range domain.TechniquesRequired {
		requiredSet[t] = true
	}

	// Check all techniques referenced in the domain's required + expected + brief
	allTechs := make(map[string]bool)
	for _, t := range domain.TechniquesRequired {
		allTechs[t] = true
	}
	for _, t := range domain.TechniquesExpected {
		allTechs[t] = true
	}
	for _, tw := range domain.TechniquesBrief {
		allTechs[tw.ID] = true
	}

	validatedCount := 0
	requiredCount := len(domain.TechniquesRequired)

	for techID := range allTechs {
		accessor, ok := techniqueFieldMap[techID]
		if !ok {
			continue // technique not yet mapped (e.g., synastry, horary — on-demand)
		}

		count, richness := accessor(fullCtx)
		status := TechniqueStatus{TechniqueID: techID, Richness: richness, EntryCount: count}

		if richness >= RichnessMention {
			result.Validated = append(result.Validated, status)
			if requiredSet[techID] {
				validatedCount++
			}
		} else {
			result.Ghosts = append(result.Ghosts, status)
			// Warn for high-weight techniques in the brief
			for _, tw := range domain.TechniquesBrief {
				if tw.ID == techID && tw.Weight >= 0.7 {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("%s (peso %.1f) sin datos — verificar context builder", techID, tw.Weight))
				}
			}
		}
	}

	if requiredCount > 0 {
		result.Coverage = float64(validatedCount) / float64(requiredCount)
	} else {
		result.Coverage = 1.0
	}

	return result
}
