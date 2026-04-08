package technique

import (
	"sort"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// ChartPair holds two charts for relational techniques.
type ChartPair struct {
	ChartA *natal.Chart
	ChartB *natal.Chart
	NameA  string
	NameB  string
}

// SynastryAspect records an inter-chart aspect.
type SynastryAspect struct {
	PlanetA  string  `json:"planet_a"`
	PlanetB  string  `json:"planet_b"`
	Aspect   string  `json:"aspect"`
	Orb      float64 `json:"orb"`
	Score    float64 `json:"score"`    // positive=harmonious, negative=tense
	Nature   string  `json:"nature"`   // "conexión" or "fricción"
	Interp   string  `json:"interp"`   // brief interpretation
}

// HouseOverlay records where one person's planet falls in the other's houses.
type HouseOverlay struct {
	Planet string `json:"planet"`
	Owner  string `json:"owner"`  // "A" or "B"
	House  int    `json:"house"`  // house number in the other person's chart
	Theme  string `json:"theme"`
}

// SynastryResult holds the full synastry analysis.
type SynastryResult struct {
	NameA         string            `json:"name_a"`
	NameB         string            `json:"name_b"`
	Aspects       []SynastryAspect  `json:"aspects"`
	Connections   []SynastryAspect  `json:"connections"`    // top 5 positive
	Frictions     []SynastryAspect  `json:"frictions"`      // top 5 negative
	HouseOverlays []HouseOverlay    `json:"house_overlays"`
	Score         int               `json:"compatibility_score"` // 0-100
	Summary       string            `json:"summary"`
}

// synastry aspect scoring
var synAspectBase = map[string]float64{
	"conjunction": 6, "trine": 8, "sextile": 5,
	"square": -5, "opposition": -3,
}

// planet importance weights
var synPlanetWeight = map[string]float64{
	"Sol": 1.5, "Luna": 1.5, "Venus": 1.3, "Marte": 1.2,
	"Mercurio": 1.0, "Júpiter": 1.1, "Saturno": 1.1,
	"ASC": 1.3, "MC": 1.0,
}

func synWeight(name string) float64 {
	if w, ok := synPlanetWeight[name]; ok {
		return w
	}
	return 1.0
}

const synOrb = 5.0

// synPlanets are the planets to check in synastry.
var synPlanets = []string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno"}

// CalcSynastry computes inter-chart aspects, house overlays, and compatibility score.
func CalcSynastry(pair *ChartPair) *SynastryResult {
	result := &SynastryResult{
		NameA: pair.NameA,
		NameB: pair.NameB,
	}

	// Collect points for each chart
	pointsA := chartPoints(pair.ChartA)
	pointsB := chartPoints(pair.ChartB)

	// Cross-aspects: A's planets to B's planets
	seen := make(map[string]bool)
	var aspects []SynastryAspect

	for nameA, lonA := range pointsA {
		for nameB, lonB := range pointsB {
			asp := astromath.FindAspect(lonA, lonB, synOrb)
			if asp == nil {
				continue
			}

			// Deduplicate symmetric pairs
			key := dedupKey(nameA, "A", nameB, "B", asp.Name)
			if seen[key] {
				continue
			}
			seen[key] = true

			base := synAspectBase[asp.Name]
			// Conjunction: depends on planet nature
			if asp.Name == "conjunction" {
				if isMalefic(nameA) || isMalefic(nameB) {
					base = -3
				} else if isBenefic(nameA) || isBenefic(nameB) {
					base = 10
				}
			}
			weight := (synWeight(nameA) + synWeight(nameB)) / 2
			score := base * weight

			nature := "conexión"
			if score < 0 {
				nature = "fricción"
			}

			aspects = append(aspects, SynastryAspect{
				PlanetA: nameA,
				PlanetB: nameB,
				Aspect:  asp.Name,
				Orb:     asp.Orb,
				Score:   score,
				Nature:  nature,
			})
		}
	}

	// Sort by absolute score descending
	sort.Slice(aspects, func(i, j int) bool {
		return abs64(aspects[i].Score) > abs64(aspects[j].Score)
	})

	result.Aspects = aspects

	// Top 5 connections and frictions
	for _, a := range aspects {
		if a.Score > 0 && len(result.Connections) < 5 {
			result.Connections = append(result.Connections, a)
		}
		if a.Score < 0 && len(result.Frictions) < 5 {
			result.Frictions = append(result.Frictions, a)
		}
	}

	// House overlays: B's planets in A's houses and vice versa
	for _, name := range synPlanets {
		if posB, ok := pair.ChartB.Planets[name]; ok {
			house := astromath.HouseForLon(posB.Lon, pair.ChartA.Cusps)
			result.HouseOverlays = append(result.HouseOverlays, HouseOverlay{
				Planet: name, Owner: "B", House: house,
				Theme: astromath.HouseThemes[house],
			})
		}
		if posA, ok := pair.ChartA.Planets[name]; ok {
			house := astromath.HouseForLon(posA.Lon, pair.ChartB.Cusps)
			result.HouseOverlays = append(result.HouseOverlays, HouseOverlay{
				Planet: name, Owner: "A", House: house,
				Theme: astromath.HouseThemes[house],
			})
		}
	}

	// Compatibility score (0-100)
	result.Score = normalizeScore(aspects)

	// Summary
	result.Summary = synSummary(result.Score, len(result.Connections), len(result.Frictions))

	return result
}

func chartPoints(chart *natal.Chart) map[string]float64 {
	pts := make(map[string]float64)
	for _, name := range synPlanets {
		if p, ok := chart.Planets[name]; ok {
			pts[name] = p.Lon
		}
	}
	pts["ASC"] = chart.ASC
	pts["MC"] = chart.MC
	return pts
}

func dedupKey(a, ownerA, b, ownerB, aspect string) string {
	if a+ownerA > b+ownerB {
		return b + ownerB + a + ownerA + aspect
	}
	return a + ownerA + b + ownerB + aspect
}

func isBenefic(name string) bool  { return name == "Venus" || name == "Júpiter" }
func isMalefic(name string) bool  { return name == "Saturno" || name == "Marte" }

func normalizeScore(aspects []SynastryAspect) int {
	if len(aspects) == 0 {
		return 50
	}
	rawSum := 0.0
	for _, a := range aspects {
		rawSum += a.Score
	}
	maxPossible := float64(len(aspects)) * 15
	minPossible := float64(len(aspects)) * -7.5
	shifted := rawSum - minPossible
	span := maxPossible - minPossible
	if span == 0 {
		return 50
	}
	normalized := (shifted / span) * 100
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 100 {
		normalized = 100
	}
	return int(normalized)
}

func synSummary(score, connections, frictions int) string {
	switch {
	case score >= 75:
		return "Alta compatibilidad — predominan las conexiones armónicas"
	case score >= 55:
		return "Compatibilidad moderada — equilibrio entre conexión y fricción"
	case score >= 35:
		return "Relación desafiante — las fricciones requieren trabajo consciente"
	default:
		return "Compatibilidad baja — diferencias significativas en la dinámica"
	}
}
