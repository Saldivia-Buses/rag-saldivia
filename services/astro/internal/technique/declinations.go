package technique

import (
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// DeclinationResult holds the declination analysis.
type DeclinationResult struct {
	Positions  []DecPosition     `json:"positions"`
	Parallels  []DecAspect       `json:"parallels"`
	OOBPlanets []DecPosition     `json:"oob_planets"` // out-of-bounds (|dec| > 23.44°)
}

// DecPosition holds a planet's declination.
type DecPosition struct {
	Planet string  `json:"planet"`
	Dec    float64 `json:"declination"` // degrees, + = North, - = South
	OOB    bool    `json:"oob"`         // out of bounds (beyond ecliptic obliquity)
}

// DecAspect records a parallel or contra-parallel.
type DecAspect struct {
	PlanetA string  `json:"planet_a"`
	PlanetB string  `json:"planet_b"`
	Type    string  `json:"type"` // "paralelo" or "contra-paralelo"
	DecA    float64 `json:"dec_a"`
	DecB    float64 `json:"dec_b"`
	Orb     float64 `json:"orb"`
}

const (
	parallelOrb    = 1.0  // degrees
	eclipticObliq  = 23.44 // approximate obliquity for OOB detection
)

// decPlanets are all planets checked for declinations.
var decPlanets = []string{
	"Sol", "Luna", "Mercurio", "Venus", "Marte",
	"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón",
	"Quirón", "Nodo Norte",
}

// CalcDeclinations computes declination positions, parallels, and out-of-bounds planets.
func CalcDeclinations(chart *natal.Chart) *DeclinationResult {
	result := &DeclinationResult{}

	// Collect declinations
	var positions []DecPosition
	for _, name := range decPlanets {
		p, ok := chart.Planets[name]
		if !ok || (p.Dec == 0 && p.RA == 0 && p.Lon == 0) {
			continue
		}
		oob := math.Abs(p.Dec) > eclipticObliq
		pos := DecPosition{
			Planet: name,
			Dec:    math.Round(p.Dec*100) / 100,
			OOB:    oob,
		}
		positions = append(positions, pos)
		if oob {
			result.OOBPlanets = append(result.OOBPlanets, pos)
		}
	}
	result.Positions = positions

	// Find parallels and contra-parallels
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			a := positions[i]
			b := positions[j]

			// Parallel: both same hemisphere, similar declination
			diffSame := math.Abs(a.Dec - b.Dec)
			if diffSame <= parallelOrb {
				result.Parallels = append(result.Parallels, DecAspect{
					PlanetA: a.Planet,
					PlanetB: b.Planet,
					Type:    "paralelo",
					DecA:    a.Dec,
					DecB:    b.Dec,
					Orb:     math.Round(diffSame*100) / 100,
				})
			}

			// Contra-parallel: opposite hemispheres, similar absolute declination
			diffOpposite := math.Abs(a.Dec + b.Dec) // + because they're on opposite sides
			if diffOpposite <= parallelOrb && a.Dec*b.Dec < 0 { // opposite signs
				result.Parallels = append(result.Parallels, DecAspect{
					PlanetA: a.Planet,
					PlanetB: b.Planet,
					Type:    "contra-paralelo",
					DecA:    a.Dec,
					DecB:    b.Dec,
					Orb:     math.Round(diffOpposite*100) / 100,
				})
			}
		}
	}

	return result
}
