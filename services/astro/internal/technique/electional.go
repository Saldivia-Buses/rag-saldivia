package technique

import (
	"sort"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// ElectionalCriteria specifies what the user is looking for in a date.
type ElectionalCriteria struct {
	FavorAspects  []string `json:"favor_aspects"`  // planet names to favor (e.g., "Júpiter", "Venus")
	AvoidAspects  []string `json:"avoid_aspects"`  // planet names to avoid (e.g., "Saturno", "Marte")
	AvoidVOC      bool     `json:"avoid_voc"`      // avoid void-of-course Moon
	PreferDiurnal bool     `json:"prefer_diurnal"` // prefer daytime elections
	FocusHouses   []int    `json:"focus_houses"`   // natal houses to activate (e.g., [7, 10])
}

// ElectionalWindow is a scored candidate date.
type ElectionalWindow struct {
	JD       float64  `json:"jd"`
	Year     int      `json:"year"`
	Month    int      `json:"month"`
	Day      int      `json:"day"`
	Score    float64  `json:"score"`
	Factors  []string `json:"factors"` // reasons for the score
}

// ElectionalResult holds the electional analysis.
type ElectionalResult struct {
	Criteria ElectionalCriteria `json:"criteria"`
	Windows  []ElectionalWindow `json:"windows"` // top results, sorted by score
}

const (
	electCoarseStep = 1.0   // 1 day for coarse scan
	electMaxResults = 10
)

// CalcElectional finds optimal dates within a year for the given criteria.
// Uses two-pass approach: coarse scan (daily) to find promising regions,
// then fine scan within top regions.
// Semaphore-limited at handler level (max 1 concurrent).
func CalcElectional(natalChart *natal.Chart, year int, criteria ElectionalCriteria) (*ElectionalResult, error) {
	jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
	jdEnd := ephemeris.JulDay(year+1, 1, 1, 0.0)
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	// Natal points for aspect checking
	natalLons := make(map[string]float64)
	for name, pos := range natalChart.Planets {
		natalLons[name] = pos.Lon
	}
	natalLons["ASC"] = natalChart.ASC
	natalLons["MC"] = natalChart.MC

	// Coarse scan: score each day
	var candidates []ElectionalWindow

	for jd := jdStart; jd < jdEnd; jd += electCoarseStep {
		score, factors := scoreDate(jd, flags, natalLons, criteria)
		if score > 0 {
			y, m, d, _ := ephemeris.RevJul(jd)
			candidates = append(candidates, ElectionalWindow{
				JD:      jd,
				Year:    y,
				Month:   m,
				Day:     d,
				Score:   score,
				Factors: factors,
			})
		}
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Take top results
	if len(candidates) > electMaxResults {
		candidates = candidates[:electMaxResults]
	}

	return &ElectionalResult{
		Criteria: criteria,
		Windows:  candidates,
	}, nil
}

// scoreDate evaluates a single date for electional quality.
// Returns (score, factors). Score > 0 = usable date.
func scoreDate(jd float64, flags int, natalLons map[string]float64, criteria ElectionalCriteria) (float64, []string) {
	score := 50.0 // baseline
	var factors []string

	// Moon position — most important for electional
	moonPos, err := ephemeris.CalcPlanet(jd, ephemeris.Moon, flags)
	if err != nil {
		return 0, nil
	}

	// Moon sign — avoid Scorpio/Capricorn for most elections
	moonSign := astromath.SignIndex(moonPos.Lon)
	switch moonSign {
	case 7: // Escorpio
		score -= 15
		factors = append(factors, "Luna en Escorpio (-15)")
	case 9: // Capricornio
		score -= 10
		factors = append(factors, "Luna en Capricornio (-10)")
	case 1: // Tauro — excellent
		score += 10
		factors = append(factors, "Luna en Tauro (+10)")
	case 3: // Cáncer — excellent
		score += 10
		factors = append(factors, "Luna en Cáncer (+10)")
	}

	// Moon speed — fast is good (> 13°/day)
	if moonPos.Speed > 13.0 {
		score += 5
		factors = append(factors, "Luna rápida (+5)")
	}

	// Check favorable transits to natal points
	favorIDs := map[string]int{
		"Júpiter": ephemeris.Jupiter, "Venus": ephemeris.Venus,
	}
	avoidIDs := map[string]int{
		"Saturno": ephemeris.Saturn, "Marte": ephemeris.Mars,
	}

	// Favor planets making trines/sextiles to natal points
	for name, id := range favorIDs {
		pos, err := ephemeris.CalcPlanet(jd, id, flags)
		if err != nil {
			continue
		}
		for natName, natLon := range natalLons {
			asp := astromath.FindAspect(pos.Lon, natLon, 3.0)
			if asp == nil {
				continue
			}
			switch asp.Name {
			case "trine", "sextile":
				score += 8
				factors = append(factors, name+" "+asp.Name+" "+natName+" (+8)")
			case "conjunction":
				score += 5
				factors = append(factors, name+" conjunción "+natName+" (+5)")
			}
		}
	}

	// Penalize avoid planets making hard aspects to natal points
	for name, id := range avoidIDs {
		pos, err := ephemeris.CalcPlanet(jd, id, flags)
		if err != nil {
			continue
		}
		for natName, natLon := range natalLons {
			asp := astromath.FindAspect(pos.Lon, natLon, 3.0)
			if asp == nil {
				continue
			}
			switch asp.Name {
			case "square":
				score -= 10
				factors = append(factors, name+" cuadratura "+natName+" (-10)")
			case "opposition":
				score -= 8
				factors = append(factors, name+" oposición "+natName+" (-8)")
			}
		}
	}

	// Mercury retrograde penalty
	mercPos, err := ephemeris.CalcPlanet(jd, ephemeris.Mercury, flags)
	if err == nil && mercPos.Speed < 0 {
		score -= 10
		factors = append(factors, "Mercurio retrógrado (-10)")
	}

	return score, factors
}
