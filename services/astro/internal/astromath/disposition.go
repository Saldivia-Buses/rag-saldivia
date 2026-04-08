package astromath

import "github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"

// DispositionResult holds the full disposition chain analysis.
type DispositionResult struct {
	Chains          []DispositionChain `json:"chains"`
	FinalDispositor string             `json:"final_dispositor"` // planet that disposes everything (if any)
	MutualReceptions []MutualReception `json:"mutual_receptions"`
}

// DispositionChain traces a planet's ruler chain to its terminus.
type DispositionChain struct {
	Planet   string   `json:"planet"`
	Chain    []string `json:"chain"`    // e.g., ["Venus", "Marte", "Saturno", "Saturno"]
	Final    string   `json:"final"`    // terminus: self-ruling planet or loop
	Complete bool     `json:"complete"` // true if chain ends in a self-ruling planet
}

// MutualReception records two planets in each other's domicile.
type MutualReception struct {
	PlanetA string `json:"planet_a"`
	SignA   string `json:"sign_a"` // sign planet A is in
	PlanetB string `json:"planet_b"`
	SignB   string `json:"sign_b"` // sign planet B is in
}

// PlanetKeywords maps planets to thematic keywords (Spanish).
var PlanetKeywords = map[string]string{
	"Sol":      "identidad, voluntad, vitalidad",
	"Luna":     "emociones, hogar, instinto",
	"Mercurio": "comunicación, contratos, pensamiento",
	"Venus":    "relaciones, valores, placer",
	"Marte":    "acción, conflicto, energía",
	"Júpiter":  "expansión, oportunidad, fe",
	"Saturno":  "estructura, límite, responsabilidad",
	"Urano":    "cambio repentino, innovación, libertad",
	"Neptuno":  "ilusión, espiritualidad, disolución",
	"Plutón":   "transformación, poder, crisis",
	"Quirón":   "herida, sanación, enseñanza",
}

// traceChain follows the dispositor chain from a planet.
func traceChain(planet string, planets map[string]*ephemeris.PlanetPos, maxDepth int) DispositionChain {
	chain := []string{planet}
	visited := map[string]bool{planet: true}
	current := planet

	for i := 0; i < maxDepth; i++ {
		p, ok := planets[current]
		if !ok {
			break
		}
		signIdx := SignIndex(p.Lon)
		ruler := DomicileOf[signIdx]

		// Self-ruling: planet is in its own domicile
		if ruler == current {
			return DispositionChain{
				Planet:   planet,
				Chain:    chain,
				Final:    current,
				Complete: true,
			}
		}

		// Loop detection
		if visited[ruler] {
			chain = append(chain, ruler)
			return DispositionChain{
				Planet:   planet,
				Chain:    chain,
				Final:    ruler,
				Complete: false,
			}
		}

		chain = append(chain, ruler)
		visited[ruler] = true
		current = ruler
	}

	return DispositionChain{
		Planet:   planet,
		Chain:    chain,
		Final:    current,
		Complete: false,
	}
}

// CalcDisposition computes disposition chains for all planets in the chart.
func CalcDisposition(planets map[string]*ephemeris.PlanetPos) *DispositionResult {
	// Only trace chains for planets that have positions (skip angles/lots)
	traceList := []string{
		"Sol", "Luna", "Mercurio", "Venus", "Marte",
		"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón",
	}

	var chains []DispositionChain
	finalCounts := make(map[string]int)

	for _, name := range traceList {
		if _, ok := planets[name]; !ok {
			continue
		}
		c := traceChain(name, planets, 12)
		chains = append(chains, c)
		if c.Complete {
			finalCounts[c.Final]++
		}
	}

	// Find final dispositor (planet that ALL chains end at)
	finalDispositor := ""
	planetCount := 0
	for _, name := range traceList {
		if _, ok := planets[name]; ok {
			planetCount++
		}
	}
	for planet, count := range finalCounts {
		if count == planetCount {
			finalDispositor = planet
			break
		}
	}

	// Detect mutual receptions (planet A in sign of B, B in sign of A)
	var receptions []MutualReception
	for i, nameA := range traceList {
		pA, okA := planets[nameA]
		if !okA {
			continue
		}
		signA := SignIndex(pA.Lon)
		rulerA := DomicileOf[signA]

		for _, nameB := range traceList[i+1:] {
			pB, okB := planets[nameB]
			if !okB {
				continue
			}
			signB := SignIndex(pB.Lon)
			rulerB := DomicileOf[signB]

			// Mutual reception: A rules B's sign AND B rules A's sign
			if rulerA == nameB && rulerB == nameA {
				receptions = append(receptions, MutualReception{
					PlanetA: nameA,
					SignA:   Signs[signA],
					PlanetB: nameB,
					SignB:   Signs[signB],
				})
			}
		}
	}

	return &DispositionResult{
		Chains:           chains,
		FinalDispositor:  finalDispositor,
		MutualReceptions: receptions,
	}
}
