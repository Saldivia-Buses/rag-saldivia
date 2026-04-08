package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// PredictiveSynastryResult holds cross-chart Solar Arc timing.
type PredictiveSynastryResult struct {
	NameA       string                 `json:"name_a"`
	NameB       string                 `json:"name_b"`
	Year        int                    `json:"year"`
	Activations []SACrossActivation    `json:"activations"`
}

// SACrossActivation records when person A's SA planet hits person B's natal point.
type SACrossActivation struct {
	Person    string  `json:"person"`     // "A→B" or "B→A"
	SAPlanet  string  `json:"sa_planet"`
	SALon     float64 `json:"sa_lon"`
	NatPlanet string  `json:"nat_planet"`
	NatLon    float64 `json:"nat_lon"`
	Aspect    string  `json:"aspect"`
	Orb       float64 `json:"orb"`
	MonthEst  int     `json:"month_est"` // estimated month of exactitude
}

const saCrossOrb = 1.5

// CalcPredictiveSynastry computes SA cross-timing between two charts for a year.
// Checks: A's SA planets → B's natal points, and B's SA planets → A's natal points.
func CalcPredictiveSynastry(pair *ChartPair, year int) *PredictiveSynastryResult {
	result := &PredictiveSynastryResult{
		NameA: pair.NameA,
		NameB: pair.NameB,
		Year:  year,
	}

	jdMid := ephemeris.JulDay(year, 7, 1, 12.0)

	// A → B
	arcA := (jdMid - pair.ChartA.JD) / 365.25 * astromath.NaibodRate
	result.Activations = append(result.Activations,
		findSACross("A→B", pair.ChartA, pair.ChartB, arcA)...)

	// B → A
	arcB := (jdMid - pair.ChartB.JD) / 365.25 * astromath.NaibodRate
	result.Activations = append(result.Activations,
		findSACross("B→A", pair.ChartB, pair.ChartA, arcB)...)

	return result
}

// findSACross finds SA activations from source chart to target chart.
func findSACross(direction string, source, target *natal.Chart, arc float64) []SACrossActivation {
	saNames := []string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno"}

	// Target natal points
	targetPts := make(map[string]float64)
	for _, name := range saNames {
		if p, ok := target.Planets[name]; ok {
			targetPts[name] = p.Lon
		}
	}
	targetPts["ASC"] = target.ASC
	targetPts["MC"] = target.MC

	timingAspects := map[string]float64{
		"conjunction": 0, "opposition": 180,
	}

	var activations []SACrossActivation

	for _, saName := range saNames {
		srcPos, ok := source.Planets[saName]
		if !ok {
			continue
		}
		saLon := astromath.Normalize360(srcPos.Lon + arc)

		for tgtName, tgtLon := range targetPts {
			for aspName, aspAngle := range timingAspects {
				targetLon := astromath.Normalize360(tgtLon + aspAngle)
				orbVal := astromath.AngDiff(saLon, targetLon)

				if orbVal <= saCrossOrb {
					// Estimate month from orb position
					// Naibod rate ~0.986°/year = ~0.082°/month
					monthOffset := orbVal / (astromath.NaibodRate / 12)
					monthEst := 6 + int(monthOffset)
					if monthEst < 1 {
						monthEst = 1
					}
					if monthEst > 12 {
						monthEst = 12
					}

					activations = append(activations, SACrossActivation{
						Person:    direction,
						SAPlanet:  saName,
						SALon:     saLon,
						NatPlanet: tgtName,
						NatLon:    tgtLon,
						Aspect:    aspName,
						Orb:       orbVal,
						MonthEst:  monthEst,
					})
				}
			}
		}
	}

	return activations
}
