package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// RelocationResult holds a relocated chart analysis.
type RelocationResult struct {
	City       string    `json:"city"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
	NewASC     float64   `json:"new_asc"`
	NewMC      float64   `json:"new_mc"`
	NewCusps   []float64 `json:"new_cusps"`
	ASCSign    string    `json:"asc_sign"`
	MCSign     string    `json:"mc_sign"`
	Score      float64   `json:"score"`      // 0-100 suitability
	Highlights []string  `json:"highlights"`
}

// RelocationCity is a candidate city for relocation analysis.
type RelocationCity struct {
	Name string
	Lat  float64
	Lon  float64
}

// CalcRelocation computes a relocated chart for a given city.
// Planets stay the same (they're geocentric/topocentric to birth location),
// but houses are RECALCULATED at the new location using the same natal JD.
// This is the astronomically correct method — Swiss Ephemeris computes
// new house cusps for the new latitude/longitude at the birth moment.
func CalcRelocation(chart *natal.Chart, city RelocationCity) (*RelocationResult, error) {
	// Recalculate house cusps at the new location using the original birth JD.
	// This is the CORRECT relocation method: same moment in time, different observer location.
	ephemeris.CalcMu.Lock()
	ephemeris.SetTopo(city.Lon, city.Lat, 0)
	cusps, ascmc, err := ephemeris.CalcHousesEx(
		chart.JD,
		ephemeris.FlagSwieph|ephemeris.FlagTopoctr,
		city.Lat, city.Lon,
		ephemeris.HouseTopocentric,
	)
	ephemeris.CalcMu.Unlock()

	if err != nil {
		return nil, err
	}

	newASC := ascmc[0]
	newMC := ascmc[1]

	result := &RelocationResult{
		City:    city.Name,
		Lat:     city.Lat,
		Lon:     city.Lon,
		NewASC:  newASC,
		NewMC:   newMC,
		NewCusps: cusps,
		ASCSign: astromath.SignName(newASC),
		MCSign:  astromath.SignName(newMC),
	}

	// Score based on benefic/malefic planets near the relocated angles
	score := 50.0
	benefics := []string{"Júpiter", "Venus"}
	malefics := []string{"Saturno", "Marte", "Plutón"}

	for _, name := range benefics {
		if p, ok := chart.Planets[name]; ok {
			ascOrb := astromath.AngDiff(p.Lon, newASC)
			mcOrb := astromath.AngDiff(p.Lon, newMC)
			if ascOrb < 8 {
				score += 15
				result.Highlights = append(result.Highlights,
					name+" conjunción ASC reubicado (orbe "+fmtDeg(ascOrb)+"°) — muy favorable")
			} else if mcOrb < 8 {
				score += 12
				result.Highlights = append(result.Highlights,
					name+" conjunción MC reubicado (orbe "+fmtDeg(mcOrb)+"°) — favorable para carrera")
			}
		}
	}
	for _, name := range malefics {
		if p, ok := chart.Planets[name]; ok {
			ascOrb := astromath.AngDiff(p.Lon, newASC)
			mcOrb := astromath.AngDiff(p.Lon, newMC)
			if ascOrb < 8 {
				score -= 12
				result.Highlights = append(result.Highlights,
					name+" conjunción ASC reubicado (orbe "+fmtDeg(ascOrb)+"°) — desafiante")
			} else if mcOrb < 8 {
				score -= 8
				result.Highlights = append(result.Highlights,
					name+" conjunción MC reubicado (orbe "+fmtDeg(mcOrb)+"°) — presión profesional")
			}
		}
	}

	// Check natal Sun and Moon in relocated houses (key quality-of-life indicators)
	if sunPos, ok := chart.Planets["Sol"]; ok {
		sunHouse := astromath.HouseForLon(sunPos.Lon, cusps)
		if sunHouse == 1 || sunHouse == 10 {
			score += 8
			result.Highlights = append(result.Highlights,
				"Sol natal en casa "+itoa2(sunHouse)+" reubicada — visibilidad y protagonismo")
		}
	}
	if moonPos, ok := chart.Planets["Luna"]; ok {
		moonHouse := astromath.HouseForLon(moonPos.Lon, cusps)
		if moonHouse == 4 || moonHouse == 1 {
			score += 5
			result.Highlights = append(result.Highlights,
				"Luna natal en casa "+itoa2(moonHouse)+" reubicada — bienestar emocional")
		}
	}

	if score > 100 { score = 100 }
	if score < 0 { score = 0 }
	result.Score = score

	return result, nil
}

func fmtDeg(d float64) string {
	i := int(d)
	f := int((d - float64(i)) * 10)
	return string(rune('0'+i)) + "." + string(rune('0'+f))
}

func itoa2(n int) string {
	if n < 10 { return string(rune('0' + n)) }
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
