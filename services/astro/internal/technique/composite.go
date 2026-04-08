package technique

import (
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// CompositeResult holds the composite chart (midpoint chart).
// The composite chart represents the relationship as its own entity.
// Reference: Rob Hand "Planets in Composite" (1975).
type CompositeResult struct {
	NameA    string                          `json:"name_a"`
	NameB    string                          `json:"name_b"`
	Planets  map[string]float64              `json:"planets"`  // planet → composite longitude
	ASC      float64                         `json:"asc"`
	MC       float64                         `json:"mc"`
	Aspects  []CompositeAspect               `json:"aspects"`
}

// CompositeAspect is an aspect within the composite chart.
type CompositeAspect struct {
	PlanetA string  `json:"planet_a"`
	PlanetB string  `json:"planet_b"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
	Nature  string  `json:"nature"`
}

// midpointLon calculates the ecliptic midpoint using the shorter arc.
func midpointLon(lonA, lonB float64) float64 {
	diff := math.Mod(lonB-lonA+360, 360)
	if diff > 180 {
		// Shorter arc goes "backwards" — midpoint is opposite of long-arc midpoint
		return astromath.Normalize360(lonA + diff/2 + 180)
	}
	return astromath.Normalize360(lonA + diff/2)
}

// compositePlanets are the planets to include in composite.
var compositePlanets = []string{
	"Sol", "Luna", "Mercurio", "Venus", "Marte",
	"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón",
}

// CalcComposite calculates the composite chart between two natal charts.
func CalcComposite(pair *ChartPair) *CompositeResult {
	compPlanets := make(map[string]float64)

	for _, name := range compositePlanets {
		posA, okA := pair.ChartA.Planets[name]
		posB, okB := pair.ChartB.Planets[name]
		if okA && okB {
			compPlanets[name] = midpointLon(posA.Lon, posB.Lon)
		}
	}

	// ASC and MC midpoints
	compASC := midpointLon(pair.ChartA.ASC, pair.ChartB.ASC)
	compMC := midpointLon(pair.ChartA.MC, pair.ChartB.MC)
	compPlanets["ASC"] = compASC
	compPlanets["MC"] = compMC

	// Calculate aspects within the composite chart
	var aspects []CompositeAspect
	names := make([]string, 0, len(compPlanets))
	for n := range compPlanets {
		names = append(names, n)
	}

	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			lonA := compPlanets[names[i]]
			lonB := compPlanets[names[j]]
			asp := astromath.FindAspect(lonA, lonB, 5.0)
			if asp != nil {
				nature := "neutral"
				switch asp.Name {
				case "trine", "sextile":
					nature = "armónico"
				case "square", "opposition":
					nature = "tenso"
				}
				aspects = append(aspects, CompositeAspect{
					PlanetA: names[i],
					PlanetB: names[j],
					Aspect:  asp.Name,
					Orb:     asp.Orb,
					Nature:  nature,
				})
			}
		}
	}

	return &CompositeResult{
		NameA:   pair.NameA,
		NameB:   pair.NameB,
		Planets: compPlanets,
		ASC:     compASC,
		MC:      compMC,
		Aspects: aspects,
	}
}

// CompositeToChart converts a CompositeResult into a natal.Chart-like structure
// so it can be used with existing technique functions (transits, etc.).
func CompositeToChart(comp *CompositeResult) *natal.Chart {
	planets := make(map[string]*ephemeris.PlanetPos)
	for name, lon := range comp.Planets {
		if name == "ASC" || name == "MC" {
			continue
		}
		planets[name] = &ephemeris.PlanetPos{Lon: lon}
	}

	// Equal-house cusps from composite ASC — this IS the correct method for composites.
	// Rob Hand (Planets in Composite): composite charts use equal houses because
	// midpoint ASC/MC are not astronomically related (no real time/place).
	cusps := make([]float64, 13)
	for i := 0; i <= 12; i++ {
		cusps[i] = astromath.Normalize360(comp.ASC + float64(i)*30)
	}

	return &natal.Chart{
		Planets: planets,
		Cusps:   cusps,
		ASC:     comp.ASC,
		MC:      comp.MC,
	}
}
