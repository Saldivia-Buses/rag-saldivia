package astromath

import "github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"

// SectAnalysis holds the complete sect analysis of a chart.
type SectAnalysis struct {
	Diurnal       bool              `json:"diurnal"`
	SectLight     string            `json:"sect_light"`      // Sol (day) or Luna (night)
	SectBenefic   string            `json:"sect_benefic"`    // Júpiter (day) or Venus (night)
	SectMalefic   string            `json:"sect_malefic"`    // Saturno (day) or Marte (night)
	ContrarySect  []string          `json:"contrary_sect"`   // planets contrary to sect
	PlanetSect    map[string]string `json:"planet_sect"`     // planet → "sect"/"contrary"
	Hayz          []string          `json:"hayz"`            // planets in hayz (perfect sect condition)
	Halb          []string          `json:"halb"`            // planets in halb (contrary sect condition)
}

// CalcSect performs a complete sect analysis.
func CalcSect(planets map[string]*ephemeris.PlanetPos, diurnal bool) *SectAnalysis {
	result := &SectAnalysis{
		Diurnal:    diurnal,
		PlanetSect: make(map[string]string),
	}

	if diurnal {
		result.SectLight = "Sol"
		result.SectBenefic = "Júpiter"
		result.SectMalefic = "Saturno"
	} else {
		result.SectLight = "Luna"
		result.SectBenefic = "Venus"
		result.SectMalefic = "Marte"
	}

	// Classify each planet
	daySect := map[string]bool{"Sol": true, "Júpiter": true, "Saturno": true}
	nightSect := map[string]bool{"Luna": true, "Venus": true, "Marte": true}

	for _, name := range ClassicalPlanets {
		p, ok := planets[name]
		if !ok {
			continue
		}
		inDaySect := daySect[name]
		inNightSect := nightSect[name]
		// Mercurio is sect-neutral (oriental=day, occidental=night)
		if name == "Mercurio" {
			if p, okS := planets["Sol"]; okS {
				// Oriental = rises before Sun = sect of chart
				if diurnal {
					inDaySect = true
				} else {
					inNightSect = true
				}
				_ = p
			}
		}

		if diurnal && inDaySect || !diurnal && inNightSect {
			result.PlanetSect[name] = "sect"
		} else {
			result.PlanetSect[name] = "contrary"
			result.ContrarySect = append(result.ContrarySect, name)
		}

		// Hayz: planet in sect, in a sign of its own gender, above/below horizon matching sect
		// Simplified: planet in sect + in compatible sign element
		signIdx := SignIndex(p.Lon)
		masculine := signIdx%2 == 0 // fire/air signs (even indices in zodiac: Aries, Gemini, Leo...)
		planetMasc := name == "Sol" || name == "Marte" || name == "Júpiter" || name == "Saturno"

		if result.PlanetSect[name] == "sect" && masculine == planetMasc {
			result.Hayz = append(result.Hayz, name)
		}
		if result.PlanetSect[name] == "contrary" && masculine != planetMasc {
			result.Halb = append(result.Halb, name)
		}
	}

	return result
}
