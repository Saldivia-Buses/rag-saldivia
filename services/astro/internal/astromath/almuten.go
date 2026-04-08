package astromath

import "github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"

// AlmutenResult holds the Almuten Figuris calculation.
type AlmutenResult struct {
	Winner          string                       `json:"winner"`
	Score           int                          `json:"score"`
	Breakdown       map[string][]PointScore      `json:"breakdown"`
	SensitivePoints map[string]float64           `json:"sensitive_points"`
	Diurnal         bool                         `json:"diurnal"`
}

// PointScore records a planet's dignity score at a sensitive point.
type PointScore struct {
	Point   string `json:"point"`
	Score   int    `json:"score"`
	Dignity string `json:"dignity"`
}

// CalcAlmuten computes the Almuten Figuris of the chart.
// Evaluates the 7 classical planets' essential dignities over 5 sensitive points:
// ASC, MC, Sol, Luna, Pars Fortunae.
// The planet with the highest total score is the Almuten Figuris.
func CalcAlmuten(
	planets map[string]*ephemeris.PlanetPos,
	asc, mc float64,
	diurnal bool,
) *AlmutenResult {
	// Resolve sensitive point longitudes
	sunLon := 0.0
	moonLon := 0.0
	fortunaLon := 0.0

	if p, ok := planets["Sol"]; ok {
		sunLon = p.Lon
	}
	if p, ok := planets["Luna"]; ok {
		moonLon = p.Lon
	}
	if p, ok := planets["Fortuna"]; ok {
		fortunaLon = p.Lon
	} else {
		fortunaLon = PartOfFortune(asc, moonLon, sunLon, diurnal)
	}

	sensitive := map[string]float64{
		"ASC":     asc,
		"MC":      mc,
		"Sol":     sunLon,
		"Luna":    moonLon,
		"Fortuna": fortunaLon,
	}

	totals := make(map[string]int)
	breakdown := make(map[string][]PointScore)

	for _, planet := range ClassicalPlanets {
		var bd []PointScore
		total := 0
		for ptName, ptLon := range sensitive {
			score, dignity := DignityScoreAt(planet, ptLon, diurnal)
			if score < 0 {
				score = 0 // debilities don't count for almuten
			}
			bd = append(bd, PointScore{Point: ptName, Score: score, Dignity: dignity})
			total += score
		}
		totals[planet] = total
		breakdown[planet] = bd
	}

	// Find winner
	winner := ""
	bestScore := -1
	for _, planet := range ClassicalPlanets {
		if totals[planet] > bestScore {
			bestScore = totals[planet]
			winner = planet
		}
	}

	return &AlmutenResult{
		Winner:          winner,
		Score:           bestScore,
		Breakdown:       breakdown,
		SensitivePoints: sensitive,
		Diurnal:         diurnal,
	}
}
