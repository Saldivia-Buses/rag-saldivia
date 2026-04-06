package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// FixedStarConjunction records a fixed star conjunct a natal point.
type FixedStarConjunction struct {
	Star     string  `json:"star"`
	Nature   string  `json:"nature"`
	NatPoint string  `json:"natal_point"`
	Orb      float64 `json:"orb"`
	StarLon  float64 `json:"star_lon"`
	NatLon   float64 `json:"natal_lon"`
}

// FindFixedStarConjunctions checks all major fixed stars against natal points.
// Stars precess ~1°/72 years, so their ecliptic longitude is calculated for the natal date.
func FindFixedStarConjunctions(chart *natal.Chart) []FixedStarConjunction {
	natalPoints := make(map[string]float64)
	for name, pos := range chart.Planets {
		natalPoints[name] = pos.Lon
	}
	natalPoints["ASC"] = chart.ASC
	natalPoints["MC"] = chart.MC
	natalPoints["Vertex"] = chart.Vertex

	var results []FixedStarConjunction

	for _, star := range astromath.MajorFixedStars {
		starLon, err := ephemeris.FixstarUT(star.SweName, chart.JD, ephemeris.FlagSwieph)
		if err != nil {
			continue // star not found in catalog, skip
		}

		for natName, natLon := range natalPoints {
			orb := astromath.AngDiff(starLon, natLon)
			if orb <= astromath.FixedStarOrb {
				results = append(results, FixedStarConjunction{
					Star:     star.Name,
					Nature:   star.Nature,
					NatPoint: natName,
					Orb:      orb,
					StarLon:  starLon,
					NatLon:   natLon,
				})
			}
		}
	}

	return results
}
