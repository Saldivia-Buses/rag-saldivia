package technique

import (
	"math"
	"sort"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// MidpointResult holds the Ebertin midpoint analysis.
type MidpointResult struct {
	Midpoints     []Midpoint          `json:"midpoints"`
	Activated     []ActivatedMidpoint `json:"activated"`     // midpoints hit by natal planets
	Sort45        []MidpointEntry     `json:"sort_45"`       // 45° sort (Ebertin dial)
}

// Midpoint represents the midpoint between two planets.
type Midpoint struct {
	PlanetA string  `json:"planet_a"`
	PlanetB string  `json:"planet_b"`
	Lon     float64 `json:"lon"`   // direct midpoint (shorter arc)
	Pos     string  `json:"pos"`   // formatted position
}

// ActivatedMidpoint records when a natal planet sits on a midpoint axis.
type ActivatedMidpoint struct {
	Planet    string  `json:"planet"`      // the activating planet
	PlanetA   string  `json:"midpoint_a"`  // midpoint planet A
	PlanetB   string  `json:"midpoint_b"`  // midpoint planet B
	MidLon    float64 `json:"midpoint_lon"`
	Orb       float64 `json:"orb"`
	Notation  string  `json:"notation"`    // "Planet = A/B" (Ebertin notation)
}

// MidpointEntry is one entry in the 45° sorted dial.
type MidpointEntry struct {
	Name   string  `json:"name"`    // planet name or "A/B"
	Lon45  float64 `json:"lon_45"`  // longitude modulo 45°
	Type   string  `json:"type"`    // "planet" or "midpoint"
}

// midpointPlanets are the planets used for midpoint analysis.
var midpointPlanets = []string{
	"Sol", "Luna", "Mercurio", "Venus", "Marte",
	"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón",
}

const midpointOrb = 1.5 // tight orb for midpoint activation (Ebertin standard)

// CalcMidpoints calculates all midpoints and finds activations.
func CalcMidpoints(chart *natal.Chart) *MidpointResult {
	result := &MidpointResult{}

	// Collect planet longitudes
	lons := make(map[string]float64)
	for _, name := range midpointPlanets {
		if p, ok := chart.Planets[name]; ok {
			lons[name] = p.Lon
		}
	}
	// Add angles
	lons["ASC"] = chart.ASC
	lons["MC"] = chart.MC

	names := make([]string, 0, len(lons))
	for n := range lons {
		names = append(names, n)
	}
	sort.Strings(names)

	// Calculate all midpoints
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			mid := midpointShorterArc(lons[names[i]], lons[names[j]])
			result.Midpoints = append(result.Midpoints, Midpoint{
				PlanetA: names[i],
				PlanetB: names[j],
				Lon:     mid,
				Pos:     astromath.PosToStr(mid),
			})
		}
	}

	// Check for activated midpoints (natal planet on a midpoint axis)
	// A midpoint axis includes the direct midpoint AND its opposite (180°)
	for _, mp := range result.Midpoints {
		for activator, actLon := range lons {
			if activator == mp.PlanetA || activator == mp.PlanetB {
				continue
			}
			// Check direct midpoint and its opposite
			orb1 := astromath.AngDiff(actLon, mp.Lon)
			orb2 := astromath.AngDiff(actLon, astromath.Normalize360(mp.Lon+180))
			// Also check 45° and 135° (semi-square axis for Ebertin 45° dial)
			orb3 := astromath.AngDiff(actLon, astromath.Normalize360(mp.Lon+45))
			orb4 := astromath.AngDiff(actLon, astromath.Normalize360(mp.Lon+135))

			bestOrb := math.Min(math.Min(orb1, orb2), math.Min(orb3, orb4))

			if bestOrb <= midpointOrb {
				result.Activated = append(result.Activated, ActivatedMidpoint{
					Planet:   activator,
					PlanetA:  mp.PlanetA,
					PlanetB:  mp.PlanetB,
					MidLon:   mp.Lon,
					Orb:      math.Round(bestOrb*100) / 100,
					Notation: activator + " = " + mp.PlanetA + "/" + mp.PlanetB,
				})
			}
		}
	}

	// Sort activated by orb
	sort.Slice(result.Activated, func(i, j int) bool {
		return result.Activated[i].Orb < result.Activated[j].Orb
	})

	// Build 45° sort (Ebertin dial)
	var entries []MidpointEntry
	for name, lon := range lons {
		entries = append(entries, MidpointEntry{
			Name:  name,
			Lon45: math.Mod(lon, 45),
			Type:  "planet",
		})
	}
	for _, mp := range result.Midpoints {
		entries = append(entries, MidpointEntry{
			Name:  mp.PlanetA + "/" + mp.PlanetB,
			Lon45: math.Mod(mp.Lon, 45),
			Type:  "midpoint",
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Lon45 < entries[j].Lon45
	})
	result.Sort45 = entries

	return result
}

// midpointShorterArc calculates the midpoint using the shorter arc.
func midpointShorterArc(lonA, lonB float64) float64 {
	diff := math.Mod(lonB-lonA+360, 360)
	if diff > 180 {
		return astromath.Normalize360(lonA + diff/2 + 180)
	}
	return astromath.Normalize360(lonA + diff/2)
}
