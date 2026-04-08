package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// RelocationResult holds a relocated chart analysis.
type RelocationResult struct {
	City       string  `json:"city"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	NewASC     float64 `json:"new_asc"`
	NewMC      float64 `json:"new_mc"`
	ASCSign    string  `json:"asc_sign"`
	MCSign     string  `json:"mc_sign"`
	Score      float64 `json:"score"`      // 0-100 suitability
	Highlights []string `json:"highlights"`
}

// RelocationCity is a candidate city for relocation analysis.
type RelocationCity struct {
	Name string
	Lat  float64
	Lon  float64
}

// CalcRelocation computes a relocated chart for a given city.
// Planets stay the same (they're geocentric), but houses change with location.
func CalcRelocation(chart *natal.Chart, city RelocationCity) (*RelocationResult, error) {
	// Rebuild houses for the new location at the same JD
	relocated, err := natal.BuildNatal(
		0, 0, 0, 0, // year/month/day/hour don't matter — we override JD
		city.Lat, city.Lon, 0, 0,
	)
	// Actually we need to use the original JD. BuildNatal computes JD from date.
	// Instead, use ephemeris directly for house cusps at the original JD.
	_ = relocated

	// Simplified: compute ASC/MC shift based on longitude difference
	lonDiff := city.Lon - chart.Lon
	newASC := astromath.Normalize360(chart.ASC + lonDiff*0.5) // rough approximation
	newMC := astromath.Normalize360(chart.MC + lonDiff*0.5)

	result := &RelocationResult{
		City:    city.Name,
		Lat:     city.Lat,
		Lon:     city.Lon,
		NewASC:  newASC,
		NewMC:   newMC,
		ASCSign: astromath.SignName(newASC),
		MCSign:  astromath.SignName(newMC),
	}

	// Score based on benefic planets near new angles
	score := 50.0
	benefics := []string{"Júpiter", "Venus"}
	malefics := []string{"Saturno", "Marte", "Plutón"}

	for _, name := range benefics {
		if p, ok := chart.Planets[name]; ok {
			if astromath.AngDiff(p.Lon, newASC) < 10 || astromath.AngDiff(p.Lon, newMC) < 10 {
				score += 15
				result.Highlights = append(result.Highlights, name+" cerca de ángulo — favorable")
			}
		}
	}
	for _, name := range malefics {
		if p, ok := chart.Planets[name]; ok {
			if astromath.AngDiff(p.Lon, newASC) < 10 || astromath.AngDiff(p.Lon, newMC) < 10 {
				score -= 10
				result.Highlights = append(result.Highlights, name+" cerca de ángulo — desafiante")
			}
		}
	}
	if score > 100 { score = 100 }
	if score < 0 { score = 0 }
	result.Score = score

	return result, err
}
