package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// SolarArcResult holds a single Solar Arc activation.
type SolarArcResult struct {
	SAplanet  string  `json:"sa_planet"`
	NatPlanet string  `json:"nat_planet"`
	Aspect    string  `json:"aspect"`
	Orb       float64 `json:"orb"`
	Nature    string  `json:"nature"`
	SALon     float64 `json:"sa_lon"`
	NatLon    float64 `json:"nat_lon"`
	Kind      string  `json:"kind"` // "direct", "antiscion", "contra-antiscion"
}

// SolarArcPositions holds all SA-directed positions for a given date.
type SolarArcPositions struct {
	ArcDeg    float64            `json:"arc_deg"`
	Positions map[string]float64 `json:"positions"` // planet → SA longitude
}

// mainPlanetNames are the planets that get SA directions (matching PLANET_IDS from Python).
var mainPlanetNames = []string{
	"Sol", "Luna", "Mercurio", "Venus", "Marte",
	"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón",
}

// CalcSolarArc computes Solar Arc directed positions for a target JD.
// arc = (jdTarget - natal.JD) / 365.25 * NaibodRate
func CalcSolarArc(chart *natal.Chart, jdTarget float64) *SolarArcPositions {
	years := (jdTarget - chart.JD) / 365.25
	arc := years * astromath.NaibodRate

	positions := make(map[string]float64)
	for _, name := range mainPlanetNames {
		if pos, ok := chart.Planets[name]; ok {
			positions[name] = astromath.Normalize360(pos.Lon + arc)
		}
	}

	return &SolarArcPositions{ArcDeg: arc, Positions: positions}
}

// FindSolarArcActivations finds SA planets aspecting natal points within orb.
// Also checks antiscia and contra-antiscia of SA positions.
func FindSolarArcActivations(chart *natal.Chart, jdTarget float64) []SolarArcResult {
	sa := CalcSolarArc(chart, jdTarget)
	orb := astromath.OrbDefaults.SolarArc
	antisciaOrb := astromath.OrbDefaults.SAAntiscia

	// Collect natal points to check against (planets + ASC + MC + Vertex)
	natalPoints := make(map[string]float64)
	for name, pos := range chart.Planets {
		natalPoints[name] = pos.Lon
	}
	natalPoints["ASC"] = chart.ASC
	natalPoints["MC"] = chart.MC
	natalPoints["Vertex"] = chart.Vertex

	var results []SolarArcResult

	for saName, saLon := range sa.Positions {
		saAntiscion := astromath.Antiscion(saLon)
		saContra := astromath.ContraAntiscion(saLon)

		for natName, natLon := range natalPoints {
			if saName == natName {
				continue // skip self-aspects
			}

			// Direct SA aspect
			if asp := astromath.FindAspect(saLon, natLon, orb); asp != nil {
				results = append(results, SolarArcResult{
					SAplanet:  saName,
					NatPlanet: natName,
					Aspect:    asp.Name,
					Orb:       asp.Orb,
					Nature:    aspectNature(asp.Name),
					SALon:     saLon,
					NatLon:    natLon,
					Kind:      "direct",
				})
			}

			// Antiscion conjunction only
			if asp := astromath.FindAspect(saAntiscion, natLon, antisciaOrb); asp != nil && asp.Name == "conjunction" {
				results = append(results, SolarArcResult{
					SAplanet:  saName,
					NatPlanet: natName,
					Aspect:    "antiscion",
					Orb:       asp.Orb,
					Nature:    "neutral",
					SALon:     saAntiscion,
					NatLon:    natLon,
					Kind:      "antiscion",
				})
			}

			// Contra-antiscion conjunction only
			if asp := astromath.FindAspect(saContra, natLon, antisciaOrb); asp != nil && asp.Name == "conjunction" {
				results = append(results, SolarArcResult{
					SAplanet:  saName,
					NatPlanet: natName,
					Aspect:    "contra-antiscion",
					Orb:       asp.Orb,
					Nature:    "neutral",
					SALon:     saContra,
					NatLon:    natLon,
					Kind:      "contra-antiscion",
				})
			}
		}
	}

	return results
}

// aspectNature returns "fácil", "tenso", or "neutral" for an aspect name.
func aspectNature(name string) string {
	switch name {
	case "trine", "sextile":
		return "fácil"
	case "square", "opposition":
		return "tenso"
	default:
		return "neutral"
	}
}

// CalcSolarArcForYear is a convenience that computes SA for mid-year (June 15).
func CalcSolarArcForYear(chart *natal.Chart, year int) *SolarArcPositions {
	jdMid := ephemeris.JulDay(year, 6, 15, 12.0)
	return CalcSolarArc(chart, jdMid)
}
