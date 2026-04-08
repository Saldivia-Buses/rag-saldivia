package intelligence

import (
	"fmt"

	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// CrossRefType classifies the convergence type.
type CrossRefType string

const (
	CrossRefRuler    CrossRefType = "ruler"    // same planet rules multiple techniques
	CrossRefPoint    CrossRefType = "point"    // same natal point activated by multiple techniques
	CrossRefTemporal CrossRefType = "temporal" // multiple techniques converge in same month
)

// CrossReference is a convergence found between techniques.
type CrossReference struct {
	Type         CrossRefType
	Techniques   []string // technique IDs involved
	Planet       string   // planet at the nexus
	NatalPoint   string   // natal point activated
	Month        int      // month of convergence (0 = year-long)
	Significance float64  // 0.0-1.0
	Description  string   // human-readable (Spanish)
}

// AnalyzeCrossReferences finds convergences between technique results.
// Fully deterministic — no LLM calls. Max 7 results.
func AnalyzeCrossReferences(fullCtx *astrocontext.FullContext) []CrossReference {
	var results []CrossReference

	// TYPE 1: Ruler convergence — same planet is lord/ruler in multiple time-lord techniques
	rulerMap := make(map[string][]string) // planet → technique IDs
	if fullCtx.Profection != nil {
		rulerMap[fullCtx.Profection.Lord] = append(rulerMap[fullCtx.Profection.Lord], TechProfections)
	}
	if fullCtx.Firdaria != nil {
		rulerMap[fullCtx.Firdaria.MajorLord] = append(rulerMap[fullCtx.Firdaria.MajorLord], TechFirdaria+"_mayor")
		if fullCtx.Firdaria.SubLord != "" {
			rulerMap[fullCtx.Firdaria.SubLord] = append(rulerMap[fullCtx.Firdaria.SubLord], TechFirdaria+"_sub")
		}
	}
	if fullCtx.ZRFortune != nil && fullCtx.ZRFortune.Level1 != nil {
		rulerMap[fullCtx.ZRFortune.Level1.Lord] = append(rulerMap[fullCtx.ZRFortune.Level1.Lord], TechZR+"_fortuna")
	}
	if fullCtx.Decennials != nil {
		rulerMap[fullCtx.Decennials.MajorPlanet] = append(rulerMap[fullCtx.Decennials.MajorPlanet], TechDecennials+"_mayor")
	}

	for planet, techs := range rulerMap {
		if len(techs) >= 2 {
			results = append(results, CrossReference{
				Type:         CrossRefRuler,
				Techniques:   techs,
				Planet:       planet,
				Significance: min(float64(len(techs))/4.0, 1.0),
				Description:  fmt.Sprintf("%s es señor del tiempo en %d técnicas: %v — tema dominante del año", planet, len(techs), techs),
			})
		}
	}

	// TYPE 2: Point convergence — same natal point activated by SA + PD + transits
	pointMap := make(map[string][]string) // natal point → technique IDs
	for _, sa := range fullCtx.SolarArc {
		pointMap[sa.NatPlanet] = append(pointMap[sa.NatPlanet], TechSolarArc)
	}
	for _, pd := range fullCtx.Directions {
		if pd.OrbDeg < 1.5 {
			pointMap[pd.Significator] = append(pointMap[pd.Significator], TechPrimaryDir)
		}
	}
	for _, tr := range fullCtx.Transits {
		pointMap[tr.Natal] = append(pointMap[tr.Natal], TechTransits)
	}
	for _, ecl := range fullCtx.Eclipses {
		pointMap[ecl.NatPoint] = append(pointMap[ecl.NatPoint], TechEclipses)
	}

	for point, techs := range pointMap {
		// Deduplicate technique IDs
		unique := dedup(techs)
		if len(unique) >= 2 {
			results = append(results, CrossReference{
				Type:         CrossRefPoint,
				Techniques:   unique,
				NatalPoint:   point,
				Significance: min(float64(len(unique))/4.0, 1.0),
				Description:  fmt.Sprintf("%s activado por %d técnicas: %v — punto focal del año", point, len(unique), unique),
			})
		}
	}

	// TYPE 3: Temporal convergence — multiple techniques converge in same month
	monthTechs := make(map[int][]string) // month → technique IDs
	for _, ecl := range fullCtx.Eclipses {
		m := ecl.Eclipse.Month
		if m >= 1 && m <= 12 {
			monthTechs[m] = append(monthTechs[m], TechEclipses)
		}
	}
	for _, st := range fullCtx.Stations {
		if st.NatPoint != "" && st.Month >= 1 && st.Month <= 12 {
			monthTechs[st.Month] = append(monthTechs[st.Month], TechStations)
		}
	}
	for _, tr := range fullCtx.Transits {
		for _, ep := range tr.EpDetails {
			if ep.MonthStart >= 1 && ep.MonthStart <= 12 {
				monthTechs[ep.MonthStart] = append(monthTechs[ep.MonthStart], TechTransits)
			}
		}
	}
	for _, et := range fullCtx.EclipseTriggers {
		if et.Month >= 1 && et.Month <= 12 {
			monthTechs[et.Month] = append(monthTechs[et.Month], TechEclTriggers)
		}
	}

	for month, techs := range monthTechs {
		unique := dedup(techs)
		if len(unique) >= 3 {
			results = append(results, CrossReference{
				Type:         CrossRefTemporal,
				Techniques:   unique,
				Month:        month,
				Significance: min(float64(len(unique))/5.0, 1.0),
				Description:  fmt.Sprintf("Mes %d: convergencia de %d técnicas — período de alta actividad", month, len(unique)),
			})
		}
	}

	// Sort by significance descending, cap at 7
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Significance > results[i].Significance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	if len(results) > 7 {
		results = results[:7]
	}

	return results
}

func dedup(ss []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
